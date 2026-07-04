# CSVForge

A comprehensive CSV processing toolkit in Go. Fast, flexible, and easy to use.

## Features

- **18 commands**: head, tail, sort, filter, sample, dedup, join, aggregate, merge, split, convert, validate, stats, info, columns, select, where, trim
- **Auto-detection**: Delimiters, encodings, headers
- **Multiple output formats**: table, csv, json, jsonl, markdown, tsv
- **SQL-like operations**: WHERE clauses, JOINs, GROUP BY aggregations
- **Streaming**: Efficient processing for large files
- **Zero dependencies**: Single binary, no runtime dependencies

## Quick Start

### Install

```bash
go install github.com/EdgarOrtegaRamirez/csvforge/cmd/csvforge@latest
```

### Build from source

```bash
git clone https://github.com/EdgarOrtegaRamirez/csvforge
cd csvforge
go build -o csvforge ./cmd/csvforge/
```

## Usage

### Basic Operations

```bash
# Show first 5 rows
csvforge head -n 5 -i data.csv

# Sort by column
csvforge sort age -i data.csv

# Filter rows
csvforge filter -c city -o eq -v "NYC" -i data.csv

# Show file info
csvforge info -i data.csv
```

### SQL-like Queries

```bash
# WHERE clause
csvforge where -c "age > 25" -i data.csv

# JOIN two files
csvforge join id -f other.csv -i data.csv

# Aggregate with GROUP BY
csvforge aggregate -g city -a "age:count:total" -a "age:avg:avg_age" -i data.csv
```

### Data Transformation

```bash
# Convert to JSON
csvforge convert --to json -i data.csv

# Convert to Markdown table
csvforge convert --to markdown -i data.csv

# Remove duplicates
csvforge dedup -c name -i data.csv

# Select columns
csvforge select -c name,age -i data.csv
```

### Data Quality

```bash
# Validate CSV
csvforge validate -i data.csv

# Show statistics
csvforge stats -i data.csv

# Trim whitespace
csvforge trim -i data.csv
```

## Commands

| Command | Description |
|---------|-------------|
| `head` | Show first N rows |
| `tail` | Show last N rows |
| `sort` | Sort by column |
| `filter` | Filter rows by condition |
| `sample` | Sample N random rows |
| `dedup` | Remove duplicate rows |
| `join` | Join with another CSV file |
| `aggregate` | Aggregate data by group |
| `merge` | Merge multiple CSV files |
| `split` | Split CSV into chunks |
| `convert` | Convert CSV to another format |
| `validate` | Validate CSV data |
| `stats` | Show column statistics |
| `info` | Show CSV file information |
| `columns` | List column names |
| `select` | Select specific columns |
| `where` | Filter rows with SQL-like WHERE clause |
| `trim` | Trim whitespace from all fields |

## Global Flags

| Flag | Description |
|------|-------------|
| `-i, --input` | Input file (default: stdin) |
| `-o, --output` | Output file (default: stdout) |
| `-d, --delimiter` | CSV delimiter (auto-detect if not set) |
| `-H, --header` | CSV has header row (default: true) |
| `-f, --format` | Output format (table, csv, json, jsonl, markdown, tsv) |

## Examples

### Analyze a dataset

```bash
# Quick overview
csvforge info -i sales.csv

# Column statistics
csvforge stats -i sales.csv

# Top 10 by revenue
csvforge sort revenue -r -i sales.csv | csvforge head -n 10
```

### Data cleaning

```bash
# Trim whitespace, remove duplicates, validate
csvforge trim -i messy.csv | csvforge dedup | csvforge validate
```

### Report generation

```bash
# Monthly sales summary
csvforge aggregate -g month -a "revenue:sum:total" -a "revenue:count:orders" -i sales.csv --to markdown
```

## Architecture

```
csvforge/
├── cmd/csvforge/     # CLI entry point
├── internal/
│   ├── csv/          # Core CSV processing engine
│   └── output/       # Output formatting (table, JSON, Markdown, etc.)
└── tests/            # Integration tests
```

## License

MIT
