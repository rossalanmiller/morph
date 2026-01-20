package serializer

import (
	"errors"
	"strings"
	"testing"
)

func TestSerializeError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *SerializeError
		contains []string
	}{
		{
			name: "basic error",
			err:  NewSerializeError("invalid data"),
			contains: []string{
				"serialize error",
				"invalid data",
			},
		},
		{
			name: "error with context",
			err:  NewSerializeError("write failed").WithContext("while writing header"),
			contains: []string{
				"serialize error",
				"write failed",
				"Context:",
				"while writing header",
			},
		},
		{
			name: "error with underlying error",
			err:  NewSerializeError("output error").WithErr(errors.New("disk full")),
			contains: []string{
				"serialize error",
				"output error",
				"Caused by:",
				"disk full",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.err.Error()
			for _, substr := range tt.contains {
				if !strings.Contains(errMsg, substr) {
					t.Errorf("error message %q does not contain %q", errMsg, substr)
				}
			}
		})
	}
}

func TestSerializeError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := NewSerializeError("wrapper").WithErr(underlying)

	unwrapped := errors.Unwrap(err)
	if unwrapped != underlying {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, underlying)
	}
}

func TestSerializeError_Chaining(t *testing.T) {
	err := NewSerializeError("test error").
		WithContext("test context").
		WithErr(errors.New("underlying"))

	if err.Message != "test error" {
		t.Errorf("Message = %q, want %q", err.Message, "test error")
	}
	if err.Context != "test context" {
		t.Errorf("Context = %q, want %q", err.Context, "test context")
	}
	if err.Err == nil {
		t.Error("Err is nil, want non-nil")
	}
}
