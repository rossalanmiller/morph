package serializer

import (
	"encoding/csv"
	"io"
	"strings"

	"github.com/user/table-converter/internal/model"
)

// CSVSerializer implements the Serializer interface for CSV format
type CSVSerializer struct {
	// Delimiter is the field delimiter (default: comma)
	Delimiter rune
	// LineTerminator is the line ending (default: \n)
	LineTerminator string
	// AlwaysQuote forces all fields to be quoted
	AlwaysQuote bool
}

// NewCSVSerializer creates a new CSV serializer with default settings
func NewCSVSerializer() *CSVSerializer {
	return &CSVSerializer{
		Delimiter:      ',',
		LineTerminator: "\n",
		AlwaysQuote:    false,
	}
}

// CSVSerializerOption is a function that configures a CSVSerializer
type CSVSerializerOption func(*CSVSerializer)

// WithDelimiter sets the field delimiter
func WithDelimiter(delim rune) CSVSerializerOption {
	return func(s *CSVSerializer) {
		s.Delimiter = delim
	}
}

// WithLineTerminator sets the line terminator
func WithLineTerminator(term string) CSVSerializerOption {
	return func(s *CSVSerializer) {
		s.LineTerminator = term
	}
}

// WithAlwaysQuote forces all fields to be quoted
func WithAlwaysQuote(quote bool) CSVSerializerOption {
	return func(s *CSVSerializer) {
		s.AlwaysQuote = quote
	}
}

// NewCSVSerializerWithOptions creates a CSV serializer with custom options
func NewCSVSerializerWithOptions(opts ...CSVSerializerOption) *CSVSerializer {
	s := NewCSVSerializer()
	for _, opt := range opts {
		opt(s)
	}
	return s
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

	// If always quoting, use custom writer
	if s.AlwaysQuote {
		return s.serializeWithQuotes(data, output)
	}

	writer := csv.NewWriter(output)
	writer.Comma = s.Delimiter
	writer.UseCRLF = s.LineTerminator == "\r\n"

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
		if len(record) == 1 && record[0] == "" {
			writer.Flush()
			if err := writer.Error(); err != nil {
				return NewSerializeError("failed to flush before special row").WithErr(err)
			}
			if _, err := output.Write([]byte("\"\"" + s.LineTerminator)); err != nil {
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

// serializeWithQuotes writes CSV with all fields quoted
func (s *CSVSerializer) serializeWithQuotes(data *model.TableData, output io.Writer) error {
	// Write headers
	if err := s.writeQuotedRow(data.Headers, output); err != nil {
		return err
	}

	// Write data rows
	for _, row := range data.Rows {
		record := make([]string, len(row))
		for j, value := range row {
			record[j] = value.String()
		}
		if err := s.writeQuotedRow(record, output); err != nil {
			return err
		}
	}

	return nil
}

// writeQuotedRow writes a single row with all fields quoted
func (s *CSVSerializer) writeQuotedRow(fields []string, output io.Writer) error {
	var builder strings.Builder

	for i, field := range fields {
		if i > 0 {
			builder.WriteRune(s.Delimiter)
		}
		builder.WriteByte('"')
		// Escape any quotes in the field
		for _, ch := range field {
			if ch == '"' {
				builder.WriteString("\"\"")
			} else {
				builder.WriteRune(ch)
			}
		}
		builder.WriteByte('"')
	}
	builder.WriteString(s.LineTerminator)

	_, err := output.Write([]byte(builder.String()))
	if err != nil {
		return NewSerializeError("failed to write CSV row").WithErr(err)
	}
	return nil
}

// ParseLineTerminator converts a string to a line terminator
func ParseLineTerminator(s string) string {
	switch strings.ToLower(s) {
	case "crlf", "\\r\\n", "\r\n", "windows":
		return "\r\n"
	case "lf", "\\n", "\n", "unix":
		return "\n"
	case "cr", "\\r", "\r", "mac":
		return "\r"
	default:
		return "\n"
	}
}
