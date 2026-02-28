package csv

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
)

// WriteFile writes the CSV data to the specified path.
// It uses atomic write (temp file + rename) to ensure file safety.
func WriteFile(data *Data, outputPath string) error {
	// Create temp file in the same directory for atomic rename
	dir := filepath.Dir(outputPath)
	tempFile, err := os.CreateTemp(dir, "csvmap-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()

	// Ensure cleanup on error
	success := false
	defer func() {
		if !success {
			tempFile.Close()
			os.Remove(tempPath)
		}
	}()

	writer := csv.NewWriter(tempFile)

	// Write header if present
	if data.HasHeader {
		if err := writer.Write(data.Headers); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}
	}

	// Write all rows
	for _, row := range data.Rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("failed to flush CSV: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, outputPath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	success = true
	return nil
}

// FileExists checks if a file exists at the given path.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// TransformColumn applies a transformation function to a specific column.
func (d *Data) TransformColumn(colIndex int, transform func(string) string) {
	for i, row := range d.Rows {
		if colIndex < len(row) {
			d.Rows[i][colIndex] = transform(row[colIndex])
		}
	}
}

// Clone creates a deep copy of the CSV data.
func (d *Data) Clone() *Data {
	clone := &Data{
		FilePath:  d.FilePath,
		HasHeader: d.HasHeader,
		Headers:   make([]string, len(d.Headers)),
		Rows:      make([][]string, len(d.Rows)),
		Comma:     d.Comma,
		Comment:   d.Comment,
	}

	copy(clone.Headers, d.Headers)

	for i, row := range d.Rows {
		clone.Rows[i] = make([]string, len(row))
		copy(clone.Rows[i], row)
	}

	return clone
}

// ReorderColumns reorders columns according to the given order.
// order[i] = j means the column at position i in output comes from original column j.
func (d *Data) ReorderColumns(order []int) {
	if len(order) != len(d.Headers) {
		return
	}

	// Check if reordering is needed
	needsReorder := false
	for i, idx := range order {
		if i != idx {
			needsReorder = true
			break
		}
	}
	if !needsReorder {
		return
	}

	// Reorder headers
	newHeaders := make([]string, len(d.Headers))
	for i, origIdx := range order {
		newHeaders[i] = d.Headers[origIdx]
	}
	d.Headers = newHeaders

	// Reorder each row
	for i, row := range d.Rows {
		newRow := make([]string, len(row))
		for j, origIdx := range order {
			if origIdx < len(row) {
				newRow[j] = row[origIdx]
			}
		}
		d.Rows[i] = newRow
	}
}

// DropRows removes the specified rows from the data.
func (d *Data) DropRows(rowsToDrop map[int]bool) {
	if len(rowsToDrop) == 0 {
		return
	}

	newRows := make([][]string, 0, len(d.Rows)-len(rowsToDrop))
	for i, row := range d.Rows {
		if !rowsToDrop[i] {
			newRows = append(newRows, row)
		}
	}
	d.Rows = newRows
}

// DropColumns removes the specified columns from the data.
// Indices can be in any order; they will be processed correctly.
func (d *Data) DropColumns(indices []int) {
	if len(indices) == 0 {
		return
	}

	// Create a set of indices to drop for O(1) lookup
	dropSet := make(map[int]bool)
	for _, idx := range indices {
		dropSet[idx] = true
	}

	// Filter headers
	newHeaders := make([]string, 0, len(d.Headers)-len(indices))
	for i, h := range d.Headers {
		if !dropSet[i] {
			newHeaders = append(newHeaders, h)
		}
	}
	d.Headers = newHeaders

	// Filter each row
	for i, row := range d.Rows {
		newRow := make([]string, 0, len(row)-len(indices))
		for j, cell := range row {
			if !dropSet[j] {
				newRow = append(newRow, cell)
			}
		}
		d.Rows[i] = newRow
	}
}
