package cli

import (
	"fmt"
	"io"

	"github.com/user/table-converter/internal/model"
	"github.com/user/table-converter/internal/parser"
	"github.com/user/table-converter/internal/registry"
	"github.com/user/table-converter/internal/serializer"
)

// Parser interface for parsing input
type parserInterface interface {
	Parse(input io.Reader) (*model.TableData, error)
}

// Serializer interface for serializing output
type serializerInterface interface {
	Serialize(data *model.TableData, output io.Writer) error
}

// ConvertOptions holds options for the conversion process
type ConvertOptions struct {
	// InputFormat is the format of the input data
	InputFormat Format
	// OutputFormat is the format of the output data
	OutputFormat Format
	// Sheet is the Excel sheet name (optional, for Excel input)
	Sheet string
	// NoHeader indicates whether to treat the first row as data
	NoHeader bool
	// CSVDelimiter is the CSV field delimiter
	CSVDelimiter string
	// CSVLineTerminator is the CSV line terminator
	CSVLineTerminator string
	// CSVQuoteAll forces all CSV fields to be quoted
	CSVQuoteAll bool
}

// Convert performs the conversion from input to output using the specified formats
// It looks up the parser and serializer from the format registry,
// parses the input to TableData, and serializes it to the output
func Convert(input io.Reader, output io.Writer, opts ConvertOptions) error {
	// Validate formats
	if opts.InputFormat == "" {
		return NewCLIError("input format is required", ExitUsageError)
	}
	if opts.OutputFormat == "" {
		return NewCLIError("output format is required", ExitUsageError)
	}

	// Get parser - use custom CSV parser if delimiter specified
	var p parserInterface
	var err error
	if opts.InputFormat == FormatCSV && opts.CSVDelimiter != "" {
		delim := parser.ParseDelimiter(opts.CSVDelimiter)
		p = parser.NewCSVParserWithDelimiter(delim)
	} else {
		p, err = registry.GetParser(registry.Format(opts.InputFormat))
		if err != nil {
			return FormatUnsupportedFormatError(string(opts.InputFormat)).WithErr(err)
		}
	}

	// Get serializer - use custom CSV serializer if options specified
	var s serializerInterface
	if opts.OutputFormat == FormatCSV && (opts.CSVDelimiter != "" || opts.CSVLineTerminator != "" || opts.CSVQuoteAll) {
		var csvOpts []serializer.CSVSerializerOption
		if opts.CSVDelimiter != "" {
			csvOpts = append(csvOpts, serializer.WithDelimiter(parser.ParseDelimiter(opts.CSVDelimiter)))
		}
		if opts.CSVLineTerminator != "" {
			csvOpts = append(csvOpts, serializer.WithLineTerminator(serializer.ParseLineTerminator(opts.CSVLineTerminator)))
		}
		if opts.CSVQuoteAll {
			csvOpts = append(csvOpts, serializer.WithAlwaysQuote(true))
		}
		s = serializer.NewCSVSerializerWithOptions(csvOpts...)
	} else {
		s, err = registry.GetSerializer(registry.Format(opts.OutputFormat))
		if err != nil {
			return FormatUnsupportedFormatError(string(opts.OutputFormat)).WithErr(err)
		}
	}

	// Parse input to TableData
	tableData, err := p.Parse(input)
	if err != nil {
		return FormatParseError(string(opts.InputFormat), err)
	}

	// Serialize TableData to output
	if err := s.Serialize(tableData, output); err != nil {
		return FormatSerializeError(string(opts.OutputFormat), err)
	}

	return nil
}

// ConvertWithConfig performs conversion using a Config struct
// This is a convenience wrapper around Convert that extracts options from Config
func ConvertWithConfig(input io.Reader, output io.Writer, config *Config) error {
	return Convert(input, output, ConvertOptions{
		InputFormat:       config.InputFormat,
		OutputFormat:      config.OutputFormat,
		Sheet:             config.Sheet,
		NoHeader:          config.NoHeader,
		CSVDelimiter:      config.CSVDelimiter,
		CSVLineTerminator: config.CSVLineTerminator,
		CSVQuoteAll:       config.CSVQuoteAll,
	})
}

// Run executes the full CLI workflow:
// 1. Parse CLI arguments
// 2. Set up input reader and output writer
// 3. Call conversion function
// 4. Handle errors and return exit code
func Run(args []string, stdout, stderr io.Writer) ExitCode {
	// Parse CLI arguments
	config, err := ParseArgsWithOutput(args, stderr)
	if err != nil {
		cliErr := FormatUsageError(err.Error())
		fmt.Fprintln(stderr, cliErr.Message)
		return cliErr.ExitCode
	}

	// Handle help flag
	if config.ShowHelp {
		printUsage(stdout)
		return ExitSuccess
	}

	// Handle version flag
	if config.ShowVersion {
		PrintVersionTo(stdout)
		return ExitSuccess
	}

	// Set up I/O handler
	ioHandler, err := NewIOHandler(config)
	if err != nil {
		cliErr := FormatError(err)
		fmt.Fprintln(stderr, cliErr.Message)
		return cliErr.ExitCode
	}
	defer ioHandler.Close()

	// Perform conversion
	if err := ConvertWithConfig(ioHandler.InputReader(), ioHandler.OutputWriter(), config); err != nil {
		cliErr := FormatError(err)
		fmt.Fprintln(stderr, cliErr.Message)
		return cliErr.ExitCode
	}

	return ExitSuccess
}
