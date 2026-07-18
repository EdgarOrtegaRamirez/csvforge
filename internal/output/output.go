package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

// Format represents output format
type Format string

const (
	FormatTable    Format = "table"
	FormatCSV      Format = "csv"
	FormatJSON     Format = "json"
	FormatJSONL    Format = "jsonl"
	FormatMarkdown Format = "markdown"
	FormatTSV      Format = "tsv"
	FormatMinimal  Format = "minimal"
)

// Writer handles output formatting
type Writer struct {
	format    Format
	delimiter rune
}

// NewWriter creates a new output writer
func NewWriter(format Format) *Writer {
	delim := ','
	switch format {
	case FormatTSV:
		delim = '\t'
	case FormatMinimal:
		delim = ' '
	}
	return &Writer{
		format:    format,
		delimiter: delim,
	}
}

// Write writes the data in the specified format
func (w *Writer) Write(header []string, records [][]string, out io.Writer) error {
	switch w.format {
	case FormatTable:
		return w.writeTable(header, records, out)
	case FormatCSV:
		return w.writeCSV(header, records, out)
	case FormatJSON:
		return w.writeJSON(header, records, out)
	case FormatJSONL:
		return w.writeJSONL(header, records, out)
	case FormatMarkdown:
		return w.writeMarkdown(header, records, out)
	case FormatTSV:
		return w.writeTSV(header, records, out)
	case FormatMinimal:
		return w.writeMinimal(header, records, out)
	default:
		return w.writeTable(header, records, out)
	}
}

// writeTable writes a formatted table
func (w *Writer) writeTable(header []string, records [][]string, out io.Writer) error {
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)

	// Write header
	for i, h := range header {
		if i > 0 {
			fmt.Fprint(tw, "\t")
		}
		fmt.Fprint(tw, strings.ToUpper(h))
	}
	fmt.Fprintln(tw)

	// Write separator
	for i := range header {
		if i > 0 {
			fmt.Fprint(tw, "\t")
		}
		fmt.Fprint(tw, strings.Repeat("-", len(strings.ToUpper(header[i]))))
	}
	fmt.Fprintln(tw)

	// Write records
	for _, rec := range records {
		for i, val := range rec {
			if i > 0 {
				fmt.Fprint(tw, "\t")
			}
			fmt.Fprint(tw, val)
		}
		fmt.Fprintln(tw)
	}

	return tw.Flush()
}

// writeCSV writes CSV format
func (w *Writer) writeCSV(header []string, records [][]string, out io.Writer) error {
	// Write header
	w.writeCSVLineMethod(out, header)

	// Write records
	for _, rec := range records {
		w.writeCSVLineMethod(out, rec)
	}

	return nil
}

// writeCSVLineMethod writes a single CSV line using the writer's delimiter
func (w *Writer) writeCSVLineMethod(out io.Writer, fields []string) {
	for i, field := range fields {
		if i > 0 {
			fmt.Fprint(out, string(w.delimiter))
		}
		// Quote fields that contain the delimiter, newline, or quotes
		if strings.ContainsAny(field, string(w.delimiter)+"\n\"\r") {
			field = strings.ReplaceAll(field, "\"", "\"\"")
			fmt.Fprintf(out, "\"%s\"", field)
		} else {
			fmt.Fprint(out, field)
		}
	}
	fmt.Fprintln(out)
}

// writeCSVLine writes a single CSV line
func (w *Writer) writeCSVLine(out io.Writer, fields []string) {
	for i, field := range fields {
		if i > 0 {
			fmt.Fprint(out, string(w.delimiter))
		}
		// Quote fields that contain the delimiter, newline, or quotes
		if strings.ContainsAny(field, string(w.delimiter)+"\n\"\r") {
			field = strings.ReplaceAll(field, "\"", "\"\"")
			fmt.Fprintf(out, "\"%s\"", field)
		} else {
			fmt.Fprint(out, field)
		}
	}
	fmt.Fprintln(out)
}

// writeJSON writes JSON array format
func (w *Writer) writeJSON(header []string, records [][]string, out io.Writer) error {
	var result []map[string]string

	for _, rec := range records {
		row := make(map[string]string)
		for i, h := range header {
			if i < len(rec) {
				row[h] = rec[i]
			} else {
				row[h] = ""
			}
		}
		result = append(result, row)
	}

	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

// writeJSONL writes JSON Lines format
func (w *Writer) writeJSONL(header []string, records [][]string, out io.Writer) error {
	for _, rec := range records {
		row := make(map[string]string)
		for i, h := range header {
			if i < len(rec) {
				row[h] = rec[i]
			} else {
				row[h] = ""
			}
		}

		data, err := json.Marshal(row)
		if err != nil {
			return err
		}
		fmt.Fprintln(out, string(data))
	}
	return nil
}

// writeMarkdown writes Markdown table format
func (w *Writer) writeMarkdown(header []string, records [][]string, out io.Writer) error {
	// Write header
	fmt.Fprint(out, "| ")
	for i, h := range header {
		if i > 0 {
			fmt.Fprint(out, " | ")
		}
		fmt.Fprint(out, h)
	}
	fmt.Fprintln(out, " |")

	// Write separator
	fmt.Fprint(out, "| ")
	for i := range header {
		if i > 0 {
			fmt.Fprint(out, " | ")
		}
		fmt.Fprint(out, "---")
	}
	fmt.Fprintln(out, " |")

	// Write records
	for _, rec := range records {
		fmt.Fprint(out, "| ")
		for i, val := range rec {
			if i > 0 {
				fmt.Fprint(out, " | ")
			}
			// Escape pipe characters
			val = strings.ReplaceAll(val, "|", "\\|")
			fmt.Fprint(out, val)
		}
		fmt.Fprintln(out, " |")
	}

	return nil
}

// writeTSV writes TSV format
func (w *Writer) writeTSV(header []string, records [][]string, out io.Writer) error {
	w.delimiter = '\t'
	return w.writeCSV(header, records, out)
}

// writeMinimal writes minimal space-separated format
func (w *Writer) writeMinimal(header []string, records [][]string, out io.Writer) error {
	// Just write values separated by spaces
	for _, rec := range records {
		for i, val := range rec {
			if i > 0 {
				fmt.Fprint(out, " ")
			}
			fmt.Fprint(out, val)
		}
		fmt.Fprintln(out)
	}
	return nil
}

// WriteToFile writes output to a file
func WriteToFile(header []string, records [][]string, filename string, format Format) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	w := NewWriter(format)
	return w.Write(header, records, f)
}

// Helper to write a CSV line (standalone version)
func writeCSVLine(out io.Writer, fields []string) {
	for i, field := range fields {
		if i > 0 {
			fmt.Fprint(out, ",")
		}
		if strings.ContainsAny(field, ",\n\"\r") {
			field = strings.ReplaceAll(field, "\"", "\"\"")
			fmt.Fprintf(out, "\"%s\"", field)
		} else {
			fmt.Fprint(out, field)
		}
	}
	fmt.Fprintln(out)
}
