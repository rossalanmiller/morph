package model

import (
	"testing"

	"pgregory.net/rapid"
)

// Feature: table-converter, Property 7: Row Normalization
// Validates: Requirements 7.1
//
// Property: For any input data with inconsistent column counts across rows,
// the parser should normalize all rows to have the same number of columns
// (padding with null values as needed).
func TestProperty_RowNormalization(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate random headers (1-20 columns)
		numCols := rapid.IntRange(1, 20).Draw(t, "numCols")
		headers := make([]string, numCols)
		for i := 0; i < numCols; i++ {
			headers[i] = rapid.StringN(1, 20, -1).Draw(t, "header")
		}

		// Generate random rows with varying column counts (0-100 rows)
		numRows := rapid.IntRange(0, 100).Draw(t, "numRows")
		rows := make([][]Value, numRows)
		for i := 0; i < numRows; i++ {
			// Each row can have 0 to 30 columns (intentionally inconsistent)
			rowLen := rapid.IntRange(0, 30).Draw(t, "rowLen")
			row := make([]Value, rowLen)
			for j := 0; j < rowLen; j++ {
				// Generate random values
				row[j] = generateRandomValue(t)
			}
			rows[i] = row
		}

		// Create TableData - this should normalize all rows
		td := NewTableData(headers, rows)

		// Property: All rows must have exactly numCols columns
		if len(td.Headers) != numCols {
			t.Fatalf("expected %d headers, got %d", numCols, len(td.Headers))
		}

		for i, row := range td.Rows {
			if len(row) != numCols {
				t.Fatalf("row %d has %d columns, expected %d (normalized to match headers)", 
					i, len(row), numCols)
			}
		}

		// Property: Rows shorter than numCols should be padded with null values
		for i := 0; i < numRows; i++ {
			originalRowLen := len(rows[i])
			normalizedRow := td.Rows[i]
			
			// Check that original values are preserved
			for j := 0; j < originalRowLen && j < numCols; j++ {
				if normalizedRow[j].Raw != rows[i][j].Raw {
					t.Fatalf("row %d, col %d: value changed during normalization", i, j)
				}
			}

			// Check that padding is null values
			if originalRowLen < numCols {
				for j := originalRowLen; j < numCols; j++ {
					if normalizedRow[j].Type != TypeNull {
						t.Fatalf("row %d, col %d: expected null padding, got type %v", 
							i, j, normalizedRow[j].Type)
					}
				}
			}
		}

		// Property: TableData should pass validation
		if err := td.Validate(); err != nil {
			t.Fatalf("normalized TableData failed validation: %v", err)
		}
	})
}

// generateRandomValue creates a random Value for testing
func generateRandomValue(t *rapid.T) Value {
	valueType := rapid.IntRange(0, 3).Draw(t, "valueType")
	
	switch valueType {
	case 0: // String
		s := rapid.String().Draw(t, "stringValue")
		return NewStringValue(s)
	case 1: // Number
		n := rapid.Float64().Draw(t, "numberValue")
		return NewNumberValue(n)
	case 2: // Boolean
		b := rapid.Bool().Draw(t, "boolValue")
		return NewBooleanValue(b)
	case 3: // Null
		return NewNullValue()
	default:
		return NewStringValue("")
	}
}
