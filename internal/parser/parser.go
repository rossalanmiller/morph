package parser

import (
	"fmt"
	"io"

	"github.com/user/table-converter/internal/model"
)

// Parser defines the interface for parsing input data into TableData
type Parser interface {
	// Parse reads data from the input reader and converts it to TableData
	Parse(input io.Reader) (*model.TableData, error)
}

// ParseError represents an error that occurred during parsing
type ParseError struct {
	// Message describes what went wrong
	Message string
	// Line is the line number where the error occurred (optional, 1-indexed)
	Line *int
	// Column is the column number where the error occurred (optional, 1-indexed)
	Column *int
	// Context provides additional context about the error
	Context string
	// Err is the underlying error if any
	Err error
}

// Error implements the error interface
func (e *ParseError) Error() string {
	msg := fmt.Sprintf("parse error: %s", e.Message)
	
	if e.Line != nil {
		msg += fmt.Sprintf(" (line %d", *e.Line)
		if e.Column != nil {
			msg += fmt.Sprintf(", column %d", *e.Column)
		}
		msg += ")"
	}
	
	if e.Context != "" {
		msg += fmt.Sprintf("\n  Context: %s", e.Context)
	}
	
	if e.Err != nil {
		msg += fmt.Sprintf("\n  Caused by: %v", e.Err)
	}
	
	return msg
}

// Unwrap returns the underlying error
func (e *ParseError) Unwrap() error {
	return e.Err
}

// NewParseError creates a new ParseError with the given message
func NewParseError(message string) *ParseError {
	return &ParseError{
		Message: message,
	}
}

// NewParseErrorWithLine creates a new ParseError with line information
func NewParseErrorWithLine(message string, line int) *ParseError {
	return &ParseError{
		Message: message,
		Line:    &line,
	}
}

// NewParseErrorWithLocation creates a new ParseError with line and column information
func NewParseErrorWithLocation(message string, line, column int) *ParseError {
	return &ParseError{
		Message: message,
		Line:    &line,
		Column:  &column,
	}
}

// WithContext adds context to a ParseError
func (e *ParseError) WithContext(context string) *ParseError {
	e.Context = context
	return e
}

// WithErr wraps an underlying error
func (e *ParseError) WithErr(err error) *ParseError {
	e.Err = err
	return e
}
