package tui

import (
	"fmt"
	"strings"

	"github.com/erikburt/csvmap/internal/csv"

	tea "github.com/charmbracelet/bubbletea"
)

// PreviewModel handles the file preview and header confirmation view.
type PreviewModel struct {
	data         *csv.Data
	askHeader    bool
	headerChoice int // 0 = yes, 1 = no
	done         bool
	hasHeader    *bool
}

// PreviewResult contains the result of the preview view.
type PreviewResult struct {
	Done      bool
	HasHeader *bool
}

// NewPreviewModel creates a new preview model.
func NewPreviewModel(data *csv.Data, askHeader bool) *PreviewModel {
	return &PreviewModel{
		data:         data,
		askHeader:    askHeader,
		headerChoice: 0,
	}
}

// Update handles input for the preview view.
func (m *PreviewModel) Update(msg tea.Msg) PreviewResult {
	if m.done {
		return PreviewResult{Done: true, HasHeader: m.hasHeader}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.askHeader && m.headerChoice > 0 {
				m.headerChoice--
			}
		case "down", "j":
			if m.askHeader && m.headerChoice < 1 {
				m.headerChoice++
			}
		case "enter":
			if m.askHeader {
				hasHeader := m.headerChoice == 0
				m.hasHeader = &hasHeader
			}
			m.done = true
			return PreviewResult{Done: true, HasHeader: m.hasHeader}
		}
	}

	return PreviewResult{Done: false}
}

// View renders the preview view.
func (m *PreviewModel) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("CSV File Preview"))
	s.WriteString("\n\n")

	// Show file path
	s.WriteString(dimStyle.Render("File: "))
	s.WriteString(m.data.FilePath)
	s.WriteString("\n")
	s.WriteString(dimStyle.Render(fmt.Sprintf("Rows: %d | Columns: %d", m.data.RowCount(), m.data.ColumnCount())))
	s.WriteString("\n\n")

	// Render table preview
	s.WriteString(m.renderTable())
	s.WriteString("\n")

	if m.askHeader {
		s.WriteString("\n")
		s.WriteString("Does the first row contain column headers?\n\n")

		yes := "  Yes, the first row is a header"
		no := "  No, all rows are data"

		if m.headerChoice == 0 {
			yes = selectedStyle.Render("▸ Yes, the first row is a header")
		} else {
			no = selectedStyle.Render("▸ No, all rows are data")
		}

		s.WriteString(yes + "\n")
		s.WriteString(no + "\n")
	}

	s.WriteString(helpStyle.Render("\n↑/↓ to select • Enter to confirm"))

	return s.String()
}

func (m *PreviewModel) renderTable() string {
	previewRows := m.data.PreviewRows(5)
	if len(previewRows) == 0 {
		return "(no data)"
	}

	// Calculate column widths
	colCount := m.data.ColumnCount()
	widths := make([]int, colCount)

	// Check header widths
	for i, h := range m.data.Headers {
		if i < colCount {
			headerText := fmt.Sprintf("[%d] %s", i, h)
			if len(headerText) > widths[i] {
				widths[i] = len(headerText)
			}
		}
	}

	// Check data widths
	for _, row := range previewRows {
		for i, cell := range row {
			if i < colCount && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Cap widths at reasonable maximum
	for i := range widths {
		if widths[i] > 30 {
			widths[i] = 30
		}
		if widths[i] < 3 {
			widths[i] = 3
		}
	}

	var table strings.Builder

	// Header row with column indices
	for i, h := range m.data.Headers {
		if i < colCount {
			header := fmt.Sprintf("[%d] %s", i, h)
			if len(header) > widths[i] {
				header = header[:widths[i]-1] + "…"
			}
			table.WriteString(highlightStyle.Render(padRight(header, widths[i])))
			if i < colCount-1 {
				table.WriteString("  ")
			}
		}
	}
	table.WriteString("\n")

	// Separator
	for i := 0; i < colCount; i++ {
		table.WriteString(strings.Repeat("─", widths[i]))
		if i < colCount-1 {
			table.WriteString("  ")
		}
	}
	table.WriteString("\n")

	// Data rows
	for _, row := range previewRows {
		for i := 0; i < colCount; i++ {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			if len(cell) > widths[i] {
				cell = cell[:widths[i]-1] + "…"
			}
			table.WriteString(padRight(cell, widths[i]))
			if i < colCount-1 {
				table.WriteString("  ")
			}
		}
		table.WriteString("\n")
	}

	if m.data.RowCount() > 5 {
		table.WriteString(dimStyle.Render(fmt.Sprintf("... and %d more rows", m.data.RowCount()-5)))
	}

	return borderStyle.Render(table.String())
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
