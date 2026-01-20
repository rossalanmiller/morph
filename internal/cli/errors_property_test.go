package cli

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/user/table-converter/internal/parser"
	"github.com/user/table-converter/internal/serializer"
	"pgregory.net/rapid"
)

// Feature: table-converter, Property 6: CLI Error Exit Codes
// Validates: Requirements 6.1, 6.4, 6.5
//
// Property: For any error condition (file not found, parse error, invalid format,
// write error), the CLI should exit with a non-zero status code and display an error message.
func TestProperty_CLIErrorExitCodes(t *testing.T) {
	t.Run("FileReadErrors_NonZeroExitCode", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Generate random file path that doesn't exist
			filename := rapid.StringMatching(`[a-z]{5,10}\.(csv|json|yaml|xml)`).Draw(t, "filename")
			filepath := "/nonexistent/path/" + filename

			cliErr := FormatFileReadError(filepath, os.ErrNotExist)

			// Verify non-zero exit code
			if cliErr.ExitCode == ExitSuccess {
				t.Fatalf("expected non-zero exit code for file read error, got %d", cliErr.ExitCode)
			}

			// Verify error message contains file path
			if cliErr.Message == "" {
				t.Fatal("expected non-empty error message")
			}

			// Verify GetExitCode returns non-zero
			if GetExitCode(cliErr) == ExitSuccess {
				t.Fatal("GetExitCode should return non-zero for file read error")
			}

			// Verify IsNonZeroExitCode returns true
			if !IsNonZeroExitCode(cliErr) {
				t.Fatal("IsNonZeroExitCode should return true for file read error")
			}
		})
	})

	t.Run("FileWriteErrors_NonZeroExitCode", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Generate random file path
			filename := rapid.StringMatching(`[a-z]{5,10}\.(csv|json|yaml|xml)`).Draw(t, "filename")
			filepath := "/readonly/path/" + filename

			cliErr := FormatFileWriteError(filepath, os.ErrPermission)

			// Verify non-zero exit code
			if cliErr.ExitCode == ExitSuccess {
				t.Fatalf("expected non-zero exit code for file write error, got %d", cliErr.ExitCode)
			}

			// Verify error message is not empty
			if cliErr.Message == "" {
				t.Fatal("expected non-empty error message")
			}

			// Verify GetExitCode returns non-zero
			if GetExitCode(cliErr) == ExitSuccess {
				t.Fatal("GetExitCode should return non-zero for file write error")
			}
		})
	})

	t.Run("ParseErrors_NonZeroExitCode", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Generate random format name
			formats := []string{"csv", "json", "yaml", "xml", "html", "markdown", "ascii", "excel"}
			formatIdx := rapid.IntRange(0, len(formats)-1).Draw(t, "formatIdx")
			format := formats[formatIdx]

			// Generate random line number
			line := rapid.IntRange(1, 1000).Draw(t, "line")

			parseErr := parser.NewParseErrorWithLine("test error", line)
			cliErr := FormatParseError(format, parseErr)

			// Verify non-zero exit code
			if cliErr.ExitCode == ExitSuccess {
				t.Fatalf("expected non-zero exit code for parse error, got %d", cliErr.ExitCode)
			}

			// Verify error message is not empty
			if cliErr.Message == "" {
				t.Fatal("expected non-empty error message")
			}

			// Verify GetExitCode returns non-zero
			if GetExitCode(cliErr) == ExitSuccess {
				t.Fatal("GetExitCode should return non-zero for parse error")
			}
		})
	})

	t.Run("UnsupportedFormatErrors_NonZeroExitCode", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Generate random invalid format name
			invalidFormat := rapid.StringMatching(`[a-z]{3,8}`).Draw(t, "invalidFormat")
			// Make sure it's not a valid format
			if IsValidFormat(invalidFormat) {
				invalidFormat = invalidFormat + "_invalid"
			}

			cliErr := FormatUnsupportedFormatError(invalidFormat)

			// Verify non-zero exit code
			if cliErr.ExitCode == ExitSuccess {
				t.Fatalf("expected non-zero exit code for unsupported format error, got %d", cliErr.ExitCode)
			}

			// Verify error message is not empty
			if cliErr.Message == "" {
				t.Fatal("expected non-empty error message")
			}

			// Verify GetExitCode returns non-zero
			if GetExitCode(cliErr) == ExitSuccess {
				t.Fatal("GetExitCode should return non-zero for unsupported format error")
			}
		})
	})

	t.Run("UsageErrors_NonZeroExitCode", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Generate random error message
			message := rapid.StringMatching(`[a-zA-Z ]{10,50}`).Draw(t, "message")

			cliErr := FormatUsageError(message)

			// Verify non-zero exit code
			if cliErr.ExitCode == ExitSuccess {
				t.Fatalf("expected non-zero exit code for usage error, got %d", cliErr.ExitCode)
			}

			// Verify error message is not empty
			if cliErr.Message == "" {
				t.Fatal("expected non-empty error message")
			}

			// Verify GetExitCode returns non-zero
			if GetExitCode(cliErr) == ExitSuccess {
				t.Fatal("GetExitCode should return non-zero for usage error")
			}
		})
	})

	t.Run("SerializeErrors_NonZeroExitCode", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Generate random format name
			formats := []string{"csv", "json", "yaml", "xml", "html", "markdown", "ascii", "excel"}
			formatIdx := rapid.IntRange(0, len(formats)-1).Draw(t, "formatIdx")
			format := formats[formatIdx]

			serializeErr := serializer.NewSerializeError("test serialize error")
			cliErr := FormatSerializeError(format, serializeErr)

			// Verify non-zero exit code
			if cliErr.ExitCode == ExitSuccess {
				t.Fatalf("expected non-zero exit code for serialize error, got %d", cliErr.ExitCode)
			}

			// Verify error message is not empty
			if cliErr.Message == "" {
				t.Fatal("expected non-empty error message")
			}

			// Verify GetExitCode returns non-zero
			if GetExitCode(cliErr) == ExitSuccess {
				t.Fatal("GetExitCode should return non-zero for serialize error")
			}
		})
	})

	t.Run("NilError_ZeroExitCode", func(t *testing.T) {
		// Verify nil error returns zero exit code
		if GetExitCode(nil) != ExitSuccess {
			t.Fatal("GetExitCode(nil) should return ExitSuccess (0)")
		}

		if IsNonZeroExitCode(nil) {
			t.Fatal("IsNonZeroExitCode(nil) should return false")
		}
	})

	t.Run("GenericErrors_NonZeroExitCode", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Generate random error message
			message := rapid.StringMatching(`[a-zA-Z ]{10,50}`).Draw(t, "message")
			genericErr := errors.New(message)

			cliErr := FormatError(genericErr)

			// Verify non-zero exit code
			if cliErr.ExitCode == ExitSuccess {
				t.Fatalf("expected non-zero exit code for generic error, got %d", cliErr.ExitCode)
			}

			// Verify error message is not empty
			if cliErr.Message == "" {
				t.Fatal("expected non-empty error message")
			}
		})
	})
}

