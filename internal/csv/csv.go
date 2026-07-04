package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

// Reader handles reading CSV files with auto-detection
type Reader struct {
	reader    *csv.Reader
	header    []string
	records   [][]string
	delim     rune
	hasHeader bool
}

// NewReader creates a new CSV reader from a file
func NewReader(filename string) (*Reader, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	return NewReaderFrom(f), nil
}

// NewReaderFrom creates a reader from an io.Reader
func NewReaderFrom(r io.Reader) *Reader {
	cr := csv.NewReader(r)
	cr.LazyQuotes = true
	cr.TrimLeadingSpace = true
	return &Reader{
		reader:    cr,
		delim:     ',',
		hasHeader: true,
	}
}

// SetDelimiter sets the CSV delimiter
func (r *Reader) SetDelimiter(d rune) {
	r.delim = d
	r.reader.Comma = d
}

// SetHeader indicates whether the CSV has a header row
func (r *Reader) SetHeader(hasHeader bool) {
	r.hasHeader = hasHeader
}

// SetFieldsPerRecord sets the expected number of fields per record (-1 for variable)
func (r *Reader) SetFieldsPerRecord(n int) {
	r.reader.FieldsPerRecord = n
}

// DetectDelimiter detects the delimiter from a sample of the file
func DetectDelimiter(data []byte) rune {
	// Count occurrences of common delimiters
	commas := 0
	tabs := 0
	semicolons := 0
	pipes := 0
	
	lines := strings.SplitN(string(data), "\n", 10)
	for _, line := range lines {
		commas += strings.Count(line, ",")
		tabs += strings.Count(line, "\t")
		semicolons += strings.Count(line, ";")
		pipes += strings.Count(line, "|")
	}
	
	// Pick the most common
	max := commas
	delim := ','
	if tabs > max {
		max = tabs
		delim = '\t'
	}
	if semicolons > max {
		max = semicolons
		delim = ';'
	}
	if pipes > max {
		delim = '|'
	}
	
	return delim
}

// DetectEncoding detects if the file likely uses a specific encoding
func DetectEncoding(data []byte) string {
	// Simple heuristic: check for BOM
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return "utf-8-bom"
	}
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xFE {
		return "utf-16-le"
	}
	return "utf-8"
}

// Read reads all records from the CSV
func (r *Reader) Read() error {
	r.header = nil
	r.records = nil
	
	// Set delimiter
	r.reader.Comma = r.delim
	
	// Read all records
	records, err := r.reader.ReadAll()
	if err != nil {
		return fmt.Errorf("read csv: %w", err)
	}
	
	if len(records) == 0 {
		return nil
	}
	
	if r.hasHeader {
		r.header = records[0]
		r.records = records[1:]
	} else {
		r.records = records
		if len(records) > 0 {
			// Generate numeric headers
			r.header = make([]string, len(records[0]))
			for i := range r.header {
				r.header[i] = fmt.Sprintf("col_%d", i+1)
			}
		}
	}
	
	return nil
}

// Header returns the header row
func (r *Reader) Header() []string {
	return r.header
}

// Records returns all data records
func (r *Reader) Records() [][]string {
	return r.records
}

// NumRows returns the number of data rows
func (r *Reader) NumRows() int {
	return len(r.records)
}

// NumCols returns the number of columns
func (r *Reader) NumCols() int {
	return len(r.header)
}

// ColumnIndex returns the index of a column by name, or -1 if not found
func (r *Reader) ColumnIndex(name string) int {
	for i, h := range r.header {
		if strings.EqualFold(h, name) {
			return i
		}
	}
	return -1
}

// GetColumn returns all values for a column by name
func (r *Reader) GetColumn(name string) ([]string, error) {
	idx := r.ColumnIndex(name)
	if idx < 0 {
		return nil, fmt.Errorf("column %q not found", name)
	}
	result := make([]string, len(r.records))
	for i, row := range r.records {
		if idx < len(row) {
			result[i] = row[idx]
		}
	}
	return result, nil
}

// Row represents a single CSV row as a map
type Row struct {
	Fields map[string]string
	Index  int
}

// ToMaps converts records to a slice of maps using the header
func (r *Reader) ToMaps() []Row {
	result := make([]Row, len(r.records))
	for i, rec := range r.records {
		m := make(map[string]string)
		for j, h := range r.header {
			if j < len(rec) {
				m[h] = rec[j]
			} else {
				m[h] = ""
			}
		}
		result[i] = Row{Fields: m, Index: i}
	}
	return result
}

