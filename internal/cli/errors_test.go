package cli

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/user/table-converter/internal/parser"
	"github.com/user/table-converter/internal/serializer"
)

func TestCLIError_Error(t *testing.T) {
	err := NewCLIError("test error", ExitError)
	if err.Error() != "test error" {
		t.Errorf("Error() = %q, want %q", err.Error(), "test error")
	}
}

func TestCLIError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := NewCLIError("test error", ExitError).WithErr(underlying)

	if err.Unwrap() != underlying {
		t.Error("Unwrap() did not return underlying error")
	}
}

func TestFormatFileReadError(t *testing.T) {
	tests := []struct {
		name     string
		filepath string
		err      error
		wantCode ExitCode
	}{
		{
			name:     "file not found",
			filepath: "/path/to/file.csv",
			err:      os.ErrNotExist,
			wantCode: ExitFileReadError,
		},
		{
			name:     "permission denied",
			filepath: "/protected/file.json",
			err:      os.ErrPermission,
			wantCode: ExitFileReadError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cliErr := FormatFileReadError(tt.filepath, tt.err)

			if cliErr.ExitCode != tt.wantCode {
				t.Errorf("ExitCode = %d, want %d", cliErr.ExitCode, tt.wantCode)
			}

			if !strings.Contains(cliErr.Message, tt.filepath) {
				t.Errorf("Message should contain filepath %q, got: %s", tt.filepath, cliErr.Message)
			}

			if !strings.Contains(cliErr.Message, "Error:") {
				t.Errorf("Message should contain 'Error:', got: %s", cliErr.Message)
			}
		})
	}
}

func TestFormatFileWriteError(t *testing.T) {
	tests := []struct {
		name     string
		filepath string
		err      error
		wantCode ExitCode
	}{
		{
			name:     "permission denied",
			filepath: "/readonly/file.csv",
			err:      os.ErrPermission,
			wantCode: ExitFileWriteError,
		},
		{
			name:     "disk full",
			filepath: "/path/to/output.json",
			err:      errors.New("no space left on device"),
			wantCode: ExitFileWriteError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cliErr := FormatFileWriteError(tt.filepath, tt.err)

			if cliErr.ExitCode != tt.wantCode {
				t.Errorf("ExitCode = %d, want %d", cliErr.ExitCode, tt.wantCode)
			}

			if !strings.Contains(cliErr.Message, tt.filepath) {
				t.Errorf("Message should contain filepath %q, got: %s", tt.filepath, cliErr.Message)
			}
		})
	}
}

func TestFormatParseError(t *testing.T) {
	t.Run("with ParseError", func(t *testing.T) {
		parseErr := parser.NewParseErrorWithLine("unclosed quote", 5).WithContext(`"John Doe`)
		cliErr := FormatParseError("csv", parseErr)

		if cliErr.ExitCode != ExitParseError {
			t.Errorf("ExitCode = %d, want %d", cliErr.ExitCode, ExitParseError)
		}

		if !strings.Contains(cliErr.Message, "csv") {
			t.Errorf("Message should contain format 'csv', got: %s", cliErr.Message)
		}

		if !strings.Contains(cliErr.Message, "Line 5") {
			t.Errorf("Message should contain 'Line 5', got: %s", cliErr.Message)
		}
	})

	t.Run("with ParseError with location", func(t *testing.T) {
		parseErr := parser.NewParseErrorWithLocation("invalid character", 10, 25)
		cliErr := FormatParseError("json", parseErr)

		if !strings.Contains(cliErr.Message, "Line 10") {
			t.Errorf("Message should contain 'Line 10', got: %s", cliErr.Message)
		}

		if !strings.Contains(cliErr.Message, "Column 25") {
			t.Errorf("Message should contain 'Column 25', got: %s", cliErr.Message)
		}
	})

	t.Run("with generic error", func(t *testing.T) {
		genericErr := errors.New("something went wrong")
		cliErr := FormatParseError("yaml", genericErr)

		if cliErr.ExitCode != ExitParseError {
			t.Errorf("ExitCode = %d, want %d", cliErr.ExitCode, ExitParseError)
		}

		if !strings.Contains(cliErr.Message, "yaml") {
			t.Errorf("Message should contain format 'yaml', got: %s", cliErr.Message)
		}
	})
}

func TestFormatSerializeError(t *testing.T) {
	t.Run("with SerializeError", func(t *testing.T) {
		serializeErr := serializer.NewSerializeError("invalid data").WithContext("row 5")
		cliErr := FormatSerializeError("json", serializeErr)

		if cliErr.ExitCode != ExitError {
			t.Errorf("ExitCode = %d, want %d", cliErr.ExitCode, ExitError)
		}

		if !strings.Contains(cliErr.Message, "json") {
			t.Errorf("Message should contain format 'json', got: %s", cliErr.Message)
		}

		if !strings.Contains(cliErr.Message, "invalid data") {
			t.Errorf("Message should contain error message, got: %s", cliErr.Message)
		}
	})

	t.Run("with generic error", func(t *testing.T) {
		genericErr := errors.New("write failed")
		cliErr := FormatSerializeError("csv", genericErr)

		if !strings.Contains(cliErr.Message, "csv") {
			t.Errorf("Message should contain format 'csv', got: %s", cliErr.Message)
		}
	})
}

