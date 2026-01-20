package parser

import (
	"bytes"
	"io"

	"github.com/user/table-converter/internal/model"
	"github.com/xuri/excelize/v2"
)

// ExcelParser implements the Parser interface for Excel (.xlsx) format
type ExcelParser struct {
	// SheetName specifies which sheet to parse (empty = first sheet)
	SheetName string
}

// NewExcelParser creates a new Excel parser that reads the first sheet
func NewExcelParser() *ExcelParser {
	return &ExcelParser{}
}

// NewExcelParserWithSheet creates a new Excel parser for a specific sheet
func NewExcelParserWithSheet(sheetName string) *ExcelParser {
	return &ExcelParser{SheetName: sheetName}
}

// Parse reads Excel data from the input reader and converts it to TableData
func (p *ExcelParser) Parse(input io.Reader) (*model.TableData, error) {
	// Read all data into buffer (excelize requires random access)
	buf, err := io.ReadAll(input)
	if err != nil {
		return nil, NewParseError("failed to read input").WithErr(err)
	}

	// Open the Excel file from buffer
	f, err := excelize.OpenReader(bytes.NewReader(buf))
	if err != nil {
		return nil, NewParseError("failed to open Excel file").WithErr(err)
	}
	defer f.Close()

	// Determine which sheet to read
	sheetName := p.SheetName
	if sheetName == "" {
		// Use first sheet
		sheetList := f.GetSheetList()
		if len(sheetList) == 0 {
			return nil, NewParseError("Excel file contains no sheets")
		}
		sheetName = sheetList[0]
	}

	// Get sheet dimensions to determine the data range
	dim, err := f.GetSheetDimension(sheetName)
	if err != nil || dim == "" {
		// Fall back to GetRows for sheets without dimension info
		return p.parseWithGetRows(f, sheetName)
	}

	// Parse dimension (e.g., "A1:C10")
	startCell, endCell, err := parseDimension(dim)
	if err != nil {
		return p.parseWithGetRows(f, sheetName)
	}

	startCol, startRow, _ := excelize.CellNameToCoordinates(startCell)
	endCol, endRow, _ := excelize.CellNameToCoordinates(endCell)

	// Handle empty sheet
	if endRow < startRow || endCol < startCol {
		return model.NewTableData([]string{}, [][]model.Value{}), nil
	}

	// Use GetRows to get the actual data, but use dimension for row count
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, NewParseError("failed to read sheet").
			WithContext(sheetName).WithErr(err)
	}

	if len(rows) == 0 {
		return model.NewTableData([]string{}, [][]model.Value{}), nil
	}

	// Determine number of columns from the first row (headers)
	// This handles the case where dimension doesn't include empty columns
	numCols := len(rows[0])
	if numCols == 0 {
		return model.NewTableData([]string{}, [][]model.Value{}), nil
	}

	// Also consider the dimension's column range
	dimCols := endCol - startCol + 1
	if dimCols > numCols {
		numCols = dimCols
	}

	// Read headers from first row
	headers := make([]string, numCols)
	for i := 0; i < numCols; i++ {
		if i < len(rows[0]) {
			headers[i] = rows[0][i]
		}
	}

	// Determine number of data rows from dimension
	numDataRows := endRow - startRow // -1 for header, but dimension is 1-indexed
	if numDataRows < 0 {
		numDataRows = 0
	}
	// Also consider actual rows returned
	actualDataRows := len(rows) - 1
	if actualDataRows > numDataRows {
		numDataRows = actualDataRows
	}

	// Read data rows
	dataRows := make([][]model.Value, numDataRows)
	for i := 0; i < numDataRows; i++ {
		values := make([]model.Value, numCols)
		rowIdx := i + 1 // Skip header row
		for j := 0; j < numCols; j++ {
			cellRef, _ := excelize.CoordinatesToCellName(startCol+j, startRow+rowIdx)
			var cellValue string
			if rowIdx < len(rows) && j < len(rows[rowIdx]) {
				cellValue = rows[rowIdx][j]
			}
			values[j] = p.parseCellValue(f, sheetName, cellRef, cellValue)
		}
		dataRows[i] = values
	}

	return model.NewTableData(headers, dataRows), nil
}

// parseWithGetRows is a fallback parser using GetRows
func (p *ExcelParser) parseWithGetRows(f *excelize.File, sheetName string) (*model.TableData, error) {
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, NewParseError("failed to read sheet").
			WithContext(sheetName).WithErr(err)
	}

	if len(rows) == 0 {
		return model.NewTableData([]string{}, [][]model.Value{}), nil
	}

	headers := rows[0]
	if len(headers) == 0 {
		return model.NewTableData([]string{}, [][]model.Value{}), nil
	}

	dataRows := make([][]model.Value, 0, len(rows)-1)
	for rowIdx := 1; rowIdx < len(rows); rowIdx++ {
		row := rows[rowIdx]
		values := make([]model.Value, len(headers))
		for colIdx := 0; colIdx < len(headers); colIdx++ {
			if colIdx < len(row) {
				cellRef, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+1)
				values[colIdx] = p.parseCellValue(f, sheetName, cellRef, row[colIdx])
			} else {
				values[colIdx] = model.NewNullValue()
			}
		}
		dataRows = append(dataRows, values)
	}

	return model.NewTableData(headers, dataRows), nil
}

// parseDimension parses an Excel dimension string like "A1:C10"
func parseDimension(dim string) (start, end string, err error) {
	for i, c := range dim {
		if c == ':' {
			return dim[:i], dim[i+1:], nil
		}
	}
	// Single cell dimension
	return dim, dim, nil
}


// parseCellValue extracts the value from a cell with type preservation
func (p *ExcelParser) parseCellValue(f *excelize.File, sheet, cellRef, rawValue string) model.Value {
	if rawValue == "" {
		return model.NewNullValue()
	}

	// Try to get the cell type
	cellType, err := f.GetCellType(sheet, cellRef)
	if err != nil {
		// Fall back to type inference
		return model.NewValue(rawValue)
	}

	switch cellType {
	case excelize.CellTypeBool:
		// Parse as boolean
		if rawValue == "TRUE" || rawValue == "true" || rawValue == "1" {
			return model.NewBooleanValue(true)
		}
		return model.NewBooleanValue(false)

	case excelize.CellTypeNumber, excelize.CellTypeDate:
		// Parse as number - GetCellValue returns formatted string
		// Try to get the raw numeric value
		val, err := f.GetCellValue(sheet, cellRef)
		if err == nil && val != "" {
			return model.NewValue(val)
		}
		return model.NewValue(rawValue)

	case excelize.CellTypeFormula:
		// For formulas, use the calculated value
		return model.NewValue(rawValue)

	case excelize.CellTypeInlineString, excelize.CellTypeSharedString:
		// Explicit string type
		return model.NewStringValue(rawValue)

	default:
		// Use type inference for unknown types
		return model.NewValue(rawValue)
	}
}
