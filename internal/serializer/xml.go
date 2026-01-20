package serializer

import (
	"encoding/xml"
	"io"
	"strings"

	"github.com/user/table-converter/internal/model"
)

// XMLSerializer implements the Serializer interface for XML format
type XMLSerializer struct {
	// Indent specifies the indentation string for pretty printing
	// If empty, output will be compact
	Indent string
}

// NewXMLSerializer creates a new XML serializer with default settings (pretty print)
func NewXMLSerializer() *XMLSerializer {
	return &XMLSerializer{
		Indent: "  ",
	}
}

// NewCompactXMLSerializer creates a new XML serializer with compact output
func NewCompactXMLSerializer() *XMLSerializer {
	return &XMLSerializer{
		Indent: "",
	}
}

// Serialize writes TableData to the output writer in XML format
// Output format: <?xml version="1.0" encoding="UTF-8"?><dataset><record>...</record></dataset>
func (s *XMLSerializer) Serialize(data *model.TableData, output io.Writer) error {
	if data == nil {
		return NewSerializeError("TableData is nil")
	}

	// Validate the table data
	if err := data.Validate(); err != nil {
		return NewSerializeError("invalid TableData").WithErr(err)
	}

	var sb strings.Builder
	indent := s.Indent
	newline := "\n"
	if indent == "" {
		newline = ""
	}

	// Write XML declaration
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString(newline)

	// Write opening dataset tag
	sb.WriteString("<dataset>")
	sb.WriteString(newline)

	// Write records
	for _, row := range data.Rows {
		sb.WriteString(indent)
		sb.WriteString("<record>")
		sb.WriteString(newline)

		for j, value := range row {
			if j < len(data.Headers) {
				header := data.Headers[j]
				sb.WriteString(indent)
				sb.WriteString(indent)
				sb.WriteString("<")
				sb.WriteString(escapeXMLName(header))
				sb.WriteString(">")
				sb.WriteString(escapeXMLContent(xmlValueToString(value)))
				sb.WriteString("</")
				sb.WriteString(escapeXMLName(header))
				sb.WriteString(">")
				sb.WriteString(newline)
			}
		}

		sb.WriteString(indent)
		sb.WriteString("</record>")
		sb.WriteString(newline)
	}

	// Write closing dataset tag
	sb.WriteString("</dataset>")
	sb.WriteString(newline)

	// Write to output
	_, err := output.Write([]byte(sb.String()))
	if err != nil {
		return NewSerializeError("failed to write XML output").WithErr(err)
	}

	return nil
}

// escapeXMLContent escapes special XML characters in content
func escapeXMLContent(s string) string {
	var sb strings.Builder
	for _, r := range s {
		switch r {
		case '<':
			sb.WriteString("&lt;")
		case '>':
			sb.WriteString("&gt;")
		case '&':
			sb.WriteString("&amp;")
		case '"':
			sb.WriteString("&quot;")
		case '\'':
			sb.WriteString("&apos;")
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// escapeXMLName ensures a string is a valid XML element name
// XML names must start with a letter or underscore and contain only letters, digits, hyphens, underscores, and periods
func escapeXMLName(s string) string {
	if s == "" {
		return "_"
	}

	var sb strings.Builder
	for i, r := range s {
		if i == 0 {
			// First character must be letter or underscore
			if isXMLNameStartChar(r) {
				sb.WriteRune(r)
			} else {
				sb.WriteRune('_')
				if isXMLNameChar(r) {
					sb.WriteRune(r)
				}
			}
		} else {
			if isXMLNameChar(r) {
				sb.WriteRune(r)
			} else {
				sb.WriteRune('_')
			}
		}
	}
	return sb.String()
}

// isXMLNameStartChar checks if a rune can start an XML name
func isXMLNameStartChar(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || r == '_'
}

// isXMLNameChar checks if a rune can be part of an XML name
func isXMLNameChar(r rune) bool {
	return isXMLNameStartChar(r) || (r >= '0' && r <= '9') || r == '-' || r == '.'
}

// xmlValueToString converts a model.Value to its string representation for XML
func xmlValueToString(val model.Value) string {
	switch val.Type {
	case model.TypeNull:
		return ""
	case model.TypeBoolean:
		if b, ok := val.Parsed.(bool); ok {
			if b {
				return "true"
			}
			return "false"
		}
		return val.Raw
	case model.TypeNumber:
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

// XMLDataset is used for marshaling TableData to XML
type XMLDataset struct {
	XMLName xml.Name    `xml:"dataset"`
	Records []XMLRecord `xml:"record"`
}

// XMLRecord represents a single record for XML marshaling
type XMLRecord struct {
	XMLName xml.Name   `xml:"record"`
	Fields  []XMLField `xml:",any"`
}

// XMLField represents a single field for XML marshaling
type XMLField struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}
