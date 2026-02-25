package parser

import (
	"bufio"
	"io"
	"strings"

	"github.com/user/table-converter/internal/model"
)

// TableStyle represents the detected or desired table style
type TableStyle string

const (
	StyleBox       TableStyle = "box"        // Traditional ASCII box with full borders
	StylePsql      TableStyle = "psql"       // PostgreSQL aligned format
	StyleMarkdown  TableStyle = "md"         // Markdown table
	StyleOrgMode   TableStyle = "org"        // Emacs org-mode
	StyleRSTGrid   TableStyle = "rst-grid"   // reStructuredText grid table
	StyleRSTSimple TableStyle = "rst-simple" // reStructuredText simple table
)

// UnifiedASCIIParser implements the Parser interface for all ASCII-style table formats
// Supports: ASCII box, psql, Markdown, Org-mode, reStructuredText (grid and simple)
type UnifiedASCIIParser struct {
	DetectedStyle TableStyle // The style that was detected during parsing
}

// NewUnifiedASCIIParser creates a new unified ASCII table parser
func NewUnifiedASCIIParser() *UnifiedASCIIParser {
	return &UnifiedASCIIParser{}
}

// Parse reads an ASCII-style table and auto-detects the format
func (p *UnifiedASCIIParser) Parse(input io.Reader) (*model.TableData, error) {
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

	// Detect the table style
	style := p.detectStyle(lines)
	p.DetectedStyle = style

	// Parse based on detected style
	switch style {
	case StyleRSTSimple:
		return p.parseRSTSimple(lines)
	default:
		// All pipe-based formats use similar parsing
		return p.parsePipeBased(lines, style)
	}
}

// detectStyle determines which table format is being used
func (p *UnifiedASCIIParser) detectStyle(lines []string) TableStyle {
	// Check for RST Simple (uses = only, no pipes)
	if p.isRSTSimple(lines) {
		return StyleRSTSimple
	}

	// Find separator lines for pipe-based formats
	var sepLines []string
	for _, line := range lines {
		if p.isSeparatorLine(line) {
			sepLines = append(sepLines, line)
		}
	}

	if len(sepLines) == 0 {
		// No separator found, default to box
		return StyleBox
	}

	// Check for RST Grid (uses = in header separator)
	// RST Grid has a line with +===+ pattern
	for _, sepLine := range sepLines {
		if strings.Contains(sepLine, "=") && strings.Contains(sepLine, "+") {
			return StyleRSTGrid
		}
	}

	// Use first separator line for other checks
	sepLine := sepLines[0]
	var sepIndex int
	for i, line := range lines {
		if line == sepLine {
			sepIndex = i
			break
		}
	}

	// Check if it's psql format (no leading border)
	trimmed := strings.TrimSpace(sepLine)
	if len(trimmed) > 0 && trimmed[0] != '|' && trimmed[0] != '+' {
		return StylePsql
	}

	// Check for full box borders
	if strings.HasPrefix(trimmed, "+") && strings.HasSuffix(trimmed, "+") {
		return StyleBox
	}

	// Distinguish between Markdown and Org-mode
	// Org-mode uses + at intersections, Markdown uses only -
	if strings.Contains(sepLine, "+") {
		// Could be org-mode or box
		// Check if there's a data line to distinguish
		if sepIndex > 0 {
			dataLine := lines[sepIndex-1]
			if strings.HasPrefix(strings.TrimSpace(dataLine), "|") {
				// Has leading pipe, check for + in separator
				if p.hasIntersectionPlus(sepLine) {
					return StyleOrgMode
				}
			}
		}
		return StyleBox
	}

	// Default to Markdown (uses | and - only)
	return StyleMarkdown
}

// isRSTSimple checks if the table uses reStructuredText simple format
func (p *UnifiedASCIIParser) isRSTSimple(lines []string) bool {
	// RST Simple uses = for separators, no | characters
	hasEquals := false
	hasPipes := false

	for _, line := range lines {
		if strings.Contains(line, "|") {
			hasPipes = true
		}
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 0 && strings.Trim(trimmed, "= \t") == "" {
			hasEquals = true
		}
	}

	return hasEquals && !hasPipes
}

// hasIntersectionPlus checks if + appears at column intersections (org-mode style)
func (p *UnifiedASCIIParser) hasIntersectionPlus(sepLine string) bool {
	// In org-mode, the separator looks like: |---------+-----+---------|
	// The + appears between columns, not at the edges
	trimmed := strings.TrimSpace(sepLine)
	if !strings.HasPrefix(trimmed, "|") {
		return false
	}
	// Remove leading and trailing |
	inner := strings.Trim(trimmed, "|")
	return strings.Contains(inner, "+")
}

// isSeparatorLine checks if a line is a separator
func (p *UnifiedASCIIParser) isSeparatorLine(line string) bool {
	for _, ch := range line {
		switch ch {
		case '+', '-', '=', '|', ':', ' ', '\t':
			// Valid separator characters
		default:
			return false
		}
	}
	return strings.ContainsAny(line, "-=")
}

