package parser

import (
	"bytes"
	"strings"
	"testing"

	"github.com/user/table-converter/internal/model"
	"github.com/user/table-converter/internal/serializer"
)

func TestCSVEmptyStringRoundTrip(t *testing.T) {
	// Create TableData with one header and one row with empty string
	headers := []string{"col1"}
	rows := [][]model.Value{
		{model.NewStringValue("")},
	}
	td := model.NewTableData(headers, rows)

	// Serialize to CSV
	var buf bytes.Buffer
	csvSerializer := serializer.NewCSVSerializer()
	err := csvSerializer.Serialize(td, &buf)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	csvOutput := buf.String()
	t.Logf("CSV output:\n%q", csvOutput)

	// Parse back
	csvParser := NewCSVParser()
	parsedTD, err := csvParser.Parse(strings.NewReader(csvOutput))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	t.Logf("Original rows: %d", len(td.Rows))
	t.Logf("Parsed rows: %d", len(parsedTD.Rows))

	if len(parsedTD.Rows) != len(td.Rows) {
		t.Errorf("row count mismatch: expected %d, got %d", len(td.Rows), len(parsedTD.Rows))
	}
}

func TestCSVMultipleRowsWithEmpty(t *testing.T) {
	// Create TableData with one header and two rows: one with "1", one with ""
	headers := []string{"col1"}
	rows := [][]model.Value{
		{model.NewNumberValue(1)},
		{model.NewStringValue("")},
	}
	td := model.NewTableData(headers, rows)

	// Serialize to CSV
	var buf bytes.Buffer
	csvSerializer := serializer.NewCSVSerializer()
	err := csvSerializer.Serialize(td, &buf)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	csvOutput := buf.String()
	t.Logf("CSV output:\n%q", csvOutput)

	// Parse back
	csvParser := NewCSVParser()
	parsedTD, err := csvParser.Parse(strings.NewReader(csvOutput))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	t.Logf("Original rows: %d", len(td.Rows))
	t.Logf("Parsed rows: %d", len(parsedTD.Rows))

	if len(parsedTD.Rows) != len(td.Rows) {
		t.Errorf("row count mismatch: expected %d, got %d", len(td.Rows), len(parsedTD.Rows))
	}

	// Check values
	for i, row := range td.Rows {
		parsedRow := parsedTD.Rows[i]
		for j, value := range row {
			parsedValue := parsedRow[j]
			if parsedValue.String() != value.String() {
				t.Errorf("row %d, col %d: expected %q, got %q", i, j, value.String(), parsedValue.String())
			}
		}
	}
}
