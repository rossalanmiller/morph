package parser

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"io"
	"strings"

	"github.com/user/table-converter/internal/model"
)

// Common delimiters to try for auto-detection
var commonDelimiters = []rune{',', '\t', ';', '|'}

// CSVParser implements the Parser interface for CSV format
type CSVParser struct {
	// Delimiter is the field delimiter. If zero, auto-detect.
	Delimiter rune
}

// NewCSVParser creates a new CSV parser with auto-detection
func NewCSVParser() *CSVParser {
	return &CSVParser{
		Delimiter: 0, // Auto-detect
	}
}

// NewCSVParserWithDelimiter creates a CSV parser with a specific delimiter
func NewCSVParserWithDelimiter(delimiter rune) *CSVParser {
	return &CSVParser{
		Delimiter: delimiter,
	}
}

// Parse reads CSV data from the input reader and converts it to TableData
func (p *CSVParser) Parse(input io.Reader) (*model.TableData, error) {
	// Read all input first (needed for delimiter detection)
	data, err := io.ReadAll(input)
	if err != nil {
		return nil, NewParseError("failed to read CSV data").WithErr(err)
	}

	if len(data) == 0 {
		return nil, NewParseError("CSV file is empty")
	}

	// Determine delimiter
	delimiter := p.Delimiter
	if delimiter == 0 {
		delimiter = detectDelimiter(data)
	}

	// Parse with detected/specified delimiter
	reader := csv.NewReader(bytes.NewReader(data))
	reader.Comma = delimiter
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	// Read all records at once
	records, err := reader.ReadAll()
	if err != nil {
		return nil, NewParseError("failed to parse CSV data").WithErr(err)
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

// detectDelimiter attempts to auto-detect the CSV delimiter
// by analyzing the first few lines of the file
func detectDelimiter(data []byte) rune {
	// Read first few lines for analysis
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var lines []string
	for i := 0; i < 5 && scanner.Scan(); i++ {
		lines = append(lines, scanner.Text())
	}

	if len(lines) == 0 {
		return ',' // Default to comma
	}

	// Score each delimiter based on consistency
	bestDelimiter := ','
	bestScore := -1

	for _, delim := range commonDelimiters {
		score := scoreDelimiter(lines, delim)
		if score > bestScore {
			bestScore = score
			bestDelimiter = delim
		}
	}

	return bestDelimiter
}

// scoreDelimiter scores a delimiter based on how consistently it splits lines
// Higher score = more likely to be the correct delimiter
func scoreDelimiter(lines []string, delim rune) int {
	if len(lines) == 0 {
		return 0
	}

	// Count fields per line
	counts := make([]int, len(lines))
	for i, line := range lines {
		counts[i] = countFields(line, delim)
	}

	// Check if first line (header) has at least 2 fields
	if counts[0] < 2 {
		return 0
	}

	// Score based on consistency: all lines should have same field count
	// and that count should be > 1
	headerCount := counts[0]
	score := headerCount * 10 // Base score from number of columns

	for i := 1; i < len(counts); i++ {
		if counts[i] == headerCount {
			score += 10 // Bonus for consistent row
		} else {
			score -= 5 // Penalty for inconsistent row
		}
	}

	return score
}

// countFields counts the number of fields in a line for a given delimiter
// This is a simple count that doesn't handle quoted fields perfectly,
// but is good enough for delimiter detection
func countFields(line string, delim rune) int {
	if line == "" {
		return 0
	}

	// Simple approach: count delimiters outside of quotes
	count := 1
	inQuotes := false

	for _, ch := range line {
		switch ch {
		case '"':
			inQuotes = !inQuotes
		case delim:
			if !inQuotes {
				count++
			}
		}
	}

	return count
}

// DetectedDelimiterName returns a human-readable name for a delimiter
func DetectedDelimiterName(delim rune) string {
	switch delim {
	case ',':
		return "comma"
	case '\t':
		return "tab"
	case ';':
		return "semicolon"
	case '|':
		return "pipe"
	default:
		return string(delim)
	}
}

// ParseDelimiter converts a string to a delimiter rune
func ParseDelimiter(s string) rune {
	s = strings.ToLower(s)
	switch s {
	case "comma", ",":
		return ','
	case "tab", "\\t", "\t":
		return '\t'
	case "semicolon", ";":
		return ';'
	case "pipe", "|":
		return '|'
	case "space", " ":
		return ' '
	default:
		if len(s) == 1 {
			return rune(s[0])
		}
		return ','
	}
}
