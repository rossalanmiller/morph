package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestMain ensures the morph binary is built before running integration tests
func TestMain(m *testing.M) {
	// Build the binary for testing
	cmd := exec.Command("go", "build", "-o", "morph_test", ".")
	cmd.Dir = "."
	if err := cmd.Run(); err != nil {
		os.Stderr.WriteString("Failed to build morph binary: " + err.Error() + "\n")
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	os.Remove("morph_test")

	os.Exit(code)
}

// runMorph executes the morph binary with the given arguments
func runMorph(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()

	cmd := exec.Command("./morph_test", args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("Failed to run morph: %v", err)
		}
	}

	return outBuf.String(), errBuf.String(), exitCode
}

// runMorphWithStdin executes the morph binary with stdin input
func runMorphWithStdin(t *testing.T, stdin string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()

	cmd := exec.Command("./morph_test", args...)
	cmd.Stdin = strings.NewReader(stdin)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("Failed to run morph: %v", err)
		}
	}

	return outBuf.String(), errBuf.String(), exitCode
}

// Test file-to-file conversion for various format pairs
func TestIntegration_FileToFileConversion(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name         string
		inputFormat  string
		outputFormat string
		inputExt     string
		outputExt    string
		inputData    string
		checkOutput  func(t *testing.T, output string)
	}{
		{
			name:         "CSV to JSON",
			inputFormat:  "csv",
			outputFormat: "json",
			inputExt:     ".csv",
			outputExt:    ".json",
			inputData:    "name,age\nAlice,30\nBob,25\n",
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "Alice") || !strings.Contains(output, "Bob") {
					t.Error("Output missing expected data")
				}
				if !strings.Contains(output, "[") || !strings.Contains(output, "]") {
					t.Error("Output is not valid JSON array")
				}
			},
		},
		{
			name:         "JSON to CSV",
			inputFormat:  "json",
			outputFormat: "csv",
			inputExt:     ".json",
			outputExt:    ".csv",
			inputData:    `[{"name":"Alice","age":"30"},{"name":"Bob","age":"25"}]`,
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "Alice") || !strings.Contains(output, "Bob") {
					t.Error("Output missing expected data")
				}
				if !strings.Contains(output, ",") {
					t.Error("Output is not valid CSV")
				}
			},
		},
		{
			name:         "CSV to YAML",
			inputFormat:  "csv",
			outputFormat: "yaml",
			inputExt:     ".csv",
			outputExt:    ".yaml",
			inputData:    "name,age\nAlice,30\n",
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "name:") || !strings.Contains(output, "Alice") {
					t.Error("Output missing expected YAML structure")
				}
			},
		},
		{
			name:         "CSV to HTML",
			inputFormat:  "csv",
			outputFormat: "html",
			inputExt:     ".csv",
			outputExt:    ".html",
			inputData:    "name,age\nAlice,30\n",
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "<table>") || !strings.Contains(output, "</table>") {
					t.Error("Output missing HTML table tags")
				}
				if !strings.Contains(output, "Alice") {
					t.Error("Output missing expected data")
				}
			},
		},
		{
			name:         "CSV to XML",
			inputFormat:  "csv",
			outputFormat: "xml",
			inputExt:     ".csv",
			outputExt:    ".xml",
			inputData:    "name,age\nAlice,30\n",
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "<dataset>") || !strings.Contains(output, "</dataset>") {
					t.Error("Output missing XML dataset tags")
				}
				if !strings.Contains(output, "Alice") {
					t.Error("Output missing expected data")
				}
			},
		},
		{
			name:         "CSV to Markdown",
			inputFormat:  "csv",
			outputFormat: "markdown",
			inputExt:     ".csv",
			outputExt:    ".md",
			inputData:    "name,age\nAlice,30\n",
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "|") {
					t.Error("Output missing markdown table pipes")
				}
				if !strings.Contains(output, "Alice") {
					t.Error("Output missing expected data")
				}
			},
		},
		{
			name:         "CSV to ASCII",
			inputFormat:  "csv",
			outputFormat: "ascii",
			inputExt:     ".csv",
			outputExt:    ".txt",
			inputData:    "name,age\nAlice,30\n",
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "+") || !strings.Contains(output, "-") {
					t.Error("Output missing ASCII table characters")
				}
				if !strings.Contains(output, "Alice") {
					t.Error("Output missing expected data")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputFile := filepath.Join(tmpDir, "input"+tt.inputExt)
			outputFile := filepath.Join(tmpDir, "output"+tt.outputExt)

			// Write input file
			if err := os.WriteFile(inputFile, []byte(tt.inputData), 0644); err != nil {
				t.Fatalf("Failed to write input file: %v", err)
			}

			// Run conversion
			_, stderr, exitCode := runMorph(t, inputFile, outputFile)

			if exitCode != 0 {
				t.Fatalf("morph exited with code %d, stderr: %s", exitCode, stderr)
			}

			// Read and check output
			output, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}

			tt.checkOutput(t, string(output))
		})
	}
}

