package parser

import (
	"bytes"
	"strings"
	"testing"

	"github.com/user/table-converter/internal/model"
	"github.com/user/table-converter/internal/serializer"
	"pgregory.net/rapid"
)

// Feature: table-converter, Property 1: Round-Trip Preservation (CSV)
// Validates: Requirements 1.1, 2.1, 3.1
//
// Property: For any valid TableData, serializing to CSV and then parsing back
// should produce equivalent TableData (same headers, same number of rows, same values).
func TestProperty_CSVRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate random TableData
		td := generateRandomTableData(t)

		// Serialize to CSV
		var buf bytes.Buffer
		csvSerializer := serializer.NewCSVSerializer()
		err := csvSerializer.Serialize(td, &buf)
		if err != nil {
			t.Fatalf("failed to serialize TableData to CSV: %v", err)
		}

		// Parse back from CSV
		csvParser := NewCSVParser()
		parsedTD, err := csvParser.Parse(&buf)
		if err != nil {
			t.Fatalf("failed to parse CSV back to TableData: %v", err)
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

		// Property: All values should be preserved (as strings in CSV)
		// Note: CSV normalizes \r\n to \n, so we normalize before comparison
		for i, row := range td.Rows {
			parsedRow := parsedTD.Rows[i]
			if len(parsedRow) != len(row) {
				t.Fatalf("row %d column count mismatch: expected %d, got %d", 
					i, len(row), len(parsedRow))
			}

			for j, value := range row {
				parsedValue := parsedRow[j]
				// CSV normalizes \r\n to \n, but preserves standalone \r
				expected := strings.ReplaceAll(value.String(), "\r\n", "\n")
				if parsedValue.String() != expected {
					t.Fatalf("row %d, col %d value mismatch: expected %q, got %q", 
						i, j, expected, parsedValue.String())
				}
			}
		}

		// Property: Parsed data should be valid
		if err := parsedTD.Validate(); err != nil {
			t.Fatalf("parsed TableData failed validation: %v", err)
		}
	})
}

// generateRandomTableData creates a random TableData for testing
func generateRandomTableData(t *rapid.T) *model.TableData {
	// Generate random headers (1-20 columns)
	numCols := rapid.IntRange(1, 20).Draw(t, "numCols")
	headers := make([]string, numCols)
	for i := 0; i < numCols; i++ {
		// Use alphanumeric strings for headers to avoid CSV issues
		headers[i] = rapid.StringMatching(`[a-zA-Z0-9_]+`).Draw(t, "header")
	}

	// Generate random rows (0-100 rows)
	numRows := rapid.IntRange(0, 100).Draw(t, "numRows")
	rows := make([][]model.Value, numRows)
	for i := 0; i < numRows; i++ {
		row := make([]model.Value, numCols)
		for j := 0; j < numCols; j++ {
			row[j] = generateRandomValue(t)
		}
		rows[i] = row
	}

	return model.NewTableData(headers, rows)
}

// generateRandomValue creates a random Value for testing
func generateRandomValue(t *rapid.T) model.Value {
	valueType := rapid.IntRange(0, 3).Draw(t, "valueType")
	
	switch valueType {
	case 0: // String
		s := rapid.String().Draw(t, "stringValue")
		return model.NewStringValue(s)
	case 1: // Number
		n := rapid.Float64().Draw(t, "numberValue")
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

// Feature: table-converter, Property 2: Special Character Preservation (CSV)
// Validates: Requirements 3.2
//
// Property: For any TableData containing special characters (quotes, commas, newlines),
// converting through CSV format should preserve these characters correctly.
func TestProperty_CSVSpecialCharacters(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate TableData with special characters
		numCols := rapid.IntRange(1, 10).Draw(t, "numCols")
		headers := make([]string, numCols)
		for i := 0; i < numCols; i++ {
			headers[i] = rapid.StringMatching(`[a-zA-Z0-9_]+`).Draw(t, "header")
		}

		numRows := rapid.IntRange(1, 50).Draw(t, "numRows")
		rows := make([][]model.Value, numRows)
		for i := 0; i < numRows; i++ {
			row := make([]model.Value, numCols)
			for j := 0; j < numCols; j++ {
				// Generate values with special characters
				row[j] = generateValueWithSpecialChars(t)
			}
			rows[i] = row
		}

		td := model.NewTableData(headers, rows)

		// Serialize to CSV
		var buf bytes.Buffer
		csvSerializer := serializer.NewCSVSerializer()
		err := csvSerializer.Serialize(td, &buf)
		if err != nil {
			t.Fatalf("failed to serialize TableData to CSV: %v", err)
		}

		// Parse back from CSV
		csvParser := NewCSVParser()
		parsedTD, err := csvParser.Parse(&buf)
		if err != nil {
			t.Fatalf("failed to parse CSV back to TableData: %v", err)
		}

		// Property: All special characters should be preserved
		// Note: CSV normalizes \r\n to \n, but preserves standalone \r
		for i, row := range td.Rows {
			parsedRow := parsedTD.Rows[i]
			for j, value := range row {
				parsedValue := parsedRow[j]
				// Normalize \r\n to \n before comparison (standalone \r is preserved)
				expected := strings.ReplaceAll(value.String(), "\r\n", "\n")
				if parsedValue.String() != expected {
					t.Fatalf("row %d, col %d: special character not preserved\nexpected: %q\ngot: %q", 
						i, j, expected, parsedValue.String())
				}
			}
		}
	})
}

// generateValueWithSpecialChars creates a Value containing CSV special characters
func generateValueWithSpecialChars(t *rapid.T) model.Value {
	// Choose what type of special character to include
	charType := rapid.IntRange(0, 6).Draw(t, "charType")
	
	baseString := rapid.String().Draw(t, "baseString")
	
	var result string
	switch charType {
	case 0: // Quote
		result = baseString + `"` + rapid.String().Draw(t, "afterQuote")
	case 1: // Comma
		result = baseString + `,` + rapid.String().Draw(t, "afterComma")
	case 2: // Newline
		result = baseString + "\n" + rapid.String().Draw(t, "afterNewline")
	case 3: // Multiple quotes
		result = `"` + baseString + `"` + rapid.String().Draw(t, "afterQuotes")
	case 4: // Combination
		result = baseString + `,"` + rapid.String().Draw(t, "combo") + "\n"
	case 5: // Unicode
		result = baseString + "🎉" + rapid.String().Draw(t, "afterEmoji")
	case 6: // Regular string (control case)
		result = baseString
	}
	
	return model.NewStringValue(result)
}

