package parser

import (
	"bytes"
	"math"
	"testing"

	"github.com/user/table-converter/internal/model"
	"github.com/user/table-converter/internal/serializer"
	"pgregory.net/rapid"
)

// Feature: table-converter, Property 1: Round-Trip Preservation (Excel)
// Validates: Requirements 1.2, 2.2, 3.1
//
// Property: For any valid TableData, serializing to Excel and then parsing back
// should produce equivalent TableData (same headers, same number of rows, same values).
func TestProperty_ExcelRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate random TableData with Excel-safe values
		td := generateExcelSafeTableData(t)

		// Serialize to Excel
		var buf bytes.Buffer
		excelSerializer := serializer.NewExcelSerializer()
		err := excelSerializer.Serialize(td, &buf)
		if err != nil {
			t.Fatalf("failed to serialize TableData to Excel: %v", err)
		}

		// Parse back from Excel
		excelParser := NewExcelParser()
		parsedTD, err := excelParser.Parse(bytes.NewReader(buf.Bytes()))
		if err != nil {
			t.Fatalf("failed to parse Excel back to TableData: %v", err)
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

		// Property: All values should be preserved
		for i, row := range td.Rows {
			parsedRow := parsedTD.Rows[i]
			if len(parsedRow) != len(row) {
				t.Fatalf("row %d column count mismatch: expected %d, got %d",
					i, len(row), len(parsedRow))
			}

			for j, value := range row {
				parsedValue := parsedRow[j]
				if !valuesEquivalent(value, parsedValue) {
					t.Fatalf("row %d, col %d value mismatch: expected %v (%v), got %v (%v)",
						i, j, value.Raw, value.Type, parsedValue.Raw, parsedValue.Type)
				}
			}
		}

		// Property: Parsed data should be valid
		if err := parsedTD.Validate(); err != nil {
			t.Fatalf("parsed TableData failed validation: %v", err)
		}
	})
}


// generateExcelSafeTableData creates a random TableData with Excel-compatible values
func generateExcelSafeTableData(t *rapid.T) *model.TableData {
	// Generate random headers (1-10 columns, smaller for Excel)
	numCols := rapid.IntRange(1, 10).Draw(t, "numCols")
	headers := make([]string, numCols)
	usedHeaders := make(map[string]bool)
	for i := 0; i < numCols; i++ {
		// Generate unique headers to avoid Excel issues
		for {
			h := rapid.StringMatching(`[a-zA-Z][a-zA-Z0-9_]{0,20}`).Draw(t, "header")
			if !usedHeaders[h] {
				headers[i] = h
				usedHeaders[h] = true
				break
			}
		}
	}

	// Generate random rows (1-50 rows - at least 1 row required for Excel
	// to properly track sheet dimensions across all columns)
	numRows := rapid.IntRange(1, 50).Draw(t, "numRows")
	rows := make([][]model.Value, numRows)
	for i := 0; i < numRows; i++ {
		row := make([]model.Value, numCols)
		for j := 0; j < numCols; j++ {
			// Last column must have non-null, non-empty value to ensure
			// Excel tracks the full column range in sheet dimensions
			if j == numCols-1 {
				row[j] = generateExcelNonEmptyValue(t)
			} else {
				row[j] = generateExcelSafeValue(t)
			}
		}
		rows[i] = row
	}

	return model.NewTableData(headers, rows)
}

// generateExcelSafeValue creates a random Value that Excel can handle
func generateExcelSafeValue(t *rapid.T) model.Value {
	valueType := rapid.IntRange(0, 3).Draw(t, "valueType")

	switch valueType {
	case 0: // String - use printable ASCII to avoid encoding issues
		s := rapid.StringMatching(`[a-zA-Z0-9 .,!?-]{0,50}`).Draw(t, "stringValue")
		return model.NewStringValue(s)
	case 1: // Number - use reasonable range to avoid precision issues
		n := rapid.Float64Range(-1e10, 1e10).Draw(t, "numberValue")
		// Round to avoid floating point precision issues
		n = math.Round(n*1e6) / 1e6
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

// generateExcelNonNullValue creates a random non-null Value for Excel
func generateExcelNonNullValue(t *rapid.T) model.Value {
	valueType := rapid.IntRange(0, 2).Draw(t, "nonNullValueType")

	switch valueType {
	case 0: // String - use printable ASCII
		s := rapid.StringMatching(`[a-zA-Z0-9 .,!?-]{1,50}`).Draw(t, "stringValue")
		return model.NewStringValue(s)
	case 1: // Number
		n := rapid.Float64Range(-1e10, 1e10).Draw(t, "numberValue")
		n = math.Round(n*1e6) / 1e6
		return model.NewNumberValue(n)
	case 2: // Boolean
		b := rapid.Bool().Draw(t, "boolValue")
		return model.NewBooleanValue(b)
	default:
		return model.NewStringValue("x")
	}
}

// generateExcelNonEmptyValue creates a value that Excel will recognize as non-empty
// This ensures the column is included in sheet dimensions
func generateExcelNonEmptyValue(t *rapid.T) model.Value {
	valueType := rapid.IntRange(0, 2).Draw(t, "nonEmptyValueType")

	switch valueType {
	case 0: // Non-empty string
		s := rapid.StringMatching(`[a-zA-Z0-9.,!?-]{1,50}`).Draw(t, "nonEmptyString")
		return model.NewStringValue(s)
	case 1: // Number
		n := rapid.Float64Range(-1e10, 1e10).Draw(t, "numberValue")
		n = math.Round(n*1e6) / 1e6
		return model.NewNumberValue(n)
	case 2: // Boolean
		b := rapid.Bool().Draw(t, "boolValue")
		return model.NewBooleanValue(b)
	default:
		return model.NewStringValue("x")
	}
}

// valuesEquivalent checks if two values are semantically equivalent
func valuesEquivalent(a, b model.Value) bool {
	// Null values
	if a.Type == model.TypeNull && b.Type == model.TypeNull {
		return true
	}
	if a.Type == model.TypeNull || b.Type == model.TypeNull {
		// One is null, other is not - check if empty string
		if a.Type == model.TypeNull && b.Raw == "" {
			return true
		}
		if b.Type == model.TypeNull && a.Raw == "" {
			return true
		}
		return false
	}

	// Boolean values
	if a.Type == model.TypeBoolean && b.Type == model.TypeBoolean {
		aBool, aOk := a.Parsed.(bool)
		bBool, bOk := b.Parsed.(bool)
		if aOk && bOk {
			return aBool == bBool
		}
	}

	// Number values - compare with tolerance for floating point
	if a.Type == model.TypeNumber && b.Type == model.TypeNumber {
		aNum, aOk := a.Parsed.(float64)
		bNum, bOk := b.Parsed.(float64)
		if aOk && bOk {
			// Use relative tolerance for comparison
			if aNum == 0 && bNum == 0 {
				return true
			}
			diff := math.Abs(aNum - bNum)
			maxVal := math.Max(math.Abs(aNum), math.Abs(bNum))
			return diff <= maxVal*1e-9
		}
	}

	// String values - direct comparison
	if a.Type == model.TypeString && b.Type == model.TypeString {
		aStr, aOk := a.Parsed.(string)
		bStr, bOk := b.Parsed.(string)
		if aOk && bOk {
			return aStr == bStr
		}
	}

	// Fall back to raw string comparison
	return a.Raw == b.Raw
}
