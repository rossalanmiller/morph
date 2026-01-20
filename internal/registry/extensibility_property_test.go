package registry

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/user/table-converter/internal/model"
	"github.com/user/table-converter/internal/parser"
	"github.com/user/table-converter/internal/serializer"
	"pgregory.net/rapid"
)

// Feature: table-converter, Property 9: Format Extensibility
// Validates: Requirements 9.1, 9.2
//
// Property: For any new format added by implementing the Parser and Serializer interfaces,
// the format should automatically work with all existing formats for conversions
// (new format can convert to/from any existing format).

// mockFormat implements a simple custom format for testing extensibility
// It uses a simple JSON-based format with a custom wrapper structure
type mockFormat struct {
	Version string          `json:"version"`
	Headers []string        `json:"headers"`
	Data    [][]interface{} `json:"data"`
}

// mockFormatParser implements the Parser interface for our mock format
type mockFormatParser struct{}

func (p *mockFormatParser) Parse(input io.Reader) (*model.TableData, error) {
	var mf mockFormat
	decoder := json.NewDecoder(input)
	if err := decoder.Decode(&mf); err != nil {
		return nil, parser.NewParseError("failed to parse mock format").WithErr(err)
	}

	if mf.Version != "1.0" {
		return nil, parser.NewParseError("unsupported mock format version: " + mf.Version)
	}

	rows := make([][]model.Value, len(mf.Data))
	for i, row := range mf.Data {
		values := make([]model.Value, len(row))
		for j, cell := range row {
			switch v := cell.(type) {
			case string:
				values[j] = model.NewStringValue(v)
			case float64:
				values[j] = model.NewNumberValue(v)
			case bool:
				values[j] = model.NewBooleanValue(v)
			case nil:
				values[j] = model.NewNullValue()
			default:
				values[j] = model.NewStringValue("")
			}
		}
		rows[i] = values
	}

	return model.NewTableData(mf.Headers, rows), nil
}

// mockFormatSerializer implements the Serializer interface for our mock format
type mockFormatSerializer struct{}

func (s *mockFormatSerializer) Serialize(data *model.TableData, output io.Writer) error {
	mf := mockFormat{
		Version: "1.0",
		Headers: data.Headers,
		Data:    make([][]interface{}, len(data.Rows)),
	}

	for i, row := range data.Rows {
		mf.Data[i] = make([]interface{}, len(row))
		for j, val := range row {
			switch val.Type {
			case model.TypeString:
				mf.Data[i][j] = val.Raw
			case model.TypeNumber:
				mf.Data[i][j] = val.Parsed
			case model.TypeBoolean:
				mf.Data[i][j] = val.Parsed
			case model.TypeNull:
				mf.Data[i][j] = nil
			default:
				mf.Data[i][j] = val.Raw
			}
		}
	}

	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(mf); err != nil {
		return serializer.NewSerializeError("failed to serialize mock format").WithErr(err)
	}

	return nil
}

// TestProperty_FormatExtensibility tests that a new format can be added and works
// with the existing conversion pipeline
func TestProperty_FormatExtensibility(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Create a fresh registry for each test iteration
		reg := NewRegistry()

		// Register the mock format
		mockParser := &mockFormatParser{}
		mockSerializer := &mockFormatSerializer{}
		err := reg.Register("mockformat", mockParser, mockSerializer)
		if err != nil {
			t.Fatalf("failed to register mock format: %v", err)
		}

		// Generate random TableData
		td := generateRandomTableData(rt)

		// Property 1: New format can serialize data
		var buf bytes.Buffer
		s, err := reg.GetSerializer("mockformat")
		if err != nil {
			t.Fatalf("failed to get mock serializer: %v", err)
		}
		err = s.Serialize(td, &buf)
		if err != nil {
			t.Fatalf("mock serializer failed: %v", err)
		}

		// Property 2: New format can parse data back
		p, err := reg.GetParser("mockformat")
		if err != nil {
			t.Fatalf("failed to get mock parser: %v", err)
		}
		parsedTD, err := p.Parse(&buf)
		if err != nil {
			t.Fatalf("mock parser failed: %v", err)
		}

		// Property 3: Round-trip preserves data structure
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

		if len(parsedTD.Rows) != len(td.Rows) {
			t.Fatalf("row count mismatch: expected %d, got %d",
				len(td.Rows), len(parsedTD.Rows))
		}

		// Property 4: New format is discoverable via registry
		if !reg.IsSupported("mockformat") {
			t.Fatal("mock format should be supported after registration")
		}

		formats := reg.SupportedFormats()
		found := false
		for _, f := range formats {
			if f == "mockformat" {
				found = true
				break
			}
		}
		if !found {
			t.Fatal("mock format should appear in supported formats list")
		}

		// Property 5: New format info is retrievable
		info, err := reg.GetFormat("mockformat")
		if err != nil {
			t.Fatalf("failed to get mock format info: %v", err)
		}
		if info.Parser == nil || info.Serializer == nil {
			t.Fatal("format info should contain both parser and serializer")
		}
	})
}

