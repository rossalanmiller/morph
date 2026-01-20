package serializer

import (
	"io"
	"strings"

	"github.com/user/table-converter/internal/model"
)

// ASCIISerializer implements the Serializer interface for ASCII box-drawing tables
// Uses simple box style with +, -, and | characters
type ASCIISerializer struct{}

// NewASCIISerializer creates a new ASCII table serializer
func NewASCIISerializer() *ASCIISerializer {
	return &ASCIISerializer{}
}

// Serialize writes TableData to the output writer as an ASCII table
func (s *ASCIISerializer) Serialize(data *model.TableData, output io.Writer) error {
	if data == nil {
		return NewSerializeError("TableData is nil")
	}

	if err := data.Validate(); err != nil {
		return NewSerializeError("invalid TableData").WithErr(err)
	}

	if len(data.Headers) == 0 {
		return nil // Empty table
	}

	// Calculate column widths
	widths := make([]int, len(data.Headers))
	for i, header := range data.Headers {
		widths[i] = len(header)
	}
	for _, row := range data.Rows {
		for i, value := range row {
			if i < len(widths) {
				cellLen := len(asciiValueToString(value))
				if cellLen > widths[i] {
					widths[i] = cellLen
				}
			}
		}
	}

	var sb strings.Builder

	// Write top border
	sb.WriteString(s.buildSeparator(widths))
	sb.WriteString("\n")

	// Write header row
	sb.WriteString(s.buildDataRow(data.Headers, widths))
	sb.WriteString("\n")

	// Write header separator
	sb.WriteString(s.buildSeparator(widths))
	sb.WriteString("\n")

	// Write data rows
	for _, row := range data.Rows {
		cells := make([]string, len(data.Headers))
		for i := 0; i < len(data.Headers); i++ {
			if i < len(row) {
				cells[i] = asciiValueToString(row[i])
			}
		}
		sb.WriteString(s.buildDataRow(cells, widths))
		sb.WriteString("\n")
	}

	// Write bottom border
	sb.WriteString(s.buildSeparator(widths))
	sb.WriteString("\n")

	_, err := output.Write([]byte(sb.String()))
	if err != nil {
		return NewSerializeError("failed to write ASCII output").WithErr(err)
	}

	return nil
}

// buildSeparator creates a separator line like +------+------+
func (s *ASCIISerializer) buildSeparator(widths []int) string {
	var sb strings.Builder
	sb.WriteString("+")
	for _, w := range widths {
		sb.WriteString(strings.Repeat("-", w+2))
		sb.WriteString("+")
	}
	return sb.String()
}

// buildDataRow creates a data row like | val1 | val2 |
func (s *ASCIISerializer) buildDataRow(cells []string, widths []int) string {
	var sb strings.Builder
	sb.WriteString("|")
	for i, cell := range cells {
		sb.WriteString(" ")
		sb.WriteString(cell)
		if i < len(widths) {
			sb.WriteString(strings.Repeat(" ", widths[i]-len(cell)))
		}
		sb.WriteString(" |")
	}
	return sb.String()
}

// asciiValueToString converts a model.Value to its string representation
func asciiValueToString(val model.Value) string {
	switch val.Type {
	case model.TypeNull:
		return ""
	case model.TypeBoolean:
		if b, ok := val.Parsed.(bool); ok {
			if b {
				return "true"
			}
			return "false"
		}
		return val.Raw
	case model.TypeNumber:
		return val.Raw
	case model.TypeString:
		if s, ok := val.Parsed.(string); ok {
			return s
		}
		return val.Raw
	default:
		return val.Raw
	}
}
