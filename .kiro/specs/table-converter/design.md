# Design Document: Table Converter

## Overview

The Table Converter is a CLI utility that transforms structured tabular data between 8 different formats: CSV, Excel, YAML, JSON, HTML, XML, Markdown tables, and ASCII tables. The design follows a pipeline architecture with three main stages: parsing input, transforming to an internal representation, and serializing to output.

The tool prioritizes simplicity, correctness, and ease of use. It supports both file-based and stream-based (stdin/stdout) operations, making it suitable for both interactive use and shell scripting pipelines.

## Architecture

### High-Level Architecture

```
Input → Parser → Internal Table Model → Serializer → Output
         ↑                                    ↑
    Format-specific              Format-specific
```

The architecture consists of:

1. **CLI Layer**: Handles argument parsing, file I/O, and user interaction
2. **Parser Layer**: Format-specific parsers that convert input to the internal model
3. **Internal Table Model**: A unified representation of tabular data
4. **Serializer Layer**: Format-specific serializers that convert from the internal model to output
5. **Format Detection**: Automatic format identification based on file extensions

### Design Principles

- **Separation of Concerns**: Each format has its own parser and serializer
- **Single Internal Representation**: All formats convert to/from the same Table_Data structure
- **Fail Fast**: Validate input early and provide clear error messages
- **Streaming Where Possible**: Use streaming for large files to minimize memory usage
- **Extensibility**: New formats can be added by implementing parser/serializer interfaces

## Components and Interfaces

### Internal Table Model

The core data structure representing tabular data:

```
Table_Data:
  headers: List<String>      // Column names
  rows: List<List<Value>>    // Data rows
  
Value:
  type: ValueType            // string, number, boolean, null
  raw: String                // Original string representation
  parsed: Any                // Typed value (string, number, boolean, null)
```

Design decisions:
- Headers are always present (even if empty strings for headerless formats)
- All rows must have the same number of columns (padded with null if needed)
- Values preserve both raw string and parsed type for format-specific handling
- Empty cells are represented as null values

### Known Limitations

**Duplicate Column Names**: Formats like CSV, Excel, and HTML allow duplicate column names, but map-based formats (JSON, YAML) do not support duplicate keys. When converting from a format with duplicate headers to JSON/YAML, only the last value for each duplicate column name is preserved. This is a known data loss scenario - users should ensure unique column names when targeting these formats.

### Parser Interface

```
interface Parser:
  parse(input: Stream) -> Result<Table_Data, ParseError>
  
ParseError:
  message: String
  line: Optional<Integer>
  column: Optional<Integer>
  context: String
```

Each format implements this interface:
- **CSVParser**: Uses standard CSV parsing library
- **ExcelParser**: Uses Excel library (e.g., openpyxl, xlsx)
- **YAMLParser**: Expects list of dictionaries
- **JSONParser**: Expects array of objects
- **HTMLParser**: Extracts first `<table>` element
- **XMLParser**: Expects `<dataset><record>...</record></dataset>` structure
- **MarkdownParser**: Parses GitHub-flavored markdown tables
- **ASCIIParser**: Parses ASCII box-drawing tables

### Serializer Interface

```
interface Serializer:
  serialize(data: Table_Data, output: Stream) -> Result<Void, SerializeError>
  
SerializeError:
  message: String
  context: String
```

Each format implements this interface:
- **CSVSerializer**: Standard CSV with proper escaping
- **ExcelSerializer**: Creates single-sheet workbook
- **YAMLSerializer**: Outputs list of dictionaries
- **JSONSerializer**: Outputs array of objects
- **HTMLSerializer**: Generates `<table>` with `<thead>` and `<tbody>`
- **XMLSerializer**: Generates `<dataset><record>...</record></dataset>`
- **MarkdownSerializer**: Generates GitHub-flavored markdown table
- **ASCIISerializer**: Generates box-drawing ASCII table

### CLI Interface

```
Command-line syntax:
  tableconv [OPTIONS] [INPUT_FILE] [OUTPUT_FILE]
  
Options:
  -in <format>      Input format (csv|excel|yaml|json|html|xml|markdown|ascii)
  -out <format>     Output format (csv|excel|yaml|json|html|xml|markdown|ascii)
  -h, --help        Show help message
  -v, --version     Show version
  --sheet <name>    Excel sheet name (default: first sheet)
  --no-header       Treat first row as data, not headers
  
Examples:
  tableconv -in json -out yaml < input.json > output.yaml
  tableconv data.csv output.xlsx
  echo '[{"a":1}]' | tableconv -in json -out csv
```

