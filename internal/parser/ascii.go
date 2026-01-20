package parser

import (
	"bufio"
	"io"
	"strings"

	"github.com/user/table-converter/internal/model"
)

// ASCIIParser implements the Parser interface for ASCII box-drawing tables
// Supports simple box style with +, -, and | characters
type ASCIIParser struct{}

// NewASCIIParser creates a new ASCII table parser
func NewASCIIParser() *ASCIIParser {
	return &ASCIIParser{}
}

// Parse reads an ASCII table from the input reader and converts it to TableData
func (p *ASCIIParser) Parse(input io.Reader) (*model.TableData, error) {
	scanner := bufio.NewScanner(input)
	var lines []string

	// Read all non-empty lines
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, NewParseError("failed to read input").WithErr(err)
	}

	if len(lines) == 0 {
		return model.NewTableData([]string{}, [][]model.Value{}), nil
	}

	// Find column boundaries from separator lines
	colBoundaries := p.findColumnBoundaries(lines)
	if len(colBoundaries) < 2 {
		return nil, NewParseError("invalid ASCII table: cannot detect column boundaries")
	}

	// Parse data rows (skip separator lines)
	var headers []string
	var rows [][]model.Value
	headerFound := false

	for _, line := range lines {
		if p.isSeparatorLine(line) {
			continue
		}

		cells := p.parseDataRow(line, colBoundaries)
		
		if !headerFound {
			headers = cells
			headerFound = true
		} else {
			values := make([]model.Value, len(cells))
			for i, cell := range cells {
				values[i] = model.NewValue(cell)
			}
			rows = append(rows, values)
		}
	}

	if len(headers) == 0 {
		return model.NewTableData([]string{}, [][]model.Value{}), nil
	}

	return model.NewTableData(headers, rows), nil
}


// isSeparatorLine checks if a line is a separator (contains only +, -, and spaces)
func (p *ASCIIParser) isSeparatorLine(line string) bool {
	for _, ch := range line {
		switch ch {
		case '+', '-', ' ', '\t':
			// Valid separator characters
		default:
			return false
		}
	}
	return strings.Contains(line, "-")
}

// findColumnBoundaries finds the positions of | characters in data rows
func (p *ASCIIParser) findColumnBoundaries(lines []string) []int {
	// Find a data row (not a separator)
	for _, line := range lines {
		if !p.isSeparatorLine(line) && strings.Contains(line, "|") {
			var boundaries []int
			for i, ch := range line {
				if ch == '|' {
					boundaries = append(boundaries, i)
				}
			}
			return boundaries
		}
	}
	return nil
}

// parseDataRow extracts cell values from a data row using column boundaries
func (p *ASCIIParser) parseDataRow(line string, boundaries []int) []string {
	var cells []string

	for i := 0; i < len(boundaries)-1; i++ {
		start := boundaries[i] + 1
		end := boundaries[i+1]

		if start < len(line) && end <= len(line) {
			cell := line[start:end]
			cells = append(cells, strings.TrimSpace(cell))
		} else if start < len(line) {
			cell := line[start:]
			cells = append(cells, strings.TrimSpace(cell))
		} else {
			cells = append(cells, "")
		}
	}

	return cells
}
