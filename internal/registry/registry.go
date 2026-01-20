package registry

import (
	"fmt"
	"strings"
	"sync"

	"github.com/user/table-converter/internal/parser"
	"github.com/user/table-converter/internal/serializer"
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

// FormatInfo holds parser and serializer for a format
type FormatInfo struct {
	Name       Format
	Parser     parser.Parser
	Serializer serializer.Serializer
}

// Registry manages the mapping of format names to parsers and serializers
type Registry struct {
	mu      sync.RWMutex
	formats map[Format]*FormatInfo
}

// NewRegistry creates a new empty format registry
func NewRegistry() *Registry {
	return &Registry{
		formats: make(map[Format]*FormatInfo),
	}
}

// Register adds a new format to the registry
func (r *Registry) Register(name Format, p parser.Parser, s serializer.Serializer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if name == "" {
		return fmt.Errorf("format name cannot be empty")
	}

	// Normalize format name to lowercase
	name = Format(strings.ToLower(string(name)))

	if _, exists := r.formats[name]; exists {
		return fmt.Errorf("format %q is already registered", name)
	}

	r.formats[name] = &FormatInfo{
		Name:       name,
		Parser:     p,
		Serializer: s,
	}

	return nil
}

// GetParser retrieves the parser for a given format
func (r *Registry) GetParser(name Format) (parser.Parser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Normalize format name to lowercase
	name = Format(strings.ToLower(string(name)))

	info, exists := r.formats[name]
	if !exists {
		return nil, fmt.Errorf("unsupported format %q", name)
	}

	if info.Parser == nil {
		return nil, fmt.Errorf("no parser registered for format %q", name)
	}

	return info.Parser, nil
}

// GetSerializer retrieves the serializer for a given format
func (r *Registry) GetSerializer(name Format) (serializer.Serializer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Normalize format name to lowercase
	name = Format(strings.ToLower(string(name)))

	info, exists := r.formats[name]
	if !exists {
		return nil, fmt.Errorf("unsupported format %q", name)
	}

	if info.Serializer == nil {
		return nil, fmt.Errorf("no serializer registered for format %q", name)
	}

	return info.Serializer, nil
}

// GetFormat retrieves the complete format info for a given format
func (r *Registry) GetFormat(name Format) (*FormatInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Normalize format name to lowercase
	name = Format(strings.ToLower(string(name)))

	info, exists := r.formats[name]
	if !exists {
		return nil, fmt.Errorf("unsupported format %q", name)
	}

	return info, nil
}

// SupportedFormats returns a list of all registered format names
func (r *Registry) SupportedFormats() []Format {
	r.mu.RLock()
	defer r.mu.RUnlock()

	formats := make([]Format, 0, len(r.formats))
	for name := range r.formats {
		formats = append(formats, name)
	}

	return formats
}

// IsSupported checks if a format is registered
func (r *Registry) IsSupported(name Format) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Normalize format name to lowercase
	name = Format(strings.ToLower(string(name)))

	_, exists := r.formats[name]
	return exists
}

// Global registry instance
var globalRegistry = NewRegistry()

// Register adds a format to the global registry
func Register(name Format, p parser.Parser, s serializer.Serializer) error {
	return globalRegistry.Register(name, p, s)
}

// GetParser retrieves a parser from the global registry
func GetParser(name Format) (parser.Parser, error) {
	return globalRegistry.GetParser(name)
}

// GetSerializer retrieves a serializer from the global registry
func GetSerializer(name Format) (serializer.Serializer, error) {
	return globalRegistry.GetSerializer(name)
}

// GetFormat retrieves format info from the global registry
func GetFormat(name Format) (*FormatInfo, error) {
	return globalRegistry.GetFormat(name)
}

// SupportedFormats returns all formats from the global registry
func SupportedFormats() []Format {
	return globalRegistry.SupportedFormats()
}

// IsSupported checks if a format is in the global registry
func IsSupported(name Format) bool {
	return globalRegistry.IsSupported(name)
}
