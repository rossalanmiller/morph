package parser

import (
	"io"
	"strings"

	"golang.org/x/net/html"

	"github.com/user/table-converter/internal/model"
)

// HTMLParser implements the Parser interface for HTML table format
type HTMLParser struct{}

// NewHTMLParser creates a new HTML parser
func NewHTMLParser() *HTMLParser {
	return &HTMLParser{}
}

// Parse reads HTML data from the input reader and converts it to TableData
// Expects input to contain at least one <table> element
func (p *HTMLParser) Parse(input io.Reader) (*model.TableData, error) {
	doc, err := html.Parse(input)
	if err != nil {
		return nil, NewParseError("failed to parse HTML").WithErr(err)
	}

	// Find the first table element
	tableNode := findFirstElement(doc, "table")
	if tableNode == nil {
		return nil, NewParseError("no <table> element found in HTML input")
	}

	// Extract headers and rows
	headers, rows, err := parseTable(tableNode)
	if err != nil {
		return nil, err
	}

	return model.NewTableData(headers, rows), nil
}

// findFirstElement recursively searches for the first element with the given tag name
func findFirstElement(n *html.Node, tagName string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tagName {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := findFirstElement(c, tagName); result != nil {
			return result
		}
	}
	return nil
}


// findAllElements finds all direct child elements with the given tag name
func findAllElements(n *html.Node, tagName string) []*html.Node {
	var elements []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == tagName {
			elements = append(elements, c)
		}
	}
	return elements
}

// findAllElementsRecursive finds all elements with the given tag name recursively
func findAllElementsRecursive(n *html.Node, tagName string) []*html.Node {
	var elements []*html.Node
	if n.Type == html.ElementNode && n.Data == tagName {
		elements = append(elements, n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		elements = append(elements, findAllElementsRecursive(c, tagName)...)
	}
	return elements
}

// getTextContent extracts all text content from a node and its children
func getTextContent(n *html.Node) string {
	if n == nil {
		return ""
	}
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(getTextContent(c))
	}
	return sb.String()
}

// parseTable extracts headers and rows from a table element
func parseTable(tableNode *html.Node) ([]string, [][]model.Value, error) {
	var headers []string
	var rows [][]model.Value

	// Look for thead element
	thead := findFirstElement(tableNode, "thead")
	tbody := findFirstElement(tableNode, "tbody")

	if thead != nil {
		// Extract headers from thead
		headerRow := findFirstElement(thead, "tr")
		if headerRow != nil {
			headers = extractCellsAsStrings(headerRow, "th")
			// If no th elements, try td
			if len(headers) == 0 {
				headers = extractCellsAsStrings(headerRow, "td")
			}
		}
	}

	// Get all tr elements for data rows
	var dataRows []*html.Node
	if tbody != nil {
		dataRows = findAllElements(tbody, "tr")
	} else {
		// No tbody, get all tr elements from table
		dataRows = findAllElementsRecursive(tableNode, "tr")
	}

	// If no headers found yet, use first row as headers
	if len(headers) == 0 && len(dataRows) > 0 {
		firstRow := dataRows[0]
		// Try th first, then td
		headers = extractCellsAsStrings(firstRow, "th")
		if len(headers) == 0 {
			headers = extractCellsAsStrings(firstRow, "td")
		}
		// Remove first row from data rows since it's headers
		if len(dataRows) > 1 {
			dataRows = dataRows[1:]
		} else {
			dataRows = nil
		}
	}

	// If still no headers, return empty table
	if len(headers) == 0 {
		return []string{}, [][]model.Value{}, nil
	}

	// Parse data rows
	for _, tr := range dataRows {
		// Skip if this is the header row in thead
		if thead != nil && isChildOf(tr, thead) {
			continue
		}
		rowValues := extractCellsAsValues(tr)
		rows = append(rows, rowValues)
	}

	return headers, rows, nil
}


// extractCellsAsStrings extracts cell text content from a row
func extractCellsAsStrings(tr *html.Node, cellTag string) []string {
	var cells []string
	for c := tr.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == cellTag {
			text := strings.TrimSpace(getTextContent(c))
			cells = append(cells, text)
		}
	}
	return cells
}

// extractCellsAsValues extracts cell values from a row (td or th)
func extractCellsAsValues(tr *html.Node) []model.Value {
	var values []model.Value
	for c := tr.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && (c.Data == "td" || c.Data == "th") {
			text := strings.TrimSpace(getTextContent(c))
			values = append(values, model.NewValue(text))
		}
	}
	return values
}

// isChildOf checks if node is a descendant of parent
func isChildOf(node, parent *html.Node) bool {
	for n := node.Parent; n != nil; n = n.Parent {
		if n == parent {
			return true
		}
	}
	return false
}
