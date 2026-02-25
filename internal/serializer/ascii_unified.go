package serializer

import (
	"fmt"
	"io"
	"strings"

	"github.com/user/table-converter/internal/model"
)

// TableStyle represents the output table style
type TableStyle string

const (
	StyleBox       TableStyle = "box"        // Traditional ASCII box with full borders (default)
	StylePsql      TableStyle = "psql"       // PostgreSQL aligned format
	StyleMarkdown  TableStyle = "md"         // Markdown table
	StyleOrgMode   TableStyle = "org"        // Emacs org-mode
	StyleRSTGrid   TableStyle = "rst-grid"   // reStructuredText grid table
	StyleRSTSimple TableStyle = "rst-simple" // reStructuredText simple table
)

// UnifiedASCIISerializer implements the Serializer interface for all ASCII-style table formats
type UnifiedASCIISerializer struct {
	Style          TableStyle
	RowSeparators  bool // Whether to add separators between data rows
}

// NewUnifiedASCIISerializer creates a new unified ASCII table serializer
func NewUnifiedASCIISerializer(style TableStyle) *UnifiedASCIISerializer {
	if style == "" {
		style = StyleBox // Default style
	}
	return &UnifiedASCIISerializer{
		Style:         style,
		RowSeparators: false,
	}
}

// SetStyle sets the table style for serialization
func (s *UnifiedASCIISerializer) SetStyle(style string) error {
	validStyles := map[string]TableStyle{
		"box":        StyleBox,
		"psql":       StylePsql,
		"md":         StyleMarkdown,
		"markdown":   StyleMarkdown,
		"org":        StyleOrgMode,
		"rst-grid":   StyleRSTGrid,
		"rst-simple": StyleRSTSimple,
	}

	if ts, ok := validStyles[style]; ok {
		s.Style = ts
		return nil
	}

	return fmt.Errorf("unsupported style %q, valid styles: box, psql, md, org, rst-grid, rst-simple", style)
}


// Serialize writes TableData to the output writer in the specified style
func (s *UnifiedASCIISerializer) Serialize(data *model.TableData, output io.Writer) error {
	if data == nil {
		return NewSerializeError("TableData is nil")
	}

	if err := data.Validate(); err != nil {
		return NewSerializeError("invalid TableData").WithErr(err)
	}

	if len(data.Headers) == 0 {
		return nil // Empty table
	}

	// Route to appropriate serializer based on style
	switch s.Style {
	case StyleRSTSimple:
		return s.serializeRSTSimple(data, output)
	case StyleMarkdown:
		return s.serializeMarkdown(data, output)
	case StyleOrgMode:
		return s.serializeOrgMode(data, output)
	case StylePsql:
		return s.serializePsql(data, output)
	case StyleRSTGrid:
		return s.serializeRSTGrid(data, output)
	case StyleBox:
		fallthrough
	default:
		return s.serializeBox(data, output)
	}
}

// calculateWidths computes the maximum width for each column
func (s *UnifiedASCIISerializer) calculateWidths(data *model.TableData) []int {
	widths := make([]int, len(data.Headers))
	for i, header := range data.Headers {
		widths[i] = len(header)
	}
	for _, row := range data.Rows {
		for i, value := range row {
			if i < len(widths) {
				cellLen := len(unifiedValueToString(value))
				if cellLen > widths[i] {
					widths[i] = cellLen
				}
			}
		}
	}
	// Ensure minimum width of 3
	for i := range widths {
		if widths[i] < 3 {
			widths[i] = 3
		}
	}
	return widths
}

// serializeBox outputs traditional ASCII box format
func (s *UnifiedASCIISerializer) serializeBox(data *model.TableData, output io.Writer) error {
	widths := s.calculateWidths(data)
	var sb strings.Builder

	// Top border
	sb.WriteString(s.buildBorder(widths, '+', '-', '+'))
	sb.WriteString("\n")

	// Header row
	sb.WriteString(s.buildRow(data.Headers, widths, '|'))
	sb.WriteString("\n")

	// Header separator
	sb.WriteString(s.buildBorder(widths, '+', '-', '+'))
	sb.WriteString("\n")

	// Data rows
	for i, row := range data.Rows {
		cells := s.rowToCells(row, data.Headers)
		sb.WriteString(s.buildRow(cells, widths, '|'))
		sb.WriteString("\n")

		if s.RowSeparators && i < len(data.Rows)-1 {
			sb.WriteString(s.buildBorder(widths, '+', '-', '+'))
			sb.WriteString("\n")
		}
	}

	// Bottom border
	sb.WriteString(s.buildBorder(widths, '+', '-', '+'))
	sb.WriteString("\n")

	_, err := output.Write([]byte(sb.String()))
	return err
}

