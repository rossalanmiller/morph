package parser

import (
	"strings"
	"testing"

	"github.com/user/table-converter/internal/model"
)

// TestASCIIParser_PsqlFormat tests parsing of PostgreSQL psql aligned format
// This format uses | for column separators in data rows and + for separator lines
func TestASCIIParser_PsqlFormat(t *testing.T) {
	input := `cspProductName |  dbProduct  | financeCode
---------------+-------------+-------------
ae             | ae          | au
camsi          | amsi        | am
dbi            | bi          | bs`

	parser := NewASCIIParser()
	td, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("failed to parse psql format: %v", err)
	}

	// Verify headers
	expectedHeaders := []string{"cspProductName", "dbProduct", "financeCode"}
	if len(td.Headers) != len(expectedHeaders) {
		t.Fatalf("expected %d headers, got %d", len(expectedHeaders), len(td.Headers))
	}
	for i, expected := range expectedHeaders {
		if td.Headers[i] != expected {
			t.Errorf("header %d: expected %q, got %q", i, expected, td.Headers[i])
		}
	}

	// Verify row count
	if len(td.Rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(td.Rows))
	}

	// Verify first row
	expectedRow1 := []string{"ae", "ae", "au"}
	for i, expected := range expectedRow1 {
		if td.Rows[0][i].Raw != expected {
			t.Errorf("row 0, col %d: expected %q, got %q", i, expected, td.Rows[0][i].Raw)
		}
	}

	// Verify second row
	expectedRow2 := []string{"camsi", "amsi", "am"}
	for i, expected := range expectedRow2 {
		if td.Rows[1][i].Raw != expected {
			t.Errorf("row 1, col %d: expected %q, got %q", i, expected, td.Rows[1][i].Raw)
		}
	}
}

// TestASCIIParser_TraditionalBoxFormat tests parsing of traditional box-drawing format
// This format uses | for all column separators and + for corners
func TestASCIIParser_TraditionalBoxFormat(t *testing.T) {
	input := `+-------+-----+--------+
| name  | age | active |
+-------+-----+--------+
| Alice | 30  | true   |
| Bob   | 25  | false  |
+-------+-----+--------+`

	parser := NewASCIIParser()
	td, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("failed to parse box format: %v", err)
	}

	// Verify headers
	expectedHeaders := []string{"name", "age", "active"}
	if len(td.Headers) != len(expectedHeaders) {
		t.Fatalf("expected %d headers, got %d", len(expectedHeaders), len(td.Headers))
	}
	for i, expected := range expectedHeaders {
		if td.Headers[i] != expected {
			t.Errorf("header %d: expected %q, got %q", i, expected, td.Headers[i])
		}
	}

	// Verify row count
	if len(td.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(td.Rows))
	}

	// Verify first row
	if td.Rows[0][0].Raw != "Alice" {
		t.Errorf("expected Alice, got %q", td.Rows[0][0].Raw)
	}
	if td.Rows[0][1].Type != model.TypeNumber {
		t.Errorf("expected number type for age, got %v", td.Rows[0][1].Type)
	}
	if td.Rows[0][2].Type != model.TypeBoolean {
		t.Errorf("expected boolean type for active, got %v", td.Rows[0][2].Type)
	}
}

// TestASCIIParser_EmptyTable tests parsing of an empty table
func TestASCIIParser_EmptyTable(t *testing.T) {
	input := ``

	parser := NewASCIIParser()
	td, err := parser.Parse(strings.NewReader(input))
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
