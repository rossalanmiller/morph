package parser

import (
	"strings"
	"testing"

	"pgregory.net/rapid"
)

// Feature: table-converter, Property 5: Invalid Input Error Handling
// Validates: Requirements 1.9, 6.2, 7.4
//
// Property: For any malformed input in a specific format, the parser should
// return a descriptive error (not crash) that indicates the nature of the problem.
func TestProperty_InvalidInputErrorHandling(t *testing.T) {
	t.Run("JSON_InvalidSyntax", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Generate invalid JSON
			invalidJSON := generateInvalidJSON(t)
			parser := NewJSONParser()

			_, err := parser.Parse(strings.NewReader(invalidJSON))

			// Should return error, not panic
			if err == nil {
				t.Fatalf("expected error for invalid JSON: %q", invalidJSON)
			}
		})
	})

	t.Run("JSON_NonArrayStructure", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Generate valid JSON that's not an array of objects
			nonArrayJSON := generateNonArrayJSON(t)
			parser := NewJSONParser()

			_, err := parser.Parse(strings.NewReader(nonArrayJSON))

			// Should return error for non-array structure
			if err == nil {
				t.Fatalf("expected error for non-array JSON: %q", nonArrayJSON)
			}
		})
	})

	t.Run("YAML_InvalidSyntax", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Generate invalid YAML
			invalidYAML := generateInvalidYAML(t)
			parser := NewYAMLParser()

			_, err := parser.Parse(strings.NewReader(invalidYAML))

			// Should return error, not panic
			if err == nil {
				t.Fatalf("expected error for invalid YAML: %q", invalidYAML)
			}
		})
	})

	t.Run("XML_InvalidSyntax", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Generate invalid XML
			invalidXML := generateInvalidXML(t)
			parser := NewXMLParser()

			_, err := parser.Parse(strings.NewReader(invalidXML))

			// Should return error, not panic
			if err == nil {
				t.Fatalf("expected error for invalid XML: %q", invalidXML)
			}
		})
	})

	t.Run("HTML_NoTable", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Generate HTML without a table
			noTableHTML := generateHTMLWithoutTable(t)
			parser := NewHTMLParser()

			// This might return empty table or error - both are acceptable
			td, err := parser.Parse(strings.NewReader(noTableHTML))
			if err == nil && td != nil && len(td.Headers) > 0 {
				t.Fatalf("expected empty table or error for HTML without table: %q", noTableHTML)
			}
		})
	})

	t.Run("Markdown_NoSeparator", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Generate markdown without separator row
			noSepMD := generateMarkdownWithoutSeparator(t)
			parser := NewMarkdownParser()

			_, err := parser.Parse(strings.NewReader(noSepMD))

			// Should return error for missing separator
			if err == nil {
				t.Fatalf("expected error for markdown without separator: %q", noSepMD)
			}
		})
	})
}


// generateInvalidJSON creates syntactically invalid JSON
func generateInvalidJSON(t *rapid.T) string {
	invalidType := rapid.IntRange(0, 4).Draw(t, "invalidType")
	switch invalidType {
	case 0: // Unclosed bracket
		return `[{"a": 1}`
	case 1: // Unclosed brace
		return `[{"a": 1]`
	case 2: // Missing comma
		return `[{"a": 1} {"b": 2}]`
	case 3: // Trailing comma
		return `[{"a": 1},]`
	case 4: // Invalid value
		return `[{"a": undefined}]`
	default:
		return `{invalid`
	}
}

// generateNonArrayJSON creates valid JSON that's not an array of objects
func generateNonArrayJSON(t *rapid.T) string {
	nonArrayType := rapid.IntRange(0, 3).Draw(t, "nonArrayType")
	switch nonArrayType {
	case 0: // Plain object
		return `{"a": 1, "b": 2}`
	case 1: // String
		return `"hello"`
	case 2: // Number
		return `42`
	case 3: // Array of non-objects
		return `[1, 2, 3]`
	default:
		return `null`
	}
}

// generateInvalidYAML creates syntactically invalid YAML
func generateInvalidYAML(t *rapid.T) string {
	invalidType := rapid.IntRange(0, 2).Draw(t, "invalidType")
	switch invalidType {
	case 0: // Bad indentation
		return "- a: 1\n  b: 2\n c: 3"
	case 1: // Invalid character
		return "- a: @invalid"
	case 2: // Tabs mixed with spaces incorrectly
		return "- a: 1\n\t- b: 2"
	default:
		return "{{{"
	}
}

// generateInvalidXML creates syntactically invalid XML
func generateInvalidXML(t *rapid.T) string {
	invalidType := rapid.IntRange(0, 3).Draw(t, "invalidType")
	switch invalidType {
	case 0: // Unclosed tag
		return `<dataset><record><a>1</a></record>`
	case 1: // Mismatched tags
		return `<dataset><record></dataset></record>`
	case 2: // Invalid character in tag name
		return `<dataset><123>value</123></dataset>`
	case 3: // Missing closing bracket
		return `<dataset<record></record></dataset>`
	default:
		return `<not valid xml`
	}
}

// generateHTMLWithoutTable creates HTML that doesn't contain a table
func generateHTMLWithoutTable(t *rapid.T) string {
	htmlType := rapid.IntRange(0, 2).Draw(t, "htmlType")
	switch htmlType {
	case 0: // Just text
		return `<html><body><p>Hello world</p></body></html>`
	case 1: // Div instead of table
		return `<html><body><div>Not a table</div></body></html>`
	case 2: // Empty body
		return `<html><body></body></html>`
	default:
		return `<html></html>`
	}
}

// generateMarkdownWithoutSeparator creates markdown table without separator row
func generateMarkdownWithoutSeparator(t *rapid.T) string {
	// Just a header row, no separator
	return `| col1 | col2 |
| data1 | data2 |`
}
