package serializer

import (
	"io"
	"strings"

	"github.com/user/table-converter/internal/model"
)

// MarkdownSerializer implements the Serializer interface for GitHub-flavored Markdown tables
type MarkdownSerializer struct{}

// NewMarkdownSerializer creates a new Markdown table serializer
func NewMarkdownSerializer() *MarkdownSerializer {
	return &MarkdownSerializer{}
}

// Serialize writes TableData to the output writer as a Markdown table
func (s *MarkdownSerializer) Serialize(data *model.TableData, output io.Writer) error {
	if data == nil {
		return NewSerializeError("TableData is nil")
	}

	if err := data.Validate(); err != nil {
		return NewSerializeError("invalid TableData").WithErr(err)
	}

	if len(data.Headers) == 0 {
		return nil // Empty table
	}

	// Calculate column widths for alignment
	widths := make([]int, len(data.Headers))
	for i, header := range data.Headers {
		widths[i] = len(escapeMarkdown(header))
	}
	for _, row := range data.Rows {
		for i, value := range row {
			if i < len(widths) {
				cellLen := len(escapeMarkdown(valueToMarkdownString(value)))
				if cellLen > widths[i] {
					widths[i] = cellLen
				}
			}
		}
	}

	// Ensure minimum width of 3 for separator
	for i := range widths {
		if widths[i] < 3 {
			widths[i] = 3
		}
	}

	var sb strings.Builder

	// Write header row
	sb.WriteString("|")
	for i, header := range data.Headers {
		sb.WriteString(" ")
		cell := escapeMarkdown(header)
		sb.WriteString(cell)
		sb.WriteString(strings.Repeat(" ", widths[i]-len(cell)))
		sb.WriteString(" |")
	}
	sb.WriteString("\n")

	// Write separator row
	sb.WriteString("|")
	for _, w := range widths {
		sb.WriteString(" ")
		sb.WriteString(strings.Repeat("-", w))
		sb.WriteString(" |")
	}
	sb.WriteString("\n")

	// Write data rows
	for _, row := range data.Rows {
		sb.WriteString("|")
		for i := 0; i < len(data.Headers); i++ {
			sb.WriteString(" ")
			var cell string
			if i < len(row) {
				cell = escapeMarkdown(valueToMarkdownString(row[i]))
			}
			sb.WriteString(cell)
			sb.WriteString(strings.Repeat(" ", widths[i]-len(cell)))
			sb.WriteString(" |")
		}
		sb.WriteString("\n")
	}

	_, err := output.Write([]byte(sb.String()))
	if err != nil {
		return NewSerializeError("failed to write Markdown output").WithErr(err)
	}

	return nil
}

// escapeMarkdown escapes pipe characters in cell values
func escapeMarkdown(s string) string {
	return strings.ReplaceAll(s, "|", "\\|")
}

// valueToMarkdownString converts a model.Value to its Markdown string representation
func valueToMarkdownString(val model.Value) string {
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
