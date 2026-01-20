package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/user/table-converter/internal/parser"
	"github.com/user/table-converter/internal/serializer"
)

// ExitCode represents CLI exit codes
type ExitCode int

const (
	// ExitSuccess indicates successful execution
	ExitSuccess ExitCode = 0
	// ExitError indicates a general error
	ExitError ExitCode = 1
	// ExitUsageError indicates invalid command-line usage
	ExitUsageError ExitCode = 2
	// ExitFileReadError indicates a file read error
	ExitFileReadError ExitCode = 3
	// ExitFileWriteError indicates a file write error
	ExitFileWriteError ExitCode = 4
	// ExitParseError indicates a parse error
	ExitParseError ExitCode = 5
	// ExitUnsupportedFormat indicates an unsupported format error
	ExitUnsupportedFormat ExitCode = 6
)

// CLIError represents a CLI-specific error with exit code
type CLIError struct {
	Message  string
	ExitCode ExitCode
	Err      error
}

// Error implements the error interface
func (e *CLIError) Error() string {
	return e.Message
}

// Unwrap returns the underlying error
func (e *CLIError) Unwrap() error {
	return e.Err
}

// NewCLIError creates a new CLIError
func NewCLIError(message string, exitCode ExitCode) *CLIError {
	return &CLIError{
		Message:  message,
		ExitCode: exitCode,
	}
}

// WithErr wraps an underlying error
func (e *CLIError) WithErr(err error) *CLIError {
	e.Err = err
	return e
}

// FormatFileReadError formats a file read error with path and reason
func FormatFileReadError(filepath string, err error) *CLIError {
	msg := fmt.Sprintf("Error: Failed to read file %q\n  Reason: %v", filepath, err)
	return &CLIError{
		Message:  msg,
		ExitCode: ExitFileReadError,
		Err:      err,
	}
}

// FormatFileWriteError formats a file write error with path and reason
func FormatFileWriteError(filepath string, err error) *CLIError {
	msg := fmt.Sprintf("Error: Failed to write file %q\n  Reason: %v", filepath, err)
	return &CLIError{
		Message:  msg,
		ExitCode: ExitFileWriteError,
		Err:      err,
	}
}

// FormatParseError formats a parse error with format and location
func FormatParseError(format string, err error) *CLIError {
	var parseErr *parser.ParseError
	var msg string

	if errors.As(err, &parseErr) {
		msg = fmt.Sprintf("Error: Failed to parse %s input\n", format)
		if parseErr.Line != nil {
			msg += fmt.Sprintf("  Line %d", *parseErr.Line)
			if parseErr.Column != nil {
				msg += fmt.Sprintf(", Column %d", *parseErr.Column)
			}
			msg += fmt.Sprintf(": %s\n", parseErr.Message)
		} else {
			msg += fmt.Sprintf("  %s\n", parseErr.Message)
		}
		if parseErr.Context != "" {
			msg += fmt.Sprintf("  Context: %s", parseErr.Context)
		}
	} else {
		msg = fmt.Sprintf("Error: Failed to parse %s input\n  %v", format, err)
	}

	return &CLIError{
		Message:  strings.TrimSpace(msg),
		ExitCode: ExitParseError,
		Err:      err,
	}
}

// FormatSerializeError formats a serialization error
func FormatSerializeError(format string, err error) *CLIError {
	var serializeErr *serializer.SerializeError
	var msg string

	if errors.As(err, &serializeErr) {
		msg = fmt.Sprintf("Error: Failed to serialize to %s format\n", format)
		msg += fmt.Sprintf("  %s", serializeErr.Message)
		if serializeErr.Context != "" {
			msg += fmt.Sprintf("\n  Context: %s", serializeErr.Context)
		}
	} else {
		msg = fmt.Sprintf("Error: Failed to serialize to %s format\n  %v", format, err)
	}

	return &CLIError{
		Message:  strings.TrimSpace(msg),
		ExitCode: ExitError,
		Err:      err,
	}
}

// FormatUnsupportedFormatError formats an unsupported format error with list of supported formats
func FormatUnsupportedFormatError(format string) *CLIError {
	formats := SupportedFormats()
	formatStrs := make([]string, len(formats))
	for i, f := range formats {
		formatStrs[i] = string(f)
	}

	msg := fmt.Sprintf("Error: Unsupported format %q\n  Supported formats: %s",
		format, strings.Join(formatStrs, ", "))

	return &CLIError{
		Message:  msg,
		ExitCode: ExitUnsupportedFormat,
	}
}

// FormatUsageError formats a usage error
func FormatUsageError(message string) *CLIError {
	return &CLIError{
		Message:  fmt.Sprintf("Error: %s\n  Use -h or --help for usage information", message),
		ExitCode: ExitUsageError,
	}
}

// FormatError converts any error to a CLIError with appropriate exit code
func FormatError(err error) *CLIError {
	if err == nil {
		return nil
	}

	// Check if it's already a CLIError
	var cliErr *CLIError
	if errors.As(err, &cliErr) {
		return cliErr
	}

	// Check for parse error
	var parseErr *parser.ParseError
	if errors.As(err, &parseErr) {
		return FormatParseError("input", err)
	}

	// Check for serialize error
	var serializeErr *serializer.SerializeError
	if errors.As(err, &serializeErr) {
		return FormatSerializeError("output", err)
	}

	// Default to general error
	return &CLIError{
		Message:  fmt.Sprintf("Error: %v", err),
		ExitCode: ExitError,
		Err:      err,
	}
}

// GetExitCode returns the exit code for an error
// Returns ExitSuccess (0) if err is nil
func GetExitCode(err error) ExitCode {
	if err == nil {
		return ExitSuccess
	}

	var cliErr *CLIError
	if errors.As(err, &cliErr) {
		return cliErr.ExitCode
	}

	// Default to general error
	return ExitError
}

// IsNonZeroExitCode returns true if the error should result in a non-zero exit code
func IsNonZeroExitCode(err error) bool {
	return GetExitCode(err) != ExitSuccess
}