func TestFormatUnsupportedFormatError(t *testing.T) {
	cliErr := FormatUnsupportedFormatError("xyz")

	if cliErr.ExitCode != ExitUnsupportedFormat {
		t.Errorf("ExitCode = %d, want %d", cliErr.ExitCode, ExitUnsupportedFormat)
	}

	if !strings.Contains(cliErr.Message, "xyz") {
		t.Errorf("Message should contain invalid format 'xyz', got: %s", cliErr.Message)
	}

	// Should list supported formats
	for _, format := range SupportedFormats() {
		if !strings.Contains(cliErr.Message, string(format)) {
			t.Errorf("Message should list supported format %q, got: %s", format, cliErr.Message)
		}
	}
}

func TestFormatUsageError(t *testing.T) {
	cliErr := FormatUsageError("missing required argument")

	if cliErr.ExitCode != ExitUsageError {
		t.Errorf("ExitCode = %d, want %d", cliErr.ExitCode, ExitUsageError)
	}

	if !strings.Contains(cliErr.Message, "missing required argument") {
		t.Errorf("Message should contain error message, got: %s", cliErr.Message)
	}

	if !strings.Contains(cliErr.Message, "-h") || !strings.Contains(cliErr.Message, "--help") {
		t.Errorf("Message should mention help flag, got: %s", cliErr.Message)
	}
}

func TestFormatError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		if FormatError(nil) != nil {
			t.Error("FormatError(nil) should return nil")
		}
	})

	t.Run("CLIError passthrough", func(t *testing.T) {
		original := NewCLIError("test", ExitFileReadError)
		result := FormatError(original)

		if result != original {
			t.Error("FormatError should return CLIError as-is")
		}
	})

	t.Run("ParseError conversion", func(t *testing.T) {
		parseErr := parser.NewParseError("test parse error")
		result := FormatError(parseErr)

		if result.ExitCode != ExitParseError {
			t.Errorf("ExitCode = %d, want %d", result.ExitCode, ExitParseError)
		}
	})

	t.Run("SerializeError conversion", func(t *testing.T) {
		serializeErr := serializer.NewSerializeError("test serialize error")
		result := FormatError(serializeErr)

		if result.ExitCode != ExitError {
			t.Errorf("ExitCode = %d, want %d", result.ExitCode, ExitError)
		}
	})

	t.Run("generic error conversion", func(t *testing.T) {
		genericErr := errors.New("generic error")
		result := FormatError(genericErr)

		if result.ExitCode != ExitError {
			t.Errorf("ExitCode = %d, want %d", result.ExitCode, ExitError)
		}
	})
}

func TestGetExitCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode ExitCode
	}{
		{
			name:     "nil error",
			err:      nil,
			wantCode: ExitSuccess,
		},
		{
			name:     "CLIError",
			err:      NewCLIError("test", ExitFileReadError),
			wantCode: ExitFileReadError,
		},
		{
			name:     "generic error",
			err:      errors.New("generic"),
			wantCode: ExitError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetExitCode(tt.err); got != tt.wantCode {
				t.Errorf("GetExitCode() = %d, want %d", got, tt.wantCode)
			}
		})
	}
}

func TestIsNonZeroExitCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "CLIError with non-zero code",
			err:  NewCLIError("test", ExitError),
			want: true,
		},
		{
			name: "generic error",
			err:  errors.New("generic"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNonZeroExitCode(tt.err); got != tt.want {
				t.Errorf("IsNonZeroExitCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExitCodes(t *testing.T) {
	// Verify exit codes are distinct and non-zero (except ExitSuccess)
	codes := []ExitCode{
		ExitSuccess,
		ExitError,
		ExitUsageError,
		ExitFileReadError,
		ExitFileWriteError,
		ExitParseError,
		ExitUnsupportedFormat,
	}

	// ExitSuccess should be 0
	if ExitSuccess != 0 {
		t.Errorf("ExitSuccess = %d, want 0", ExitSuccess)
	}

	// All other codes should be non-zero
	for _, code := range codes[1:] {
		if code == 0 {
			t.Errorf("Exit code %d should be non-zero", code)
		}
	}

	// All codes should be distinct
	seen := make(map[ExitCode]bool)
	for _, code := range codes {
		if seen[code] {
			t.Errorf("Duplicate exit code: %d", code)
		}
		seen[code] = true
	}
}
