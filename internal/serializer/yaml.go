package serializer

import (
	"fmt"
	"io"
	"strings"

	"github.com/user/table-converter/internal/model"
	"gopkg.in/yaml.v3"
)

// YAMLSerializer implements the Serializer interface for YAML format
type YAMLSerializer struct{}

// NewYAMLSerializer creates a new YAML serializer
func NewYAMLSerializer() *YAMLSerializer {
	return &YAMLSerializer{}
}

// Serialize writes TableData to the output writer in YAML format
// Output is a list of maps: [{header1: value1, header2: value2}, ...]
func (s *YAMLSerializer) Serialize(data *model.TableData, output io.Writer) error {
	if data == nil {
		return NewSerializeError("TableData is nil")
	}

	// Validate the table data
	if err := data.Validate(); err != nil {
		return NewSerializeError("invalid TableData").WithErr(err)
	}

	// Build YAML node tree to control scalar styles
	var rootNode yaml.Node
	rootNode.Kind = yaml.SequenceNode

	for _, row := range data.Rows {
		mapNode := yaml.Node{
			Kind: yaml.MappingNode,
		}

		for j, value := range row {
			if j < len(data.Headers) {
				// Key node
				keyNode := yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: data.Headers[j],
				}

				// Value node with appropriate style
				valueNode := modelValueToYAMLNode(value)

				mapNode.Content = append(mapNode.Content, &keyNode, &valueNode)
			}
		}

		rootNode.Content = append(rootNode.Content, &mapNode)
	}

	// Create encoder
	encoder := yaml.NewEncoder(output)
	encoder.SetIndent(2)
	defer encoder.Close()

	// Encode to output
	if err := encoder.Encode(&rootNode); err != nil {
		return NewSerializeError("failed to encode YAML").WithErr(err)
	}

	return nil
}

// modelValueToYAMLNode converts a model.Value to a yaml.Node with appropriate style
func modelValueToYAMLNode(val model.Value) yaml.Node {
	node := yaml.Node{
		Kind: yaml.ScalarNode,
	}

	switch val.Type {
	case model.TypeNull:
		// Use "null" which YAML parsers recognize without explicit tag
		node.Value = "null"
	case model.TypeBoolean:
		if b, ok := val.Parsed.(bool); ok {
			if b {
				node.Value = "true"
			} else {
				node.Value = "false"
			}
		} else {
			node.Value = val.Raw
		}
	case model.TypeNumber:
		if n, ok := val.Parsed.(float64); ok {
			node.Value = formatFloat(n)
		} else {
			node.Value = val.Raw
		}
	case model.TypeString:
		if s, ok := val.Parsed.(string); ok {
			node.Value = s
			// Use double-quoted style for:
			// - Empty strings (to distinguish from null)
			// - Strings containing newlines (to preserve them correctly on round-trip)
			// - Strings that look like numbers, booleans, or null (to preserve string type)
			if s == "" || strings.Contains(s, "\n") || looksLikeYAMLScalar(s) {
				node.Style = yaml.DoubleQuotedStyle
			}
		} else {
			node.Value = val.Raw
		}
	default:
		node.Value = val.Raw
	}

	return node
}

// formatFloat formats a float64 for YAML output
func formatFloat(n float64) string {
	// Use strconv for precise formatting
	if n == float64(int64(n)) {
		return fmt.Sprintf("%d", int64(n))
	}
	return fmt.Sprintf("%g", n)
}

// looksLikeYAMLScalar checks if a string looks like a YAML scalar value
// that would be auto-converted to a non-string type during parsing
func looksLikeYAMLScalar(s string) bool {
	// Check for null-like values
	switch strings.ToLower(s) {
	case "null", "~", "":
		return true
	}

	// Check for boolean-like values
	switch strings.ToLower(s) {
	case "true", "false", "yes", "no", "on", "off", "y", "n":
		return true
	}

	// Check if it looks like a number (integer or float)
	// This includes scientific notation, hex, octal, etc.
	if len(s) > 0 {
		// Try to see if YAML would parse this as a number
		var v interface{}
		if err := yaml.Unmarshal([]byte(s), &v); err == nil {
			switch v.(type) {
			case int, int64, float64:
				return true
			}
		}
	}

	return false
}
