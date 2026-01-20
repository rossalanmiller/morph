package parser

import (
	"bytes"
	"testing"

	"github.com/user/table-converter/internal/model"
	"github.com/user/table-converter/internal/serializer"
	"pgregory.net/rapid"
)

// Feature: table-converter, Property 1: Round-Trip Preservation (HTML)
// Validates: Requirements 1.5, 2.5, 3.1
//
// Property: For any valid TableData, serializing to HTML and then parsing back
// should produce equivalent TableData (same headers, same number of rows, same values).
// Note: HTML format treats all values as strings, so type information is lost.
func TestProperty_HTMLRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate random TableData with HTML-safe values
		td := generateHTMLSafeTableData(t)

		// Serialize to HTML
		var buf bytes.Buffer
		htmlSerializer := serializer.NewHTMLSerializer()
		err := htmlSerializer.Serialize(td, &buf)
		if err != nil {
			t.Fatalf("failed to serialize TableData to HTML: %v", err)
		}

		// Parse back from HTML
		htmlParser := NewHTMLParser()
		parsedTD, err := htmlParser.Parse(&buf)
		if err != nil {
			t.Fatalf("failed to parse HTML back to TableData: %v", err)
		}

		// Property: Headers should be identical
		if len(parsedTD.Headers) != len(td.Headers) {
			t.Fatalf("header count mismatch: expected %d, got %d",
				len(td.Headers), len(parsedTD.Headers))
		}
		for i, header := range td.Headers {
			if parsedTD.Headers[i] != header {
				t.Fatalf("header %d mismatch: expected %q, got %q",
					i, header, parsedTD.Headers[i])
			}
		}

		// Property: Row count should be identical
		if len(parsedTD.Rows) != len(td.Rows) {
			t.Fatalf("row count mismatch: expected %d, got %d",
				len(td.Rows), len(parsedTD.Rows))
		}

		// Property: All values should be preserved (as strings in HTML)
		for i, row := range td.Rows {
			parsedRow := parsedTD.Rows[i]
			if len(parsedRow) != len(row) {
				t.Fatalf("row %d column count mismatch: expected %d, got %d",
					i, len(row), len(parsedRow))
			}

			for j, value := range row {
				parsedValue := parsedRow[j]
				// HTML treats all values as strings, compare string representations
				expected := valueToHTMLString(value)
				actual := parsedValue.String()
				if actual != expected {
					t.Fatalf("row %d, col %d value mismatch: expected %q, got %q",
						i, j, expected, actual)
				}
			}
		}

		// Property: Parsed data should be valid
		if err := parsedTD.Validate(); err != nil {
			t.Fatalf("parsed TableData failed validation: %v", err)
		}
	})
}


// generateHTMLSafeTableData creates a random TableData with HTML-safe values
func generateHTMLSafeTableData(t *rapid.T) *model.TableData {
	// Generate random headers (1-20 columns)
	// Use alphanumeric strings for headers
	numCols := rapid.IntRange(1, 20).Draw(t, "numCols")
	headers := make([]string, numCols)
	for i := 0; i < numCols; i++ {
		headers[i] = rapid.StringMatching(`[a-zA-Z][a-zA-Z0-9_]*`).Draw(t, "header")
	}

	// Generate random rows (0-100 rows)
	numRows := rapid.IntRange(0, 100).Draw(t, "numRows")
	rows := make([][]model.Value, numRows)
	for i := 0; i < numRows; i++ {
		row := make([]model.Value, numCols)
		for j := 0; j < numCols; j++ {
			row[j] = generateHTMLSafeValue(t)
		}
		rows[i] = row
	}

	return model.NewTableData(headers, rows)
}

