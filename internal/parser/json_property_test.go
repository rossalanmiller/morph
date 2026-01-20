package parser

import (
	"bytes"
	"math"
	"testing"

	"github.com/user/table-converter/internal/model"
	"github.com/user/table-converter/internal/serializer"
	"pgregory.net/rapid"
)

// Feature: table-converter, Property 1: Round-Trip Preservation (JSON)
// Validates: Requirements 1.4, 2.4, 3.1
//
// Property: For any valid TableData with at least one row, serializing to JSON and then parsing back
// should produce equivalent TableData (same headers, same number of rows, same values).
// Note: JSON format cannot preserve headers for empty tables since headers are derived from object keys.
func TestProperty_JSONRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate random TableData with JSON-safe values
		// Must have at least 1 row since JSON derives headers from object keys
		td := generateJSONSafeTableDataWithRows(t)

		// Serialize to JSON
		var buf bytes.Buffer
		jsonSerializer := serializer.NewJSONSerializer()
		err := jsonSerializer.Serialize(td, &buf)
		if err != nil {
			t.Fatalf("failed to serialize TableData to JSON: %v", err)
		}

		// Parse back from JSON
		jsonParser := NewJSONParser()
		parsedTD, err := jsonParser.Parse(&buf)
		if err != nil {
			t.Fatalf("failed to parse JSON back to TableData: %v", err)
		}

		// Property: Header count should be identical
		if len(parsedTD.Headers) != len(td.Headers) {
			t.Fatalf("header count mismatch: expected %d, got %d",
				len(td.Headers), len(parsedTD.Headers))
		}

		// Property: Headers should be identical (JSON sorts headers alphabetically)
		// Create a map for header lookup since JSON parser sorts headers
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
				if !valuesEqual(value, parsedValue) {
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

// Feature: table-converter, Property 4: Numeric Precision Preservation
// Validates: Requirements 3.4
//
// Property: For any TableData containing numeric values (integers, floats),
// converting through JSON format should preserve the numeric value and precision.
func TestProperty_JSONNumericPrecision(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate TableData with numeric values
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
				row[j] = generateNumericValue(t)
			}
			rows[i] = row
		}

		td := model.NewTableData(headers, rows)

		// Serialize to JSON
		var buf bytes.Buffer
		jsonSerializer := serializer.NewJSONSerializer()
		err := jsonSerializer.Serialize(td, &buf)
		if err != nil {
			t.Fatalf("failed to serialize TableData to JSON: %v", err)
		}

		// Parse back from JSON
		jsonParser := NewJSONParser()
		parsedTD, err := jsonParser.Parse(&buf)
		if err != nil {
			t.Fatalf("failed to parse JSON back to TableData: %v", err)
		}

		// Create header index map for parsed data
		parsedHeaderMap := make(map[string]int)
		for i, h := range parsedTD.Headers {
			parsedHeaderMap[h] = i
		}

		// Property: Numeric values should be preserved with precision
		for i, row := range td.Rows {
			parsedRow := parsedTD.Rows[i]
			for j, value := range row {
				header := td.Headers[j]
				parsedColIdx := parsedHeaderMap[header]
				parsedValue := parsedRow[parsedColIdx]

				if value.Type == model.TypeNumber {
					// Both should be numbers
					if parsedValue.Type != model.TypeNumber {
						t.Fatalf("row %d, col %q: expected number type, got %d",
							i, header, parsedValue.Type)
					}

					// Compare numeric values
					origNum, ok1 := value.Parsed.(float64)
					parsedNum, ok2 := parsedValue.Parsed.(float64)
					if !ok1 || !ok2 {
						t.Fatalf("row %d, col %q: failed to get numeric values", i, header)
					}

					// Allow small floating point differences
					if !floatsEqual(origNum, parsedNum) {
						t.Fatalf("row %d, col %q: numeric precision lost\nexpected: %v\ngot: %v",
							i, header, origNum, parsedNum)
					}
				}
			}
		}
	})
}

// generateJSONSafeTableData creates a random TableData with JSON-safe values
func generateJSONSafeTableData(t *rapid.T) *model.TableData {
	// Generate random headers (1-20 columns)
	// Use valid JSON object keys (alphanumeric starting with letter)
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

	// Generate random rows (0-100 rows)
	numRows := rapid.IntRange(0, 100).Draw(t, "numRows")
	rows := make([][]model.Value, numRows)
	for i := 0; i < numRows; i++ {
		row := make([]model.Value, numCols)
		for j := 0; j < numCols; j++ {
			row[j] = generateJSONSafeValue(t)
		}
		rows[i] = row
	}

	return model.NewTableData(headers, rows)
}

// generateJSONSafeTableDataWithRows creates a random TableData with at least 1 row
// This is needed for JSON round-trip testing since headers are derived from object keys
func generateJSONSafeTableDataWithRows(t *rapid.T) *model.TableData {
	// Generate random headers (1-20 columns)
	// Use valid JSON object keys (alphanumeric starting with letter)
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
			row[j] = generateJSONSafeValue(t)
		}
		rows[i] = row
	}

	return model.NewTableData(headers, rows)
}

// generateJSONSafeValue creates a random Value that is safe for JSON
func generateJSONSafeValue(t *rapid.T) model.Value {
	valueType := rapid.IntRange(0, 3).Draw(t, "valueType")

	switch valueType {
	case 0: // String
		s := rapid.String().Draw(t, "stringValue")
		return model.NewStringValue(s)
	case 1: // Number (finite values only for JSON)
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

// generateNumericValue creates a random numeric Value
func generateNumericValue(t *rapid.T) model.Value {
	numType := rapid.IntRange(0, 2).Draw(t, "numType")

	switch numType {
	case 0: // Integer-like float
		n := float64(rapid.IntRange(-1000000, 1000000).Draw(t, "intValue"))
		return model.NewNumberValue(n)
	case 1: // Float with decimals
		n := rapid.Float64Range(-1e10, 1e10).Draw(t, "floatValue")
		return model.NewNumberValue(n)
	case 2: // Small precise float
		n := rapid.Float64Range(-1000, 1000).Draw(t, "smallFloat")
		return model.NewNumberValue(n)
	default:
		return model.NewNumberValue(0)
	}
}

// valuesEqual compares two model.Value instances for equality
func valuesEqual(a, b model.Value) bool {
	// Null values
	if a.Type == model.TypeNull && b.Type == model.TypeNull {
		return true
	}

	// For JSON, types should match
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
		return ok1 && ok2 && floatsEqual(aNum, bNum)
	case model.TypeBoolean:
		aBool, ok1 := a.Parsed.(bool)
		bBool, ok2 := b.Parsed.(bool)
		return ok1 && ok2 && aBool == bBool
	default:
		return a.Raw == b.Raw
	}
}

// floatsEqual compares two float64 values with tolerance for floating point errors
func floatsEqual(a, b float64) bool {
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
