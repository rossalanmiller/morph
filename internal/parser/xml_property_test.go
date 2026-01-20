package parser

import (
	"bytes"
	"testing"

	"github.com/user/table-converter/internal/model"
	"github.com/user/table-converter/internal/serializer"
	"pgregory.net/rapid"
)

// Feature: table-converter, Property 1: Round-Trip Preservation (XML)
// Validates: Requirements 1.6, 2.6, 3.1
//
// Property: For any valid TableData with at least one row, serializing to XML and then parsing back
// should produce equivalent TableData (same headers, same number of rows, same values).
// Note: XML format cannot preserve headers for empty tables since headers are derived from element names.
// Note: XML treats all values as strings, so type information is lost.
func TestProperty_XMLRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate random TableData with XML-safe values
		// Must have at least 1 row since XML derives headers from element names
		td := generateXMLSafeTableDataWithRows(t)

		// Serialize to XML
		var buf bytes.Buffer
		xmlSerializer := serializer.NewXMLSerializer()
		err := xmlSerializer.Serialize(td, &buf)
		if err != nil {
			t.Fatalf("failed to serialize TableData to XML: %v", err)
		}

		// Parse back from XML
		xmlParser := NewXMLParser()
		parsedTD, err := xmlParser.Parse(&buf)
		if err != nil {
			t.Fatalf("failed to parse XML back to TableData: %v", err)
		}

		// Property: Header count should be identical
		if len(parsedTD.Headers) != len(td.Headers) {
			t.Fatalf("header count mismatch: expected %d, got %d",
				len(td.Headers), len(parsedTD.Headers))
		}

		// Property: Headers should be identical (XML sorts headers alphabetically)
		// Create a map for header lookup since XML parser sorts headers
		headerMap := make(map[string]int)
		for i, h := range td.Headers {
			headerMap[h] = i
		}
		parsedHeaderMap := make(map[string]int)
		for i, h := range parsedTD.Headers {
			parsedHeaderMap[h] = i
		}

		// Check all original headers exist in parsed
		for header := range headerMap {
			if _, exists := parsedHeaderMap[header]; !exists {
				t.Fatalf("header %q missing in parsed data", header)
			}
		}

		// Property: Row count should be identical
		if len(parsedTD.Rows) != len(td.Rows) {
			t.Fatalf("row count mismatch: expected %d, got %d",
				len(td.Rows), len(parsedTD.Rows))
		}

		// Property: All values should be preserved (as strings in XML)
		for i, row := range td.Rows {
			parsedRow := parsedTD.Rows[i]

			for j, value := range row {
				// Find the corresponding column in parsed data
				header := td.Headers[j]
				parsedColIdx := parsedHeaderMap[header]
				parsedValue := parsedRow[parsedColIdx]

				// XML treats all values as strings, compare string representations
				expected := xmlValueToString(value)
				actual := parsedValue.String()
				if actual != expected {
					t.Fatalf("row %d, col %q value mismatch:\nexpected: %q\ngot: %q",
						i, header, expected, actual)
				}
			}
		}

		// Property: Parsed data should be valid
		if err := parsedTD.Validate(); err != nil {
			t.Fatalf("parsed TableData failed validation: %v", err)
		}
	})
}

// generateXMLSafeTableDataWithRows creates a random TableData with at least 1 row
// This is needed for XML round-trip testing since headers are derived from element names
func generateXMLSafeTableDataWithRows(t *rapid.T) *model.TableData {
	// Generate random headers (1-20 columns)
	// Use valid XML element names (alphanumeric starting with letter)
	numCols := rapid.IntRange(1, 20).Draw(t, "numCols")
	headers := make([]string, numCols)
	usedHeaders := make(map[string]bool)
	for i := 0; i < numCols; i++ {
		for {
			h := rapid.StringMatching(`[a-zA-Z][a-zA-Z0-9_]*`).Draw(t, "header")
			if !usedHeaders[h] {
				headers[i] = h
				usedHeaders[h] = true
				break
			}
		}
	}

	// Generate random rows (1-100 rows) - at least 1 row required
	numRows := rapid.IntRange(1, 100).Draw(t, "numRows")
	rows := make([][]model.Value, numRows)
	for i := 0; i < numRows; i++ {
		row := make([]model.Value, numCols)
		for j := 0; j < numCols; j++ {
			row[j] = generateXMLSafeValue(t)
		}
		rows[i] = row
	}

	return model.NewTableData(headers, rows)
}