// serializePsql outputs PostgreSQL aligned format
func (s *UnifiedASCIISerializer) serializePsql(data *model.TableData, output io.Writer) error {
	widths := s.calculateWidths(data)
	var sb strings.Builder

	// Header row (no leading/trailing borders)
	sb.WriteString(s.buildPsqlRow(data.Headers, widths))
	sb.WriteString("\n")

	// Header separator
	sb.WriteString(s.buildPsqlSeparator(widths))
	sb.WriteString("\n")

	// Data rows
	for _, row := range data.Rows {
		cells := s.rowToCells(row, data.Headers)
		sb.WriteString(s.buildPsqlRow(cells, widths))
		sb.WriteString("\n")
	}

	_, err := output.Write([]byte(sb.String()))
	return err
}

// serializeMarkdown outputs Markdown table format
func (s *UnifiedASCIISerializer) serializeMarkdown(data *model.TableData, output io.Writer) error {
	widths := s.calculateWidths(data)
	var sb strings.Builder

	// Header row
	sb.WriteString(s.buildRow(data.Headers, widths, '|'))
	sb.WriteString("\n")

	// Separator row (all dashes, no +)
	sb.WriteString("|")
	for _, w := range widths {
		sb.WriteString(" ")
		sb.WriteString(strings.Repeat("-", w))
		sb.WriteString(" |")
	}
	sb.WriteString("\n")

	// Data rows
	for _, row := range data.Rows {
		cells := s.rowToCells(row, data.Headers)
		sb.WriteString(s.buildRow(cells, widths, '|'))
		sb.WriteString("\n")
	}

	_, err := output.Write([]byte(sb.String()))
	return err
}

// serializeOrgMode outputs Emacs org-mode format
func (s *UnifiedASCIISerializer) serializeOrgMode(data *model.TableData, output io.Writer) error {
	widths := s.calculateWidths(data)
	var sb strings.Builder

	// Header row
	sb.WriteString(s.buildRow(data.Headers, widths, '|'))
	sb.WriteString("\n")

	// Separator row (with + at intersections)
	sb.WriteString("|")
	for i, w := range widths {
		sb.WriteString(strings.Repeat("-", w+2))
		if i < len(widths)-1 {
			sb.WriteString("+")
		} else {
			sb.WriteString("|")
		}
	}
	sb.WriteString("\n")

	// Data rows
	for _, row := range data.Rows {
		cells := s.rowToCells(row, data.Headers)
		sb.WriteString(s.buildRow(cells, widths, '|'))
		sb.WriteString("\n")
	}

	_, err := output.Write([]byte(sb.String()))
	return err
}

// serializeRSTGrid outputs reStructuredText grid table format
func (s *UnifiedASCIISerializer) serializeRSTGrid(data *model.TableData, output io.Writer) error {
	widths := s.calculateWidths(data)
	var sb strings.Builder

	// Top border
	sb.WriteString(s.buildBorder(widths, '+', '-', '+'))
	sb.WriteString("\n")

	// Header row
	sb.WriteString(s.buildRow(data.Headers, widths, '|'))
	sb.WriteString("\n")

	// Header separator (uses = instead of -)
	sb.WriteString(s.buildBorder(widths, '+', '=', '+'))
	sb.WriteString("\n")

	// Data rows
	for i, row := range data.Rows {
		cells := s.rowToCells(row, data.Headers)
		sb.WriteString(s.buildRow(cells, widths, '|'))
		sb.WriteString("\n")

		if s.RowSeparators && i < len(data.Rows)-1 {
			sb.WriteString(s.buildBorder(widths, '+', '-', '+'))
			sb.WriteString("\n")
		}
	}

	// Bottom border
	sb.WriteString(s.buildBorder(widths, '+', '-', '+'))
	sb.WriteString("\n")

	_, err := output.Write([]byte(sb.String()))
	return err
}

