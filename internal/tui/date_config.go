package tui

import (
	"fmt"
	"strings"

	"github.com/erikburt/csvmap/internal/dateformat"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// DateConfigPhase represents the current phase of date configuration.
type DateConfigPhase int

const (
	DatePhaseConfirmInput DateConfigPhase = iota
	DatePhaseSelectOutput
	DatePhaseCustomOutput
)

// DateConfigModel handles date format configuration.
type DateConfigModel struct {
	values       []string
	inferResult  *dateformat.InferResult
	phase        DateConfigPhase
	cursor       int
	inputFormat  string
	outputFormat string
	customInput  textinput.Model
	done         bool
	cancelled    bool
}

// DateConfigResult contains the result of date configuration.
type DateConfigResult struct {
	Done         bool
	Cancelled    bool
	InputFormat  string
	OutputFormat string
}

// NewDateConfigModel creates a new date configuration model.
func NewDateConfigModel(values []string) *DateConfigModel {
	inferResult := dateformat.InferFormat(values)

	customInput := textinput.New()
	customInput.Placeholder = "e.g., 2006-01-02"
	customInput.CharLimit = 50

	m := &DateConfigModel{
		values:      values,
		inferResult: inferResult,
		phase:       DatePhaseConfirmInput,
		customInput: customInput,
	}

	if inferResult != nil {
		m.inputFormat = inferResult.Format.Layout
	}

	return m
}

// Update handles input for date configuration.
func (m *DateConfigModel) Update(msg tea.Msg) DateConfigResult {
	if m.done {
		return DateConfigResult{
			Done:         true,
			Cancelled:    m.cancelled,
			InputFormat:  m.inputFormat,
			OutputFormat: m.outputFormat,
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.done = true
			m.cancelled = true
			return DateConfigResult{Done: true, Cancelled: true}
		}
	}

	switch m.phase {
	case DatePhaseConfirmInput:
		return m.updateConfirmInput(msg)
	case DatePhaseSelectOutput:
		return m.updateSelectOutput(msg)
	case DatePhaseCustomOutput:
		return m.updateCustomOutput(msg)
	}

	return DateConfigResult{Done: false}
}

func (m *DateConfigModel) updateConfirmInput(msg tea.Msg) DateConfigResult {
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
			if m.cursor == 0 {
				// Accept detected format
				m.phase = DatePhaseSelectOutput
				m.cursor = 0
			} else {
				// Custom format - would need text input
				// For simplicity, show common formats to choose from
				m.phase = DatePhaseSelectOutput
				m.cursor = 0
			}
		}
	}
	return DateConfigResult{Done: false}
}

func (m *DateConfigModel) updateSelectOutput(msg tea.Msg) DateConfigResult {
	formats := dateformat.GetFormatCheatsheet()
	maxCursor := len(formats) // +1 for custom option

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < maxCursor {
				m.cursor++
			}
		case "enter":
			if m.cursor < len(formats) {
				m.outputFormat = formats[m.cursor].Layout
				m.done = true
				return DateConfigResult{
					Done:         true,
					InputFormat:  m.inputFormat,
					OutputFormat: m.outputFormat,
				}
			} else {
				// Custom format
				m.phase = DatePhaseCustomOutput
				m.customInput.Focus()
			}
		}
	}
	return DateConfigResult{Done: false}
}

func (m *DateConfigModel) updateCustomOutput(msg tea.Msg) DateConfigResult {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			format := m.customInput.Value()
			if dateformat.ValidateLayout(format) {
				m.outputFormat = format
				m.done = true
				return DateConfigResult{
					Done:         true,
					InputFormat:  m.inputFormat,
					OutputFormat: m.outputFormat,
				}
			}
			// Invalid format, stay on this screen
		}
	}

	var cmd tea.Cmd
	m.customInput, cmd = m.customInput.Update(msg)
	_ = cmd

	return DateConfigResult{Done: false}
}

// View renders the date configuration view.
func (m *DateConfigModel) View() string {
	switch m.phase {
	case DatePhaseConfirmInput:
		return m.viewConfirmInput()
	case DatePhaseSelectOutput:
		return m.viewSelectOutput()
	case DatePhaseCustomOutput:
		return m.viewCustomOutput()
	}
	return ""
}