// parsePipeBased parses pipe-based table formats (box, psql, markdown, org, rst-grid)
func (p *UnifiedASCIIParser) parsePipeBased(lines []string, style TableStyle) (*model.TableData, error) {
	// Find column boundaries
	colBoundaries := p.findColumnBoundaries(lines, style)
	if len(colBoundaries) < 2 {
		return nil, NewParseError("invalid table: cannot detect column boundaries")
	}

	// Parse data rows (skip separator lines)
	var headers []string
	var rows [][]model.Value
	headerFound := false

	for _, line := range lines {
		if p.isSeparatorLine(line) {
			continue
		}

		cells := p.parseDataRow(line, colBoundaries, style)

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

// findColumnBoundaries finds column separator positions
func (p *UnifiedASCIIParser) findColumnBoundaries(lines []string, style TableStyle) []int {
	var maxLen int
	for _, line := range lines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}

	// For psql format, look for + in separator line
	if style == StylePsql {
		for _, line := range lines {
			if p.isSeparatorLine(line) && strings.Contains(line, "+") {
				boundaries := []int{0}
				for i, ch := range line {
					if ch == '+' {
						boundaries = append(boundaries, i)
					}
				}
				boundaries = append(boundaries, maxLen)
				return boundaries
			}
		}
	}

	// For other formats, find | in data rows
	for _, line := range lines {
		if !p.isSeparatorLine(line) && strings.Contains(line, "|") {
			var boundaries []int
			for i, ch := range line {
				if ch == '|' {
					boundaries = append(boundaries, i)
				}
			}
			if len(boundaries) >= 2 {
				return boundaries
			}
		}
	}

	return nil
}

// parseDataRow extracts cell values from a data row
func (p *UnifiedASCIIParser) parseDataRow(line string, boundaries []int, style TableStyle) []string {
	var cells []string
	isPsqlFormat := style == StylePsql

	for i := 0; i < len(boundaries)-1; i++ {
		start := boundaries[i]
		end := boundaries[i+1]

		if isPsqlFormat {
			// psql format: boundaries mark + positions, extract between them
			if start < len(line) && end <= len(line) {
				cell := line[start:end]
				cell = strings.Trim(cell, "|")
				cells = append(cells, strings.TrimSpace(cell))
			} else if start < len(line) {
				cell := line[start:]
				cell = strings.Trim(cell, "|")
				cells = append(cells, strings.TrimSpace(cell))
			} else {
				cells = append(cells, "")
			}
		} else {
			// Other formats: boundaries mark | positions
			start++ // Skip the | character
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
	}

	return cells
}

// parseRSTSimple parses reStructuredText simple table format
func (p *UnifiedASCIIParser) parseRSTSimple(lines []string) (*model.TableData, error) {
	// Find separator lines (lines with only = and spaces)
	var separatorIndices []int
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 0 && strings.Trim(trimmed, "= \t") == "" {
			separatorIndices = append(separatorIndices, i)
		}
	}

	if len(separatorIndices) < 2 {
		return nil, NewParseError("invalid RST simple table: need at least 2 separator lines")
	}

	// Use first separator to find column boundaries
	sepLine := lines[separatorIndices[0]]
	colBoundaries := p.findRSTSimpleColumns(sepLine)

	if len(colBoundaries) == 0 {
		return nil, NewParseError("invalid RST simple table: cannot detect columns")
	}

	// Header is between first and second separator
	var headers []string
	if separatorIndices[0]+1 < separatorIndices[1] {
		headerLine := lines[separatorIndices[0]+1]
		headers = p.parseRSTSimpleRow(headerLine, colBoundaries)
	}

	// Data rows are between second separator and last separator (or end)
	var rows [][]model.Value
	startRow := separatorIndices[1] + 1
	endRow := len(lines)
	if len(separatorIndices) > 2 {
		endRow = separatorIndices[len(separatorIndices)-1]
	}

	for i := startRow; i < endRow; i++ {
		cells := p.parseRSTSimpleRow(lines[i], colBoundaries)
		values := make([]model.Value, len(cells))
		for j, cell := range cells {
			values[j] = model.NewValue(cell)
		}
		rows = append(rows, values)
	}

	return model.NewTableData(headers, rows), nil
}

// findRSTSimpleColumns finds column boundaries from = separator line
func (p *UnifiedASCIIParser) findRSTSimpleColumns(sepLine string) [][]int {
	// Find sequences of = characters
	var columns [][]int
	inColumn := false
	start := 0

	for i, ch := range sepLine {
		if ch == '=' {
			if !inColumn {
				start = i
				inColumn = true
			}
		} else {
			if inColumn {
				columns = append(columns, []int{start, i})
				inColumn = false
			}
		}
	}

	if inColumn {
		columns = append(columns, []int{start, len(sepLine)})
	}

	return columns
}

// parseRSTSimpleRow extracts cells from an RST simple table row
func (p *UnifiedASCIIParser) parseRSTSimpleRow(line string, colBoundaries [][]int) []string {
	var cells []string

	for _, bounds := range colBoundaries {
		start, end := bounds[0], bounds[1]
		if start < len(line) {
			if end > len(line) {
				end = len(line)
			}
			cell := line[start:end]
			cells = append(cells, strings.TrimSpace(cell))
		} else {
			cells = append(cells, "")
		}
	}

	return cells
}
