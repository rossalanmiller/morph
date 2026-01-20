package cli

import (
	"testing"
)

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name     string
		filepath string
		want     Format
		wantErr  bool
	}{
		// CSV
		{"csv lowercase", "data.csv", FormatCSV, false},
		{"csv uppercase", "DATA.CSV", FormatCSV, false},
		{"csv with path", "/path/to/file.csv", FormatCSV, false},

		// Excel
		{"xlsx", "data.xlsx", FormatExcel, false},
		{"xls", "data.xls", FormatExcel, false},
		{"xlsx uppercase", "DATA.XLSX", FormatExcel, false},

		// YAML
		{"yaml", "config.yaml", FormatYAML, false},
		{"yml", "config.yml", FormatYAML, false},

		// JSON
		{"json", "data.json", FormatJSON, false},

		// HTML
		{"html", "page.html", FormatHTML, false},
		{"htm", "page.htm", FormatHTML, false},

		// XML
		{"xml", "data.xml", FormatXML, false},

		// Markdown
		{"md", "table.md", FormatMarkdown, false},

		// ASCII (txt)
		{"txt", "table.txt", FormatASCII, false},

		// Error cases
		{"unknown extension", "data.xyz", "", true},
		{"no extension", "datafile", "", true},
		{"empty string", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectFormat(tt.filepath)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DetectFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidFormat(t *testing.T) {
	tests := []struct {
		format string
		want   bool
	}{
		{"csv", true},
		{"CSV", true},
		{"excel", true},
		{"yaml", true},
		{"json", true},
		{"html", true},
		{"xml", true},
		{"markdown", true},
		{"ascii", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			if got := IsValidFormat(tt.format); got != tt.want {
				t.Errorf("IsValidFormat(%q) = %v, want %v", tt.format, got, tt.want)
			}
		})
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		format  string
		want    Format
		wantErr bool
	}{
		{"csv", FormatCSV, false},
		{"CSV", FormatCSV, false},
		{"excel", FormatExcel, false},
		{"yaml", FormatYAML, false},
		{"json", FormatJSON, false},
		{"html", FormatHTML, false},
		{"xml", FormatXML, false},
		{"markdown", FormatMarkdown, false},
		{"ascii", FormatASCII, false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			got, err := ParseFormat(tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSupportedFormats(t *testing.T) {
	formats := SupportedFormats()

	// Should have 8 formats
	if len(formats) != 8 {
		t.Errorf("SupportedFormats() returned %d formats, want 8", len(formats))
	}

	// Check all expected formats are present
	expected := map[Format]bool{
		FormatCSV:      true,
		FormatExcel:    true,
		FormatYAML:     true,
		FormatJSON:     true,
		FormatHTML:     true,
		FormatXML:      true,
		FormatMarkdown: true,
		FormatASCII:    true,
	}

	for _, f := range formats {
		if !expected[f] {
			t.Errorf("Unexpected format in SupportedFormats(): %v", f)
		}
		delete(expected, f)
	}

	if len(expected) > 0 {
		t.Errorf("Missing formats in SupportedFormats(): %v", expected)
	}
}
