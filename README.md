# Morph - Table Format Converter

A command-line utility that converts structured tabular data between multiple formats including CSV, Excel, YAML, JSON, HTML tables, XML, Markdown tables, and ASCII tables.

## Features

- Convert between 8 different table formats
- Automatic format detection from file extensions
- Support for stdin/stdout piping
- Preserves data types (numbers, booleans, nulls) where supported
- Handles special characters and escaping correctly
- Single binary with no runtime dependencies

## Supported Formats

| Format   | Extensions      | Description                          |
|----------|-----------------|--------------------------------------|
| CSV      | `.csv`          | Comma-separated values (RFC 4180)    |
| Excel    | `.xlsx`, `.xls` | Microsoft Excel spreadsheet          |
| JSON     | `.json`         | Array of objects                     |
| YAML     | `.yaml`, `.yml` | List of dictionaries                 |
| HTML     | `.html`, `.htm` | HTML table element                   |
| XML      | `.xml`          | Dataset/record structure             |
| Markdown | `.md`           | GitHub-flavored markdown table       |
| ASCII    | `.txt`          | ASCII tables (auto-detects: box, psql, markdown, org-mode, rst) |

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/user/morph.git
cd morph

# Build the binary
go build -o morph ./cmd/morph

# Optional: Install to your PATH
go install ./cmd/morph
```

### Cross-Compilation

Build for different platforms:

```bash
# Windows (64-bit)
GOOS=windows GOARCH=amd64 go build -o morph.exe ./cmd/morph

# Linux (64-bit)
GOOS=linux GOARCH=amd64 go build -o morph-linux ./cmd/morph

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o morph-darwin-arm64 ./cmd/morph

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o morph-darwin-amd64 ./cmd/morph
```

### Build Requirements

- Go 1.21 or later

## Usage

### Basic Syntax

```bash
morph [OPTIONS] [INPUT_FILE] [OUTPUT_FILE]
```

### Options

| Flag           | Description                                      |
|----------------|--------------------------------------------------|
| `-in <format>` | Input format (csv, excel, yaml, json, html, xml, markdown, ascii) |
| `-out <format>`| Output format (csv, excel, yaml, json, html, xml, markdown, ascii) |
| `-f <style>`   | ASCII table style (md, psql, box, org, rst-grid, rst-simple) |
| `-h`, `--help` | Show help message                                |
| `-v`, `--version` | Show version                                  |

### Format Aliases

For convenience, you can use shorthand aliases for formats:

| Format   | Aliases              |
|----------|----------------------|
| excel    | `xlsx`, `xls`, `xl`  |
| yaml     | `yml`                |
| json     | `js`                 |
| html     | `htm`                |
| markdown | `md`                 |
| ascii    | `txt`, `table`       |

### Examples

#### File to File Conversion

```bash
# Convert CSV to JSON (formats auto-detected from extensions)
morph data.csv output.json

# Convert Excel to YAML
morph spreadsheet.xlsx output.yaml

# Convert JSON to Markdown table
morph data.json table.md
```

#### Explicit Format Specification

```bash
# Specify formats explicitly
morph -in csv -out json data.csv output.json

# Useful when extensions don't match the format
morph -in json -out yaml data.txt output.txt
```

#### Using stdin/stdout

```bash
# Pipe JSON to CSV
echo '[{"name":"Alice","age":30}]' | morph -in json -out csv

# Read from stdin, write to file
cat data.json | morph -in json output.csv

# Read from file, write to stdout
morph -out yaml data.csv

# Full pipeline
curl -s https://api.example.com/data | morph -in json -out csv > data.csv
```

#### ASCII Table Styles

The ASCII format supports multiple visual styles via the `-f` flag:

```bash
# Markdown style
morph data.csv -out ascii -f md

# PostgreSQL psql style
morph data.csv -out ascii -f psql

# Traditional box style (default)
morph data.csv -out ascii -f box

# Emacs org-mode style
morph data.csv -out ascii -f org

# reStructuredText grid style
morph data.csv -out ascii -f rst-grid

# reStructuredText simple style
morph data.csv -out ascii -f rst-simple
```

#### Converting PostgreSQL Query Results

```bash
# Copy psql output and convert to JSON
psql -d mydb -c "SELECT * FROM users" | morph -in ascii -out json > users.json

