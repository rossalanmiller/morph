package registry

import (
	"io"
	"testing"

	"github.com/user/table-converter/internal/model"
)

// Mock parser for testing
type mockParser struct{}

func (m *mockParser) Parse(input io.Reader) (*model.TableData, error) {
	return nil, nil
}

// Mock serializer for testing
type mockSerializer struct{}

func (m *mockSerializer) Serialize(data *model.TableData, output io.Writer) error {
	return nil
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()
	p := &mockParser{}
	s := &mockSerializer{}

	// Test successful registration
	err := r.Register("test", p, s)
	if err != nil {
		t.Errorf("Register() error = %v, want nil", err)
	}

	// Test duplicate registration
	err = r.Register("test", p, s)
	if err == nil {
		t.Error("Register() duplicate should return error, got nil")
	}

	// Test empty format name
	err = r.Register("", p, s)
	if err == nil {
		t.Error("Register() with empty name should return error, got nil")
	}
}

func TestRegistry_GetParser(t *testing.T) {
	r := NewRegistry()
	p := &mockParser{}
	s := &mockSerializer{}

	r.Register("test", p, s)

	// Test getting existing parser
	retrieved, err := r.GetParser("test")
	if err != nil {
		t.Errorf("GetParser() error = %v, want nil", err)
	}
	if retrieved != p {
		t.Error("GetParser() returned wrong parser")
	}

	// Test case insensitivity
	retrieved, err = r.GetParser("TEST")
	if err != nil {
		t.Errorf("GetParser() with uppercase error = %v, want nil", err)
	}
	if retrieved != p {
		t.Error("GetParser() should be case insensitive")
	}

	// Test getting non-existent parser
	_, err = r.GetParser("nonexistent")
	if err == nil {
		t.Error("GetParser() for non-existent format should return error, got nil")
	}
}

func TestRegistry_GetSerializer(t *testing.T) {
	r := NewRegistry()
	p := &mockParser{}
	s := &mockSerializer{}

	r.Register("test", p, s)

	// Test getting existing serializer
	retrieved, err := r.GetSerializer("test")
	if err != nil {
		t.Errorf("GetSerializer() error = %v, want nil", err)
	}
	if retrieved != s {
		t.Error("GetSerializer() returned wrong serializer")
	}

	// Test case insensitivity
	retrieved, err = r.GetSerializer("TEST")
	if err != nil {
		t.Errorf("GetSerializer() with uppercase error = %v, want nil", err)
	}
	if retrieved != s {
		t.Error("GetSerializer() should be case insensitive")
	}

	// Test getting non-existent serializer
	_, err = r.GetSerializer("nonexistent")
	if err == nil {
		t.Error("GetSerializer() for non-existent format should return error, got nil")
	}
}

func TestRegistry_IsSupported(t *testing.T) {
	r := NewRegistry()
	p := &mockParser{}
	s := &mockSerializer{}

	r.Register("test", p, s)

	// Test supported format
	if !r.IsSupported("test") {
		t.Error("IsSupported() = false, want true for registered format")
	}

	// Test case insensitivity
	if !r.IsSupported("TEST") {
		t.Error("IsSupported() should be case insensitive")
	}

	// Test unsupported format
	if r.IsSupported("nonexistent") {
		t.Error("IsSupported() = true, want false for non-existent format")
	}
}

func TestRegistry_SupportedFormats(t *testing.T) {
	r := NewRegistry()
	p := &mockParser{}
	s := &mockSerializer{}

	// Empty registry
	formats := r.SupportedFormats()
	if len(formats) != 0 {
		t.Errorf("SupportedFormats() = %v, want empty slice", formats)
	}

	// Register some formats
	r.Register("csv", p, s)
	r.Register("json", p, s)

	formats = r.SupportedFormats()
	if len(formats) != 2 {
		t.Errorf("SupportedFormats() length = %d, want 2", len(formats))
	}

	// Check that both formats are present
	found := make(map[Format]bool)
	for _, f := range formats {
		found[f] = true
	}
	if !found["csv"] || !found["json"] {
		t.Errorf("SupportedFormats() = %v, want [csv, json]", formats)
	}
}

func TestRegistry_GetFormat(t *testing.T) {
	r := NewRegistry()
	p := &mockParser{}
	s := &mockSerializer{}

	r.Register("test", p, s)

	// Test getting existing format
	info, err := r.GetFormat("test")
	if err != nil {
		t.Errorf("GetFormat() error = %v, want nil", err)
	}
	if info.Name != "test" {
		t.Errorf("GetFormat() Name = %q, want %q", info.Name, "test")
	}
	if info.Parser != p {
		t.Error("GetFormat() returned wrong parser")
	}
	if info.Serializer != s {
		t.Error("GetFormat() returned wrong serializer")
	}

	// Test getting non-existent format
	_, err = r.GetFormat("nonexistent")
	if err == nil {
		t.Error("GetFormat() for non-existent format should return error, got nil")
	}
}

func TestGlobalRegistry(t *testing.T) {
	// Note: This test uses the global registry, so it may affect other tests
	// In a real scenario, you might want to reset the global registry between tests
	
	p := &mockParser{}
	s := &mockSerializer{}

	// Test global Register
	err := Register("globaltest", p, s)
	if err != nil {
		t.Errorf("Register() error = %v, want nil", err)
	}

	// Test global IsSupported
	if !IsSupported("globaltest") {
		t.Error("IsSupported() = false, want true")
	}

	// Test global GetParser
	_, err = GetParser("globaltest")
	if err != nil {
		t.Errorf("GetParser() error = %v, want nil", err)
	}

	// Test global GetSerializer
	_, err = GetSerializer("globaltest")
	if err != nil {
		t.Errorf("GetSerializer() error = %v, want nil", err)
	}
}