// generateXMLSafeValue creates a random Value that is safe for XML round-trip
func generateXMLSafeValue(t *rapid.T) model.Value {
	valueType := rapid.IntRange(0, 3).Draw(t, "valueType")

	switch valueType {
	case 0: // String - avoid whitespace-only strings and XML special chars for basic round-trip
		s := rapid.StringMatching(`[a-zA-Z0-9][a-zA-Z0-9 ]*[a-zA-Z0-9]|[a-zA-Z0-9]`).Draw(t, "stringValue")
		return model.NewStringValue(s)
	case 1: // Number
		n := rapid.Float64Range(-1e10, 1e10).Draw(t, "numberValue")
		return model.NewNumberValue(n)
	case 2: // Boolean
		b := rapid.Bool().Draw(t, "boolValue")
		return model.NewBooleanValue(b)
	case 3: // Null (becomes empty string in XML)
		return model.NewNullValue()
	default:
		return model.NewStringValue("default")
	}
}

// xmlValueToString converts a model.Value to its XML string representation
func xmlValueToString(val model.Value) string {
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


// Feature: table-converter, Property 8: Character Escaping (XML)
// Validates: Requirements 7.2
//
// Property: For any TableData containing characters that are special in XML (<, >, &, ", '),
// the serializer should properly escape them so the output is valid XML,
// and the parser should decode them back to the original characters.
func TestProperty_XMLCharacterEscaping(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate TableData with XML special characters
		numCols := rapid.IntRange(1, 10).Draw(t, "numCols")
		headers := make([]string, numCols)
		usedHeaders := make(map[string]bool)
		for i := 0; i < numCols; i++ {
			for {
				h := rapid.StringMatching(`[a-zA-Z][a-zA-Z0-9_]*`).Draw(t, "header")
				if !usedHeaders[h] {
					headers[i] = h
					usedHeaders[h] = true
					break
				}
			}
		}

		numRows := rapid.IntRange(1, 50).Draw(t, "numRows")
		rows := make([][]model.Value, numRows)
		for i := 0; i < numRows; i++ {
			row := make([]model.Value, numCols)
			for j := 0; j < numCols; j++ {
				// Generate values with XML special characters
				row[j] = generateValueWithXMLSpecialChars(t)
			}
			rows[i] = row
		}

		td := model.NewTableData(headers, rows)

		// Serialize to XML
		var buf bytes.Buffer
		xmlSerializer := serializer.NewXMLSerializer()
		err := xmlSerializer.Serialize(td, &buf)
		if err != nil {
			t.Fatalf("failed to serialize TableData to XML: %v", err)
		}

		xmlOutput := buf.String()

		// Property: The XML output should be valid (parseable)
		// Parse back from XML
		xmlParser := NewXMLParser()
		parsedTD, err := xmlParser.Parse(bytes.NewBufferString(xmlOutput))
		if err != nil {
			t.Fatalf("failed to parse XML back to TableData: %v", err)
		}

		// Create header index map for parsed data
		parsedHeaderMap := make(map[string]int)
		for i, h := range parsedTD.Headers {
			parsedHeaderMap[h] = i
		}

		// Property: All special characters should be preserved after round-trip
		for i, row := range td.Rows {
			parsedRow := parsedTD.Rows[i]
			for j, value := range row {
				header := td.Headers[j]
				parsedColIdx := parsedHeaderMap[header]
				parsedValue := parsedRow[parsedColIdx]
				expected := xmlValueToString(value)
				actual := parsedValue.String()
				if actual != expected {
					t.Fatalf("row %d, col %q: special character not preserved\nexpected: %q\ngot: %q",
						i, header, expected, actual)
				}
			}
		}
	})
}

// generateValueWithXMLSpecialChars creates a Value containing XML special characters
func generateValueWithXMLSpecialChars(t *rapid.T) model.Value {
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
