package tui

import (
	"fmt"
	"strings"

	"github.com/erikburt/csvmap/internal/config"
	"github.com/erikburt/csvmap/internal/csv"

	tea "github.com/charmbracelet/bubbletea"
)

// ReviewModel handles the review and confirmation view.
type ReviewModel struct {
	data        *csv.Data
	operations  []ColumnOperation
	columnOrder []int
	inputPath   string
	outputPath  string
	cursor      int
	done        bool
	confirmed   bool
	overwrite   bool
}

// ReviewResult contains the result of the review view.
type ReviewResult struct {
	Done      bool
	Confirmed bool
}

// NewReviewModel creates a new review model.
func NewReviewModel(data *csv.Data, operations []ColumnOperation, inputPath string, columnOrder []int) *ReviewModel {
	outputPath := config.OutputFilePath(inputPath)
	return &ReviewModel{
		data:        data,
		operations:  operations,
		columnOrder: columnOrder,
		inputPath:   inputPath,
		outputPath:  outputPath,
		overwrite:   csv.FileExists(outputPath),
	}
}

// Update handles input for the review view.
func (m *ReviewModel) Update(msg tea.Msg) ReviewResult {
	if m.done {
		return ReviewResult{Done: true, Confirmed: m.confirmed}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < 1 {
				m.cursor++
			}
		case "enter":
			m.done = true
			m.confirmed = m.cursor == 0
			return ReviewResult{Done: true, Confirmed: m.confirmed}
		case "esc":
			m.done = true
			m.confirmed = false
			return ReviewResult{Done: true, Confirmed: false}
		}
	}

	return ReviewResult{Done: false}
}

// View renders the review view.
func (m *ReviewModel) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Review & Confirm"))
	s.WriteString("\n\n")

	// Input/output files
	s.WriteString(subtitleStyle.Render("Input file:"))
	s.WriteString("\n")
	s.WriteString(fmt.Sprintf("  %s\n", m.inputPath))
	s.WriteString(subtitleStyle.Render("Output file:"))
	s.WriteString("\n")
	s.WriteString(fmt.Sprintf("  %s\n", m.outputPath))

	if m.overwrite {
		s.WriteString(errorStyle.Render("  (file exists - will be overwritten)"))
		s.WriteString("\n")
	}
	s.WriteString("\n")

	// Column order
	columnsReordered := false
	for i, idx := range m.columnOrder {
		if i != idx {
			columnsReordered = true
			break
		}
	}
	if columnsReordered {
		s.WriteString(subtitleStyle.Render("Column order:"))
		s.WriteString("\n")
		for _, idx := range m.columnOrder {
			s.WriteString(fmt.Sprintf("  %s\n", m.data.Headers[idx]))
		}
		s.WriteString("\n")
	}

	// Operations summary
	if len(m.operations) == 0 && !columnsReordered {
		s.WriteString(dimStyle.Render("No operations configured."))
		s.WriteString("\n")
	} else if len(m.operations) > 0 {
		s.WriteString(subtitleStyle.Render("Operations to apply:"))
		s.WriteString("\n\n")

		for _, op := range m.operations {
			s.WriteString(fmt.Sprintf("  Column: %s\n", highlightStyle.Render(op.ColumnName)))

			switch op.Operation {
			case OpDateReformat:
				s.WriteString("    Operation: Date Reformat\n")
				s.WriteString(fmt.Sprintf("    From: %s\n", op.InputDateFormat))
				s.WriteString(fmt.Sprintf("    To:   %s\n", op.OutputDateFormat))

			case OpStringMapping:
				s.WriteString("    Operation: String Mapping\n")
				if op.MappingFile != nil {
					s.WriteString(fmt.Sprintf("    Mapping file: %s\n", op.MappingFile.Name()))
				}
				s.WriteString(fmt.Sprintf("    Mapped: %d | Skipped: %d | Unknown: %d\n",
					op.MappedCount, op.SkippedCount, op.UnknownCount))

			case OpDropColumn:
				s.WriteString(errorStyle.Render("    Operation: Drop Column") + "\n")
				s.WriteString("    This column will be removed from the output\n")

			case OpFilterRows:
				s.WriteString("    Operation: Filter Rows\n")
				s.WriteString(fmt.Sprintf("    Rows to drop: %s\n", errorStyle.Render(fmt.Sprintf("%d", op.RowsToDropCount))))

			case OpInvertSign:
				s.WriteString("    Operation: Invert Sign\n")
				s.WriteString("    Flip positive/negative on all values\n")
			}
			s.WriteString("\n")
		}
	}

	// Confirm options
	s.WriteString("\n")

	options := []string{"Confirm and write output", "Go back to column selection"}

	for i, opt := range options {
		cursor := "  "
		style := dimStyle
		if m.cursor == i {
			cursor = selectedStyle.Render("▸ ")
			if i == 0 {
				style = successStyle
			} else {
				style = selectedStyle
			}
		}
		s.WriteString(cursor + style.Render(opt) + "\n")
	}

	s.WriteString(helpStyle.Render("\n↑/↓ to navigate • Enter to select • Esc to go back"))

	return s.String()
}
