package tui

import (
	"fmt"
	"strings"

	"github.com/erikburt/csvmap/internal/csv"

	tea "github.com/charmbracelet/bubbletea"
)

// ColumnSelectModel handles column selection.
type ColumnSelectModel struct {
	data           *csv.Data
	operations     []ColumnOperation
	columnOrder    []int // Current column order (display pos -> original index)
	cursor         int
	done           bool
	selectedColumn int
	quickDrop      bool
	removeOpIndex  int // -1 if not removing
	// Move mode
	movingColumnPos int // Position in columnOrder being moved (-1 if not moving)
}

// ColumnSelectResult contains the result of column selection.
type ColumnSelectResult struct {
	Done           bool
	SelectedColumn int    // -1 if user chose to finish or removing op
	QuickDrop      bool   // true if user pressed x on a column
	RemoveOpIndex  int    // >= 0 if removing an operation
	ColumnOrder    []int  // Updated column order
}

// NewColumnSelectModel creates a new column selection model.
func NewColumnSelectModel(data *csv.Data, operations []ColumnOperation, columnOrder []int) *ColumnSelectModel {
	// Start cursor at first column (after operations section)
	startCursor := len(operations)
	if len(operations) == 0 {
		startCursor = 0
	}

	// Copy column order
	order := make([]int, len(columnOrder))
	copy(order, columnOrder)

	return &ColumnSelectModel{
		data:            data,
		operations:      operations,
		columnOrder:     order,
		cursor:          startCursor,
		selectedColumn:  -1,
		removeOpIndex:   -1,
		movingColumnPos: -1,
	}
}

// cursorSection returns which section the cursor is in and the index within that section.
// Returns: section (0=operations, 1=columns, 2=done), index within section
func (m *ColumnSelectModel) cursorSection() (int, int) {
	opCount := len(m.operations)
	colCount := len(m.columnOrder)

	if opCount > 0 && m.cursor < opCount {
		return 0, m.cursor // In operations section
	} else if m.cursor < opCount+colCount {
		return 1, m.cursor - opCount // In columns section (position in columnOrder)
	}
	return 2, 0 // Done option
}

// totalOptions returns the total number of navigable options.
func (m *ColumnSelectModel) totalOptions() int {
	return len(m.operations) + len(m.columnOrder) + 1 // +1 for Done
}

// Update handles input for column selection.
func (m *ColumnSelectModel) Update(msg tea.Msg) ColumnSelectResult {
	if m.done {
		return ColumnSelectResult{
			Done:           true,
			SelectedColumn: m.selectedColumn,
			QuickDrop:      m.quickDrop,
			RemoveOpIndex:  m.removeOpIndex,
			ColumnOrder:    m.columnOrder,
		}
	}

	totalOpts := m.totalOptions()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.movingColumnPos >= 0 {
				// Moving a column up
				if m.movingColumnPos > 0 {
					// Swap with previous
					m.columnOrder[m.movingColumnPos], m.columnOrder[m.movingColumnPos-1] =
						m.columnOrder[m.movingColumnPos-1], m.columnOrder[m.movingColumnPos]
					m.movingColumnPos--
					m.cursor--
				}
			} else {
				if m.cursor > 0 {
					m.cursor--
				}
			}
		case "down", "j":
			if m.movingColumnPos >= 0 {
				// Moving a column down
				if m.movingColumnPos < len(m.columnOrder)-1 {
					// Swap with next
					m.columnOrder[m.movingColumnPos], m.columnOrder[m.movingColumnPos+1] =
						m.columnOrder[m.movingColumnPos+1], m.columnOrder[m.movingColumnPos]
					m.movingColumnPos++
					m.cursor++
				}
			} else {
				if m.cursor < totalOpts-1 {
					m.cursor++
				}
			}
		case " ":
			// Toggle move mode for columns
			section, idx := m.cursorSection()
			if section == 1 {
				if m.movingColumnPos >= 0 {
					// Drop the column at current position
					m.movingColumnPos = -1
				} else {
					// Pick up this column
					m.movingColumnPos = idx
				}
			}
		case "enter":
			if m.movingColumnPos >= 0 {
				// Drop column if in move mode
				m.movingColumnPos = -1
				return ColumnSelectResult{Done: false}
			}
			section, idx := m.cursorSection()
			m.done = true
			switch section {
			case 0: // Operations - enter does nothing useful
				m.done = false
			case 1: // Columns
				m.selectedColumn = m.columnOrder[idx] // Return original column index
			case 2: // Done
				m.selectedColumn = -1
			}
			if m.done {
				return ColumnSelectResult{
					Done:           true,
					SelectedColumn: m.selectedColumn,
					QuickDrop:      false,
					RemoveOpIndex:  -1,
					ColumnOrder:    m.columnOrder,
				}
			}
		case "x", "X":
			if m.movingColumnPos >= 0 {
				// Cancel move mode
				m.movingColumnPos = -1
				return ColumnSelectResult{Done: false}
			}
			section, idx := m.cursorSection()
			m.done = true
			switch section {
			case 0: // Operations - remove this operation
				m.removeOpIndex = idx
				return ColumnSelectResult{
					Done:          true,
					RemoveOpIndex: idx,
					ColumnOrder:   m.columnOrder,
				}
			case 1: // Columns - quick drop
				m.selectedColumn = m.columnOrder[idx] // Return original column index
				m.quickDrop = true
				return ColumnSelectResult{
					Done:           true,
					SelectedColumn: m.selectedColumn,
					QuickDrop:      true,
					RemoveOpIndex:  -1,
					ColumnOrder:    m.columnOrder,
				}
			case 2: // Done - ignore x
				m.done = false
			}
		case "esc":
			if m.movingColumnPos >= 0 {
				// Cancel move mode
				m.movingColumnPos = -1
			}
		}
	}

	return ColumnSelectResult{Done: false}
}

