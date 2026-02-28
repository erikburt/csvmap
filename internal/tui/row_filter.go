package tui

import (
	"fmt"
	"strings"

	"csvmap/internal/csv"

	tea "github.com/charmbracelet/bubbletea"
)

// RowFilterModel handles interactive row filtering.
type RowFilterModel struct {
	data        *csv.Data
	columnIndex int
	currentRow  int
	rowsToDrop  map[int]bool
	done        bool
	cursor      int // 0 = keep, 1 = drop
}

// RowFilterResult contains the result of row filtering.
type RowFilterResult struct {
	Done       bool
	RowsToDrop map[int]bool
}

// NewRowFilterModel creates a new row filter model.
func NewRowFilterModel(data *csv.Data, columnIndex int) *RowFilterModel {
	return &RowFilterModel{
		data:        data,
		columnIndex: columnIndex,
		currentRow:  0,
		rowsToDrop:  make(map[int]bool),
		cursor:      0, // Default to "Keep"
	}
}

// Update handles input for row filtering.
func (m *RowFilterModel) Update(msg tea.Msg) RowFilterResult {
	if m.done {
		return RowFilterResult{Done: true, RowsToDrop: m.rowsToDrop}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k", "left", "h":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j", "right", "l":
			if m.cursor < 1 {
				m.cursor++
			}
		case "enter", " ":
			if m.cursor == 1 {
				// Drop this row
				m.rowsToDrop[m.currentRow] = true
			}
			// Move to next row
			m.currentRow++
			if m.currentRow >= m.data.RowCount() {
				m.done = true
				return RowFilterResult{Done: true, RowsToDrop: m.rowsToDrop}
			}
			m.cursor = 0 // Reset to "Keep" for next row
		case "d":
			// Quick drop
			m.rowsToDrop[m.currentRow] = true
			m.currentRow++
			if m.currentRow >= m.data.RowCount() {
				m.done = true
				return RowFilterResult{Done: true, RowsToDrop: m.rowsToDrop}
			}
			m.cursor = 0
		case "s", "n":
			// Quick skip/keep (next)
			m.currentRow++
			if m.currentRow >= m.data.RowCount() {
				m.done = true
				return RowFilterResult{Done: true, RowsToDrop: m.rowsToDrop}
			}
			m.cursor = 0
		case "q", "esc":
			// Finish early, keep remaining rows
			m.done = true
			return RowFilterResult{Done: true, RowsToDrop: m.rowsToDrop}
		}
	}

	return RowFilterResult{Done: false}
}

// View renders the row filter view.
func (m *RowFilterModel) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Filter Rows"))
	s.WriteString("\n\n")

	// Progress
	s.WriteString(dimStyle.Render(fmt.Sprintf("Row %d of %d | Marked for removal: %d",
		m.currentRow+1, m.data.RowCount(), len(m.rowsToDrop))))
	s.WriteString("\n\n")

	// Show current row
	row := m.data.GetRow(m.currentRow)
	if row == nil {
		s.WriteString("No more rows.\n")
		return s.String()
	}

	s.WriteString(subtitleStyle.Render("Current row:"))
	s.WriteString("\n")

	// Display row data with all columns
	for i, header := range m.data.Headers {
		value := ""
		if i < len(row) {
			value = row[i]
		}
		// Truncate long values
		if len(value) > 50 {
			value = value[:47] + "..."
		}
		if i == m.columnIndex {
			s.WriteString(fmt.Sprintf("  %s: %s\n", highlightStyle.Render(header), highlightStyle.Render(value)))
		} else {
			s.WriteString(fmt.Sprintf("  %s: %s\n", dimStyle.Render(header), value))
		}
	}

	s.WriteString("\n")

	// Action options
	keepStyle := dimStyle
	dropStyle := dimStyle
	keepCursor := "  "
	dropCursor := "  "

	if m.cursor == 0 {
		keepCursor = selectedStyle.Render("▸ ")
		keepStyle = selectedStyle
	} else {
		dropCursor = selectedStyle.Render("▸ ")
		dropStyle = errorStyle
	}

	s.WriteString(keepCursor + keepStyle.Render("Keep") + "  ")
	s.WriteString(dropCursor + dropStyle.Render("Drop") + "\n")

	s.WriteString(helpStyle.Render("\n←/→ or ↑/↓ to select • Enter/Space to confirm • d=drop • s/n=skip • q/Esc=finish"))

	return s.String()
}