// FilterFunc is a function that filters rows
type FilterFunc func(row map[string]string) bool

// Filter filters records using a filter function
func (r *Reader) Filter(fn FilterFunc) *Reader {
	var filtered [][]string
	for _, rec := range r.records {
		m := make(map[string]string)
		for j, h := range r.header {
			if j < len(rec) {
				m[h] = rec[j]
			}
		}
		if fn(m) {
			filtered = append(filtered, rec)
		}
	}
	return &Reader{
		header:  r.header,
		records: filtered,
		delim:   r.delim,
	}
}

// FilterExpr represents a simple filter expression
type FilterExpr struct {
	Column    string
	Operator  string // eq, neq, gt, gte, lt, lte, contains, starts_with, ends_with
	Value     string
}

// Matches checks if a row matches the filter expression
func (f FilterExpr) Matches(row map[string]string) bool {
	val, ok := row[f.Column]
	if !ok {
		return false
	}
	
	switch f.Operator {
	case "eq", "=":
		return val == f.Value
	case "neq", "!=":
		return val != f.Value
	case "gt", ">":
		return val > f.Value
	case "gte", ">=":
		return val >= f.Value
	case "lt", "<":
		return val < f.Value
	case "lte", "<=":
		return val <= f.Value
	case "contains":
		return strings.Contains(val, f.Value)
	case "starts_with":
		return strings.HasPrefix(val, f.Value)
	case "ends_with":
		return strings.HasSuffix(val, f.Value)
	case "empty":
		return val == ""
	case "not_empty":
		return val != ""
	default:
		return false
	}
}

// SortFunc is a comparison function for sorting
type SortFunc func(a, b []string) int

// Sort sorts records by a column
func (r *Reader) Sort(column string, descending bool) *Reader {
	idx := r.ColumnIndex(column)
	if idx < 0 {
		return r
	}
	
	sorted := make([][]string, len(r.records))
	copy(sorted, r.records)
	
	// Simple insertion sort (good enough for most CSV files)
	for i := 1; i < len(sorted); i++ {
		key := sorted[i]
		j := i - 1
		
		for j >= 0 {
			a := ""
			b := ""
			if idx < len(sorted[j]) {
				a = sorted[j][idx]
			}
			if idx < len(key) {
				b = key[idx]
			}
			
			compare := strings.Compare(a, b)
			if descending {
				compare = -compare
			}
			
			if compare > 0 {
				sorted[j+1] = sorted[j]
				j--
			} else {
				break
			}
		}
		sorted[j+1] = key
	}
	
	return &Reader{
		header:  r.header,
		records: sorted,
		delim:   r.delim,
	}
}

// Head returns the first n rows
func (r *Reader) Head(n int) *Reader {
	if n > len(r.records) {
		n = len(r.records)
	}
	return &Reader{
		header:  r.header,
		records: r.records[:n],
		delim:   r.delim,
	}
}

// Tail returns the last n rows
func (r *Reader) Tail(n int) *Reader {
	start := len(r.records) - n
	if start < 0 {
		start = 0
	}
	return &Reader{
		header:  r.header,
		records: r.records[start:],
		delim:   r.delim,
	}
}

// Sample returns n random rows
func (r *Reader) Sample(n int, seed int64) *Reader {
	if n >= len(r.records) {
		return r
	}
	
	// Simple deterministic sampling using seed
	result := make([][]string, n)
	used := make(map[int]bool)
	
	for i := 0; i < n; i++ {
		// Simple hash-based selection
		idx := int((seed + int64(i*7919)) % int64(len(r.records)))
		for used[idx] {
			idx = (idx + 1) % len(r.records)
		}
		used[idx] = true
		result[i] = r.records[idx]
	}
	
	return &Reader{
		header:  r.header,
		records: result,
		delim:   r.delim,
	}
}

