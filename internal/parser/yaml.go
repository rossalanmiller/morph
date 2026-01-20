package parser

import (
	"fmt"
	"io"
	"sort"

	"github.com/user/table-converter/internal/model"
	"gopkg.in/yaml.v3"
)

// YAMLParser implements the Parser interface for YAML format
type YAMLParser struct{}

// NewYAMLParser creates a new YAML parser
func NewYAMLParser() *YAMLParser {
	return &YAMLParser{}
}

// Parse reads YAML data from the input reader and converts it to TableData
// Expects input to be a list of maps: [{key: value}, ...]
func (p *YAMLParser) Parse(input io.Reader) (*model.TableData, error) {
	// Read all input
	data, err := io.ReadAll(input)
	if err != nil {
		return nil, NewParseError("failed to read YAML data").WithErr(err)
	}

	// Check for empty input
	if len(data) == 0 {
		return nil, NewParseError("YAML input is empty")
	}

	// Parse YAML into a slice of maps
	var records []map[string]interface{}
	if err := yaml.Unmarshal(data, &records); err != nil {
		// Check if it's a non-list structure
		var singleObj map[string]interface{}
		if yaml.Unmarshal(data, &singleObj) == nil {
			return nil, NewParseError("invalid YAML structure: expected list of maps, got map").
				WithContext("YAML input must be a list of maps, e.g., [{key: value}]")
		}
		return nil, NewParseError("failed to parse YAML").WithErr(err)
	}

	// Handle empty list
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
				row[j] = yamlValueToModelValue(val)
			}
		}
		rows[i] = row
	}

	return model.NewTableData(headers, rows), nil
}

// yamlValueToModelValue converts a YAML value to a model.Value
func yamlValueToModelValue(val interface{}) model.Value {
	switch v := val.(type) {
	case nil:
		return model.NewNullValue()
	case bool:
		return model.NewBooleanValue(v)
	case int:
		return model.NewNumberValue(float64(v))
	case int64:
		return model.NewNumberValue(float64(v))
	case float64:
		return model.NewNumberValue(v)
	case string:
		return model.NewStringValue(v)
	default:
		// For complex types (arrays, nested objects), convert to YAML string
		yamlBytes, err := yaml.Marshal(v)
		if err != nil {
			return model.NewStringValue(fmt.Sprintf("%v", v))
		}
		return model.NewStringValue(string(yamlBytes))
	}
}
