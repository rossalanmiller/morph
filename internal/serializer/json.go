package serializer

import (
	"encoding/json"
	"io"

	"github.com/user/table-converter/internal/model"
)

// JSONSerializer implements the Serializer interface for JSON format
type JSONSerializer struct {
	// Indent specifies the indentation string for pretty printing
	// If empty, output will be compact
	Indent string
}

// NewJSONSerializer creates a new JSON serializer with default settings (pretty print)
func NewJSONSerializer() *JSONSerializer {
	return &JSONSerializer{
		Indent: "  ",
	}
}

// NewCompactJSONSerializer creates a new JSON serializer with compact output
func NewCompactJSONSerializer() *JSONSerializer {
	return &JSONSerializer{
		Indent: "",
	}
}

// Serialize writes TableData to the output writer in JSON format
// Output is an array of objects: [{"header1": "value1", "header2": "value2"}, ...]
func (s *JSONSerializer) Serialize(data *model.TableData, output io.Writer) error {
	if data == nil {
		return NewSerializeError("TableData is nil")
	}

	// Validate the table data
	if err := data.Validate(); err != nil {
		return NewSerializeError("invalid TableData").WithErr(err)
	}

	// Convert TableData to slice of maps
	records := make([]map[string]interface{}, len(data.Rows))
	for i, row := range data.Rows {
		record := make(map[string]interface{})
		for j, value := range row {
			if j < len(data.Headers) {
				record[data.Headers[j]] = modelValueToJSONValue(value)
			}
		}
		records[i] = record
	}

	// Create encoder
	encoder := json.NewEncoder(output)
	if s.Indent != "" {
		encoder.SetIndent("", s.Indent)
	}

	// Encode to output
	if err := encoder.Encode(records); err != nil {
		return NewSerializeError("failed to encode JSON").WithErr(err)
	}

	return nil
}

// modelValueToJSONValue converts a model.Value to a JSON-compatible value
func modelValueToJSONValue(val model.Value) interface{} {
	switch val.Type {
	case model.TypeNull:
		return nil
	case model.TypeBoolean:
		if b, ok := val.Parsed.(bool); ok {
			return b
		}
		return val.Raw
	case model.TypeNumber:
		if n, ok := val.Parsed.(float64); ok {
			return n
		}
		return val.Raw
	case model.TypeString:
		if s, ok := val.Parsed.(string); ok {
			return s
		}
		return val.Raw
	default:
		return val.Raw
	}
}