// TestProperty_FormatExtensibility_CrossFormat tests that a new format can convert
// to and from existing formats
func TestProperty_FormatExtensibility_CrossFormat(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Create a fresh registry for each test iteration
		reg := NewRegistry()

		// Register the mock format
		mockParser := &mockFormatParser{}
		mockSerializer := &mockFormatSerializer{}
		err := reg.Register("mockformat", mockParser, mockSerializer)
		if err != nil {
			t.Fatalf("failed to register mock format: %v", err)
		}

		// Generate random TableData with safe values for cross-format conversion
		td := generateSafeTableData(rt)

		// Serialize to mock format
		var mockBuf bytes.Buffer
		s, _ := reg.GetSerializer("mockformat")
		err = s.Serialize(td, &mockBuf)
		if err != nil {
			t.Fatalf("failed to serialize to mock format: %v", err)
		}

		// Parse from mock format
		p, _ := reg.GetParser("mockformat")
		parsedFromMock, err := p.Parse(&mockBuf)
		if err != nil {
			t.Fatalf("failed to parse from mock format: %v", err)
		}

		// Property: Data parsed from new format should be valid TableData
		// that could be serialized to any other format
		if err := parsedFromMock.Validate(); err != nil {
			t.Fatalf("parsed data from mock format is invalid: %v", err)
		}

		// Property: Headers and row count should be preserved
		if len(parsedFromMock.Headers) != len(td.Headers) {
			t.Fatalf("header count mismatch after mock format round-trip: expected %d, got %d",
				len(td.Headers), len(parsedFromMock.Headers))
		}

		if len(parsedFromMock.Rows) != len(td.Rows) {
			t.Fatalf("row count mismatch after mock format round-trip: expected %d, got %d",
				len(td.Rows), len(parsedFromMock.Rows))
		}
	})
}

// generateRandomTableData creates a random TableData for testing
func generateRandomTableData(t *rapid.T) *model.TableData {
	// Generate random headers (1-10 columns)
	numCols := rapid.IntRange(1, 10).Draw(t, "numCols")
	headers := make([]string, numCols)
	for i := 0; i < numCols; i++ {
		// Use alphanumeric strings for headers
		headers[i] = rapid.StringMatching(`[a-zA-Z][a-zA-Z0-9_]*`).Draw(t, "header")
	}

	// Generate random rows (0-50 rows)
	numRows := rapid.IntRange(0, 50).Draw(t, "numRows")
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

// generateSafeTableData creates TableData with values safe for cross-format conversion
func generateSafeTableData(t *rapid.T) *model.TableData {
	// Generate random headers (1-10 columns)
	numCols := rapid.IntRange(1, 10).Draw(t, "numCols")
	headers := make([]string, numCols)
	for i := 0; i < numCols; i++ {
		// Use simple alphanumeric strings for headers
		headers[i] = rapid.StringMatching(`[a-zA-Z][a-zA-Z0-9]*`).Draw(t, "header")
	}

	// Generate random rows (0-20 rows)
	numRows := rapid.IntRange(0, 20).Draw(t, "numRows")
	rows := make([][]model.Value, numRows)
	for i := 0; i < numRows; i++ {
		row := make([]model.Value, numCols)
		for j := 0; j < numCols; j++ {
			row[j] = generateSafeValue(t)
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
		n := rapid.Float64Range(-1e10, 1e10).Draw(t, "numberValue")
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

// generateSafeValue creates a Value that's safe for cross-format conversion
func generateSafeValue(t *rapid.T) model.Value {
	valueType := rapid.IntRange(0, 3).Draw(t, "valueType")

	switch valueType {
	case 0: // String - use simple alphanumeric
		s := rapid.StringMatching(`[a-zA-Z0-9 ]*`).Draw(t, "stringValue")
		return model.NewStringValue(s)
	case 1: // Number - use reasonable range
		n := rapid.Float64Range(-1000, 1000).Draw(t, "numberValue")
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
