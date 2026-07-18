package csv

import (
	"strings"
	"testing"
)

func TestNewReader(t *testing.T) {
	csvData := "name,age,city\nAlice,30,NYC\nBob,25,LA\n"
	r := NewReaderFrom(strings.NewReader(csvData))

	if err := r.Read(); err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if r.NumRows() != 2 {
		t.Errorf("NumRows() = %d, want 2", r.NumRows())
	}

	if r.NumCols() != 3 {
		t.Errorf("NumCols() = %d, want 3", r.NumCols())
	}

	if r.Header()[0] != "name" {
		t.Errorf("Header()[0] = %q, want %q", r.Header()[0], "name")
	}
}

func TestNoHeader(t *testing.T) {
	csvData := "Alice,30,NYC\nBob,25,LA\n"
	r := NewReaderFrom(strings.NewReader(csvData))
	r.SetHeader(false)

	if err := r.Read(); err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if r.NumRows() != 2 {
		t.Errorf("NumRows() = %d, want 2", r.NumRows())
	}

	if r.Header()[0] != "col_1" {
		t.Errorf("Header()[0] = %q, want %q", r.Header()[0], "col_1")
	}
}

func TestFilter(t *testing.T) {
	csvData := "name,age,city\nAlice,30,NYC\nBob,25,LA\nCharlie,35,NYC\n"
	r := NewReaderFrom(strings.NewReader(csvData))

	if err := r.Read(); err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	filtered := r.Filter(func(row map[string]string) bool {
		return row["city"] == "NYC"
	})

	if filtered.NumRows() != 2 {
		t.Errorf("NumRows() = %d, want 2", filtered.NumRows())
	}
}