// Test stdin to stdout conversion
func TestIntegration_StdinToStdout(t *testing.T) {
	tests := []struct {
		name        string
		inFormat    string
		outFormat   string
		input       string
		checkOutput func(t *testing.T, output string)
	}{
		{
			name:      "JSON to CSV via stdin/stdout",
			inFormat:  "json",
			outFormat: "csv",
			input:     `[{"name":"Alice","age":"30"}]`,
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "Alice") {
					t.Error("Output missing expected data")
				}
			},
		},
		{
			name:      "CSV to JSON via stdin/stdout",
			inFormat:  "csv",
			outFormat: "json",
			input:     "name,age\nAlice,30\n",
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "Alice") {
					t.Error("Output missing expected data")
				}
				if !strings.Contains(output, "[") {
					t.Error("Output is not JSON array")
				}
			},
		},
		{
			name:      "YAML to JSON via stdin/stdout",
			inFormat:  "yaml",
			outFormat: "json",
			input:     "- name: Alice\n  age: 30\n",
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "Alice") {
					t.Error("Output missing expected data")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, exitCode := runMorphWithStdin(t, tt.input, "-in", tt.inFormat, "-out", tt.outFormat)

			if exitCode != 0 {
				t.Fatalf("morph exited with code %d, stderr: %s", exitCode, stderr)
			}

			tt.checkOutput(t, stdout)
		})
	}
}

// Test format auto-detection from file extensions
func TestIntegration_FormatAutoDetection(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		inputExt  string
		outputExt string
		inputData string
	}{
		{
			name:      "CSV auto-detection",
			inputExt:  ".csv",
			outputExt: ".json",
			inputData: "name,age\nAlice,30\n",
		},
		{
			name:      "JSON auto-detection",
			inputExt:  ".json",
			outputExt: ".csv",
			inputData: `[{"name":"Alice","age":"30"}]`,
		},
		{
			name:      "YAML auto-detection (.yaml)",
			inputExt:  ".yaml",
			outputExt: ".json",
			inputData: "- name: Alice\n  age: 30\n",
		},
		{
			name:      "YAML auto-detection (.yml)",
			inputExt:  ".yml",
			outputExt: ".json",
			inputData: "- name: Alice\n  age: 30\n",
		},
		{
			name:      "HTML auto-detection",
			inputExt:  ".html",
			outputExt: ".csv",
			inputData: "<table><thead><tr><th>name</th></tr></thead><tbody><tr><td>Alice</td></tr></tbody></table>",
		},
		{
			name:      "XML auto-detection",
			inputExt:  ".xml",
			outputExt: ".csv",
			inputData: "<?xml version=\"1.0\"?><dataset><record><name>Alice</name></record></dataset>",
		},
		{
			name:      "Markdown auto-detection",
			inputExt:  ".md",
			outputExt: ".csv",
			inputData: "| name |\n|------|\n| Alice |\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputFile := filepath.Join(tmpDir, "input"+tt.inputExt)
			outputFile := filepath.Join(tmpDir, "output"+tt.outputExt)

			// Write input file
			if err := os.WriteFile(inputFile, []byte(tt.inputData), 0644); err != nil {
				t.Fatalf("Failed to write input file: %v", err)
			}

			// Run conversion without explicit format flags
			_, stderr, exitCode := runMorph(t, inputFile, outputFile)

			if exitCode != 0 {
				t.Fatalf("morph exited with code %d, stderr: %s", exitCode, stderr)
			}

			// Verify output file was created
			if _, err := os.Stat(outputFile); os.IsNotExist(err) {
				t.Error("Output file was not created")
			}
		})
	}
}