// Dedup removes duplicate rows by specified columns
func (r *Reader) Dedup(columns []string) *Reader {
	if len(columns) == 0 {
		columns = r.header
	}
	
	// Get column indices
	indices := make([]int, 0, len(columns))
	for _, col := range columns {
		if idx := r.ColumnIndex(col); idx >= 0 {
			indices = append(indices, idx)
		}
	}
	
	seen := make(map[string]bool)
	var deduped [][]string
	
	for _, rec := range r.records {
		// Build key from specified columns
		key := ""
		for _, idx := range indices {
			if idx < len(rec) {
				key += rec[idx] + "\x00"
			}
		}
		
		if !seen[key] {
			seen[key] = true
			deduped = append(deduped, rec)
		}
	}
	
	return &Reader{
		header:  r.header,
		records: deduped,
		delim:   r.delim,
	}
}

// Join joins two readers on specified columns
func (r *Reader) Join(other *Reader, joinCol string, joinType string) *Reader {
	idx1 := r.ColumnIndex(joinCol)
	idx2 := other.ColumnIndex(joinCol)
	
	if idx1 < 0 || idx2 < 0 {
		return r
	}
	
	// Build lookup from other reader
	lookup := make(map[string][][]string)
	for _, rec := range other.records {
		if idx2 < len(rec) {
			key := rec[idx2]
			lookup[key] = append(lookup[key], rec)
		}
	}
	
	// Merge headers (avoid duplicates)
	mergedHeader := make([]string, 0, len(r.header)+len(other.header))
	mergedHeader = append(mergedHeader, r.header...)
	otherHeaderSet := make(map[string]bool)
	for _, h := range r.header {
		otherHeaderSet[h] = true
	}
	for _, h := range other.header {
		if !otherHeaderSet[h] {
			mergedHeader = append(mergedHeader, h)
		}
	}
	
	// Perform join
	var joined [][]string
	matched := make(map[string]bool)
	
	for _, rec := range r.records {
		key := ""
		if idx1 < len(rec) {
			key = rec[idx1]
		}
		
		matches, ok := lookup[key]
		if ok {
			matched[key] = true
			for _, match := range matches {
				merged := make([]string, len(mergedHeader))
				copy(merged, rec)
				// Copy fields from match (excluding join column)
				for j, h := range other.header {
					if h != joinCol {
						// Find position in merged header
						for k, mh := range mergedHeader {
							if mh == h && k >= len(r.header) {
								if j < len(match) {
									merged[k] = match[j]
								}
								break
							}
						}
					}
				}
				joined = append(joined, merged)
			}
		} else if joinType == "left" || joinType == "full" {
			merged := make([]string, len(mergedHeader))
			copy(merged, rec)
			joined = append(joined, merged)
		}
	}
	
	// For right/full joins, add unmatched rows from other
	if joinType == "right" || joinType == "full" {
		for key, matches := range lookup {
			if !matched[key] {
				for _, match := range matches {
					merged := make([]string, len(mergedHeader))
					// Copy fields from match
					for j, h := range other.header {
						for k, mh := range mergedHeader {
							if mh == h {
								if j < len(match) {
									merged[k] = match[j]
								}
								break
							}
						}
					}
					joined = append(joined, merged)
				}
			}
		}
	}
	
	return &Reader{
		header:  mergedHeader,
		records: joined,
		delim:   r.delim,
	}
}

// Aggregate performs aggregation on grouped data
type AggFunc struct {
	Column   string
	Function string // count, sum, avg, min, max, first, last
	Alias    string
}

// Aggregate groups by specified columns and applies aggregation functions
func (r *Reader) Aggregate(groupBy []string, aggs []AggFunc) *Reader {
	// Get group column indices
	groupIndices := make([]int, 0, len(groupBy))
	for _, col := range groupBy {
		if idx := r.ColumnIndex(col); idx >= 0 {
			groupIndices = append(groupIndices, idx)
		}
	}
	
	// Group records
	groups := make(map[string][][]string)
	groupOrder := make([]string, 0)
	
	for _, rec := range r.records {
		key := ""
		for _, idx := range groupIndices {
			if idx < len(rec) {
				key += rec[idx] + "\x00"
			}
		}
		
		if _, exists := groups[key]; !exists {
			groupOrder = append(groupOrder, key)
		}
		groups[key] = append(groups[key], rec)
	}
	
	// Build result header
	resultHeader := make([]string, 0, len(groupBy)+len(aggs))
	resultHeader = append(resultHeader, groupBy...)
	for _, agg := range aggs {
		if agg.Alias != "" {
			resultHeader = append(resultHeader, agg.Alias)
		} else {
			resultHeader = append(resultHeader, agg.Function+"_"+agg.Column)
		}
	}
	
	// Apply aggregations
	var results [][]string
	for _, key := range groupOrder {
		records := groups[key]
		row := make([]string, len(resultHeader))
		
		// Copy group values
		if len(records) > 0 {
			for i, idx := range groupIndices {
				if i < len(groupBy) && idx < len(records[0]) {
					row[i] = records[0][idx]
				}
			}
		}
		
		// Apply aggregation functions
		for j, agg := range aggs {
			colIdx := r.ColumnIndex(agg.Column)
			if colIdx < 0 {
				continue
			}
			
			values := make([]string, len(records))
			for i, rec := range records {
				if colIdx < len(rec) {
					values[i] = rec[colIdx]
				}
			}
			
			result := applyAggFunc(values, agg.Function)
			row[len(groupBy)+j] = result
		}
		
		results = append(results, row)
	}
	
	return &Reader{
		header:  resultHeader,
		records: results,
		delim:   r.delim,
	}
}

