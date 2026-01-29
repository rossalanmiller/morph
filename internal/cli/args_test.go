package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseArgs_ValidFlagCombinations(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantInputFile  string
		wantOutputFile string
		wantInFormat   Format
		wantOutFormat  Format
	}{
		{
			name:           "file to file with auto-detection",
			args:           []string{"input.csv", "output.json"},
			wantInputFile:  "input.csv",
			wantOutputFile: "output.json",
			wantInFormat:   FormatCSV,
			wantOutFormat:  FormatJSON,
		},
		{
			name:           "explicit formats override detection",
			args:           []string{"-in", "yaml", "-out", "xml", "input.csv", "output.json"},
			wantInputFile:  "input.csv",
			wantOutputFile: "output.json",
			wantInFormat:   FormatYAML,
			wantOutFormat:  FormatXML,
		},
		{
			name:          "stdin to stdout with explicit formats",
			args:          []string{"-in", "json", "-out", "csv"},
			wantInFormat:  FormatJSON,
			wantOutFormat: FormatCSV,
		},
		{
			name:           "file input to stdout",
			args:           []string{"-out", "yaml", "input.csv"},
			wantInputFile:  "input.csv",
			wantInFormat:   FormatCSV,
			wantOutFormat:  FormatYAML,
		},
		{
			name:           "stdin to file output using dash",
			args:           []string{"-in", "json", "-", "output.xlsx"},
			wantInputFile:  "-",
			wantOutputFile: "output.xlsx",
			wantInFormat:   FormatJSON,
			wantOutFormat:  FormatExcel,
		},
		{
			name:         "all formats - csv",
			args:         []string{"-in", "csv", "-out", "csv"},
			wantInFormat: FormatCSV,
			wantOutFormat: FormatCSV,
		},
		{
			name:         "all formats - excel",
			args:         []string{"-in", "excel", "-out", "excel"},
			wantInFormat: FormatExcel,
			wantOutFormat: FormatExcel,
		},
		{
			name:         "all formats - yaml",
			args:         []string{"-in", "yaml", "-out", "yaml"},
			wantInFormat: FormatYAML,
			wantOutFormat: FormatYAML,
		},
		{
			name:         "all formats - json",
			args:         []string{"-in", "json", "-out", "json"},
			wantInFormat: FormatJSON,
			wantOutFormat: FormatJSON,
		},
		{
			name:         "all formats - html",
			args:         []string{"-in", "html", "-out", "html"},
			wantInFormat: FormatHTML,
			wantOutFormat: FormatHTML,
		},
		{
			name:         "all formats - xml",
			args:         []string{"-in", "xml", "-out", "xml"},
			wantInFormat: FormatXML,
			wantOutFormat: FormatXML,
		},
		{
			name:         "all formats - markdown",
			args:         []string{"-in", "markdown", "-out", "markdown"},
			wantInFormat: FormatMarkdown,
			wantOutFormat: FormatMarkdown,
		},
		{
			name:         "all formats - ascii",
			args:         []string{"-in", "ascii", "-out", "ascii"},
			wantInFormat: FormatASCII,
			wantOutFormat: FormatASCII,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("ParseArgs() unexpected error: %v", err)
			}

			if config.InputFile != tt.wantInputFile {
				t.Errorf("InputFile = %q, want %q", config.InputFile, tt.wantInputFile)
			}
			if config.OutputFile != tt.wantOutputFile {
				t.Errorf("OutputFile = %q, want %q", config.OutputFile, tt.wantOutputFile)
			}
			if config.InputFormat != tt.wantInFormat {
				t.Errorf("InputFormat = %q, want %q", config.InputFormat, tt.wantInFormat)
			}
			if config.OutputFormat != tt.wantOutFormat {
				t.Errorf("OutputFormat = %q, want %q", config.OutputFormat, tt.wantOutFormat)
			}
		})
	}
}

func TestParseArgs_InvalidFlagCombinations(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErrMsg  string
	}{
		{
			name:       "stdin without input format",
			args:       []string{"-out", "csv"},
			wantErrMsg: "input format required when reading from stdin",
		},
		{
			name:       "stdout without output format",
			args:       []string{"-in", "csv"},
			wantErrMsg: "output format required when writing to stdout",
		},
		{
			name:       "invalid input format",
			args:       []string{"-in", "invalid", "-out", "csv"},
			wantErrMsg: "unsupported format",
		},
		{
			name:       "invalid output format",
			args:       []string{"-in", "csv", "-out", "invalid"},
			wantErrMsg: "unsupported format",
		},
		{
			name:       "unknown file extension for input",
			args:       []string{"input.xyz", "output.csv"},
			wantErrMsg: "cannot determine input format",
		},
		{
			name:       "unknown file extension for output",
			args:       []string{"input.csv", "output.xyz"},
			wantErrMsg: "cannot determine output format",
		},

		{
			name:       "too many positional arguments",
			args:       []string{"input.csv", "output.json", "extra.txt"},
			wantErrMsg: "too many arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseArgs(tt.args)
			if err == nil {
				t.Fatal("ParseArgs() expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Errorf("error = %q, want to contain %q", err.Error(), tt.wantErrMsg)
			}
		})
	}
}

func TestParseArgs_HelpFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"short help", []string{"-h"}},
		{"long help", []string{"--help"}},
		{"help with other args", []string{"-h", "input.csv"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("ParseArgs() unexpected error: %v", err)
			}
			if !config.ShowHelp {
				t.Error("ShowHelp = false, want true")
			}
		})
	}
}

func TestParseArgs_VersionFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"short version", []string{"-v"}},
		{"long version", []string{"--version"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("ParseArgs() unexpected error: %v", err)
			}
			if !config.ShowVersion {
				t.Error("ShowVersion = false, want true")
			}
		})
	}
}

func TestPrintHelp(t *testing.T) {
	var buf bytes.Buffer
	printUsage(&buf)
	output := buf.String()

	// Check that help contains key information
	expectedStrings := []string{
		"morph",
		"Usage:",
		"-in",
		"-out",
		"-h, --help",
		"-v, --version",
		"csv",
		"excel",
		"yaml",
		"json",
		"html",
		"xml",
		"markdown",
		"ascii",
		"Examples:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("help output missing %q", expected)
		}
	}
}

func TestPrintVersion(t *testing.T) {
	var buf bytes.Buffer
	PrintVersionTo(&buf)
	output := buf.String()

	if !strings.Contains(output, "morph") {
		t.Error("version output missing 'morph'")
	}
	if !strings.Contains(output, Version) {
		t.Errorf("version output missing version %q", Version)
	}
}

func TestParseArgsWithOutput(t *testing.T) {
	var buf bytes.Buffer
	
	// Test that help flag writes to output
	config, err := ParseArgsWithOutput([]string{"-h"}, &buf)
	if err != nil {
		t.Fatalf("ParseArgsWithOutput() unexpected error: %v", err)
	}
	if !config.ShowHelp {
		t.Error("ShowHelp = false, want true")
	}
}
