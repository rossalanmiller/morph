package parser

import (
	"bytes"
	"math"
	"testing"

	"github.com/user/table-converter/internal/model"
	"github.com/user/table-converter/internal/serializer"
	"pgregory.net/rapid"
)

// Feature: table-converter, Property 1: Round-Trip Preservation (YAML)
// Validates: Requirements 1.3, 2.3, 3.1
//
// Property: For any valid TableData with at least one row, serializing to YAML and then parsing back
// should produce equivalent TableData (same headers, same number of rows, same values).
// Note: YAML format cannot preserve headers for empty tables since headers are derived from map keys.
func TestProperty_YAMLRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate random TableData with YAML-safe values
		// Must have at least 1 row since YAML derives headers from map keys
		td := generateYAMLSafeTableDataWithRows(t)

		// Serialize to YAML
		var buf bytes.Buffer
		yamlSerializer := serializer.NewYAMLSerializer()
		err := yamlSerializer.Serialize(td, &buf)
		if err != nil {
			t.Fatalf("failed to serialize TableData to YAML: %v", err)
		}

		// Parse back from YAML
		yamlParser := NewYAMLParser()
		parsedTD, err := yamlParser.Parse(&buf)
		if err != nil {
			t.Fatalf("failed to parse YAML back to TableData: %v", err)
		}

		// Property: Header count should be identical
		if len(parsedTD.Headers) != len(td.Headers) {
			t.Fatalf("header count mismatch: expected %d, got %d",
				len(td.Headers), len(parsedTD.Headers))
		}

		// Property: Headers should be identical (YAML sorts headers alphabetically)
		// Create a map for header lookup since YAML parser sorts headers
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

		// Property: All values should be preserved (with type preservation)
		for i, row := range td.Rows {
			parsedRow := parsedTD.Rows[i]

			for j, value := range row {
				// Find the corresponding column in parsed data
				header := td.Headers[j]
				parsedColIdx := parsedHeaderMap[header]
				parsedValue := parsedRow[parsedColIdx]

				// Compare values based on type
				if !yamlValuesEqual(value, parsedValue) {
					t.Fatalf("row %d, col %q value mismatch:\nexpected type=%d, raw=%q, parsed=%v\ngot type=%d, raw=%q, parsed=%v",
						i, header, value.Type, value.Raw, value.Parsed,
						parsedValue.Type, parsedValue.Raw, parsedValue.Parsed)
				}
			}
		}

		// Property: Parsed data should be valid
		if err := parsedTD.Validate(); err != nil {
			t.Fatalf("parsed TableData failed validation: %v", err)
		}
	})
}

// generateYAMLSafeTableDataWithRows creates a random TableData with at least 1 row
// This is needed for YAML round-trip testing since headers are derived from map keys
func generateYAMLSafeTableDataWithRows(t *rapid.T) *model.TableData {
	// Generate random headers (1-20 columns)
	// Use valid YAML map keys (alphanumeric starting with letter)
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
			row[j] = generateYAMLSafeValue(t)
		}
		rows[i] = row
	}

	return model.NewTableData(headers, rows)
}

// generateYAMLSafeValue creates a random Value that is safe for YAML
// Note: YAML has limitations with whitespace-only strings - they get normalized/stripped
// So we generate strings that either have non-whitespace content or are empty
func generateYAMLSafeValue(t *rapid.T) model.Value {
	valueType := rapid.IntRange(0, 3).Draw(t, "valueType")

	switch valueType {
	case 0: // String - avoid whitespace-only strings as YAML normalizes them
		s := generateYAMLSafeString(t)
		return model.NewStringValue(s)
	case 1: // Number (finite values only for YAML)
		n := rapid.Float64Range(-1e15, 1e15).Draw(t, "numberValue")
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

// generateYAMLSafeString generates a string that YAML can round-trip correctly
// With double-quoted style for strings containing newlines, all strings should work
func generateYAMLSafeString(t *rapid.T) string {
	// Generate any string - the serializer now handles newlines correctly
	return rapid.String().Draw(t, "stringValue")
}

// yamlValuesEqual compares two model.Value instances for equality
func yamlValuesEqual(a, b model.Value) bool {
	// Null values
	if a.Type == model.TypeNull && b.Type == model.TypeNull {
		return true
	}

	// For YAML, types should match
	if a.Type != b.Type {
		return false
	}

	switch a.Type {
	case model.TypeString:
		aStr, ok1 := a.Parsed.(string)
		bStr, ok2 := b.Parsed.(string)
		return ok1 && ok2 && aStr == bStr
	case model.TypeNumber:
		aNum, ok1 := a.Parsed.(float64)
		bNum, ok2 := b.Parsed.(float64)
		return ok1 && ok2 && yamlFloatsEqual(aNum, bNum)
	case model.TypeBoolean:
		aBool, ok1 := a.Parsed.(bool)
		bBool, ok2 := b.Parsed.(bool)
		return ok1 && ok2 && aBool == bBool
	default:
		return a.Raw == b.Raw
	}
}

// yamlFloatsEqual compares two float64 values with tolerance for floating point errors
func yamlFloatsEqual(a, b float64) bool {
	// Handle special cases
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	if math.IsInf(a, 1) && math.IsInf(b, 1) {
		return true
	}
	if math.IsInf(a, -1) && math.IsInf(b, -1) {
		return true
	}

	// For very small numbers, use absolute comparison
	if math.Abs(a) < 1e-10 && math.Abs(b) < 1e-10 {
		return math.Abs(a-b) < 1e-15
	}

	// For larger numbers, use relative comparison
	diff := math.Abs(a - b)
	avg := (math.Abs(a) + math.Abs(b)) / 2
	return diff/avg < 1e-10
}