// applyAggFunc applies an aggregation function to values
func applyAggFunc(values []string, fn string) string {
	switch fn {
	case "count":
		return fmt.Sprintf("%d", len(values))
	case "sum":
		sum := 0.0
		for _, v := range values {
			var f float64
			fmt.Sscanf(v, "%f", &f)
			sum += f
		}
		return fmt.Sprintf("%.2f", sum)
	case "avg":
		if len(values) == 0 {
			return "0"
		}
		sum := 0.0
		count := 0
		for _, v := range values {
			var f float64
			if _, err := fmt.Sscanf(v, "%f", &f); err == nil {
				sum += f
				count++
			}
		}
		if count == 0 {
			return "0"
		}
		return fmt.Sprintf("%.2f", sum/float64(count))
	case "min":
		min := ""
		for _, v := range values {
			if v != "" && (min == "" || v < min) {
				min = v
			}
		}
		return min
	case "max":
		max := ""
		for _, v := range values {
			if v != "" && (max == "" || v > max) {
				max = v
			}
		}
		return max
	case "first":
		if len(values) > 0 {
			return values[0]
		}
		return ""
	case "last":
		if len(values) > 0 {
			return values[len(values)-1]
		}
		return ""
	default:
		return ""
	}
}

// Merge concatenates multiple readers
func Merge(readers ...*Reader) *Reader {
	if len(readers) == 0 {
		return &Reader{}
	}
	
	header := readers[0].header
	var allRecords [][]string
	
	for _, r := range readers {
		for _, rec := range r.records {
			// Ensure record matches header length
			if len(rec) < len(header) {
				padded := make([]string, len(header))
				copy(padded, rec)
				rec = padded
			}
			allRecords = append(allRecords, rec)
		}
	}
	
	return &Reader{
		header:  header,
		records: allRecords,
		delim:   readers[0].delim,
	}
}

// Split splits a reader into chunks of size n
func (r *Reader) Split(n int) []*Reader {
	var chunks []*Reader
	for i := 0; i < len(r.records); i += n {
		end := i + n
		if end > len(r.records) {
			end = len(r.records)
		}
		chunks = append(chunks, &Reader{
			header:  r.header,
			records: r.records[i:end],
			delim:   r.delim,
		})
	}
	return chunks
}

// Stats computes column statistics
type ColumnStats struct {
	Name     string
	Type     string // string, number, mixed
	Count    int
	Nulls    int
	Unique   int
	Min      string
	Max      string
	AvgLen   float64
	Samples  []string
}

// ComputeStats computes statistics for each column
func (r *Reader) ComputeStats() []ColumnStats {
	stats := make([]ColumnStats, len(r.header))
	
	for i, h := range r.header {
		stats[i].Name = h
		unique := make(map[string]bool)
		totalLen := 0
		isNumeric := true
		min := ""
		max := ""
		
		for _, rec := range r.records {
			val := ""
			if i < len(rec) {
				val = rec[i]
			}
			
			stats[i].Count++
			if val == "" {
				stats[i].Nulls++
			} else {
				unique[val] = true
				totalLen += len(val)
				if min == "" || val < min {
					min = val
				}
				if max == "" || val > max {
					max = val
				}
				// Check if numeric
				var f float64
				if _, err := fmt.Sscanf(val, "%f", &f); err != nil {
					isNumeric = false
				}
			}
		}
		
		stats[i].Unique = len(unique)
		stats[i].Min = min
		stats[i].Max = max
		if stats[i].Count-stats[i].Nulls > 0 {
			stats[i].AvgLen = float64(totalLen) / float64(stats[i].Count-stats[i].Nulls)
		}
		
		if isNumeric {
			stats[i].Type = "number"
		} else {
			stats[i].Type = "string"
		}
		
		// Collect samples
		sampleCount := 0
		for _, rec := range r.records {
			if sampleCount >= 5 {
				break
			}
			if i < len(rec) && rec[i] != "" {
				stats[i].Samples = append(stats[i].Samples, rec[i])
				sampleCount++
			}
		}
	}
	
	return stats
}

