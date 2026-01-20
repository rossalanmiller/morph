# Implementation Plan: Table Converter

## Overview

This implementation plan breaks down the table converter CLI utility into discrete coding tasks. The approach follows a layered architecture: first establishing the core data model and interfaces, then implementing format-specific parsers and serializers, and finally wiring everything together through the CLI. Each task builds incrementally on previous work, with property-based tests integrated throughout to validate correctness.

## Tasks

- [x] 1. Set up Go project structure and core data model
  - Initialize Go module with `go mod init`
  - Create directory structure: `cmd/`, `internal/model/`, `internal/parser/`, `internal/serializer/`, `internal/cli/`
  - Define `TableData` struct with headers and rows
  - Define `Value` type with raw string and parsed value
  - Create basic constructor and validation methods
  - _Requirements: 1.1-1.9, 2.1-2.8, 9.4_

- [x] 1.1 Write property test for TableData normalization
  - **Property 7: Row Normalization**
  - **Validates: Requirements 7.1**

- [x] 2. Implement Parser and Serializer interfaces
  - Define `Parser` interface with `Parse(io.Reader) (*TableData, error)` method
  - Define `Serializer` interface with `Serialize(*TableData, io.Writer) error` method
  - Create error types: `ParseError` and `SerializeError` with context
  - Create format registry for mapping format names to implementations
  - _Requirements: 9.1, 9.2, 9.4_

- [x] 3. Implement CSV parser and serializer
  - [x] 3.1 Implement CSV parser using `encoding/csv`
    - Handle quoted fields, escaped characters
    - Parse headers from first row
    - Normalize row lengths
    - _Requirements: 1.1_
  
  - [x] 3.2 Write property test for CSV round-trip
    - **Property 1: Round-Trip Preservation (CSV)**
    - **Validates: Requirements 1.1, 2.1, 3.1**
  
  - [x] 3.3 Implement CSV serializer
    - Properly escape quotes, commas, newlines
    - Write headers as first row
    - _Requirements: 2.1_
  
  - [x] 3.4 Write property test for CSV special character handling
    - **Property 2: Special Character Preservation (CSV)**
    - **Validates: Requirements 3.2**

- [x] 4. Implement JSON parser and serializer
  - [x] 4.1 Implement JSON parser using `encoding/json`
    - Validate input is array of objects
    - Extract headers from object keys (union of all keys)
    - Handle missing keys as null values
    - Preserve numeric types
    - _Requirements: 1.4, 7.4_
  
  - [x] 4.2 Write property test for JSON round-trip
    - **Property 1: Round-Trip Preservation (JSON)**
    - **Validates: Requirements 1.4, 2.4, 3.1**
  
  - [x] 4.3 Implement JSON serializer
    - Output array of objects
    - Preserve numeric and boolean types
    - Handle null values
    - _Requirements: 2.4_
  
  - [x] 4.4 Write property test for JSON numeric precision
    - **Property 4: Numeric Precision Preservation**
    - **Validates: Requirements 3.4**

- [x] 5. Implement YAML parser and serializer
  - [x] 5.1 Implement YAML parser using `gopkg.in/yaml.v3`
    - Parse list of maps structure
    - Extract headers from map keys
    - Preserve types (numbers, booleans, nulls)
    - _Requirements: 1.3_
  
  - [x] 5.2 Write property test for YAML round-trip
    - **Property 1: Round-Trip Preservation (YAML)**
    - **Validates: Requirements 1.3, 2.3, 3.1**
  
  - [x] 5.3 Implement YAML serializer
    - Output list of maps
    - Preserve types
    - _Requirements: 2.3_

- [x] 6. Implement HTML parser and serializer
  - [x] 6.1 Implement HTML parser using `golang.org/x/net/html`
    - Find first `<table>` element
    - Extract headers from `<thead>` or first `<tr>`
    - Parse rows from `<tbody>` or remaining `<tr>` elements
    - Decode HTML entities
    - _Requirements: 1.5_
  
  - [x] 6.2 Write property test for HTML round-trip
    - **Property 1: Round-Trip Preservation (HTML)**
    - **Validates: Requirements 1.5, 2.5, 3.1**
  
  - [x] 6.3 Implement HTML serializer
    - Generate `<table>` with `<thead>` and `<tbody>`
    - Encode special characters (`<`, `>`, `&`)
    - _Requirements: 2.5_
  
  - [x] 6.4 Write property test for HTML character escaping
    - **Property 8: Character Escaping (HTML)**
    - **Validates: Requirements 7.2**

