package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// OperationSelectModel handles operation type selection.
type OperationSelectModel struct {
	columnName string
	cursor     int
	done       bool
	operation  OperationType
}

// OperationSelectResult contains the result of operation selection.
type OperationSelectResult struct {
	Done      bool
	Operation OperationType
}

// NewOperationSelectModel creates a new operation selection model.
func NewOperationSelectModel(columnName string) *OperationSelectModel {
	return &OperationSelectModel{
		columnName: columnName,
		cursor:     0,
	}
}

// Update handles input for operation selection.
func (m *OperationSelectModel) Update(msg tea.Msg) OperationSelectResult {
	if m.done {
		return OperationSelectResult{Done: true, Operation: m.operation}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < 4 { // 5 options: date, mapping, drop, filter, cancel
				m.cursor++
			}
		case "enter":
			m.done = true
			switch m.cursor {
			case 0:
				m.operation = OpDateReformat
			case 1:
				m.operation = OpStringMapping
			case 2:
				m.operation = OpDropColumn
			case 3:
				m.operation = OpFilterRows
			case 4:
				m.operation = OpNone // Cancel
			}
			return OperationSelectResult{Done: true, Operation: m.operation}
		case "esc":
			m.done = true
			m.operation = OpNone
			return OperationSelectResult{Done: true, Operation: OpNone}
		}
	}

	return OperationSelectResult{Done: false}
}

// View renders the operation selection view.
func (m *OperationSelectModel) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Select Operation"))
	s.WriteString("\n\n")
	s.WriteString(fmt.Sprintf("Column: %s\n\n", highlightStyle.Render(m.columnName)))

	options := []struct {
		name string
		desc string
	}{
		{"Date Reformat", "Convert date values to a different format"},
		{"String Mapping", "Map values using pattern matching (e.g., merchant names)"},
		{"Drop Column", "Remove this column from the output"},
		{"Filter Rows", "Interactively review and drop rows"},
		{"Cancel", "Go back to column selection"},
	}

	for i, opt := range options {
		cursor := "  "
		style := dimStyle
		if m.cursor == i {
			cursor = selectedStyle.Render("▸ ")
			style = selectedStyle
		}

		s.WriteString(cursor + style.Render(opt.name) + "\n")
		if m.cursor == i {
			s.WriteString("    " + dimStyle.Render(opt.desc) + "\n")
		}
	}

	s.WriteString(helpStyle.Render("\n↑/↓ to navigate • Enter to select • Esc to cancel"))

	return s.String()
}