// Validate checks CSV data against basic rules
type ValidationError struct {
	Row     int
	Column  string
	Message string
}

// Validate checks for common CSV issues
func (r *Reader) Validate() []ValidationError {
	var errors []ValidationError
	
	for i, rec := range r.records {
		// Check for wrong number of fields
		if len(rec) != len(r.header) {
			errors = append(errors, ValidationError{
				Row:     i + 1,
				Column:  "*",
				Message: fmt.Sprintf("expected %d fields, got %d", len(r.header), len(rec)),
			})
		}
		
		// Check for empty rows
		allEmpty := true
		for _, field := range rec {
			if strings.TrimSpace(field) != "" {
				allEmpty = false
				break
			}
		}
		if allEmpty {
			errors = append(errors, ValidationError{
				Row:     i + 1,
				Column:  "*",
				Message: "empty row",
			})
		}
		
		// Check for whitespace-only fields
		for j, field := range rec {
			if field != "" && strings.TrimSpace(field) != field {
				if j < len(r.header) {
					errors = append(errors, ValidationError{
						Row:     i + 1,
						Column:  r.header[j],
						Message: "leading/trailing whitespace",
					})
				}
			}
		}
		
		// Check for null bytes
		for j, field := range rec {
			if strings.ContainsRune(field, '\x00') {
				if j < len(r.header) {
					errors = append(errors, ValidationError{
						Row:     i + 1,
						Column:  r.header[j],
						Message: "contains null byte",
					})
				}
			}
		}
	}
	
	return errors
}

// TrimSpace trims whitespace from all fields
func (r *Reader) TrimSpace() *Reader {
	trimmed := make([][]string, len(r.records))
	for i, rec := range r.records {
		trimmed[i] = make([]string, len(rec))
		for j, field := range rec {
			trimmed[i][j] = strings.TrimSpace(field)
		}
	}
	return &Reader{
		header:  r.header,
		records: trimmed,
		delim:   r.delim,
	}
}

// SelectColumns returns only the specified columns
func (r *Reader) SelectColumns(columns []string) *Reader {
	indices := make([]int, 0, len(columns))
	for _, col := range columns {
		if idx := r.ColumnIndex(col); idx >= 0 {
			indices = append(indices, idx)
		}
	}
	
	newHeader := make([]string, len(indices))
	for i, idx := range indices {
		newHeader[i] = r.header[idx]
	}
	
	newRecords := make([][]string, len(r.records))
	for i, rec := range r.records {
		newRecords[i] = make([]string, len(indices))
		for j, idx := range indices {
			if idx < len(rec) {
				newRecords[i][j] = rec[idx]
			}
		}
	}
	
	return &Reader{
		header:  newHeader,
		records: newRecords,
		delim:   r.delim,
	}
}

// AddColumn adds a new column with a default value or expression
func (r *Reader) AddColumn(name string, defaultValue string) *Reader {
	newHeader := make([]string, len(r.header)+1)
	copy(newHeader, r.header)
	newHeader[len(r.header)] = name
	
	newRecords := make([][]string, len(r.records))
	for i, rec := range r.records {
		newRecords[i] = make([]string, len(rec)+1)
		copy(newRecords[i], rec)
		newRecords[i][len(rec)] = defaultValue
	}
	
	return &Reader{
		header:  newHeader,
		records: newRecords,
		delim:   r.delim,
	}
}

