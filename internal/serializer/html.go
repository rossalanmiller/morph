package serializer

import (
	"html"
	"io"
	"strings"

	"github.com/user/table-converter/internal/model"
)

// HTMLSerializer implements the Serializer interface for HTML table format
type HTMLSerializer struct {
	// Indent specifies the indentation string for pretty printing
	// If empty, output will be compact
	Indent string
}

// NewHTMLSerializer creates a new HTML serializer with default settings (pretty print)
func NewHTMLSerializer() *HTMLSerializer {
	return &HTMLSerializer{
		Indent: "  ",
	}
}

// NewCompactHTMLSerializer creates a new HTML serializer with compact output
func NewCompactHTMLSerializer() *HTMLSerializer {
	return &HTMLSerializer{
		Indent: "",
	}
}

// Serialize writes TableData to the output writer in HTML table format
func (s *HTMLSerializer) Serialize(data *model.TableData, output io.Writer) error {
	if data == nil {
		return NewSerializeError("TableData is nil")
	}

	// Validate the table data
	if err := data.Validate(); err != nil {
		return NewSerializeError("invalid TableData").WithErr(err)
	}

	var sb strings.Builder
	indent := s.Indent
	newline := "\n"
	if indent == "" {
		newline = ""
	}

	// Write opening table tag
	sb.WriteString("<table>")
	sb.WriteString(newline)

	// Write thead with headers
	if len(data.Headers) > 0 {
		sb.WriteString(indent)
		sb.WriteString("<thead>")
		sb.WriteString(newline)
		sb.WriteString(indent)
		sb.WriteString(indent)
		sb.WriteString("<tr>")
		for _, header := range data.Headers {
			sb.WriteString("<th>")
			sb.WriteString(escapeHTML(header))
			sb.WriteString("</th>")
		}
		sb.WriteString("</tr>")
		sb.WriteString(newline)
		sb.WriteString(indent)
		sb.WriteString("</thead>")
		sb.WriteString(newline)
	}

	// Write tbody with data rows
	sb.WriteString(indent)
	sb.WriteString("<tbody>")
	sb.WriteString(newline)
	for _, row := range data.Rows {
		sb.WriteString(indent)
		sb.WriteString(indent)
		sb.WriteString("<tr>")
		for _, value := range row {
			sb.WriteString("<td>")
			sb.WriteString(escapeHTML(valueToString(value)))
			sb.WriteString("</td>")
		}
		sb.WriteString("</tr>")
		sb.WriteString(newline)
	}
	sb.WriteString(indent)
	sb.WriteString("</tbody>")
	sb.WriteString(newline)

	// Write closing table tag
	sb.WriteString("</table>")
	sb.WriteString(newline)

	// Write to output
	_, err := output.Write([]byte(sb.String()))
	if err != nil {
		return NewSerializeError("failed to write HTML output").WithErr(err)
	}

	return nil
}

// escapeHTML escapes special HTML characters
func escapeHTML(s string) string {
	return html.EscapeString(s)
}

// valueToString converts a model.Value to its string representation
func valueToString(val model.Value) string {
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