// View renders the column selection view.
func (m *ColumnSelectModel) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Select Column to Transform"))
	s.WriteString("\n\n")

	opCount := len(m.operations)
	cursorPos := 0

	// Show existing operations (navigable)
	if opCount > 0 {
		s.WriteString(subtitleStyle.Render("Configured operations:"))
		s.WriteString(" " + dimStyle.Render("(x to remove)"))
		s.WriteString("\n")
		for i, op := range m.operations {
			cursor := "  "
			style := dimStyle
			if m.cursor == cursorPos {
				cursor = selectedStyle.Render("▸ ")
				style = selectedStyle
			}

			var opType string
			switch op.Operation {
			case OpDateReformat:
				opType = "Date Reformat"
			case OpStringMapping:
				opType = "String Mapping"
			case OpDropColumn:
				opType = "Drop Column"
			case OpFilterRows:
				opType = fmt.Sprintf("Filter Rows (%d to drop)", op.RowsToDropCount)
			case OpInvertSign:
				opType = "Invert Sign"
			}
			_ = i
			s.WriteString(fmt.Sprintf("%s%s: %s\n", cursor, style.Render(op.ColumnName), highlightStyle.Render(opType)))
			cursorPos++
		}
		s.WriteString("\n")
	}

	s.WriteString(subtitleStyle.Render("Columns:"))
	if m.movingColumnPos >= 0 {
		s.WriteString(" " + highlightStyle.Render("(moving - ↑/↓ to reorder, space/enter to drop)"))
	} else {
		s.WriteString(" " + dimStyle.Render("(x=drop, space=move)"))
	}
	s.WriteString("\n\n")

	// List columns in current order
	for pos, origIdx := range m.columnOrder {
		cursor := "  "
		style := dimStyle
		isMoving := m.movingColumnPos == pos

		if m.cursor == cursorPos {
			cursor = selectedStyle.Render("▸ ")
			style = selectedStyle
		}

		if isMoving {
			style = highlightStyle
			cursor = highlightStyle.Render("≡ ")
		}

		header := m.data.Headers[origIdx]

		// Check if column already has an operation
		hasOp := false
		for _, op := range m.operations {
			if op.ColumnIndex == origIdx {
				hasOp = true
				break
			}
		}

		line := fmt.Sprintf("[%d] %s", origIdx, header)
		if hasOp {
			line += " (configured)"
			if !isMoving {
				style = dimStyle
			}
		}

		s.WriteString(cursor + style.Render(line) + "\n")

		// Show sample values for this column
		samples := m.getSampleValues(origIdx, 3)
		if len(samples) > 0 {
			s.WriteString("      " + dimStyle.Render("samples: "+samples) + "\n")
		}
		cursorPos++
	}

	// Done option
	s.WriteString("\n")
	cursor := "  "
	style := successStyle
	if m.cursor == cursorPos {
		cursor = selectedStyle.Render("▸ ")
		style = selectedStyle
	}
	s.WriteString(cursor + style.Render("Done - proceed to review") + "\n")

	s.WriteString(helpStyle.Render("\n↑/↓ navigate • Enter select • x drop/remove • space move"))

	return s.String()
}

// getSampleValues returns a string with sample values from a column.
func (m *ColumnSelectModel) getSampleValues(colIndex int, maxSamples int) string {
	values := m.data.ColumnValues(colIndex)
	var samples []string
	seen := make(map[string]bool)

	for _, v := range values {
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true

		// Truncate long values
		display := v
		if len(display) > 25 {
			display = display[:22] + "..."
		}
		samples = append(samples, "\""+display+"\"")

		if len(samples) >= maxSamples {
			break
		}
	}

	if len(samples) == 0 {
		return ""
	}

	result := strings.Join(samples, ", ")
	if len(values) > len(samples) {
		result += ", ..."
	}
	return result
}
