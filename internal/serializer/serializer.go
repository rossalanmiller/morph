package serializer

import (
	"fmt"
	"io"

	"github.com/user/table-converter/internal/model"
)

// Serializer defines the interface for serializing TableData to output
type Serializer interface {
	// Serialize writes TableData to the output writer in a specific format
	Serialize(data *model.TableData, output io.Writer) error
}

// SerializeError represents an error that occurred during serialization
type SerializeError struct {
	// Message describes what went wrong
	Message string
	// Context provides additional context about the error
	Context string
	// Err is the underlying error if any
	Err error
}

// Error implements the error interface
func (e *SerializeError) Error() string {
	msg := fmt.Sprintf("serialize error: %s", e.Message)
	
	if e.Context != "" {
		msg += fmt.Sprintf("\n  Context: %s", e.Context)
	}
	
	if e.Err != nil {
		msg += fmt.Sprintf("\n  Caused by: %v", e.Err)
	}
	
	return msg
}

// Unwrap returns the underlying error
func (e *SerializeError) Unwrap() error {
	return e.Err
}

// NewSerializeError creates a new SerializeError with the given message
func NewSerializeError(message string) *SerializeError {
	return &SerializeError{
		Message: message,
	}
}

// WithContext adds context to a SerializeError
func (e *SerializeError) WithContext(context string) *SerializeError {
	e.Context = context
	return e
}

// WithErr wraps an underlying error
func (e *SerializeError) WithErr(err error) *SerializeError {
	e.Err = err
	return e
}
