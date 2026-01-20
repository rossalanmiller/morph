package parser

import (
	"errors"
	"strings"
	"testing"
)

func TestParseError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ParseError
		contains []string
	}{
		{
			name: "basic error",
			err:  NewParseError("invalid format"),
			contains: []string{
				"parse error",
				"invalid format",
			},
		},
		{
			name: "error with line",
			err:  NewParseErrorWithLine("unexpected token", 42),
			contains: []string{
				"parse error",
				"unexpected token",
				"line 42",
			},
		},
		{
			name: "error with line and column",
			err:  NewParseErrorWithLocation("syntax error", 10, 5),
			contains: []string{
				"parse error",
				"syntax error",
				"line 10",
				"column 5",
			},
		},
		{
			name: "error with context",
			err:  NewParseError("malformed data").WithContext("while parsing header row"),
			contains: []string{
				"parse error",
				"malformed data",
				"Context:",
				"while parsing header row",
			},
		},
		{
			name: "error with underlying error",
			err:  NewParseError("failed to read").WithErr(errors.New("io error")),
			contains: []string{
				"parse error",
				"failed to read",
				"Caused by:",
				"io error",
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

func TestParseError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := NewParseError("wrapper").WithErr(underlying)

	unwrapped := errors.Unwrap(err)
	if unwrapped != underlying {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, underlying)
	}
}

func TestParseError_Chaining(t *testing.T) {
	err := NewParseError("test error").
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
