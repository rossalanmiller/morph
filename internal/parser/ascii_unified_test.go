package parser

import (
	"strings"
	"testing"
)

// TestUnifiedASCIIParser_AllFormats tests parsing of all supported ASCII table formats
func TestUnifiedASCIIParser_AllFormats(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedStyle TableStyle
		expectedRows  int
		expectedCols  int
	}{
		{
			name: "psql format",
			input: `Name    | Age | City
--------+-----+----------
Alice   | 30  | New York
Bob     | 25  | London`,
			expectedStyle: StylePsql,
			expectedRows:  2,
			expectedCols:  3,
		},
		{
			name: "markdown format",
			input: `| Name  | Age | City     |
|-------|-----|----------|
| Alice | 30  | New York |
| Bob   | 25  | London   |`,
			expectedStyle: StyleMarkdown,
			expectedRows:  2,
			expectedCols:  3,
		},
		{
			name: "org-mode format",
			input: `| Name  | Age | City     |
|-------+-----+----------|
| Alice | 30  | New York |
| Bob   | 25  | London   |`,
			expectedStyle: StyleOrgMode,
			expectedRows:  2,
			expectedCols:  3,
		},
		{
			name: "ASCII box format",
			input: `+-------+-----+----------+
| Name  | Age | City     |
+-------+-----+----------+
| Alice | 30  | New York |
| Bob   | 25  | London   |
+-------+-----+----------+`,
			expectedStyle: StyleBox,
			expectedRows:  2,
			expectedCols:  3,
		},
		{
			name: "RST grid format",
			input: `+-------+-----+----------+
| Name  | Age | City     |
+=======+=====+==========+
| Alice | 30  | New York |
| Bob   | 25  | London   |
+-------+-----+----------+`,
			expectedStyle: StyleRSTGrid,
			expectedRows:  2,
			expectedCols:  3,
		},
		{
			name: "RST simple format",
			input: `=====  ===  ========
Name   Age  City
=====  ===  ========
Alice  30   New York
Bob    25   London
=====  ===  ========`,
			expectedStyle: StyleRSTSimple,
			expectedRows:  2,
			expectedCols:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewUnifiedASCIIParser()
			td, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("failed to parse %s: %v", tt.name, err)
			}

			// Check detected style
			if parser.DetectedStyle != tt.expectedStyle {
				t.Errorf("expected style %s, got %s", tt.expectedStyle, parser.DetectedStyle)
			}

			// Check dimensions
			if len(td.Headers) != tt.expectedCols {
				t.Errorf("expected %d columns, got %d", tt.expectedCols, len(td.Headers))
			}
			if len(td.Rows) != tt.expectedRows {
				t.Errorf("expected %d rows, got %d", tt.expectedRows, len(td.Rows))
			}

			// Check first cell
			if len(td.Rows) > 0 && len(td.Rows[0]) > 0 {
				if td.Rows[0][0].Raw != "Alice" {
					t.Errorf("expected first cell to be 'Alice', got %q", td.Rows[0][0].Raw)
				}
			}
		})
	}
}

// TestUnifiedASCIIParser_WithRowSeparators tests parsing tables with row separators
func TestUnifiedASCIIParser_WithRowSeparators(t *testing.T) {
	input := `+-------+-----+
| Name  | Age |
+-------+-----+
| Alice | 30  |
+-------+-----+
| Bob   | 25  |
+-------+-----+`

	parser := NewUnifiedASCIIParser()
	td, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("failed to parse table with row separators: %v", err)
	}

	if len(td.Rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(td.Rows))
	}
}

// TestUnifiedASCIIParser_EmptyTable tests parsing an empty table
func TestUnifiedASCIIParser_EmptyTable(t *testing.T) {
	parser := NewUnifiedASCIIParser()
	td, err := parser.Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("failed to parse empty table: %v", err)
	}

	if len(td.Headers) != 0 {
		t.Errorf("expected 0 headers, got %d", len(td.Headers))
	}
	if len(td.Rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(td.Rows))
	}
}
