package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteTable(t *testing.T) {
	header := []string{"name", "age", "city"}
	records := [][]string{
		{"Alice", "30", "NYC"},
		{"Bob", "25", "LA"},
	}

	var buf bytes.Buffer
	w := NewWriter(FormatTable)
	if err := w.Write(header, records, &buf); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "NAME") {
		t.Error("Expected uppercase header")
	}
	if !strings.Contains(output, "Alice") {
		t.Error("Expected data row")
	}
}

func TestWriteJSON(t *testing.T) {
	header := []string{"name", "age"}
	records := [][]string{
		{"Alice", "30"},
		{"Bob", "25"},
	}

	var buf bytes.Buffer
	w := NewWriter(FormatJSON)
	if err := w.Write(header, records, &buf); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"name": "Alice"`) {
		t.Error("Expected JSON output with name field")
	}
	if !strings.Contains(output, `"age": "30"`) {
		t.Error("Expected JSON output with age field")
	}
}

func TestWriteJSONL(t *testing.T) {
	header := []string{"name", "age"}
	records := [][]string{
		{"Alice", "30"},
		{"Bob", "25"},
	}

	var buf bytes.Buffer
	w := NewWriter(FormatJSONL)
	if err := w.Write(header, records, &buf); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 JSONL lines, got %d", len(lines))
	}
	if !strings.Contains(lines[0], `"name":"Alice"`) {
		t.Error("Expected JSONL line with name field")
	}
}

func TestWriteMarkdown(t *testing.T) {
	header := []string{"name", "age"}
	records := [][]string{
		{"Alice", "30"},
		{"Bob", "25"},
	}

	var buf bytes.Buffer
	w := NewWriter(FormatMarkdown)
	if err := w.Write(header, records, &buf); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "| name | age |") {
		t.Error("Expected Markdown table header")
	}
	if !strings.Contains(output, "| --- | --- |") {
		t.Error("Expected Markdown table separator")
	}
	if !strings.Contains(output, "| Alice | 30 |") {
		t.Error("Expected Markdown table row")
	}
}

func TestWriteCSV(t *testing.T) {
	header := []string{"name", "age"}
	records := [][]string{
		{"Alice", "30"},
		{"Bob", "25"},
	}

	var buf bytes.Buffer
	w := NewWriter(FormatCSV)
	if err := w.Write(header, records, &buf); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "name,age") {
		t.Error("Expected CSV header")
	}
	if !strings.Contains(output, "Alice,30") {
		t.Error("Expected CSV row")
	}
}

func TestWriteCSVWithQuotes(t *testing.T) {
	header := []string{"name", "description"}
	records := [][]string{
		{"Alice", "She said \"hello\""},
		{"Bob", "Comma, here"},
	}

	var buf bytes.Buffer
	w := NewWriter(FormatCSV)
	if err := w.Write(header, records, &buf); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"She said ""hello"""`) {
		t.Error("Expected quoted field")
	}
	if !strings.Contains(output, `"Comma, here"`) {
		t.Error("Expected quoted field with comma")
	}
}

func TestWriteTSV(t *testing.T) {
	header := []string{"name", "age"}
	records := [][]string{
		{"Alice", "30"},
	}

	var buf bytes.Buffer
	w := NewWriter(FormatTSV)
	if err := w.Write(header, records, &buf); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "name\tage") {
		t.Error("Expected TSV header with tab delimiter")
	}
}

func TestWriteMinimal(t *testing.T) {
	header := []string{"name", "age"}
	records := [][]string{
		{"Alice", "30"},
	}

	var buf bytes.Buffer
	w := NewWriter(FormatMinimal)
	if err := w.Write(header, records, &buf); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Alice 30") {
		t.Error("Expected space-separated output")
	}
}

func TestFormatConstants(t *testing.T) {
	if FormatTable != "table" {
		t.Error("FormatTable should be 'table'")
	}
	if FormatCSV != "csv" {
		t.Error("FormatCSV should be 'csv'")
	}
	if FormatJSON != "json" {
		t.Error("FormatJSON should be 'json'")
	}
	if FormatJSONL != "jsonl" {
		t.Error("FormatJSONL should be 'jsonl'")
	}
	if FormatMarkdown != "markdown" {
		t.Error("FormatMarkdown should be 'markdown'")
	}
}
