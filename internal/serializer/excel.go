package serializer

import (
	"io"

	"github.com/user/table-converter/internal/model"
	"github.com/xuri/excelize/v2"
)

// ExcelSerializer implements the Serializer interface for Excel (.xlsx) format
type ExcelSerializer struct {
	// SheetName specifies the name of the sheet to create
	SheetName string
}

// NewExcelSerializer creates a new Excel serializer with default sheet name
func NewExcelSerializer() *ExcelSerializer {
	return &ExcelSerializer{SheetName: "Sheet1"}
}

// NewExcelSerializerWithSheet creates a new Excel serializer with a custom sheet name
func NewExcelSerializerWithSheet(sheetName string) *ExcelSerializer {
	return &ExcelSerializer{SheetName: sheetName}
}

// Serialize writes TableData to the output writer in Excel format
func (s *ExcelSerializer) Serialize(data *model.TableData, output io.Writer) error {
	if data == nil {
		return NewSerializeError("TableData is nil")
	}

	if err := data.Validate(); err != nil {
		return NewSerializeError("invalid TableData").WithErr(err)
	}

	// Create a new Excel file
	f := excelize.NewFile()
	defer f.Close()

	// Get the default sheet name and rename it
	defaultSheet := f.GetSheetName(0)
	sheetName := s.SheetName
	if sheetName == "" {
		sheetName = "Sheet1"
	}

	if defaultSheet != sheetName {
		if err := f.SetSheetName(defaultSheet, sheetName); err != nil {
			return NewSerializeError("failed to set sheet name").WithErr(err)
		}
	}

	// Write headers in first row
	for colIdx, header := range data.Headers {
		cellRef, err := excelize.CoordinatesToCellName(colIdx+1, 1)
		if err != nil {
			return NewSerializeError("failed to create cell reference").WithErr(err)
		}
		if err := f.SetCellValue(sheetName, cellRef, header); err != nil {
			return NewSerializeError("failed to write header").WithErr(err)
		}
	}

	// Write data rows
	for rowIdx, row := range data.Rows {
		excelRow := rowIdx + 2 // Excel rows are 1-indexed, +1 for header
		for colIdx, value := range row {
			cellRef, err := excelize.CoordinatesToCellName(colIdx+1, excelRow)
			if err != nil {
				return NewSerializeError("failed to create cell reference").WithErr(err)
			}
			if err := s.setCellValue(f, sheetName, cellRef, value); err != nil {
				return NewSerializeError("failed to write cell").WithErr(err)
			}
		}
	}

	// Write to output
	if err := f.Write(output); err != nil {
		return NewSerializeError("failed to write Excel file").WithErr(err)
	}

	return nil
}


// setCellValue writes a model.Value to an Excel cell with type preservation
func (s *ExcelSerializer) setCellValue(f *excelize.File, sheet, cellRef string, val model.Value) error {
	switch val.Type {
	case model.TypeNull:
		// Leave cell empty for null values
		return nil

	case model.TypeBoolean:
		if b, ok := val.Parsed.(bool); ok {
			return f.SetCellBool(sheet, cellRef, b)
		}
		return f.SetCellValue(sheet, cellRef, val.Raw)

	case model.TypeNumber:
		if n, ok := val.Parsed.(float64); ok {
			return f.SetCellFloat(sheet, cellRef, n, -1, 64)
		}
		return f.SetCellValue(sheet, cellRef, val.Raw)

	case model.TypeString:
		if str, ok := val.Parsed.(string); ok {
			return f.SetCellStr(sheet, cellRef, str)
		}
		return f.SetCellValue(sheet, cellRef, val.Raw)

	default:
		return f.SetCellValue(sheet, cellRef, val.Raw)
	}
}