func TestFilterExpr(t *testing.T) {
	tests := []struct {
		name   string
		expr   FilterExpr
		value  string
		expect bool
	}{
		{"eq", FilterExpr{Column: "age", Operator: "eq", Value: "30"}, "30", true},
		{"neq", FilterExpr{Column: "age", Operator: "neq", Value: "30"}, "25", true},
		{"gt", FilterExpr{Column: "age", Operator: "gt", Value: "25"}, "30", true},
		{"gte", FilterExpr{Column: "age", Operator: "gte", Value: "30"}, "30", true},
		{"lt", FilterExpr{Column: "age", Operator: "lt", Value: "30"}, "25", true},
		{"lte", FilterExpr{Column: "age", Operator: "lte", Value: "30"}, "30", true},
		{"contains", FilterExpr{Column: "name", Operator: "contains", Value: "li"}, "Alice", true},
		{"starts_with", FilterExpr{Column: "name", Operator: "starts_with", Value: "A"}, "Alice", true},
		{"ends_with", FilterExpr{Column: "name", Operator: "ends_with", Value: "e"}, "Alice", true},
		{"empty", FilterExpr{Column: "name", Operator: "empty", Value: ""}, "", true},
		{"not_empty", FilterExpr{Column: "name", Operator: "not_empty", Value: ""}, "Alice", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := map[string]string{"age": tt.value, "name": tt.value}
			if got := tt.expr.Matches(row); got != tt.expect {
				t.Errorf("Matches() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestSort(t *testing.T) {
	csvData := "name,age\nCharlie,35\nAlice,30\nBob,25\n"
	r := NewReaderFrom(strings.NewReader(csvData))

	if err := r.Read(); err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	sorted := r.Sort("age", false)

	if sorted.Records()[0][0] != "Bob" {
		t.Errorf("Sort asc: first = %q, want %q", sorted.Records()[0][0], "Bob")
	}

	sortedDesc := r.Sort("age", true)
	if sortedDesc.Records()[0][0] != "Charlie" {
		t.Errorf("Sort desc: first = %q, want %q", sortedDesc.Records()[0][0], "Charlie")
	}
}

func TestHead(t *testing.T) {
	csvData := "name,age\nA,1\nB,2\nC,3\nD,4\nE,5\n"
	r := NewReaderFrom(strings.NewReader(csvData))

	if err := r.Read(); err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	head := r.Head(3)
	if head.NumRows() != 3 {
		t.Errorf("NumRows() = %d, want 3", head.NumRows())
	}
}

func TestTail(t *testing.T) {
	csvData := "name,age\nA,1\nB,2\nC,3\nD,4\nE,5\n"
	r := NewReaderFrom(strings.NewReader(csvData))

	if err := r.Read(); err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	tail := r.Tail(2)
	if tail.NumRows() != 2 {
		t.Errorf("NumRows() = %d, want 2", tail.NumRows())
	}
	if tail.Records()[0][0] != "D" {
		t.Errorf("First record = %q, want %q", tail.Records()[0][0], "D")
	}
}

func TestSample(t *testing.T) {
	csvData := "name,age\nA,1\nB,2\nC,3\nD,4\nE,5\n"
	r := NewReaderFrom(strings.NewReader(csvData))

	if err := r.Read(); err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	sampled := r.Sample(3, 42)
	if sampled.NumRows() != 3 {
		t.Errorf("NumRows() = %d, want 3", sampled.NumRows())
	}
}

func TestDedup(t *testing.T) {
	csvData := "name,age\nAlice,30\nBob,25\nAlice,30\nCharlie,35\n"
	r := NewReaderFrom(strings.NewReader(csvData))

	if err := r.Read(); err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	deduped := r.Dedup([]string{"name"})
	if deduped.NumRows() != 3 {
		t.Errorf("NumRows() = %d, want 3", deduped.NumRows())
	}
}

func TestMerge(t *testing.T) {
	csv1 := "name,age\nAlice,30\nBob,25\n"
	csv2 := "name,age\nCharlie,35\nDave,40\n"

	r1 := NewReaderFrom(strings.NewReader(csv1))
	r2 := NewReaderFrom(strings.NewReader(csv2))

	r1.Read()
	r2.Read()

	merged := Merge(r1, r2)
	if merged.NumRows() != 4 {
		t.Errorf("NumRows() = %d, want 4", merged.NumRows())
	}
}

func TestSplit(t *testing.T) {
	csvData := "name,age\nA,1\nB,2\nC,3\nD,4\nE,5\n"
	r := NewReaderFrom(strings.NewReader(csvData))

	if err := r.Read(); err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	chunks := r.Split(2)
	if len(chunks) != 3 {
		t.Errorf("len(chunks) = %d, want 3", len(chunks))
	}
	if chunks[0].NumRows() != 2 {
		t.Errorf("chunks[0].NumRows() = %d, want 2", chunks[0].NumRows())
	}
}

func TestTrimSpace(t *testing.T) {
	csvData := "name,age\n  Alice  , 30 \n Bob ,25 \n"
	r := NewReaderFrom(strings.NewReader(csvData))

	if err := r.Read(); err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	trimmed := r.TrimSpace()
	if trimmed.Records()[0][0] != "Alice" {
		t.Errorf("Trimmed name = %q, want %q", trimmed.Records()[0][0], "Alice")
	}
}

func TestSelectColumns(t *testing.T) {
	csvData := "name,age,city\nAlice,30,NYC\nBob,25,LA\n"
	r := NewReaderFrom(strings.NewReader(csvData))

	if err := r.Read(); err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	selected := r.SelectColumns([]string{"name", "city"})
	if selected.NumCols() != 2 {
		t.Errorf("NumCols() = %d, want 2", selected.NumCols())
	}
	if selected.Header()[0] != "name" {
		t.Errorf("Header()[0] = %q, want %q", selected.Header()[0], "name")
	}
}

func TestValidate(t *testing.T) {
	csvData := "name,age\nAlice,30\n Bob ,25\nCharlie, 35 \n"
	r := NewReaderFrom(strings.NewReader(csvData))

	if err := r.Read(); err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	errors := r.Validate()
	// Should have validation errors for whitespace-only fields
	if len(errors) == 0 {
		t.Error("Expected validation errors")
	}
}

func TestComputeStats(t *testing.T) {
	csvData := "name,age\nAlice,30\nBob,25\nCharlie,35\n"
	r := NewReaderFrom(strings.NewReader(csvData))

	if err := r.Read(); err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	stats := r.ComputeStats()
	if len(stats) != 2 {
		t.Errorf("len(stats) = %d, want 2", len(stats))
	}

	if stats[0].Name != "name" {
		t.Errorf("stats[0].Name = %q, want %q", stats[0].Name, "name")
	}

	if stats[1].Type != "number" {
		t.Errorf("stats[1].Type = %q, want %q", stats[1].Type, "number")
	}
}

func TestWhere(t *testing.T) {
	csvData := "name,age,city\nAlice,30,NYC\nBob,25,LA\nCharlie,35,NYC\n"
	r := NewReaderFrom(strings.NewReader(csvData))

	if err := r.Read(); err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	filtered := r.Where("city = NYC")
	if filtered.NumRows() != 2 {
		t.Errorf("NumRows() = %d, want 2", filtered.NumRows())
	}
}

func TestJoin(t *testing.T) {
	csv1 := "id,name\n1,Alice\n2,Bob\n3,Charlie\n"
	csv2 := "id,score\n1,90\n2,85\n"

	r1 := NewReaderFrom(strings.NewReader(csv1))
	r2 := NewReaderFrom(strings.NewReader(csv2))

	r1.Read()
	r2.Read()

	joined := r1.Join(r2, "id", "inner")
	if joined.NumRows() != 2 {
		t.Errorf("NumRows() = %d, want 2", joined.NumRows())
	}

	if joined.NumCols() != 3 {
		t.Errorf("NumCols() = %d, want 3", joined.NumCols())
	}
}

func TestAggregate(t *testing.T) {
	csvData := "city,age\nNYC,30\nLA,25\nNYC,35\nLA,40\n"
	r := NewReaderFrom(strings.NewReader(csvData))

	if err := r.Read(); err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	aggregated := r.Aggregate(
		[]string{"city"},
		[]AggFunc{
			{Column: "age", Function: "count", Alias: "count"},
			{Column: "age", Function: "avg", Alias: "avg_age"},
		},
	)

	if aggregated.NumRows() != 2 {
		t.Errorf("NumRows() = %d, want 2", aggregated.NumRows())
	}

	if aggregated.NumCols() != 3 {
		t.Errorf("NumCols() = %d, want 3", aggregated.NumCols())
	}
}

func TestDetectDelimiter(t *testing.T) {
	tests := []struct {
		name string
		data string
		want rune
	}{
		{"comma", "a,b,c\n1,2,3\n", ','},
		{"tab", "a\tb\tc\n1\t2\t3\n", '\t'},
		{"semicolon", "a;b;c\n1;2;3\n", ';'},
		{"pipe", "a|b|c\n1|2|3\n", '|'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectDelimiter([]byte(tt.data))
			if got != tt.want {
				t.Errorf("DetectDelimiter() = %q, want %q", string(got), string(tt.want))
			}
		})
	}
}

func TestAddColumn(t *testing.T) {
	csvData := "name,age\nAlice,30\nBob,25\n"
	r := NewReaderFrom(strings.NewReader(csvData))

	if err := r.Read(); err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	added := r.AddColumn("status", "active")
	if added.NumCols() != 3 {
		t.Errorf("NumCols() = %d, want 3", added.NumCols())
	}
	if added.Header()[2] != "status" {
		t.Errorf("Header()[2] = %q, want %q", added.Header()[2], "status")
	}
	if added.Records()[0][2] != "active" {
		t.Errorf("Records()[0][2] = %q, want %q", added.Records()[0][2], "active")
	}
}

func TestDropColumns(t *testing.T) {
	csvData := "name,age,city\nAlice,30,NYC\nBob,25,LA\n"
	r := NewReaderFrom(strings.NewReader(csvData))

	if err := r.Read(); err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	dropped := r.DropColumns([]string{"age"})
	if dropped.NumCols() != 2 {
		t.Errorf("NumCols() = %d, want 2", dropped.NumCols())
	}
	if dropped.Header()[0] != "name" {
		t.Errorf("Header()[0] = %q, want %q", dropped.Header()[0], "name")
	}
}

func TestRenameColumn(t *testing.T) {
	csvData := "name,age\nAlice,30\n"
	r := NewReaderFrom(strings.NewReader(csvData))

	if err := r.Read(); err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	renamed := r.RenameColumn("age", "years")
	if renamed.Header()[1] != "years" {
		t.Errorf("Header()[1] = %q, want %q", renamed.Header()[1], "years")
	}
}

func TestToMaps(t *testing.T) {
	csvData := "name,age\nAlice,30\nBob,25\n"
	r := NewReaderFrom(strings.NewReader(csvData))

	if err := r.Read(); err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	maps := r.ToMaps()
	if len(maps) != 2 {
		t.Errorf("len(maps) = %d, want 2", len(maps))
	}
	if maps[0].Fields["name"] != "Alice" {
		t.Errorf("maps[0].Fields[name] = %q, want %q", maps[0].Fields["name"], "Alice")
	}
}

func TestGetColumn(t *testing.T) {
	csvData := "name,age\nAlice,30\nBob,25\n"
	r := NewReaderFrom(strings.NewReader(csvData))

	if err := r.Read(); err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	cols, err := r.GetColumn("name")
	if err != nil {
		t.Fatalf("GetColumn() error = %v", err)
	}

	if len(cols) != 2 {
		t.Errorf("len(cols) = %d, want 2", len(cols))
	}
	if cols[0] != "Alice" {
		t.Errorf("cols[0] = %q, want %q", cols[0], "Alice")
	}
}

func TestJoinLeft(t *testing.T) {
	csv1 := "id,name\n1,Alice\n2,Bob\n3,Charlie\n"
	csv2 := "id,score\n1,90\n2,85\n"

	r1 := NewReaderFrom(strings.NewReader(csv1))
	r2 := NewReaderFrom(strings.NewReader(csv2))

	r1.Read()
	r2.Read()

	joined := r1.Join(r2, "id", "left")
	if joined.NumRows() != 3 {
		t.Errorf("NumRows() = %d, want 3", joined.NumRows())
	}
}

func TestJoinFull(t *testing.T) {
	csv1 := "id,name\n1,Alice\n2,Bob\n"
	csv2 := "id,score\n1,90\n3,75\n"

	r1 := NewReaderFrom(strings.NewReader(csv1))
	r2 := NewReaderFrom(strings.NewReader(csv2))

	r1.Read()
	r2.Read()

	joined := r1.Join(r2, "id", "full")
	if joined.NumRows() != 3 {
		t.Errorf("NumRows() = %d, want 3", joined.NumRows())
	}
}
