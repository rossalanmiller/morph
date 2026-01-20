package parser

import (
	"bytes"
	"testing"

	"github.com/user/table-converter/internal/model"
	"github.com/user/table-converter/internal/serializer"
	"pgregory.net/rapid"
)

// Feature: table-converter, Property 3: Empty Cell Preservation
// Validates: Requirements 3.3
//
// Property: For any TableData containing empty/null cells, converting through
// any format should preserve the empty cells (they should remain empty).
func TestProperty_EmptyCellPreservation(t *testing.T) {
	// Test each format that supports empty cells
	formats := []struct {
		name          string
		parser        Parser
		serializer    serializer.Serializer
		preserveOrder bool // Whether format preserves column order
	}{
		{"CSV", NewCSVParser(), serializer.NewCSVSerializer(), true},
		{"JSON", NewJSONParser(), serializer.NewJSONSerializer(), false},
		{"YAML", NewYAMLParser(), serializer.NewYAMLSerializer(), false},
		{"HTML", NewHTMLParser(), serializer.NewHTMLSerializer(), true},
		{"XML", NewXMLParser(), serializer.NewXMLSerializer(), false},
		// Note: Excel, Markdown, ASCII have whitespace trimming behavior
	}

	for _, fmt := range formats {
		t.Run(fmt.name, func(t *testing.T) {
			rapid.Check(t, func(t *rapid.T) {
				// Generate TableData with empty cells
				td := generateTableWithEmptyCells(t)

				// Serialize
				var buf bytes.Buffer
				err := fmt.serializer.Serialize(td, &buf)
				if err != nil {
					t.Fatalf("failed to serialize: %v", err)
				}

				// Parse back
				parsedTD, err := fmt.parser.Parse(&buf)
				if err != nil {
					t.Fatalf("failed to parse: %v", err)
				}

				// Build header index map for parsed data
				parsedHeaderIdx := make(map[string]int)
				for i, h := range parsedTD.Headers {
					parsedHeaderIdx[h] = i
				}

				// Verify empty cells are preserved
				for i, row := range td.Rows {
					if i >= len(parsedTD.Rows) {
						t.Fatalf("row %d missing", i)
					}
					parsedRow := parsedTD.Rows[i]

					for j, value := range row {
						header := td.Headers[j]
						parsedIdx, ok := parsedHeaderIdx[header]
						if !ok {
							t.Fatalf("header %q not found in parsed data", header)
						}
						if parsedIdx >= len(parsedRow) {
							t.Fatalf("row %d, col %d missing", i, parsedIdx)
						}
						parsedValue := parsedRow[parsedIdx]

						// Check if original was empty/null
						if value.Type == model.TypeNull || value.Raw == "" {
							// Parsed should also be empty/null
							if parsedValue.Type != model.TypeNull && parsedValue.Raw != "" {
								t.Fatalf("row %d, header %q: empty cell not preserved, got %q",
									i, header, parsedValue.Raw)
							}
						}
					}
				}
			})
		})
	}
}

// generateTableWithEmptyCells creates a TableData with some empty/null cells
func generateTableWithEmptyCells(t *rapid.T) *model.TableData {
	numCols := rapid.IntRange(1, 5).Draw(t, "numCols")
	headers := make([]string, numCols)
	usedHeaders := make(map[string]bool)
	for i := 0; i < numCols; i++ {
		// Generate unique headers to avoid map-based format issues
		for {
			h := rapid.StringMatching(`[a-zA-Z][a-zA-Z0-9]{0,10}`).Draw(t, "header")
			if !usedHeaders[h] {
				headers[i] = h
				usedHeaders[h] = true
				break
			}
		}
	}

	numRows := rapid.IntRange(1, 20).Draw(t, "numRows")
	rows := make([][]model.Value, numRows)
	for i := 0; i < numRows; i++ {
		row := make([]model.Value, numCols)
		for j := 0; j < numCols; j++ {
			// 30% chance of empty cell
			if rapid.IntRange(0, 9).Draw(t, "emptyChance") < 3 {
				row[j] = model.NewNullValue()
			} else {
				// Non-empty value (simple alphanumeric)
				s := rapid.StringMatching(`[a-zA-Z0-9]{1,10}`).Draw(t, "value")
				row[j] = model.NewStringValue(s)
			}
		}
		rows[i] = row
	}

	return model.NewTableData(headers, rows)
}