### Format Detection

```
detect_format(filepath: String) -> Optional<Format>:
  extension = get_extension(filepath)
  return EXTENSION_MAP.get(extension)
  
EXTENSION_MAP:
  .csv -> CSV
  .xlsx, .xls -> Excel
  .yaml, .yml -> YAML
  .json -> JSON
  .html, .htm -> HTML
  .xml -> XML
  .md -> Markdown
```

## Data Models

### Format-Specific Structures

#### CSV Format
- Standard RFC 4180 CSV
- Comma-separated, quoted strings for values containing commas/newlines
- First row is headers (unless --no-header flag)

#### Excel Format
- Single worksheet (first sheet by default)
- First row is headers
- Preserves basic cell types (string, number, boolean)

#### YAML Format
```yaml
- c1: a
  c2: 1
- c1: b
  c2: 2
```

#### JSON Format
```json
[
  {"c1": "a", "c2": "1"},
  {"c1": "b", "c2": "2"}
]
```

#### HTML Format
```html
<table>
  <thead>
    <tr><th>c1</th><th>c2</th></tr>
  </thead>
  <tbody>
    <tr><td>a</td><td>1</td></tr>
    <tr><td>b</td><td>2</td></tr>
  </tbody>
</table>
```

#### XML Format
```xml
<?xml version="1.0" encoding="UTF-8"?>
<dataset>
  <record><c1>a</c1><c2>1</c2></record>
  <record><c1>b</c1><c2>2</c2></record>
</dataset>
```

#### Markdown Format
```markdown
| c1 | c2 |
|----|-----|
| a  | 1   |
| b  | 2   |
```

#### ASCII Format
```
+----+----+
| c1 | c2 |
+----+----+
| a  | 1  |
| b  | 2  |
+----+----+
```

### Type Handling

Value type inference rules:
1. Try parsing as number (integer or float)
2. Try parsing as boolean (true/false, yes/no, 1/0)
3. Check for null/empty
4. Default to string

Type preservation across formats:
- Numbers remain numbers where format supports it (JSON, YAML, Excel)
- Booleans remain booleans where format supports it (JSON, YAML, Excel)
- All values become strings in text formats (CSV, HTML, XML, Markdown, ASCII)


## Correctness Properties

A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.

### Property 1: Round-Trip Preservation

*For any* valid Table_Data and any supported format, serializing the data to that format and then parsing it back should produce equivalent Table_Data (same headers, same number of rows, same values).

**Validates: Requirements 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 1.8, 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 2.8, 3.1**

This is the most critical property. It ensures that:
- Parsers correctly read format-specific input
- Serializers correctly write format-specific output
- No data is lost or corrupted during conversion
- The internal Table_Data model is sufficient to represent all formats

### Property 2: Special Character Preservation

*For any* Table_Data containing special characters (quotes, commas, newlines, unicode, HTML entities, XML special chars), converting through any format should preserve these characters correctly.

**Validates: Requirements 3.2**

Special characters to test:
- CSV: quotes, commas, newlines
- HTML/XML: `<`, `>`, `&`, quotes
- All formats: unicode characters, emojis, control characters

### Property 3: Empty Cell Preservation

*For any* Table_Data containing empty/null cells, converting through any format should preserve the empty cells (they should remain empty, not become empty strings or be omitted).

**Validates: Requirements 3.3**

### Property 4: Numeric Precision Preservation

*For any* Table_Data containing numeric values (integers, floats, scientific notation), converting through formats that support numeric types (JSON, YAML, Excel) should preserve the numeric value and precision.

**Validates: Requirements 3.4**

### Property 5: Invalid Input Error Handling

*For any* malformed input in a specific format, the parser should return a descriptive error (not crash) that indicates the format and nature of the problem.

**Validates: Requirements 1.9, 6.2, 7.4**

Test cases:
- Malformed CSV (unclosed quotes, inconsistent columns)
- Invalid JSON (syntax errors, non-array structure)
- Invalid YAML (syntax errors, non-list structure)
- Invalid HTML (no table element, malformed tags)
- Invalid XML (malformed tags, wrong structure)
- Invalid Markdown (malformed table syntax)
- Invalid Excel (corrupted file)

### Property 6: CLI Error Exit Codes

*For any* error condition (file not found, parse error, invalid format, write error), the CLI should exit with a non-zero status code and display an error message.

**Validates: Requirements 6.1, 6.4, 6.5**

### Property 7: Row Normalization