// generateHTMLSafeValue creates a random Value that is safe for HTML round-trip
func generateHTMLSafeValue(t *rapid.T) model.Value {
	valueType := rapid.IntRange(0, 3).Draw(t, "valueType")

	switch valueType {
	case 0: // String - avoid whitespace-only strings as they get trimmed
		s := rapid.StringMatching(`[a-zA-Z0-9][a-zA-Z0-9 ]*[a-zA-Z0-9]|[a-zA-Z0-9]`).Draw(t, "stringValue")
		return model.NewStringValue(s)
	case 1: // Number
		n := rapid.Float64Range(-1e10, 1e10).Draw(t, "numberValue")
		return model.NewNumberValue(n)
	case 2: // Boolean
		b := rapid.Bool().Draw(t, "boolValue")
		return model.NewBooleanValue(b)
	case 3: // Null (becomes empty string in HTML)
		return model.NewNullValue()
	default:
		return model.NewStringValue("default")
	}
}

// valueToHTMLString converts a model.Value to its HTML string representation
func valueToHTMLString(val model.Value) string {
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


// Feature: table-converter, Property 8: Character Escaping (HTML)
// Validates: Requirements 7.2
//
// Property: For any TableData containing characters that are special in HTML (<, >, &),
// the serializer should properly escape them so the output is valid HTML,
// and the parser should decode them back to the original characters.
func TestProperty_HTMLCharacterEscaping(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate TableData with HTML special characters
		numCols := rapid.IntRange(1, 10).Draw(t, "numCols")
		headers := make([]string, numCols)
		for i := 0; i < numCols; i++ {
			headers[i] = rapid.StringMatching(`[a-zA-Z][a-zA-Z0-9_]*`).Draw(t, "header")
		}

		numRows := rapid.IntRange(1, 50).Draw(t, "numRows")
		rows := make([][]model.Value, numRows)
		for i := 0; i < numRows; i++ {
			row := make([]model.Value, numCols)
			for j := 0; j < numCols; j++ {
				// Generate values with HTML special characters
				row[j] = generateValueWithHTMLSpecialChars(t)
			}
			rows[i] = row
		}

		td := model.NewTableData(headers, rows)

		// Serialize to HTML
		var buf bytes.Buffer
		htmlSerializer := serializer.NewHTMLSerializer()
		err := htmlSerializer.Serialize(td, &buf)
		if err != nil {
			t.Fatalf("failed to serialize TableData to HTML: %v", err)
		}

		htmlOutput := buf.String()

		// Property: The HTML output should not contain unescaped special characters in cell values
		// (except in the HTML tags themselves)
		// This is implicitly tested by successful parsing

		// Parse back from HTML
		htmlParser := NewHTMLParser()
		parsedTD, err := htmlParser.Parse(bytes.NewBufferString(htmlOutput))
		if err != nil {
			t.Fatalf("failed to parse HTML back to TableData: %v", err)
		}

		// Property: All special characters should be preserved after round-trip
		for i, row := range td.Rows {
			parsedRow := parsedTD.Rows[i]
			for j, value := range row {
				parsedValue := parsedRow[j]
				expected := valueToHTMLString(value)
				actual := parsedValue.String()
				if actual != expected {
					t.Fatalf("row %d, col %d: special character not preserved\nexpected: %q\ngot: %q",
						i, j, expected, actual)
				}
			}
		}
	})
}

// generateValueWithHTMLSpecialChars creates a Value containing HTML special characters
func generateValueWithHTMLSpecialChars(t *rapid.T) model.Value {
	// Choose what type of special character to include
	charType := rapid.IntRange(0, 6).Draw(t, "charType")

	// Use alphanumeric base strings to avoid whitespace trimming issues
	baseString := rapid.StringMatching(`[a-zA-Z0-9]+`).Draw(t, "baseString")
	afterString := rapid.StringMatching(`[a-zA-Z0-9]+`).Draw(t, "afterString")

	var result string
	switch charType {
	case 0: // Less than
		result = baseString + "<" + afterString
	case 1: // Greater than
		result = baseString + ">" + afterString
	case 2: // Ampersand
		result = baseString + "&" + afterString
	case 3: // Double quote
		result = baseString + `"` + afterString
	case 4: // Single quote
		result = baseString + "'" + afterString
	case 5: // Combination of special chars
		result = baseString + "<>&\"'" + afterString
	case 6: // Regular string (control case)
		result = baseString
	}

	return model.NewStringValue(result)
}
