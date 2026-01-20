package parser

import (
	"encoding/xml"
	"io"
	"sort"

	"github.com/user/table-converter/internal/model"
)

// XMLParser implements the Parser interface for XML format
type XMLParser struct{}

// NewXMLParser creates a new XML parser
func NewXMLParser() *XMLParser {
	return &XMLParser{}
}

// Parse reads XML data from the input reader and converts it to TableData
// Expects input to be in the format: <dataset><record>...</record></dataset>
func (p *XMLParser) Parse(input io.Reader) (*model.TableData, error) {
	// Read all input
	data, err := io.ReadAll(input)
	if err != nil {
		return nil, NewParseError("failed to read XML data").WithErr(err)
	}

	// Check for empty input
	if len(data) == 0 {
		return nil, NewParseError("XML input is empty")
	}

	// Parse XML into generic structure
	var dataset Dataset
	if err := xml.Unmarshal(data, &dataset); err != nil {
		return nil, NewParseError("failed to parse XML").WithErr(err)
	}

	// Handle empty dataset
	if len(dataset.Records) == 0 {
		return model.NewTableData([]string{}, [][]model.Value{}), nil
	}

	// Extract headers from union of all element names across all records
	headerSet := make(map[string]bool)
	for _, record := range dataset.Records {
		for _, field := range record.Fields {
			headerSet[field.XMLName.Local] = true
		}
	}

	// Sort headers for consistent ordering
	headers := make([]string, 0, len(headerSet))
	for key := range headerSet {
		headers = append(headers, key)
	}
	sort.Strings(headers)

	// Parse rows
	rows := make([][]model.Value, len(dataset.Records))
	for i, record := range dataset.Records {
		// Create a map of field name to value for this record
		fieldMap := make(map[string]string)
		for _, field := range record.Fields {
			fieldMap[field.XMLName.Local] = field.Value
		}

		// Build row in header order
		row := make([]model.Value, len(headers))
		for j, header := range headers {
			val, exists := fieldMap[header]
			if !exists || val == "" {
				row[j] = model.NewNullValue()
			} else {
				row[j] = model.NewValue(val)
			}
		}
		rows[i] = row
	}

	return model.NewTableData(headers, rows), nil
}

// Dataset represents the root XML element
type Dataset struct {
	XMLName xml.Name `xml:"dataset"`
	Records []Record `xml:"record"`
}

// Record represents a single record in the dataset
type Record struct {
	XMLName xml.Name `xml:"record"`
	Fields  []Field  `xml:",any"`
}

// Field represents a single field within a record
type Field struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}