*For any* input data with inconsistent column counts across rows, the parser should normalize all rows to have the same number of columns (padding with null values as needed).

**Validates: Requirements 7.1**

### Property 8: Character Escaping

*For any* Table_Data containing characters that are invalid or special in the target format, the serializer should properly escape or encode them so the output is valid in that format.

**Validates: Requirements 7.2**

Examples:
- CSV: escape quotes and commas
- HTML: encode `<`, `>`, `&`
- XML: encode `<`, `>`, `&`, quotes
- JSON: escape quotes, backslashes, control characters

### Property 9: Format Extensibility

*For any* new format added by implementing the Parser and Serializer interfaces, the format should automatically work with all existing formats for conversions (new format can convert to/from any existing format).

**Validates: Requirements 9.1, 9.2**

This property ensures that:
- The internal Table_Data model is sufficient for any tabular format
- New formats integrate seamlessly without modifying existing code
- The conversion pipeline is truly format-agnostic

## Error Handling

### Error Categories

1. **Input Errors**
   - File not found or not readable
   - Invalid file format
   - Parse errors (malformed input)
   - Unsupported format specified

2. **Output Errors**
   - Output file not writable
   - Disk full
   - Permission denied

3. **Data Errors**
   - Inconsistent row lengths
   - Invalid data types for target format
   - Data too large for format constraints

### Error Response Strategy

All errors should:
- Display a clear, actionable error message
- Include context (file path, line number, format)
- Exit with non-zero status code
- Not expose internal stack traces to users (log them for debugging)

Example error messages:
```
Error: Failed to parse CSV input
  Line 5: Unclosed quote in field
  Context: "John Doe","123 Main St

Error: File not found: data.csv

Error: Unsupported format 'txt'
  Supported formats: csv, excel, yaml, json, html, xml, markdown, ascii

Error: Invalid JSON structure
  Expected: array of objects
  Got: object
```

## Testing Strategy

### Dual Testing Approach

The testing strategy uses both unit tests and property-based tests:

- **Unit tests**: Verify specific examples, edge cases, and error conditions
- **Property tests**: Verify universal properties across all inputs

Both are complementary and necessary for comprehensive coverage. Unit tests catch concrete bugs with specific examples, while property tests verify general correctness across many generated inputs.

### Property-Based Testing

We will use a property-based testing library appropriate for the implementation language (e.g., Hypothesis for Python, fast-check for TypeScript, QuickCheck for Haskell).

**Configuration**:
- Minimum 100 iterations per property test
- Each property test must reference its design document property
- Tag format: **Feature: table-converter, Property {number}: {property_text}**

**Test Data Generation**:
- Generate random Table_Data with varying:
  - Number of columns (1-20)
  - Number of rows (0-100)
  - Column names (valid identifiers, special characters)
  - Cell values (strings, numbers, booleans, nulls, empty strings)
  - Special characters (quotes, commas, newlines, unicode)
- Generate format-specific invalid inputs for error testing

### Unit Testing

Unit tests should focus on:
- Specific format examples (one test per format showing basic conversion)
- Edge cases (empty tables, single cell, very wide tables, very long tables)
- Error conditions (specific malformed inputs)
- CLI argument parsing
- Format detection logic
- File I/O operations

### Integration Testing

Integration tests should verify:
- End-to-end CLI workflows (file input → conversion → file output)
- Stdin/stdout piping
- Format auto-detection with real files
- Error message formatting and exit codes

### Test Organization

```
tests/
  unit/
    test_parsers.py          # Unit tests for each parser
    test_serializers.py      # Unit tests for each serializer
    test_table_model.py      # Unit tests for Table_Data
    test_cli.py              # Unit tests for CLI argument parsing
    test_format_detection.py # Unit tests for format detection
  
  property/
    test_roundtrip.py        # Property 1: Round-trip preservation
    test_special_chars.py    # Property 2: Special character preservation
    test_empty_cells.py      # Property 3: Empty cell preservation
    test_numeric.py          # Property 4: Numeric precision
    test_errors.py           # Property 5: Error handling
    test_cli_errors.py       # Property 6: CLI error codes
    test_normalization.py    # Property 7: Row normalization
    test_escaping.py         # Property 8: Character escaping
  
  integration/
    test_cli_workflows.py    # End-to-end CLI tests
    test_piping.py           # Stdin/stdout piping tests
```

### Coverage Goals

- Line coverage: >90%
- Branch coverage: >85%
- All 8 correctness properties must have passing property tests
- All error paths must be tested
