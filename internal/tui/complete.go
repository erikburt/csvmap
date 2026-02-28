package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// CompleteModel handles the completion view.
type CompleteModel struct {
	outputPath string
	done       bool
}

// CompleteResult contains the result of the complete view.
type CompleteResult struct {
	Done bool
}

// NewCompleteModel creates a new completion model.
func NewCompleteModel(outputPath string) *CompleteModel {
	return &CompleteModel{
		outputPath: outputPath,
	}
}

// Update handles input for the completion view.
func (m *CompleteModel) Update(msg tea.Msg) CompleteResult {
	if m.done {
		return CompleteResult{Done: true}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "q", "esc":
			m.done = true
			return CompleteResult{Done: true}
		}
	}

	return CompleteResult{Done: false}
}

// View renders the completion view.
func (m *CompleteModel) View() string {
	var s strings.Builder

	s.WriteString(successStyle.Render("Complete!"))
	s.WriteString("\n\n")

	s.WriteString("Output written to:\n")
	s.WriteString(fmt.Sprintf("  %s\n", highlightStyle.Render(m.outputPath)))
	s.WriteString("\n")

	s.WriteString(helpStyle.Render("Press Enter or Q to exit"))

	return s.String()
}
