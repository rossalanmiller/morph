package cli

import (
	"fmt"
	"io"
	"os"
)

// IOHandler manages input and output streams for the CLI
type IOHandler struct {
	inputReader  io.ReadCloser
	outputWriter io.WriteCloser
	inputFile    string
	outputFile   string
}

// NewIOHandler creates a new IOHandler based on the config
func NewIOHandler(config *Config) (*IOHandler, error) {
	handler := &IOHandler{
		inputFile:  config.InputFile,
		outputFile: config.OutputFile,
	}

	// Set up input reader
	reader, err := createInputReader(config.InputFile)
	if err != nil {
		return nil, err
	}
	handler.inputReader = reader

	// Set up output writer
	writer, err := createOutputWriter(config.OutputFile)
	if err != nil {
		// Clean up input reader if output fails
		handler.inputReader.Close()
		return nil, err
	}
	handler.outputWriter = writer

	return handler, nil
}

// InputReader returns the input reader
func (h *IOHandler) InputReader() io.Reader {
	return h.inputReader
}

// OutputWriter returns the output writer
func (h *IOHandler) OutputWriter() io.Writer {
	return h.outputWriter
}

// Close closes both input and output streams
// It returns the first error encountered, but attempts to close both
func (h *IOHandler) Close() error {
	var firstErr error

	if h.inputReader != nil {
		if err := h.inputReader.Close(); err != nil {
			firstErr = err
		}
	}

	if h.outputWriter != nil {
		if err := h.outputWriter.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// IsStdin returns true if reading from stdin
func (h *IOHandler) IsStdin() bool {
	return h.inputFile == "" || h.inputFile == "-"
}

// IsStdout returns true if writing to stdout
func (h *IOHandler) IsStdout() bool {
	return h.outputFile == "" || h.outputFile == "-"
}

// createInputReader creates an input reader based on the file path
// If the path is empty or "-", it returns stdin
// Otherwise, it opens the file for reading
func createInputReader(filepath string) (io.ReadCloser, error) {
	// Use stdin if no file specified or "-" is used
	if filepath == "" || filepath == "-" {
		return io.NopCloser(os.Stdin), nil
	}

	// Open the file
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open input file %q: %w", filepath, err)
	}

	return file, nil
}

// createOutputWriter creates an output writer based on the file path
// If the path is empty or "-", it returns stdout
// Otherwise, it creates/truncates the file for writing
func createOutputWriter(filepath string) (io.WriteCloser, error) {
	// Use stdout if no file specified or "-" is used
	if filepath == "" || filepath == "-" {
		return nopWriteCloser{os.Stdout}, nil
	}

	// Create/truncate the file
	file, err := os.Create(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file %q: %w", filepath, err)
	}

	return file, nil
}

// nopWriteCloser wraps a Writer to provide a no-op Close method
type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error {
	return nil
}

// CreateInputReader is a standalone function to create an input reader
// Useful for testing or when not using the full IOHandler
func CreateInputReader(filepath string) (io.ReadCloser, error) {
	return createInputReader(filepath)
}

// CreateOutputWriter is a standalone function to create an output writer
// Useful for testing or when not using the full IOHandler
func CreateOutputWriter(filepath string) (io.WriteCloser, error) {
	return createOutputWriter(filepath)
}
