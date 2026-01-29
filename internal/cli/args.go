package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// Version is the application version
const Version = "1.0.0"

// Config holds the parsed CLI configuration
type Config struct {
	InputFile    string // Input file path (empty for stdin)
	OutputFile   string // Output file path (empty for stdout)
	InputFormat  Format // Input format
	OutputFormat Format // Output format
	ShowHelp     bool   // Show help message
	ShowVersion  bool   // Show version
}

// ParseArgs parses command-line arguments and returns a Config
// It uses the provided args slice (typically os.Args[1:])
func ParseArgs(args []string) (*Config, error) {
	return ParseArgsWithOutput(args, io.Discard)
}

// ParseArgsWithOutput parses command-line arguments with custom output for help/version
func ParseArgsWithOutput(args []string, output io.Writer) (*Config, error) {
	fs := flag.NewFlagSet("morph", flag.ContinueOnError)
	fs.SetOutput(output)

	config := &Config{}

	// Define flags
	var inFormat, outFormat string
	fs.StringVar(&inFormat, "in", "", "Input format (csv|excel|yaml|json|html|xml|markdown|ascii)")
	fs.StringVar(&outFormat, "out", "", "Output format (csv|excel|yaml|json|html|xml|markdown|ascii)")

	// Custom help and version flags
	var showHelp, showVersion bool
	fs.BoolVar(&showHelp, "h", false, "Show help message")
	fs.BoolVar(&showHelp, "help", false, "Show help message")
	fs.BoolVar(&showVersion, "v", false, "Show version")
	fs.BoolVar(&showVersion, "version", false, "Show version")

	// Custom usage function
	fs.Usage = func() {
		printUsage(output)
	}

	// Parse flags
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			config.ShowHelp = true
			return config, nil
		}
		return nil, err
	}

	config.ShowHelp = showHelp
	config.ShowVersion = showVersion

	// If help or version requested, return early
	if config.ShowHelp || config.ShowVersion {
		return config, nil
	}

	// Parse positional arguments (input and output files)
	positionalArgs := fs.Args()
	if len(positionalArgs) > 0 {
		config.InputFile = positionalArgs[0]
	}
	if len(positionalArgs) > 1 {
		config.OutputFile = positionalArgs[1]
	}
	if len(positionalArgs) > 2 {
		return nil, fmt.Errorf("too many arguments: expected at most 2 positional arguments (input and output files), got %d", len(positionalArgs))
	}

	// Treat "-" as stdin/stdout (empty string internally)
	isStdin := config.InputFile == "" || config.InputFile == "-"
	isStdout := config.OutputFile == "" || config.OutputFile == "-"

	// Parse and validate input format
	if inFormat != "" {
		format, err := ParseFormat(inFormat)
		if err != nil {
			return nil, err
		}
		config.InputFormat = format
	} else if !isStdin {
		// Try to detect format from file extension
		format, err := DetectFormat(config.InputFile)
		if err != nil {
			return nil, fmt.Errorf("cannot determine input format: %w (use -in flag to specify format)", err)
		}
		config.InputFormat = format
	}

	// Parse and validate output format
	if outFormat != "" {
		format, err := ParseFormat(outFormat)
		if err != nil {
			return nil, err
		}
		config.OutputFormat = format
	} else if !isStdout {
		// Try to detect format from file extension
		format, err := DetectFormat(config.OutputFile)
		if err != nil {
			return nil, fmt.Errorf("cannot determine output format: %w (use -out flag to specify format)", err)
		}
		config.OutputFormat = format
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

// validateConfig validates the parsed configuration
func validateConfig(config *Config) error {
	isStdin := config.InputFile == "" || config.InputFile == "-"
	isStdout := config.OutputFile == "" || config.OutputFile == "-"

	// If reading from stdin, input format must be specified
	if isStdin && config.InputFormat == "" {
		return errors.New("input format required when reading from stdin (use -in flag)")
	}

	// If writing to stdout, output format must be specified
	if isStdout && config.OutputFormat == "" {
		return errors.New("output format required when writing to stdout (use -out flag)")
	}

	return nil
}

// printUsage prints the usage information
func printUsage(w io.Writer) {
	usage := `morph - Convert tabular data between formats

Usage:
  morph [OPTIONS] [INPUT_FILE] [OUTPUT_FILE]

Options:
  -in <format>      Input format (csv|excel|yaml|json|html|xml|markdown|ascii)
  -out <format>     Output format (csv|excel|yaml|json|html|xml|markdown|ascii)
  -h, --help        Show help message
  -v, --version     Show version

Examples:
  morph data.csv output.json
  morph -in json -out yaml < input.json > output.yaml
  echo '[{"a":1}]' | morph -in json -out csv

Supported formats:
  csv       - Comma-separated values
  excel     - Microsoft Excel (.xlsx, .xls)  [aliases: xlsx, xls, xl]
  yaml      - YAML format                    [aliases: yml]
  json      - JSON array of objects          [aliases: js]
  html      - HTML table                     [aliases: htm]
  xml       - XML dataset
  markdown  - GitHub-flavored markdown table [aliases: md]
  ascii     - ASCII box-drawing table        [aliases: txt, table]
`
	fmt.Fprint(w, usage)
}

// PrintHelp prints the help message to stdout
func PrintHelp() {
	printUsage(os.Stdout)
}

// PrintVersion prints the version to stdout
func PrintVersion() {
	fmt.Printf("morph version %s\n", Version)
}

// PrintVersionTo prints the version to the specified writer
func PrintVersionTo(w io.Writer) {
	fmt.Fprintf(w, "morph version %s\n", Version)
}

// FormatListString returns a formatted string of supported formats
func FormatListString() string {
	formats := SupportedFormats()
	strs := make([]string, len(formats))
	for i, f := range formats {
		strs[i] = string(f)
	}
	return strings.Join(strs, "|")
}