- [x] 7. Implement XML parser and serializer
  - [x] 7.1 Implement XML parser using `encoding/xml`
    - Parse `<dataset><record>...</record></dataset>` structure
    - Extract headers from child element names in first record
    - Handle missing elements as null
    - Decode XML entities
    - _Requirements: 1.6_
  
  - [x] 7.2 Write property test for XML round-trip
    - **Property 1: Round-Trip Preservation (XML)**
    - **Validates: Requirements 1.6, 2.6, 3.1**
  
  - [x] 7.3 Implement XML serializer
    - Generate `<dataset><record>...</record></dataset>` structure
    - Encode special characters (`<`, `>`, `&`, quotes)
    - Include XML declaration
    - _Requirements: 2.6_
  
  - [x] 7.4 Write property test for XML character escaping
    - **Property 8: Character Escaping (XML)**
    - **Validates: Requirements 7.2**

- [x] 8. Checkpoint - Ensure core format tests pass
  - Run all property tests for CSV, JSON, YAML, HTML, XML
  - Verify round-trip properties hold
  - Ask user if questions arise

- [x] 9. Implement Excel parser and serializer
  - [x] 9.1 Implement Excel parser using `github.com/xuri/excelize/v2`
    - Read first sheet (or specified sheet via option)
    - Parse headers from first row
    - Read all data rows
    - Preserve cell types (string, number, boolean)
    - _Requirements: 1.2, 7.3_
  
  - [x] 9.2 Write property test for Excel round-trip
    - **Property 1: Round-Trip Preservation (Excel)**
    - **Validates: Requirements 1.2, 2.2, 3.1**
  
  - [x] 9.3 Implement Excel serializer
    - Create single-sheet workbook
    - Write headers in first row
    - Write data rows
    - Preserve cell types
    - _Requirements: 2.2_

- [x] 10. Implement Markdown parser and serializer
  - [x] 10.1 Implement Markdown table parser
    - Parse GitHub-flavored markdown tables
    - Extract headers from first row
    - Parse separator row (ignore alignment)
    - Parse data rows
    - Handle escaped pipes
    - _Requirements: 1.7_
  
  - [x] 10.2 Write property test for Markdown round-trip
    - **Property 1: Round-Trip Preservation (Markdown)**
    - **Validates: Requirements 1.7, 2.7, 3.1**
  
  - [x] 10.3 Implement Markdown serializer
    - Generate GitHub-flavored markdown table
    - Calculate column widths for alignment
    - Escape pipe characters in cell values
    - _Requirements: 2.7_

- [x] 11. Implement ASCII table parser and serializer
  - [x] 11.1 Implement ASCII table parser
    - Parse box-drawing ASCII tables
    - Detect column boundaries from separator lines
    - Extract headers from first data row
    - Parse remaining data rows
    - Trim whitespace from cells
    - _Requirements: 1.8_
  
  - [x] 11.2 Write property test for ASCII round-trip
    - **Property 1: Round-Trip Preservation (ASCII)**
    - **Validates: Requirements 1.8, 2.8, 3.1**
  
  - [x] 11.3 Implement ASCII serializer
    - Calculate column widths based on content
    - Generate box-drawing characters
    - Pad cells for alignment
    - _Requirements: 2.8_

- [x] 12. Checkpoint - Ensure all format tests pass
  - Run all property tests for all 8 formats
  - Verify round-trip properties hold for all formats
  - Ask user if questions arise

- [x] 13. Write property tests for empty cells and error handling
  - [x] 13.1 Write property test for empty cell preservation
    - **Property 3: Empty Cell Preservation**
    - **Validates: Requirements 3.3**
  
  - [x] 13.2 Write property test for invalid input error handling
    - **Property 5: Invalid Input Error Handling**
    - **Validates: Requirements 1.9, 6.2, 7.4**