// serializeRSTSimple outputs reStructuredText simple table format
func (s *UnifiedASCIISerializer) serializeRSTSimple(data *model.TableData, output io.Writer) error {
	widths := s.calculateWidths(data)
	var sb strings.Builder

	// Top separator
	sb.WriteString(s.buildRSTSimpleSeparator(widths))
	sb.WriteString("\n")

	// Header row
	sb.WriteString(s.buildRSTSimpleRow(data.Headers, widths))
	sb.WriteString("\n")

	// Header separator
	sb.WriteString(s.buildRSTSimpleSeparator(widths))
	sb.WriteString("\n")

	// Data rows
	for _, row := range data.Rows {
		cells := s.rowToCells(row, data.Headers)
		sb.WriteString(s.buildRSTSimpleRow(cells, widths))
		sb.WriteString("\n")
	}

	// Bottom separator
	sb.WriteString(s.buildRSTSimpleSeparator(widths))
	sb.WriteString("\n")

	_, err := output.Write([]byte(sb.String()))
	return err
}

// buildBorder creates a border line like +------+------+
func (s *UnifiedASCIISerializer) buildBorder(widths []int, corner, fill, sep rune) string {
	var sb strings.Builder
	sb.WriteRune(corner)
	for i, w := range widths {
		sb.WriteString(strings.Repeat(string(fill), w+2))
		if i < len(widths)-1 {
			sb.WriteRune(sep)
		} else {
			sb.WriteRune(corner)
		}
	}
	return sb.String()
}

// buildRow creates a data row like | val1 | val2 |
func (s *UnifiedASCIISerializer) buildRow(cells []string, widths []int, border rune) string {
	var sb strings.Builder
	sb.WriteRune(border)
	for i, cell := range cells {
		sb.WriteString(" ")
		sb.WriteString(cell)
		if i < len(widths) {
			sb.WriteString(strings.Repeat(" ", widths[i]-len(cell)))
		}
		sb.WriteString(" ")
		sb.WriteRune(border)
	}
	return sb.String()
}

// buildPsqlRow creates a psql-style row without leading/trailing borders
func (s *UnifiedASCIISerializer) buildPsqlRow(cells []string, widths []int) string {
	var sb strings.Builder
	for i, cell := range cells {
		sb.WriteString(cell)
		if i < len(widths) {
			sb.WriteString(strings.Repeat(" ", widths[i]-len(cell)))
		}
		if i < len(cells)-1 {
			sb.WriteString(" | ")
		}
	}
	return sb.String()
}

// buildPsqlSeparator creates a psql-style separator line
func (s *UnifiedASCIISerializer) buildPsqlSeparator(widths []int) string {
	var sb strings.Builder
	for i, w := range widths {
		sb.WriteString(strings.Repeat("-", w))
		if i < len(widths)-1 {
			sb.WriteString("-+-")
		}
	}
	return sb.String()
}

// buildRSTSimpleSeparator creates an RST simple separator line
func (s *UnifiedASCIISerializer) buildRSTSimpleSeparator(widths []int) string {
	var sb strings.Builder
	for i, w := range widths {
		sb.WriteString(strings.Repeat("=", w))
		if i < len(widths)-1 {
			sb.WriteString("  ")
		}
	}
	return sb.String()
}

// buildRSTSimpleRow creates an RST simple data row
func (s *UnifiedASCIISerializer) buildRSTSimpleRow(cells []string, widths []int) string {
	var sb strings.Builder
	for i, cell := range cells {
		sb.WriteString(cell)
		if i < len(widths) {
			sb.WriteString(strings.Repeat(" ", widths[i]-len(cell)))
		}
		if i < len(cells)-1 {
			sb.WriteString("  ")
		}
	}
	return sb.String()
}

// rowToCells converts a row of Values to strings
func (s *UnifiedASCIISerializer) rowToCells(row []model.Value, headers []string) []string {
	cells := make([]string, len(headers))
	for i := 0; i < len(headers); i++ {
		if i < len(row) {
			cells[i] = unifiedValueToString(row[i])
		}
	}
	return cells
}

// unifiedValueToString converts a model.Value to its string representation
func unifiedValueToString(val model.Value) string {
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
