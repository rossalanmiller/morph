package parser

import (
	"encoding/csv"
	"io"

	"github.com/user/table-converter/internal/model"
)

// CSVParser implements the Parser interface for CSV format
type CSVParser struct{}

// NewCSVParser creates a new CSV parser
func NewCSVParser() *CSVParser {
	return &CSVParser{}
}

// Parse reads CSV data from the input reader and converts it to TableData
func (p *CSVParser) Parse(input io.Reader) (*model.TableData, error) {
	reader := csv.NewReader(input)
	reader.FieldsPerRecord = -1 // Allow variable number of fields
	
	// Read all records at once
	records, err := reader.ReadAll()
	if err != nil {
		return nil, NewParseError("failed to read CSV data").WithErr(err)
	}

	// Check if we have any data
	if len(records) == 0 {
		return nil, NewParseError("CSV file is empty")
	}

	// First row is headers
	headers := records[0]
	if len(headers) == 0 {
		return nil, NewParseError("CSV file has no columns")
	}

	// Parse remaining rows as data
	rows := make([][]model.Value, 0, len(records)-1)
	for i := 1; i < len(records); i++ {
		record := records[i]
		row := make([]model.Value, len(record))
		
		for j, field := range record {
			row[j] = model.NewValue(field)
		}
		
		rows = append(rows, row)
	}

	// NewTableData will normalize row lengths
	return model.NewTableData(headers, rows), nil
}
