package parser

import (
	"bufio"
	"io"
	"strings"

	"github.com/user/table-converter/internal/model"
)

// MarkdownParser implements the Parser interface for GitHub-flavored Markdown tables
type MarkdownParser struct{}

// NewMarkdownParser creates a new Markdown table parser
func NewMarkdownParser() *MarkdownParser {
	return &MarkdownParser{}
}

// Parse reads a Markdown table from the input reader and converts it to TableData
func (p *MarkdownParser) Parse(input io.Reader) (*model.TableData, error) {
	scanner := bufio.NewScanner(input)
	var lines []string

	// Read all lines
	for scanner.Scan() {
		line := scanner.Text()
		// Skip empty lines before/after table
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, NewParseError("failed to read input").WithErr(err)
	}

	if len(lines) == 0 {
		return model.NewTableData([]string{}, [][]model.Value{}), nil
	}

	// Need at least 2 lines: header and separator
	if len(lines) < 2 {
		return nil, NewParseError("invalid Markdown table: missing separator row").
			WithContext(lines[0])
	}

	// Parse header row
	headers := p.parseRow(lines[0])
	if len(headers) == 0 {
		return nil, NewParseError("invalid Markdown table: empty header row")
	}

	// Validate separator row (line 2)
	if !p.isSeparatorRow(lines[1]) {
		return nil, NewParseError("invalid Markdown table: second row must be separator").
			WithContext(lines[1])
	}

	// Parse data rows (lines 3+)
	var rows [][]model.Value
	for i := 2; i < len(lines); i++ {
		cells := p.parseRow(lines[i])
		values := make([]model.Value, len(cells))
		for j, cell := range cells {
			values[j] = model.NewValue(cell)
		}
		rows = append(rows, values)
	}

	return model.NewTableData(headers, rows), nil
}


// parseRow parses a Markdown table row into cells
// Handles: | cell1 | cell2 | cell3 |
// Also handles escaped pipes: \|
func (p *MarkdownParser) parseRow(line string) []string {
	line = strings.TrimSpace(line)

	// Remove leading/trailing pipes if present
	if strings.HasPrefix(line, "|") {
		line = line[1:]
	}
	if strings.HasSuffix(line, "|") {
		line = line[:len(line)-1]
	}

	// Split by unescaped pipes
	var cells []string
	var current strings.Builder
	escaped := false

	for i := 0; i < len(line); i++ {
		ch := line[i]
		if escaped {
			// Previous char was backslash
			if ch == '|' {
				current.WriteByte('|')
			} else {
				// Not an escaped pipe, keep the backslash
				current.WriteByte('\\')
				current.WriteByte(ch)
			}
			escaped = false
		} else if ch == '\\' {
			escaped = true
		} else if ch == '|' {
			cells = append(cells, strings.TrimSpace(current.String()))
			current.Reset()
		} else {
			current.WriteByte(ch)
		}
	}

	// Don't forget the last cell
	if current.Len() > 0 || len(cells) > 0 {
		cells = append(cells, strings.TrimSpace(current.String()))
	}

	return cells
}

// isSeparatorRow checks if a line is a Markdown table separator
// Valid separators: |---|---|, |:--|--:|, etc.
func (p *MarkdownParser) isSeparatorRow(line string) bool {
	line = strings.TrimSpace(line)

	// Must contain at least one dash
	if !strings.Contains(line, "-") {
		return false
	}

	// Remove pipes and check remaining chars
	for _, ch := range line {
		switch ch {
		case '|', '-', ':', ' ', '\t':
			// Valid separator characters
		default:
			return false
		}
	}

	return true
}
