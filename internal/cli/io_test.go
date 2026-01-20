package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateInputReader_File(t *testing.T) {
	// Create a temp file with test content
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_input.txt")
	testContent := "test content"
	
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	reader, err := CreateInputReader(testFile)
	if err != nil {
		t.Fatalf("CreateInputReader() error = %v", err)
	}
	defer reader.Close()

	// Read content
	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read content: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("content = %q, want %q", string(content), testContent)
	}
}

func TestCreateInputReader_Stdin(t *testing.T) {
	tests := []struct {
		name     string
		filepath string
	}{
		{"empty string", ""},
		{"dash", "-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := CreateInputReader(tt.filepath)
			if err != nil {
				t.Fatalf("CreateInputReader() error = %v", err)
			}
			defer reader.Close()

			// Just verify we got a reader (can't easily test stdin content)
			if reader == nil {
				t.Error("CreateInputReader() returned nil reader")
			}
		})
	}
}

func TestCreateInputReader_FileNotFound(t *testing.T) {
	_, err := CreateInputReader("/nonexistent/path/file.txt")
	if err == nil {
		t.Fatal("CreateInputReader() expected error for nonexistent file")
	}

	if !strings.Contains(err.Error(), "failed to open input file") {
		t.Errorf("error = %q, want to contain 'failed to open input file'", err.Error())
	}
}

func TestCreateOutputWriter_File(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_output.txt")
	testContent := "output content"

	writer, err := CreateOutputWriter(testFile)
	if err != nil {
		t.Fatalf("CreateOutputWriter() error = %v", err)
	}

	// Write content
	_, err = writer.Write([]byte(testContent))
	if err != nil {
		t.Fatalf("failed to write content: %v", err)
	}

	// Close the writer
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	// Verify content was written
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("content = %q, want %q", string(content), testContent)
	}
}

func TestCreateOutputWriter_Stdout(t *testing.T) {
	tests := []struct {
		name     string
		filepath string
	}{
		{"empty string", ""},
		{"dash", "-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer, err := CreateOutputWriter(tt.filepath)
			if err != nil {
				t.Fatalf("CreateOutputWriter() error = %v", err)
			}
			defer writer.Close()

			// Just verify we got a writer
			if writer == nil {
				t.Error("CreateOutputWriter() returned nil writer")
			}
		})
	}
}

func TestCreateOutputWriter_InvalidPath(t *testing.T) {
	_, err := CreateOutputWriter("/nonexistent/directory/file.txt")
	if err == nil {
		t.Fatal("CreateOutputWriter() expected error for invalid path")
	}

	if !strings.Contains(err.Error(), "failed to create output file") {
		t.Errorf("error = %q, want to contain 'failed to create output file'", err.Error())
	}
}

func TestNewIOHandler_FileToFile(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.txt")
	outputFile := filepath.Join(tmpDir, "output.txt")
	testContent := "test data"

	// Create input file
	if err := os.WriteFile(inputFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}

	config := &Config{
		InputFile:    inputFile,
		OutputFile:   outputFile,
		InputFormat:  FormatCSV,
		OutputFormat: FormatJSON,
	}

	handler, err := NewIOHandler(config)
	if err != nil {
		t.Fatalf("NewIOHandler() error = %v", err)
	}
	defer handler.Close()

	// Verify IsStdin/IsStdout
	if handler.IsStdin() {
		t.Error("IsStdin() = true, want false")
	}
	if handler.IsStdout() {
		t.Error("IsStdout() = true, want false")
	}

	// Read from input
	content, err := io.ReadAll(handler.InputReader())
	if err != nil {
		t.Fatalf("failed to read input: %v", err)
	}
	if string(content) != testContent {
		t.Errorf("input content = %q, want %q", string(content), testContent)
	}

	// Write to output
	outputContent := "output data"
	_, err = handler.OutputWriter().Write([]byte(outputContent))
	if err != nil {
		t.Fatalf("failed to write output: %v", err)
	}

	// Close handler
	if err := handler.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// Verify output file content
	written, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if string(written) != outputContent {
		t.Errorf("output content = %q, want %q", string(written), outputContent)
	}
}

func TestNewIOHandler_StdinStdout(t *testing.T) {
	config := &Config{
		InputFile:    "",
		OutputFile:   "",
		InputFormat:  FormatJSON,
		OutputFormat: FormatCSV,
	}

	handler, err := NewIOHandler(config)
	if err != nil {
		t.Fatalf("NewIOHandler() error = %v", err)
	}
	defer handler.Close()

	// Verify IsStdin/IsStdout
	if !handler.IsStdin() {
		t.Error("IsStdin() = false, want true")
	}
	if !handler.IsStdout() {
		t.Error("IsStdout() = false, want true")
	}
}

func TestNewIOHandler_DashAsStdinStdout(t *testing.T) {
	config := &Config{
		InputFile:    "-",
		OutputFile:   "-",
		InputFormat:  FormatJSON,
		OutputFormat: FormatCSV,
	}

	handler, err := NewIOHandler(config)
	if err != nil {
		t.Fatalf("NewIOHandler() error = %v", err)
	}
	defer handler.Close()

	// Verify IsStdin/IsStdout
	if !handler.IsStdin() {
		t.Error("IsStdin() = false, want true for '-'")
	}
	if !handler.IsStdout() {
		t.Error("IsStdout() = false, want true for '-'")
	}
}

func TestNewIOHandler_InputFileNotFound(t *testing.T) {
	config := &Config{
		InputFile:    "/nonexistent/file.txt",
		OutputFile:   "",
		InputFormat:  FormatCSV,
		OutputFormat: FormatJSON,
	}

	_, err := NewIOHandler(config)
	if err == nil {
		t.Fatal("NewIOHandler() expected error for nonexistent input file")
	}

	if !strings.Contains(err.Error(), "failed to open input file") {
		t.Errorf("error = %q, want to contain 'failed to open input file'", err.Error())
	}
}

func TestNewIOHandler_OutputPathInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.txt")
	
	// Create input file
	if err := os.WriteFile(inputFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}

	config := &Config{
		InputFile:    inputFile,
		OutputFile:   "/nonexistent/directory/output.txt",
		InputFormat:  FormatCSV,
		OutputFormat: FormatJSON,
	}

	_, err := NewIOHandler(config)
	if err == nil {
		t.Fatal("NewIOHandler() expected error for invalid output path")
	}

	if !strings.Contains(err.Error(), "failed to create output file") {
		t.Errorf("error = %q, want to contain 'failed to create output file'", err.Error())
	}
}

func TestIOHandler_Close(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.txt")
	outputFile := filepath.Join(tmpDir, "output.txt")

	// Create input file
	if err := os.WriteFile(inputFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}

	config := &Config{
		InputFile:    inputFile,
		OutputFile:   outputFile,
		InputFormat:  FormatCSV,
		OutputFormat: FormatJSON,
	}

	handler, err := NewIOHandler(config)
	if err != nil {
		t.Fatalf("NewIOHandler() error = %v", err)
	}

	// Close should not return error
	if err := handler.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Double close should also not panic or error
	if err := handler.Close(); err != nil {
		// This might error on some systems, but shouldn't panic
		t.Logf("Double close error (expected on some systems): %v", err)
	}
}
