package model

import (
	"fmt"
	"strconv"
	"strings"
)

// ValueType represents the type of a value in a table cell
type ValueType int

const (
	TypeString ValueType = iota
	TypeNumber
	TypeBoolean
	TypeNull
)

// Value represents a single cell value with both raw and parsed representations
type Value struct {
	Type   ValueType
	Raw    string
	Parsed interface{} // string, float64, bool, or nil
}

// NewValue creates a new Value by inferring the type from the raw string
func NewValue(raw string) Value {
	// Check for null/empty
	if raw == "" {
		return Value{
			Type:   TypeNull,
			Raw:    raw,
			Parsed: nil,
		}
	}

	trimmed := strings.TrimSpace(raw)
	
	// Try parsing as boolean
	lower := strings.ToLower(trimmed)
	if lower == "true" || lower == "yes" || lower == "1" {
		return Value{
			Type:   TypeBoolean,
			Raw:    raw,
			Parsed: true,
		}
	}
	if lower == "false" || lower == "no" || lower == "0" {
		return Value{
			Type:   TypeBoolean,
			Raw:    raw,
			Parsed: false,
		}
	}

	// Try parsing as number
	if num, err := strconv.ParseFloat(trimmed, 64); err == nil {
		return Value{
			Type:   TypeNumber,
			Raw:    raw,
			Parsed: num,
		}
	}

	// Default to string
	return Value{
		Type:   TypeString,
		Raw:    raw,
		Parsed: raw,
	}
}

// NewStringValue creates a Value with explicit string type
func NewStringValue(s string) Value {
	return Value{
		Type:   TypeString,
		Raw:    s,
		Parsed: s,
	}
}

// NewNumberValue creates a Value with explicit number type
func NewNumberValue(n float64) Value {
	raw := strconv.FormatFloat(n, 'f', -1, 64)
	return Value{
		Type:   TypeNumber,
		Raw:    raw,
		Parsed: n,
	}
}

// NewBooleanValue creates a Value with explicit boolean type
func NewBooleanValue(b bool) Value {
	raw := "false"
	if b {
		raw = "true"
	}
	return Value{
		Type:   TypeBoolean,
		Raw:    raw,
		Parsed: b,
	}
}

// NewNullValue creates a Value representing null/empty
func NewNullValue() Value {
	return Value{
		Type:   TypeNull,
		Raw:    "",
		Parsed: nil,
	}
}

// String returns the string representation of the value
func (v Value) String() string {
	return v.Raw
}

// TableData represents structured tabular data with headers and rows
type TableData struct {
	Headers []string
	Rows    [][]Value
}

// NewTableData creates a new TableData with the given headers and rows
// It normalizes all rows to have the same number of columns as headers
func NewTableData(headers []string, rows [][]Value) *TableData {
	td := &TableData{
		Headers: headers,
		Rows:    make([][]Value, len(rows)),
	}

	// Normalize rows to match header count
	numCols := len(headers)
	for i, row := range rows {
		td.Rows[i] = normalizeRow(row, numCols)
	}

	return td
}

// normalizeRow ensures a row has exactly numCols columns
// If row is shorter, it pads with null values
// If row is longer, it truncates
func normalizeRow(row []Value, numCols int) []Value {
	if len(row) == numCols {
		return row
	}

	normalized := make([]Value, numCols)
	
	// Copy existing values
	copyLen := len(row)
	if copyLen > numCols {
		copyLen = numCols
	}
	copy(normalized, row[:copyLen])

	// Pad with null values if needed
	for i := copyLen; i < numCols; i++ {
		normalized[i] = NewNullValue()
	}

	return normalized
}

// Validate checks if the TableData is valid
func (td *TableData) Validate() error {
	if td == nil {
		return fmt.Errorf("TableData is nil")
	}

	numCols := len(td.Headers)
	
	// Check that all rows have the correct number of columns
	for i, row := range td.Rows {
		if len(row) != numCols {
			return fmt.Errorf("row %d has %d columns, expected %d", i, len(row), numCols)
		}
	}

	return nil
}

// NumRows returns the number of rows in the table
func (td *TableData) NumRows() int {
	if td == nil {
		return 0
	}
	return len(td.Rows)
}

// NumCols returns the number of columns in the table
func (td *TableData) NumCols() int {
	if td == nil {
		return 0
	}
	return len(td.Headers)
}

// IsEmpty returns true if the table has no rows
func (td *TableData) IsEmpty() bool {
	return td == nil || len(td.Rows) == 0
}