# Convert psql output to CSV
psql -d mydb -c "SELECT * FROM products" | morph -in ascii -out csv > products.csv

# Or save psql output to a file first
psql -d mydb -c "SELECT * FROM orders" > orders.txt
morph orders.txt orders.xlsx
```

#### Working with Excel

```bash
# Convert Excel to CSV
morph report.xlsx report.csv

# Convert CSV to Excel
morph data.csv output.xlsx
```

### Format Examples

#### CSV
```csv
name,age,active
Alice,30,true
Bob,25,false
```

#### JSON
```json
[
  {"name": "Alice", "age": 30, "active": true},
  {"name": "Bob", "age": 25, "active": false}
]
```

#### YAML
```yaml
- name: Alice
  age: 30
  active: true
- name: Bob
  age: 25
  active: false
```

#### HTML
```html
<table>
  <thead>
    <tr><th>name</th><th>age</th><th>active</th></tr>
  </thead>
  <tbody>
    <tr><td>Alice</td><td>30</td><td>true</td></tr>
    <tr><td>Bob</td><td>25</td><td>false</td></tr>
  </tbody>
</table>
```

#### XML
```xml
<?xml version="1.0" encoding="UTF-8"?>
<dataset>
  <record><name>Alice</name><age>30</age><active>true</active></record>
  <record><name>Bob</name><age>25</age><active>false</active></record>
</dataset>
```

#### Markdown
```markdown
| name  | age | active |
|-------|-----|--------|
| Alice | 30  | true   |
| Bob   | 25  | false  |
```

#### ASCII

The ASCII format auto-detects and supports multiple table styles:

**Traditional Box (default):**
```
+-------+-----+--------+
| name  | age | active |
+-------+-----+--------+
| Alice | 30  | true   |
| Bob   | 25  | false  |
+-------+-----+--------+
```

**PostgreSQL psql format:**
```
name  | age | active
------+-----+--------
Alice | 30  | true
Bob   | 25  | false
```

**Markdown:**
```
| name  | age | active |
|-------|-----|--------|
| Alice | 30  | true   |
| Bob   | 25  | false  |
```

**Emacs org-mode:**
```
| name  | age | active |
|-------+-----+--------|
| Alice | 30  | true   |
| Bob   | 25  | false  |
```

**reStructuredText Grid:**
```
+-------+-----+--------+
| name  | age | active |
+=======+=====+========+
| Alice | 30  | true   |
| Bob   | 25  | false  |
+-------+-----+--------+
```

**reStructuredText Simple:**
```
=====  ===  ======
name   age  active
=====  ===  ======
Alice  30   true
Bob    25   false
=====  ===  ======
```

The parser automatically detects which format is being used. Use the `-f` flag to specify the output style.

## Error Handling

Morph provides clear error messages for common issues:

```bash
# File not found
$ morph nonexistent.csv output.json
Error: Failed to read file: nonexistent.csv
  Reason: no such file or directory

# Unsupported format
$ morph -in txt data.txt output.json
Error: Unsupported format 'txt'
  Supported formats: csv, excel, yaml, json, html, xml, markdown, ascii

# Parse error
$ morph malformed.json output.csv
Error: Failed to parse JSON input
  Line 3: unexpected end of JSON input
```

## Known Limitations

- **Duplicate Column Names**: When converting from formats that allow duplicate column names (CSV, Excel, HTML) to map-based formats (JSON, YAML), only the last value for each duplicate column name is preserved.
- **Excel**: Only the first sheet is processed.
- **Large Files**: Files over 100MB may take longer to process. Consider using streaming-friendly formats like CSV for very large datasets.

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/parser/...

# Run property-based tests only
go test ./... -run "Property"
```

### Project Structure

```
morph/
├── cmd/
│   └── morph/           # CLI entry point
├── internal/
│   ├── cli/             # CLI argument parsing and I/O
│   ├── model/           # TableData internal representation
│   ├── parser/          # Format-specific parsers
│   ├── serializer/      # Format-specific serializers
│   └── registry/        # Format registry
├── go.mod
└── README.md
```

### Adding a New Format

1. Implement the `Parser` interface in `internal/parser/`
2. Implement the `Serializer` interface in `internal/serializer/`
3. Register the format in `internal/registry/registry.go`
4. Add file extension mapping in `internal/cli/format.go`

## License

MIT License - see LICENSE file for details.