// DropColumns removes specified columns
func (r *Reader) DropColumns(columns []string) *Reader {
	dropSet := make(map[string]bool)
	for _, col := range columns {
		dropSet[strings.ToLower(col)] = true
	}
	
	var keepIndices []int
	var newHeader []string
	for i, h := range r.header {
		if !dropSet[strings.ToLower(h)] {
			keepIndices = append(keepIndices, i)
			newHeader = append(newHeader, h)
		}
	}
	
	newRecords := make([][]string, len(r.records))
	for i, rec := range r.records {
		newRecords[i] = make([]string, len(keepIndices))
		for j, idx := range keepIndices {
			if idx < len(rec) {
				newRecords[i][j] = rec[idx]
			}
		}
	}
	
	return &Reader{
		header:  newHeader,
		records: newRecords,
		delim:   r.delim,
	}
}

// RenameColumn renames a column
func (r *Reader) RenameColumn(oldName, newName string) *Reader {
	newHeader := make([]string, len(r.header))
	copy(newHeader, r.header)
	
	for i, h := range r.header {
		if strings.EqualFold(h, oldName) {
			newHeader[i] = newName
			break
		}
	}
	
	return &Reader{
		header:  newHeader,
		records: r.records,
		delim:   r.delim,
	}
}

// Where applies a SQL-like WHERE clause
func (r *Reader) Where(condition string) *Reader {
	// Simple parser for conditions like "column = value" or "column > value"
	// Supports: =, !=, >, <, >=, <=, LIKE, IS NULL, IS NOT NULL
	
	parts := parseWhereCondition(condition)
	if parts == nil {
		return r
	}
	
	return r.Filter(func(row map[string]string) bool {
		val, ok := row[parts.Column]
		if !ok {
			return false
		}
		return parts.Matches(val)
	})
}

// WhereCondition represents a parsed WHERE condition
type WhereCondition struct {
	Column   string
	Operator string
	Value    string
}

// Matches checks if a value matches the condition
func (w *WhereCondition) Matches(val string) bool {
	switch w.Operator {
	case "=":
		return val == w.Value
	case "!=":
		return val != w.Value
	case ">":
		return val > w.Value
	case ">=":
		return val >= w.Value
	case "<":
		return val < w.Value
	case "<=":
		return val <= w.Value
	case "LIKE":
		// Simple pattern matching with % wildcards
		trimmed := strings.Trim(w.Value, "%")
		if strings.HasPrefix(w.Value, "%") && strings.HasSuffix(w.Value, "%") {
			return strings.Contains(val, trimmed)
		} else if strings.HasPrefix(w.Value, "%") {
			return strings.HasSuffix(val, trimmed)
		} else if strings.HasSuffix(w.Value, "%") {
			return strings.HasPrefix(val, trimmed)
		}
		return val == w.Value
	case "IS NULL":
		return val == ""
	case "IS NOT NULL":
		return val != ""
	default:
		return false
	}
}

// parseWhereCondition parses a WHERE condition string
func parseWhereCondition(condition string) *WhereCondition {
	condition = strings.TrimSpace(condition)
	
	// Try operators in order of length (longest first)
	operators := []string{">=", "<=", "!=", "LIKE", "IS NULL", "IS NOT NULL", "=", ">", "<"}
	
	for _, op := range operators {
		idx := strings.Index(strings.ToUpper(condition), op)
		if idx > 0 {
			column := strings.TrimSpace(condition[:idx])
			value := strings.TrimSpace(condition[idx+len(op):])
			// Remove quotes if present
			value = strings.Trim(value, "'\"")
			return &WhereCondition{
				Column:   column,
				Operator: op,
				Value:    value,
			}
		}
	}
	
	return nil
}

// Write writes the CSV to a writer
func (r *Reader) Write(w io.Writer, delim rune) error {
	cw := csv.NewWriter(w)
	cw.Comma = delim
	
	// Write header
	if err := cw.Write(r.header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	
	// Write records
	for _, rec := range r.records {
		if err := cw.Write(rec); err != nil {
			return fmt.Errorf("write record: %w", err)
		}
	}
	
	cw.Flush()
	return cw.Error()
}

// WriteToFile writes the CSV to a file
func (r *Reader) WriteToFile(filename string, delim rune) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()
	return r.Write(f, delim)
}

// String returns a simple string representation
func (r *Reader) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CSV: %d rows, %d columns\n", len(r.records), len(r.header)))
	sb.WriteString("Headers: ")
	sb.WriteString(strings.Join(r.header, ", "))
	return sb.String()
}

// Helper function to check if a rune is a number
func isNumber(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) && r != '.' && r != '-' && r != '+' {
			return false
		}
	}
	return true
}