// TestProperty_CLIErrorMessages verifies that error messages contain useful information
func TestProperty_CLIErrorMessages(t *testing.T) {
	t.Run("FileReadError_ContainsPath", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Generate random file path
			filename := rapid.StringMatching(`[a-z]{5,10}\.(csv|json|yaml)`).Draw(t, "filename")
			path := "/some/path/" + filename

			cliErr := FormatFileReadError(path, os.ErrNotExist)

			// Verify error message contains the file path
			if !containsSubstring(cliErr.Message, path) {
				t.Fatalf("error message should contain file path %q, got: %s", path, cliErr.Message)
			}
		})
	})

	t.Run("FileWriteError_ContainsPath", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Generate random file path
			filename := rapid.StringMatching(`[a-z]{5,10}\.(csv|json|yaml)`).Draw(t, "filename")
			path := "/some/path/" + filename

			cliErr := FormatFileWriteError(path, os.ErrPermission)

			// Verify error message contains the file path
			if !containsSubstring(cliErr.Message, path) {
				t.Fatalf("error message should contain file path %q, got: %s", path, cliErr.Message)
			}
		})
	})

	t.Run("UnsupportedFormat_ListsSupportedFormats", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Generate random invalid format
			invalidFormat := rapid.StringMatching(`[a-z]{3,8}_invalid`).Draw(t, "invalidFormat")

			cliErr := FormatUnsupportedFormatError(invalidFormat)

			// Verify error message contains the invalid format
			if !containsSubstring(cliErr.Message, invalidFormat) {
				t.Fatalf("error message should contain invalid format %q, got: %s", invalidFormat, cliErr.Message)
			}

			// Verify error message lists at least some supported formats
			supportedFormats := SupportedFormats()
			foundFormat := false
			for _, f := range supportedFormats {
				if containsSubstring(cliErr.Message, string(f)) {
					foundFormat = true
					break
				}
			}
			if !foundFormat {
				t.Fatalf("error message should list supported formats, got: %s", cliErr.Message)
			}
		})
	})
}

// TestProperty_RealFileErrors tests with actual file system operations
func TestProperty_RealFileErrors(t *testing.T) {
	t.Run("NonexistentFile_ReturnsError", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Generate random filename
			filename := rapid.StringMatching(`[a-z]{5,10}\.csv`).Draw(t, "filename")
			tmpDir := os.TempDir()
			nonexistentPath := filepath.Join(tmpDir, "nonexistent_dir_12345", filename)

			// Try to create input reader for nonexistent file
			_, err := CreateInputReader(nonexistentPath)

			// Should return error
			if err == nil {
				t.Fatal("expected error for nonexistent file")
			}

			// Format the error and verify non-zero exit code
			cliErr := FormatFileReadError(nonexistentPath, err)
			if cliErr.ExitCode == ExitSuccess {
				t.Fatal("expected non-zero exit code for file read error")
			}
		})
	})

	t.Run("InvalidOutputPath_ReturnsError", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Generate random filename
			filename := rapid.StringMatching(`[a-z]{5,10}\.csv`).Draw(t, "filename")
			invalidPath := filepath.Join("/nonexistent_dir_12345", filename)

			// Try to create output writer for invalid path
			_, err := CreateOutputWriter(invalidPath)

			// Should return error
			if err == nil {
				t.Fatal("expected error for invalid output path")
			}

			// Format the error and verify non-zero exit code
			cliErr := FormatFileWriteError(invalidPath, err)
			if cliErr.ExitCode == ExitSuccess {
				t.Fatal("expected non-zero exit code for file write error")
			}
		})
	})
}

// containsSubstring checks if s contains substr
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