// Test error scenarios
func TestIntegration_ErrorScenarios(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("missing input file", func(t *testing.T) {
		_, stderr, exitCode := runMorph(t, "/nonexistent/file.csv", filepath.Join(tmpDir, "output.json"))

		if exitCode == 0 {
			t.Error("Expected non-zero exit code for missing file")
		}
		if !strings.Contains(stderr, "Error") {
			t.Error("Expected error message in stderr")
		}
	})

	t.Run("invalid input format", func(t *testing.T) {
		_, stderr, exitCode := runMorph(t, "-in", "invalid", "-out", "csv")

		if exitCode == 0 {
			t.Error("Expected non-zero exit code for invalid format")
		}
		if !strings.Contains(stderr, "unsupported format") && !strings.Contains(stderr, "Error") {
			t.Errorf("Expected error message about unsupported format, got: %s", stderr)
		}
	})

	t.Run("invalid output format", func(t *testing.T) {
		_, stderr, exitCode := runMorph(t, "-in", "csv", "-out", "invalid")

		if exitCode == 0 {
			t.Error("Expected non-zero exit code for invalid format")
		}
		if !strings.Contains(stderr, "unsupported format") && !strings.Contains(stderr, "Error") {
			t.Errorf("Expected error message about unsupported format, got: %s", stderr)
		}
	})

	t.Run("malformed CSV input", func(t *testing.T) {
		inputFile := filepath.Join(tmpDir, "malformed.csv")
		// Create a CSV with unclosed quote
		if err := os.WriteFile(inputFile, []byte("name,age\n\"Alice,30\n"), 0644); err != nil {
			t.Fatalf("Failed to write input file: %v", err)
		}

		_, stderr, exitCode := runMorph(t, inputFile, filepath.Join(tmpDir, "output.json"))

		if exitCode == 0 {
			t.Error("Expected non-zero exit code for malformed input")
		}
		if !strings.Contains(stderr, "Error") {
			t.Errorf("Expected error message, got: %s", stderr)
		}
	})

	t.Run("malformed JSON input", func(t *testing.T) {
		inputFile := filepath.Join(tmpDir, "malformed.json")
		// Create invalid JSON
		if err := os.WriteFile(inputFile, []byte("{invalid json}"), 0644); err != nil {
			t.Fatalf("Failed to write input file: %v", err)
		}

		_, stderr, exitCode := runMorph(t, inputFile, filepath.Join(tmpDir, "output.csv"))

		if exitCode == 0 {
			t.Error("Expected non-zero exit code for malformed input")
		}
		if !strings.Contains(stderr, "Error") {
			t.Errorf("Expected error message, got: %s", stderr)
		}
	})

	t.Run("stdin without input format", func(t *testing.T) {
		_, stderr, exitCode := runMorphWithStdin(t, "some data", "-out", "csv")

		if exitCode == 0 {
			t.Error("Expected non-zero exit code when stdin format not specified")
		}
		if !strings.Contains(stderr, "input format required") && !strings.Contains(stderr, "Error") {
			t.Errorf("Expected error about input format, got: %s", stderr)
		}
	})

	t.Run("stdout without output format", func(t *testing.T) {
		inputFile := filepath.Join(tmpDir, "input.csv")
		if err := os.WriteFile(inputFile, []byte("name\nAlice\n"), 0644); err != nil {
			t.Fatalf("Failed to write input file: %v", err)
		}

		_, stderr, exitCode := runMorph(t, inputFile)

		if exitCode == 0 {
			t.Error("Expected non-zero exit code when stdout format not specified")
		}
		if !strings.Contains(stderr, "output format required") && !strings.Contains(stderr, "Error") {
			t.Errorf("Expected error about output format, got: %s", stderr)
		}
	})

	t.Run("unknown file extension", func(t *testing.T) {
		inputFile := filepath.Join(tmpDir, "input.xyz")
		if err := os.WriteFile(inputFile, []byte("data"), 0644); err != nil {
			t.Fatalf("Failed to write input file: %v", err)
		}

		_, stderr, exitCode := runMorph(t, inputFile, filepath.Join(tmpDir, "output.json"))

		if exitCode == 0 {
			t.Error("Expected non-zero exit code for unknown extension")
		}
		if !strings.Contains(stderr, "cannot determine") && !strings.Contains(stderr, "Error") {
			t.Errorf("Expected error about format detection, got: %s", stderr)
		}
	})

	t.Run("output to invalid path", func(t *testing.T) {
		inputFile := filepath.Join(tmpDir, "input.csv")
		if err := os.WriteFile(inputFile, []byte("name\nAlice\n"), 0644); err != nil {
			t.Fatalf("Failed to write input file: %v", err)
		}

		_, stderr, exitCode := runMorph(t, inputFile, "/nonexistent/directory/output.json")

		if exitCode == 0 {
			t.Error("Expected non-zero exit code for invalid output path")
		}
		if !strings.Contains(stderr, "Error") {
			t.Errorf("Expected error message, got: %s", stderr)
		}
	})
}

