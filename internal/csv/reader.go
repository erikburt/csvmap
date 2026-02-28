// Package csv provides CSV reading and writing functionality for csvmap.
package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

// Data represents a loaded CSV file with its contents and metadata.
type Data struct {
	FilePath  string
	HasHeader bool
	Headers   []string // Column names (from header row or generated)
	Rows      [][]string
	// Original reader settings for preservation
	Comma   rune
	Comment rune
}

// ReadFile reads a CSV file and returns the parsed data.
// hasHeader indicates whether the first row should be treated as headers.
// If hasHeader is nil, the caller should determine this later.
func ReadFile(path string, hasHeader *bool) (*Data, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Allow variable field counts
	reader.LazyQuotes = true    // Be lenient with quoting

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	data := &Data{
		FilePath: path,
		Comma:    ',',
	}

	// Determine column count from first row
	colCount := len(records[0])

	if hasHeader != nil && *hasHeader {
		data.HasHeader = true
		data.Headers = records[0]
		data.Rows = records[1:]
	} else if hasHeader != nil && !*hasHeader {
		data.HasHeader = false
		data.Headers = generateColumnNames(colCount)
		data.Rows = records
	} else {
		// Header status unknown - store all as rows, headers will be set later
		data.HasHeader = false
		data.Headers = generateColumnNames(colCount)
		data.Rows = records
	}

	return data, nil
}

// SetHasHeader updates the data to treat the first row as a header or not.
func (d *Data) SetHasHeader(hasHeader bool) {
	if hasHeader == d.HasHeader {
		return
	}

	if hasHeader && len(d.Rows) > 0 {
		// Convert first row to headers
		d.HasHeader = true
		d.Headers = d.Rows[0]
		d.Rows = d.Rows[1:]
	} else if !hasHeader && d.HasHeader {
		// Convert headers back to first row
		d.HasHeader = false
		d.Rows = append([][]string{d.Headers}, d.Rows...)
		d.Headers = generateColumnNames(len(d.Headers))
	}
}

// PreviewRows returns up to n rows for preview purposes.
func (d *Data) PreviewRows(n int) [][]string {
	if n > len(d.Rows) {
		n = len(d.Rows)
	}
	return d.Rows[:n]
}

// AllRows returns all rows including the header if present.
func (d *Data) AllRows() [][]string {
	return d.Rows
}

// ColumnCount returns the number of columns.
func (d *Data) ColumnCount() int {
	return len(d.Headers)
}

// RowCount returns the number of data rows (excluding header).
func (d *Data) RowCount() int {
	return len(d.Rows)
}

// ColumnValues returns all values in a specific column.
func (d *Data) ColumnValues(colIndex int) []string {
	if colIndex < 0 || colIndex >= d.ColumnCount() {
		return nil
	}

	values := make([]string, 0, len(d.Rows))
	for _, row := range d.Rows {
		if colIndex < len(row) {
			values = append(values, row[colIndex])
		} else {
			values = append(values, "")
		}
	}
	return values
}

// UniqueColumnValues returns unique non-empty values in a column.
func (d *Data) UniqueColumnValues(colIndex int) []string {
	seen := make(map[string]bool)
	var unique []string

	for _, row := range d.Rows {
		if colIndex < len(row) {
			val := row[colIndex]
			if val != "" && !seen[val] {
				seen[val] = true
				unique = append(unique, val)
			}
		}
	}
	return unique
}

// RowsWithValue returns all rows where the specified column has the given value.
func (d *Data) RowsWithValue(colIndex int, value string) []int {
	var rowIndices []int
	for i, row := range d.Rows {
		if colIndex < len(row) && row[colIndex] == value {
			rowIndices = append(rowIndices, i)
		}
	}
	return rowIndices
}

// GetRow returns a specific row by index.
func (d *Data) GetRow(index int) []string {
	if index < 0 || index >= len(d.Rows) {
		return nil
	}
	return d.Rows[index]
}

// ReadFileForPreview reads just enough of a file for preview purposes.
func ReadFileForPreview(path string, maxRows int) ([][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	var rows [][]string
	for i := 0; i < maxRows; i++ {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV: %w", err)
		}
		rows = append(rows, record)
	}

	return rows, nil
}

func generateColumnNames(count int) []string {
	names := make([]string, count)
	for i := 0; i < count; i++ {
		names[i] = fmt.Sprintf("Column %d", i)
	}
	return names
}
