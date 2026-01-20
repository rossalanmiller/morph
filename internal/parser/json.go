package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/user/table-converter/internal/model"
)

// JSONParser implements the Parser interface for JSON format
type JSONParser struct{}

// NewJSONParser creates a new JSON parser
func NewJSONParser() *JSONParser {
	return &JSONParser{}
}

// Parse reads JSON data from the input reader and converts it to TableData
// Expects input to be an array of objects: [{"key": "value"}, ...]
func (p *JSONParser) Parse(input io.Reader) (*model.TableData, error) {
	// Read all input
	data, err := io.ReadAll(input)
	if err != nil {
		return nil, NewParseError("failed to read JSON data").WithErr(err)
	}

	// Check for empty input
	if len(data) == 0 {
		return nil, NewParseError("JSON input is empty")
	}

	// Parse JSON into a slice of maps
	var records []map[string]interface{}
	if err := json.Unmarshal(data, &records); err != nil {
		// Check if it's a non-array structure
		var singleObj map[string]interface{}
		if json.Unmarshal(data, &singleObj) == nil {
			return nil, NewParseError("invalid JSON structure: expected array of objects, got object").
				WithContext("JSON input must be an array of objects, e.g., [{\"key\": \"value\"}]")
		}
		return nil, NewParseError("failed to parse JSON").WithErr(err)
	}

	// Handle empty array
	if len(records) == 0 {
		return model.NewTableData([]string{}, [][]model.Value{}), nil
	}

	// Extract headers from union of all keys across all records
	headerSet := make(map[string]bool)
	for _, record := range records {
		for key := range record {
			headerSet[key] = true
		}
	}

	// Sort headers for consistent ordering
	headers := make([]string, 0, len(headerSet))
	for key := range headerSet {
		headers = append(headers, key)
	}
	sort.Strings(headers)

	// Parse rows
	rows := make([][]model.Value, len(records))
	for i, record := range records {
		row := make([]model.Value, len(headers))
		for j, header := range headers {
			val, exists := record[header]
			if !exists || val == nil {
				row[j] = model.NewNullValue()
			} else {
				row[j] = jsonValueToModelValue(val)
			}
		}
		rows[i] = row
	}

	return model.NewTableData(headers, rows), nil
}

// jsonValueToModelValue converts a JSON value to a model.Value
func jsonValueToModelValue(val interface{}) model.Value {
	switch v := val.(type) {
	case nil:
		return model.NewNullValue()
	case bool:
		return model.NewBooleanValue(v)
	case float64:
		return model.NewNumberValue(v)
	case string:
		return model.NewStringValue(v)
	case json.Number:
		// Try to parse as float64
		if f, err := v.Float64(); err == nil {
			return model.NewNumberValue(f)
		}
		// Fall back to string
		return model.NewStringValue(string(v))
	default:
		// For complex types (arrays, nested objects), convert to JSON string
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return model.NewStringValue(fmt.Sprintf("%v", v))
		}
		return model.NewStringValue(string(jsonBytes))
	}
}
