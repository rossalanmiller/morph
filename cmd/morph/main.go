package main

import (
	"os"

	"github.com/user/table-converter/internal/cli"
	"github.com/user/table-converter/internal/parser"
	"github.com/user/table-converter/internal/registry"
	"github.com/user/table-converter/internal/serializer"
)

func init() {
	// Register all supported formats with the global registry
	registerFormats()
}

func registerFormats() {
	// CSV format
	registry.Register(registry.FormatCSV, parser.NewCSVParser(), serializer.NewCSVSerializer())

	// Excel format
	registry.Register(registry.FormatExcel, parser.NewExcelParser(), serializer.NewExcelSerializer())

	// YAML format
	registry.Register(registry.FormatYAML, parser.NewYAMLParser(), serializer.NewYAMLSerializer())

	// JSON format
	registry.Register(registry.FormatJSON, parser.NewJSONParser(), serializer.NewJSONSerializer())

	// HTML format
	registry.Register(registry.FormatHTML, parser.NewHTMLParser(), serializer.NewHTMLSerializer())

	// XML format
	registry.Register(registry.FormatXML, parser.NewXMLParser(), serializer.NewXMLSerializer())

	// Markdown format
	registry.Register(registry.FormatMarkdown, parser.NewMarkdownParser(), serializer.NewMarkdownSerializer())

	// ASCII format
	registry.Register(registry.FormatASCII, parser.NewASCIIParser(), serializer.NewASCIISerializer())
}

func main() {
	exitCode := cli.Run(os.Args[1:], os.Stdout, os.Stderr)
	os.Exit(int(exitCode))
}
