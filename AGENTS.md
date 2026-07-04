# CSVForge - AI Agent Guide

## Overview

CSVForge is a comprehensive CSV processing toolkit in Go with 18 CLI commands for sorting, filtering, joining, aggregating, and converting CSV data.

## Building

```bash
go build -o csvforge ./cmd/csvforge/
```

## Running Tests

```bash
go test ./...
```

## Project Structure

- `cmd/csvforge/main.go` - CLI entry point with all 18 commands
- `internal/csv/csv.go` - Core CSV processing engine (reader, filter, sort, join, aggregate, stats)
- `internal/output/output.go` - Output formatting (table, CSV, JSON, JSONL, Markdown, TSV)

## Key Patterns

- Reader handles CSV parsing with auto-detection of delimiters
- All operations return new Reader instances (immutable pattern)
- Filter/Where support SQL-like expressions
- Join supports inner, left, right, and full joins
- Aggregate supports count, sum, avg, min, max, first, last

## Commands

18 commands: head, tail, sort, filter, sample, dedup, join, aggregate, merge, split, convert, validate, stats, info, columns, select, where, trim

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- Standard library only for CSV processing
