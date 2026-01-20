package serializer

import (
	"encoding/csv"
	"io"

	"github.com/user/table-converter/internal/model"
)

// CSVSerializer implements the Serializer interface for CSV format
type CSVSerializer struct{}

// NewCSVSerializer creates a new CSV serializer
func NewCSVSerializer() *CSVSerializer {
	return &CSVSerializer{}
}

// Serialize writes TableData to the output writer in CSV format
func (s *CSVSerializer) Serialize(data *model.TableData, output io.Writer) error {
	if data == nil {
		return NewSerializeError("TableData is nil")
	}

	// Validate the table data
	if err := data.Validate(); err != nil {
		return NewSerializeError("invalid TableData").WithErr(err)
	}

	writer := csv.NewWriter(output)
	
	// Write headers
	if err := writer.Write(data.Headers); err != nil {
		return NewSerializeError("failed to write CSV headers").WithErr(err)
	}

	// Flush headers before writing special case rows
	writer.Flush()
	if err := writer.Error(); err != nil {
		return NewSerializeError("failed to flush CSV headers").WithErr(err)
	}

	// Write data rows
	for _, row := range data.Rows {
		record := make([]string, len(row))
		for j, value := range row {
			record[j] = value.String()
		}
		
		// Special case: if all fields are empty and we have only one column,
		// we need to ensure the row is distinguishable from an empty line.
		// We do this by using a quoted empty string.
		if len(record) == 1 && record[0] == "" {
			// Flush any pending writes first
			writer.Flush()
			if err := writer.Error(); err != nil {
				return NewSerializeError("failed to flush before special row").WithErr(err)
			}
			// Write a quoted empty string manually
			if _, err := output.Write([]byte("\"\"\n")); err != nil {
				return NewSerializeError("failed to write CSV row").WithErr(err)
			}
		} else {
			if err := writer.Write(record); err != nil {
				return NewSerializeError("failed to write CSV row").WithErr(err)
			}
		}
	}

	// Final flush
	writer.Flush()
	if err := writer.Error(); err != nil {
		return NewSerializeError("failed to flush CSV data").WithErr(err)
	}

	return nil
}