func (m *DateConfigModel) viewConfirmInput() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Date Format Configuration"))
	s.WriteString("\n\n")

	if m.inferResult != nil {
		s.WriteString(subtitleStyle.Render("Detected input format:"))
		s.WriteString("\n")
		s.WriteString(fmt.Sprintf("  %s (%s)\n",
			highlightStyle.Render(m.inferResult.Format.Layout),
			m.inferResult.Format.Description))
		s.WriteString(dimStyle.Render(fmt.Sprintf("  Confidence: %d/%d samples parsed\n",
			m.inferResult.SamplesParsed, m.inferResult.SamplesTotal)))

		if m.inferResult.Ambiguous {
			s.WriteString(errorStyle.Render("  Note: Multiple formats could match\n"))
		}
	} else {
		s.WriteString(errorStyle.Render("Could not detect date format automatically."))
		s.WriteString("\n")
	}

	// Show sample values
	s.WriteString("\n")
	s.WriteString(subtitleStyle.Render("Sample values:"))
	s.WriteString("\n")
	count := 0
	for _, v := range m.values {
		if v != "" && count < 3 {
			s.WriteString(fmt.Sprintf("  • %s\n", v))
			count++
		}
	}

	s.WriteString("\n")
	options := []string{"Accept detected format", "Choose different format"}

	for i, opt := range options {
		cursor := "  "
		style := dimStyle
		if m.cursor == i {
			cursor = selectedStyle.Render("▸ ")
			style = selectedStyle
		}
		s.WriteString(cursor + style.Render(opt) + "\n")
	}

	s.WriteString(helpStyle.Render("\n↑/↓ to navigate • Enter to select • Esc to cancel"))

	return s.String()
}

func (m *DateConfigModel) viewSelectOutput() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Select Output Date Format"))
	s.WriteString("\n\n")

	s.WriteString(subtitleStyle.Render("Input format: "))
	s.WriteString(highlightStyle.Render(m.inputFormat))
	s.WriteString("\n\n")

	s.WriteString(subtitleStyle.Render("Go time layout reference:"))
	s.WriteString("\n")
	s.WriteString(dimStyle.Render("  2006=year, 01=month, 02=day, 15=hour, 04=min, 05=sec"))
	s.WriteString("\n\n")

	formats := dateformat.GetFormatCheatsheet()
	for i, f := range formats {
		cursor := "  "
		style := dimStyle
		if m.cursor == i {
			cursor = selectedStyle.Render("▸ ")
			style = selectedStyle
		}
		s.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(f.Description)))
	}

	// Custom option
	cursor := "  "
	style := dimStyle
	if m.cursor == len(formats) {
		cursor = selectedStyle.Render("▸ ")
		style = selectedStyle
	}
	s.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render("Enter custom format...")))

	s.WriteString(helpStyle.Render("\n↑/↓ to navigate • Enter to select • Esc to cancel"))

	return s.String()
}

func (m *DateConfigModel) viewCustomOutput() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Enter Custom Date Format"))
	s.WriteString("\n\n")

	s.WriteString(subtitleStyle.Render("Go time layout reference:"))
	s.WriteString("\n")
	s.WriteString(dimStyle.Render("  Year:   2006 (4-digit), 06 (2-digit)"))
	s.WriteString("\n")
	s.WriteString(dimStyle.Render("  Month:  01 (numeric), Jan, January"))
	s.WriteString("\n")
	s.WriteString(dimStyle.Render("  Day:    02 (zero-padded), 2 (no padding)"))
	s.WriteString("\n")
	s.WriteString(dimStyle.Render("  Hour:   15 (24h), 03 (12h)"))
	s.WriteString("\n")
	s.WriteString(dimStyle.Render("  Minute: 04"))
	s.WriteString("\n")
	s.WriteString(dimStyle.Render("  Second: 05"))
	s.WriteString("\n\n")

	s.WriteString("Format: ")
	s.WriteString(m.customInput.View())
	s.WriteString("\n")

	// Show preview if valid
	if m.customInput.Value() != "" && dateformat.ValidateLayout(m.customInput.Value()) {
		s.WriteString("\n")
		s.WriteString(successStyle.Render("Valid format"))
	} else if m.customInput.Value() != "" {
		s.WriteString("\n")
		s.WriteString(errorStyle.Render("Invalid format"))
	}

	s.WriteString(helpStyle.Render("\n\nEnter to confirm • Esc to cancel"))

	return s.String()
}