- [x] 14. Implement format detection
  - Create `DetectFormat(filepath string) (string, error)` function
  - Map file extensions to format names (.csv → "csv", .xlsx/.xls → "excel", etc.)
  - Return error for unknown extensions
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7, 5.9_

- [x] 14.1 Write unit tests for format detection
  - Test each file extension mapping
  - Test unknown extension error
  - _Requirements: 5.1-5.7, 5.9_

- [x] 15. Implement CLI argument parsing
  - [x] 15.1 Create CLI package with flag parsing
    - Use `flag` package or `github.com/spf13/cobra`
    - Define flags: `-in`, `-out`, `-h/--help`, `-v/--version`, `--sheet`, `--no-header`
    - Parse positional arguments for input/output file paths
    - Validate flag combinations
    - _Requirements: 4.1, 4.7_
  
  - [x] 15.2 Write unit tests for CLI argument parsing
    - Test valid flag combinations
    - Test invalid flag combinations
    - Test help output
    - _Requirements: 4.1, 4.7, 4.8_

- [x] 16. Implement CLI I/O handling
  - [x] 16.1 Create input reader function
    - If input file specified, open file
    - If no input file, use stdin
    - Detect format from file extension or require `-in` flag
    - Return error if format cannot be determined
    - _Requirements: 4.2, 4.4, 4.6, 5.8_
  
  - [x] 16.2 Create output writer function
    - If output file specified, create file
    - If no output file, use stdout
    - Detect format from file extension or require `-out` flag
    - _Requirements: 4.3, 4.5_
  
  - [x] 16.3 Write unit tests for I/O handling
    - Test file input/output
    - Test stdin/stdout
    - Test format detection requirements
    - _Requirements: 4.2, 4.3, 4.4, 4.5, 4.6, 5.8_

- [x] 17. Implement CLI error handling
  - Create error formatting function for user-friendly messages
  - Handle file read errors with path and reason
  - Handle parse errors with format and location
  - Handle unsupported format errors with list of supported formats
  - Handle file write errors with path and reason
  - Ensure all errors exit with non-zero status code
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

- [x] 17.1 Write property test for CLI error exit codes
  - **Property 6: CLI Error Exit Codes**
  - **Validates: Requirements 6.1, 6.4, 6.5**

- [x] 18. Wire CLI components together
  - [x] 18.1 Create main conversion function
    - Accept input reader, output writer, input format, output format
    - Look up parser from format registry
    - Parse input to TableData
    - Look up serializer from format registry
    - Serialize TableData to output
    - Return errors with context
    - _Requirements: 4.1, 9.1, 9.2_
  
  - [x] 18.2 Create main CLI entry point
    - Parse CLI arguments
    - Set up input reader and output writer
    - Call conversion function
    - Handle errors and exit codes
    - _Requirements: 4.1-4.8_

- [x] 19. Write integration tests for end-to-end CLI workflows
  - Test file-to-file conversion for each format pair
  - Test stdin-to-stdout conversion
  - Test format auto-detection
  - Test error scenarios (missing files, invalid formats, malformed input)
  - _Requirements: 4.1-4.8, 5.1-5.9, 6.1-6.5_

- [x] 20. Write property test for format extensibility
  - **Property 9: Format Extensibility**
  - **Validates: Requirements 9.1, 9.2**
  - Create a mock format implementation
  - Verify it works with existing conversion pipeline

- [x] 21. Final checkpoint and documentation
  - Run all tests (unit, property, integration)
  - Verify all 9 correctness properties pass
  - Create README with usage examples
  - Document supported formats and CLI flags
  - Add build instructions for creating single binary
  - Ask user if questions arise

## Notes

- Each task references specific requirements for traceability
- Property tests validate universal correctness properties with minimum 100 iterations
- Unit tests validate specific examples and edge cases
- Integration tests verify end-to-end CLI workflows
- Go's standard library provides most needed functionality (csv, json, xml, html)
- External dependencies: `gopkg.in/yaml.v3` for YAML, `github.com/xuri/excelize/v2` for Excel
- Property-based testing library: `github.com/leanovate/gopter` or `pgregory.net/rapid`
