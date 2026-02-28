package csv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadFile(t *testing.T) {
	// Create a temporary CSV file
	tmpDir, err := os.MkdirTemp("", "csvmap-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	csvContent := `Name,Age,City
Alice,30,New York
Bob,25,Los Angeles
Charlie,35,Chicago`

	path := filepath.Join(tmpDir, "test.csv")
	if err := os.WriteFile(path, []byte(csvContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Test with header
	hasHeader := true
	data, err := ReadFile(path, &hasHeader)
	if err != nil {
		t.Fatal(err)
	}

	if data.ColumnCount() != 3 {
		t.Errorf("expected 3 columns, got %d", data.ColumnCount())
	}

	if data.RowCount() != 3 {
		t.Errorf("expected 3 rows, got %d", data.RowCount())
	}

	if data.Headers[0] != "Name" {
		t.Errorf("expected header 'Name', got %q", data.Headers[0])
	}

	// Test column values
	names := data.ColumnValues(0)
	if len(names) != 3 || names[0] != "Alice" {
		t.Errorf("unexpected column values: %v", names)
	}
}

func TestUniqueColumnValues(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "csvmap-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	csvContent := `Category
Food
Shopping
Food
Entertainment
Shopping`

	path := filepath.Join(tmpDir, "test.csv")
	if err := os.WriteFile(path, []byte(csvContent), 0644); err != nil {
		t.Fatal(err)
	}

	hasHeader := true
	data, err := ReadFile(path, &hasHeader)
	if err != nil {
		t.Fatal(err)
	}

	unique := data.UniqueColumnValues(0)
	if len(unique) != 3 {
		t.Errorf("expected 3 unique values, got %d: %v", len(unique), unique)
	}
}

func TestTransformColumn(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "csvmap-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	csvContent := `Value
hello
world`

	path := filepath.Join(tmpDir, "test.csv")
	if err := os.WriteFile(path, []byte(csvContent), 0644); err != nil {
		t.Fatal(err)
	}

	hasHeader := true
	data, err := ReadFile(path, &hasHeader)
	if err != nil {
		t.Fatal(err)
	}

	// Transform to uppercase
	data.TransformColumn(0, func(s string) string {
		return s + "!"
	})

	values := data.ColumnValues(0)
	if values[0] != "hello!" || values[1] != "world!" {
		t.Errorf("unexpected transformed values: %v", values)
	}
}

func TestReorderColumns(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "csvmap-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	csvContent := `A,B,C,D
1,2,3,4
5,6,7,8`

	path := filepath.Join(tmpDir, "test.csv")
	if err := os.WriteFile(path, []byte(csvContent), 0644); err != nil {
		t.Fatal(err)
	}

	hasHeader := true
	data, err := ReadFile(path, &hasHeader)
	if err != nil {
		t.Fatal(err)
	}

	// Reorder: D, B, A, C (indices 3, 1, 0, 2)
	data.ReorderColumns([]int{3, 1, 0, 2})

	expectedHeaders := []string{"D", "B", "A", "C"}
	for i, h := range expectedHeaders {
		if data.Headers[i] != h {
			t.Errorf("header[%d] = %s, want %s", i, data.Headers[i], h)
		}
	}

	expectedRow0 := []string{"4", "2", "1", "3"}
	for i, v := range expectedRow0 {
		if data.Rows[0][i] != v {
			t.Errorf("row[0][%d] = %s, want %s", i, data.Rows[0][i], v)
		}
	}
}

func TestDropRows(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "csvmap-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	csvContent := `Name,Age
Alice,30
Bob,25
Charlie,35
Diana,28`

	path := filepath.Join(tmpDir, "test.csv")
	if err := os.WriteFile(path, []byte(csvContent), 0644); err != nil {
		t.Fatal(err)
	}

	hasHeader := true
	data, err := ReadFile(path, &hasHeader)
	if err != nil {
		t.Fatal(err)
	}

	// Drop rows 1 (Bob) and 3 (Diana)
	data.DropRows(map[int]bool{1: true, 3: true})

	if data.RowCount() != 2 {
		t.Errorf("expected 2 rows, got %d", data.RowCount())
	}

	if data.Rows[0][0] != "Alice" {
		t.Errorf("expected first row to be Alice, got %s", data.Rows[0][0])
	}

	if data.Rows[1][0] != "Charlie" {
		t.Errorf("expected second row to be Charlie, got %s", data.Rows[1][0])
	}
}

func TestDropColumns(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "csvmap-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	csvContent := `Name,Age,City,Country
Alice,30,New York,USA
Bob,25,London,UK`

	path := filepath.Join(tmpDir, "test.csv")
	if err := os.WriteFile(path, []byte(csvContent), 0644); err != nil {
		t.Fatal(err)
	}

	hasHeader := true
	data, err := ReadFile(path, &hasHeader)
	if err != nil {
		t.Fatal(err)
	}

	// Drop columns 1 (Age) and 3 (Country)
	data.DropColumns([]int{1, 3})

	if data.ColumnCount() != 2 {
		t.Errorf("expected 2 columns, got %d", data.ColumnCount())
	}

	if data.Headers[0] != "Name" || data.Headers[1] != "City" {
		t.Errorf("unexpected headers: %v", data.Headers)
	}

	if data.Rows[0][0] != "Alice" || data.Rows[0][1] != "New York" {
		t.Errorf("unexpected row 0: %v", data.Rows[0])
	}

	if data.Rows[1][0] != "Bob" || data.Rows[1][1] != "London" {
		t.Errorf("unexpected row 1: %v", data.Rows[1])
	}
}
