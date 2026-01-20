package parser

import (
	"bytes"
	"strings"
	"testing"

	"github.com/user/table-converter/internal/model"
	"github.com/user/table-converter/internal/serializer"
	"pgregory.net/rapid"
)

// Feature: table-converter, Property 1: Round-Trip Preservation (Markdown)
// Validates: Requirements 1.7, 2.7, 3.1
//
// Property: For any valid TableData, serializing to Markdown and then parsing back
// should produce equivalent TableData (same headers, same number of rows, same values).
func TestProperty_MarkdownRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate random TableData with Markdown-safe values
		td := generateMarkdownSafeTableData(t)

		// Serialize to Markdown
		var buf bytes.Buffer
		mdSerializer := serializer.NewMarkdownSerializer()
		err := mdSerializer.Serialize(td, &buf)
		if err != nil {
			t.Fatalf("failed to serialize TableData to Markdown: %v", err)
		}

		// Parse back from Markdown
		mdParser := NewMarkdownParser()
		parsedTD, err := mdParser.Parse(&buf)
		if err != nil {
			t.Fatalf("failed to parse Markdown back to TableData: %v\nMarkdown:\n%s", err, buf.String())
		}

		// Property: Headers should be identical (after trimming whitespace)
		if len(parsedTD.Headers) != len(td.Headers) {
			t.Fatalf("header count mismatch: expected %d, got %d",
				len(td.Headers), len(parsedTD.Headers))
		}
		for i, header := range td.Headers {
			expected := strings.TrimSpace(header)
			if parsedTD.Headers[i] != expected {
				t.Fatalf("header %d mismatch: expected %q, got %q",
					i, expected, parsedTD.Headers[i])
			}
		}

		// Property: Row count should be identical
		if len(parsedTD.Rows) != len(td.Rows) {
			t.Fatalf("row count mismatch: expected %d, got %d",
				len(td.Rows), len(parsedTD.Rows))
		}

		// Property: All values should be preserved (as strings in Markdown)
		// Note: Markdown trims whitespace-only values to empty strings
		for i, row := range td.Rows {
			parsedRow := parsedTD.Rows[i]
			if len(parsedRow) != len(row) {
				t.Fatalf("row %d column count mismatch: expected %d, got %d",
					i, len(row), len(parsedRow))
			}

			for j, value := range row {
				parsedValue := parsedRow[j]
				expected := normalizeWhitespace(valueToString(value))
				got := parsedValue.Raw
				if got != expected {
					t.Fatalf("row %d, col %d value mismatch: expected %q, got %q",
						i, j, expected, got)
				}
			}
		}

		// Property: Parsed data should be valid
		if err := parsedTD.Validate(); err != nil {
			t.Fatalf("parsed TableData failed validation: %v", err)
		}
	})
}

// generateMarkdownSafeTableData creates a random TableData with Markdown-compatible values
func generateMarkdownSafeTableData(t *rapid.T) *model.TableData {
	// Generate random headers (1-10 columns)
	numCols := rapid.IntRange(1, 10).Draw(t, "numCols")
	headers := make([]string, numCols)
	for i := 0; i < numCols; i++ {
		// Use alphanumeric strings for headers (no pipes or newlines)
		headers[i] = rapid.StringMatching(`[a-zA-Z][a-zA-Z0-9_ ]{0,15}`).Draw(t, "header")
	}

	// Generate random rows (0-50 rows)
	numRows := rapid.IntRange(0, 50).Draw(t, "numRows")
	rows := make([][]model.Value, numRows)
	for i := 0; i < numRows; i++ {
		row := make([]model.Value, numCols)
		for j := 0; j < numCols; j++ {
			row[j] = generateMarkdownSafeValue(t)
		}
		rows[i] = row
	}

	return model.NewTableData(headers, rows)
}

// generateMarkdownSafeValue creates a random Value that Markdown can handle
func generateMarkdownSafeValue(t *rapid.T) model.Value {
	valueType := rapid.IntRange(0, 3).Draw(t, "valueType")

	switch valueType {
	case 0: // String - no pipes or newlines
		s := rapid.StringMatching(`[a-zA-Z0-9 .,!?-]{0,30}`).Draw(t, "stringValue")
		return model.NewStringValue(s)
	case 1: // Number
		n := rapid.Float64Range(-1e6, 1e6).Draw(t, "numberValue")
		return model.NewNumberValue(n)
	case 2: // Boolean
		b := rapid.Bool().Draw(t, "boolValue")
		return model.NewBooleanValue(b)
	case 3: // Null
		return model.NewNullValue()
	default:
		return model.NewStringValue("")
	}
}

// valueToString converts a model.Value to its string representation for comparison
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


// normalizeWhitespace trims leading/trailing whitespace from strings
// Markdown tables trim cell values, so this is expected behavior
func normalizeWhitespace(s string) string {
	return strings.TrimSpace(s)
}
