package cli

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Format represents a supported data format
type Format string

const (
	FormatCSV      Format = "csv"
	FormatExcel    Format = "excel"
	FormatYAML     Format = "yaml"
	FormatJSON     Format = "json"
	FormatHTML     Format = "html"
	FormatXML      Format = "xml"
	FormatMarkdown Format = "markdown"
	FormatASCII    Format = "ascii"
)

// extensionMap maps file extensions to formats
var extensionMap = map[string]Format{
	".csv":  FormatCSV,
	".xlsx": FormatExcel,
	".xls":  FormatExcel,
	".yaml": FormatYAML,
	".yml":  FormatYAML,
	".json": FormatJSON,
	".html": FormatHTML,
	".htm":  FormatHTML,
	".xml":  FormatXML,
	".md":   FormatMarkdown,
	".txt":  FormatASCII, // ASCII tables often in .txt files
}

// aliasMap maps shorthand aliases to canonical format names
var aliasMap = map[string]Format{
	// Excel aliases
	"xlsx": FormatExcel,
	"xls":  FormatExcel,
	"xl":   FormatExcel,
	// YAML aliases
	"yml": FormatYAML,
	// Markdown aliases
	"md": FormatMarkdown,
	// HTML aliases
	"htm": FormatHTML,
	// ASCII aliases
	"txt":   FormatASCII,
	"table": FormatASCII,
	// JSON alias
	"js": FormatJSON,
}

// SupportedFormats returns a list of all supported format names
func SupportedFormats() []Format {
	return []Format{
		FormatCSV,
		FormatExcel,
		FormatYAML,
		FormatJSON,
		FormatHTML,
		FormatXML,
		FormatMarkdown,
		FormatASCII,
	}
}

// DetectFormat determines the format from a file path based on its extension
// Returns an error if the extension is unknown
func DetectFormat(filepath string) (Format, error) {
	ext := strings.ToLower(getExtension(filepath))
	if ext == "" {
		return "", fmt.Errorf("cannot detect format: file has no extension")
	}

	format, ok := extensionMap[ext]
	if !ok {
		return "", fmt.Errorf("unknown file extension %q, supported extensions: %s",
			ext, supportedExtensions())
	}

	return format, nil
}

// getExtension extracts the file extension from a path
func getExtension(path string) string {
	return filepath.Ext(path)
}

// supportedExtensions returns a comma-separated list of supported extensions
func supportedExtensions() string {
	exts := make([]string, 0, len(extensionMap))
	seen := make(map[Format]bool)
	for ext, format := range extensionMap {
		if !seen[format] {
			exts = append(exts, ext)
			seen[format] = true
		}
	}
	return strings.Join(exts, ", ")
}

// IsValidFormat checks if a format string is valid
func IsValidFormat(format string) bool {
	f := strings.ToLower(format)
	
	// Check canonical format names
	for _, supported := range SupportedFormats() {
		if f == string(supported) {
			return true
		}
	}
	
	// Check aliases
	_, ok := aliasMap[f]
	return ok
}

// ParseFormat converts a string to a Format, returning an error if invalid
func ParseFormat(format string) (Format, error) {
	f := strings.ToLower(format)
	
	// Check canonical format names first
	for _, supported := range SupportedFormats() {
		if f == string(supported) {
			return supported, nil
		}
	}
	
	// Check aliases
	if canonical, ok := aliasMap[f]; ok {
		return canonical, nil
	}
	
	return "", fmt.Errorf("unsupported format %q, supported formats: %s",
		format, formatList())
}

// formatList returns a comma-separated list of supported formats
func formatList() string {
	formats := SupportedFormats()
	strs := make([]string, len(formats))
	for i, f := range formats {
		strs[i] = string(f)
	}
	return strings.Join(strs, ", ")
}