// Test help and version flags
func TestIntegration_HelpAndVersion(t *testing.T) {
	t.Run("help flag -h", func(t *testing.T) {
		stdout, _, exitCode := runMorph(t, "-h")

		if exitCode != 0 {
			t.Errorf("Expected exit code 0 for help, got %d", exitCode)
		}
		if !strings.Contains(stdout, "Usage:") {
			t.Error("Help output missing 'Usage:'")
		}
		if !strings.Contains(stdout, "-in") || !strings.Contains(stdout, "-out") {
			t.Error("Help output missing format flags")
		}
	})

	t.Run("help flag --help", func(t *testing.T) {
		stdout, _, exitCode := runMorph(t, "--help")

		if exitCode != 0 {
			t.Errorf("Expected exit code 0 for help, got %d", exitCode)
		}
		if !strings.Contains(stdout, "Usage:") {
			t.Error("Help output missing 'Usage:'")
		}
	})

	t.Run("version flag -v", func(t *testing.T) {
		stdout, _, exitCode := runMorph(t, "-v")

		if exitCode != 0 {
			t.Errorf("Expected exit code 0 for version, got %d", exitCode)
		}
		if !strings.Contains(stdout, "morph") || !strings.Contains(stdout, "version") {
			t.Error("Version output missing expected content")
		}
	})

	t.Run("version flag --version", func(t *testing.T) {
		stdout, _, exitCode := runMorph(t, "--version")

		if exitCode != 0 {
			t.Errorf("Expected exit code 0 for version, got %d", exitCode)
		}
		if !strings.Contains(stdout, "morph") {
			t.Error("Version output missing 'morph'")
		}
	})
}

// Test Excel format conversion (requires actual Excel file handling)
func TestIntegration_ExcelConversion(t *testing.T) {
	tmpDir := t.TempDir()

	// First create an Excel file by converting from CSV
	csvFile := filepath.Join(tmpDir, "input.csv")
	excelFile := filepath.Join(tmpDir, "output.xlsx")
	csvOutput := filepath.Join(tmpDir, "roundtrip.csv")

	csvData := "name,age\nAlice,30\nBob,25\n"
	if err := os.WriteFile(csvFile, []byte(csvData), 0644); err != nil {
		t.Fatalf("Failed to write CSV file: %v", err)
	}

	// CSV to Excel
	_, stderr, exitCode := runMorph(t, csvFile, excelFile)
	if exitCode != 0 {
		t.Fatalf("CSV to Excel conversion failed: %s", stderr)
	}

	// Verify Excel file was created
	if _, err := os.Stat(excelFile); os.IsNotExist(err) {
		t.Fatal("Excel file was not created")
	}

	// Excel back to CSV
	_, stderr, exitCode = runMorph(t, excelFile, csvOutput)
	if exitCode != 0 {
		t.Fatalf("Excel to CSV conversion failed: %s", stderr)
	}

	// Verify round-trip data
	output, err := os.ReadFile(csvOutput)
	if err != nil {
		t.Fatalf("Failed to read output CSV: %v", err)
	}

	if !strings.Contains(string(output), "Alice") || !strings.Contains(string(output), "Bob") {
		t.Error("Round-trip data missing expected values")
	}
}
