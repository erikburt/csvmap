package tui

import (
	"fmt"
	"strings"

	"github.com/erikburt/csvmap/internal/csv"
	"github.com/erikburt/csvmap/internal/mapping"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// MappingInteractivePhase represents the current phase of interactive mapping.
type MappingInteractivePhase int

const (
	MappingPhaseAction MappingInteractivePhase = iota
	MappingPhaseEnterReplacement
	MappingPhasePatternType
	MappingPhaseEnterPattern
)

// MappingAction represents the action for an unmapped value.
type MappingAction int

const (
	ActionMap MappingAction = iota
	ActionSkip
	ActionUnknown
)

// MappingInteractiveModel handles interactive value mapping.
type MappingInteractiveModel struct {
	data             *csv.Data
	columnIndex      int
	mappingFile      *mapping.File
	uniqueValues     []string
	currentIndex     int
	phase            MappingInteractivePhase
	cursor           int
	replacementInput textinput.Model
	patternInput     textinput.Model
	currentValue     string
	replacement      string
	transformations  map[string]string
	mappedCount      int
	skippedCount     int
	unknownCount     int
	done             bool
	err              error
}

// MappingInteractiveResult contains the result of interactive mapping.
type MappingInteractiveResult struct {
	Done            bool
	Transformations map[string]string
	MappedCount     int
	SkippedCount    int
	UnknownCount    int
}

// NewMappingInteractiveModel creates a new interactive mapping model.
func NewMappingInteractiveModel(data *csv.Data, columnIndex int, mf *mapping.File, uniqueValues []string) *MappingInteractiveModel {
	replacementInput := textinput.New()
	replacementInput.Placeholder = "Enter replacement value"
	replacementInput.CharLimit = 100

	patternInput := textinput.New()
	patternInput.Placeholder = "Enter glob pattern"
	patternInput.CharLimit = 100

	m := &MappingInteractiveModel{
		data:             data,
		columnIndex:      columnIndex,
		mappingFile:      mf,
		uniqueValues:     uniqueValues,
		transformations:  make(map[string]string),
		replacementInput: replacementInput,
		patternInput:     patternInput,
	}

	// Pre-map values using existing mappings
	m.preMapValues()

	// Move to first unmapped value
	m.findNextUnmapped()

	return m
}

// preMapValues applies existing mappings to values.
func (m *MappingInteractiveModel) preMapValues() {
	for _, value := range m.uniqueValues {
		if replacement, found := m.mappingFile.Match(value); found {
			m.transformations[value] = replacement
			m.mappedCount++
		}
	}
}

// findNextUnmapped finds the next unmapped value.
func (m *MappingInteractiveModel) findNextUnmapped() {
	for m.currentIndex < len(m.uniqueValues) {
		value := m.uniqueValues[m.currentIndex]
		if _, mapped := m.transformations[value]; !mapped && value != "" {
			m.currentValue = value
			return
		}
		m.currentIndex++
	}
	// No more unmapped values
	m.done = true
}

// remainingCount returns the number of remaining unmapped values.
func (m *MappingInteractiveModel) remainingCount() int {
	count := 0
	for i := m.currentIndex; i < len(m.uniqueValues); i++ {
		value := m.uniqueValues[i]
		if _, mapped := m.transformations[value]; !mapped && value != "" {
			count++
		}
	}
	return count
}

// totalUnmappedCount returns total unique non-empty values that weren't pre-mapped.
func (m *MappingInteractiveModel) totalUnmappedCount() int {
	count := 0
	for _, value := range m.uniqueValues {
		if value != "" {
			if _, mapped := m.transformations[value]; !mapped {
				count++
			}
		}
	}
	return count + (m.skippedCount + m.unknownCount)
}

// Update handles input for interactive mapping.
func (m *MappingInteractiveModel) Update(msg tea.Msg) MappingInteractiveResult {
	if m.done {
		return MappingInteractiveResult{
			Done:            true,
			Transformations: m.transformations,
			MappedCount:     m.mappedCount,
			SkippedCount:    m.skippedCount,
			UnknownCount:    m.unknownCount,
		}
	}

	switch m.phase {
	case MappingPhaseAction:
		return m.updateAction(msg)
	case MappingPhaseEnterReplacement:
		return m.updateEnterReplacement(msg)
	case MappingPhasePatternType:
		return m.updatePatternType(msg)
	case MappingPhaseEnterPattern:
		return m.updateEnterPattern(msg)
	}

	return MappingInteractiveResult{Done: false}
}

func (m *MappingInteractiveModel) updateAction(msg tea.Msg) MappingInteractiveResult {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < 2 {
				m.cursor++
			}
		case "enter":
			switch m.cursor {
			case 0: // Map
				m.phase = MappingPhaseEnterReplacement
				m.replacementInput.Focus()
				m.replacementInput.SetValue("")
			case 1: // Skip
				m.skippedCount++
				m.currentIndex++
				m.findNextUnmapped()
				m.cursor = 0
			case 2: // Unknown
				m.transformations[m.currentValue] = m.currentValue + " (unknown)"
				m.unknownCount++
				m.currentIndex++
				m.findNextUnmapped()
				m.cursor = 0
			}
		}
	}

	if m.done {
		return MappingInteractiveResult{
			Done:            true,
			Transformations: m.transformations,
			MappedCount:     m.mappedCount,
			SkippedCount:    m.skippedCount,
			UnknownCount:    m.unknownCount,
		}
	}

	return MappingInteractiveResult{Done: false}
}

func (m *MappingInteractiveModel) updateEnterReplacement(msg tea.Msg) MappingInteractiveResult {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.phase = MappingPhaseAction
			return MappingInteractiveResult{Done: false}
		case "enter":
			m.replacement = m.replacementInput.Value()
			if m.replacement != "" {
				m.phase = MappingPhasePatternType
				m.cursor = 0
			}
		}
	}

	var cmd tea.Cmd
	m.replacementInput, cmd = m.replacementInput.Update(msg)
	_ = cmd

	return MappingInteractiveResult{Done: false}
}

func (m *MappingInteractiveModel) updatePatternType(msg tea.Msg) MappingInteractiveResult {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.phase = MappingPhaseEnterReplacement
			return MappingInteractiveResult{Done: false}
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
				// Exact match
				err := m.mappingFile.AddMapping(m.currentValue, m.replacement)
				if err != nil {
					m.err = err
					return MappingInteractiveResult{Done: false}
				}
				if err := m.mappingFile.Save(); err != nil {
					m.err = err
					return MappingInteractiveResult{Done: false}
				}
				m.transformations[m.currentValue] = m.replacement
				m.mappedCount++
				m.currentIndex++
				m.findNextUnmapped()
				m.phase = MappingPhaseAction
				m.cursor = 0
			} else {
				// Glob pattern
				m.phase = MappingPhaseEnterPattern
				m.patternInput.Focus()
				// Pre-fill with suggested pattern
				m.patternInput.SetValue(mapping.SuggestPattern(m.currentValue))
			}
		}
	}

	if m.done {
		return MappingInteractiveResult{
			Done:            true,
			Transformations: m.transformations,
			MappedCount:     m.mappedCount,
			SkippedCount:    m.skippedCount,
			UnknownCount:    m.unknownCount,
		}
	}

	return MappingInteractiveResult{Done: false}
}

func (m *MappingInteractiveModel) updateEnterPattern(msg tea.Msg) MappingInteractiveResult {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.phase = MappingPhasePatternType
			return MappingInteractiveResult{Done: false}
		case "enter":
			pattern := m.patternInput.Value()
			if pattern == "" {
				return MappingInteractiveResult{Done: false}
			}

			// Validate pattern
			if err := mapping.ValidateGlobPattern(pattern); err != nil {
				m.err = fmt.Errorf("invalid pattern: %w", err)
				return MappingInteractiveResult{Done: false}
			}

			// Add mapping
			if err := m.mappingFile.AddMapping(pattern, m.replacement); err != nil {
				m.err = err
				return MappingInteractiveResult{Done: false}
			}
			if err := m.mappingFile.Save(); err != nil {
				m.err = err
				return MappingInteractiveResult{Done: false}
			}

			// Apply to current value
			m.transformations[m.currentValue] = m.replacement
			m.mappedCount++

			// Apply to any other matching values
			m.applyNewPatternToRemaining(pattern, m.replacement)

			m.currentIndex++
			m.findNextUnmapped()
			m.phase = MappingPhaseAction
			m.cursor = 0
		}
	}

	var cmd tea.Cmd
	m.patternInput, cmd = m.patternInput.Update(msg)
	_ = cmd

	if m.done {
		return MappingInteractiveResult{
			Done:            true,
			Transformations: m.transformations,
			MappedCount:     m.mappedCount,
			SkippedCount:    m.skippedCount,
			UnknownCount:    m.unknownCount,
		}
	}

	return MappingInteractiveResult{Done: false}
}

// applyNewPatternToRemaining applies a new pattern to remaining unmapped values.
func (m *MappingInteractiveModel) applyNewPatternToRemaining(pattern, replacement string) {
	for i := m.currentIndex + 1; i < len(m.uniqueValues); i++ {
		value := m.uniqueValues[i]
		if _, mapped := m.transformations[value]; !mapped && value != "" {
			if matched, _ := m.mappingFile.Match(value); matched != "" {
				m.transformations[value] = replacement
				m.mappedCount++
			}
		}
	}
}

// View renders the interactive mapping view.
func (m *MappingInteractiveModel) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Map Values"))
	s.WriteString("\n\n")

	// Progress
	remaining := m.remainingCount()
	total := m.mappedCount + m.skippedCount + m.unknownCount + remaining
	processed := total - remaining
	s.WriteString(dimStyle.Render(fmt.Sprintf("Progress: %d of %d unique values", processed, total)))
	s.WriteString("\n")
	s.WriteString(dimStyle.Render(fmt.Sprintf("Mapped: %d | Skipped: %d | Unknown: %d | Remaining: %d",
		m.mappedCount, m.skippedCount, m.unknownCount, remaining)))
	s.WriteString("\n\n")

	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())))
		s.WriteString("\n\n")
		m.err = nil
	}

	// Current value context
	s.WriteString(subtitleStyle.Render("Current value:"))
	s.WriteString("\n")
	s.WriteString(highlightStyle.Render(fmt.Sprintf("  %s", m.currentValue)))
	s.WriteString("\n\n")

	// Show row context
	rowIndices := m.data.RowsWithValue(m.columnIndex, m.currentValue)
	if len(rowIndices) > 0 {
		s.WriteString(subtitleStyle.Render("Row context:"))
		s.WriteString("\n")
		// Show first matching row
		row := m.data.GetRow(rowIndices[0])
		for i, header := range m.data.Headers {
			if i < len(row) {
				if i == m.columnIndex {
					s.WriteString(fmt.Sprintf("  %s: %s\n", header, highlightStyle.Render(row[i])))
				} else {
					s.WriteString(fmt.Sprintf("  %s: %s\n", dimStyle.Render(header), row[i]))
				}
			}
		}
		if len(rowIndices) > 1 {
			s.WriteString(dimStyle.Render(fmt.Sprintf("  (and %d more rows with this value)\n", len(rowIndices)-1)))
		}
		s.WriteString("\n")
	}

	// Phase-specific rendering
	switch m.phase {
	case MappingPhaseAction:
		s.WriteString(m.viewAction())
	case MappingPhaseEnterReplacement:
		s.WriteString(m.viewEnterReplacement())
	case MappingPhasePatternType:
		s.WriteString(m.viewPatternType())
	case MappingPhaseEnterPattern:
		s.WriteString(m.viewEnterPattern())
	}

	return s.String()
}

func (m *MappingInteractiveModel) viewAction() string {
	var s strings.Builder

	s.WriteString("What would you like to do?\n\n")

	actions := []struct {
		name string
		desc string
	}{
		{"Map it", "Define a replacement value"},
		{"Skip", "Leave the original value unchanged"},
		{"Unknown", "Append (unknown) to the original value"},
	}

	for i, action := range actions {
		cursor := "  "
		style := dimStyle
		if m.cursor == i {
			cursor = selectedStyle.Render("▸ ")
			style = selectedStyle
		}
		s.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(action.name)))
		if m.cursor == i {
			s.WriteString(fmt.Sprintf("    %s\n", dimStyle.Render(action.desc)))
		}
	}

	s.WriteString(helpStyle.Render("\n↑/↓ to navigate • Enter to select"))

	return s.String()
}

func (m *MappingInteractiveModel) viewEnterReplacement() string {
	var s strings.Builder

	s.WriteString("Enter replacement value:\n\n")
	s.WriteString("Replacement: ")
	s.WriteString(m.replacementInput.View())

	s.WriteString(helpStyle.Render("\n\nEnter to confirm • Esc to go back"))

	return s.String()
}

func (m *MappingInteractiveModel) viewPatternType() string {
	var s strings.Builder

	s.WriteString(fmt.Sprintf("Map to: %s\n\n", highlightStyle.Render(m.replacement)))
	s.WriteString("How would you like to match this value?\n\n")

	options := []struct {
		name string
		desc string
	}{
		{"Exact match", "Only match this exact string"},
		{"Glob pattern", "Define a pattern with wildcards (* and ?)"},
	}

	for i, opt := range options {
		cursor := "  "
		style := dimStyle
		if m.cursor == i {
			cursor = selectedStyle.Render("▸ ")
			style = selectedStyle
		}
		s.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(opt.name)))
		if m.cursor == i {
			s.WriteString(fmt.Sprintf("    %s\n", dimStyle.Render(opt.desc)))
		}
	}

	s.WriteString(helpStyle.Render("\n↑/↓ to navigate • Enter to select • Esc to go back"))

	return s.String()
}

func (m *MappingInteractiveModel) viewEnterPattern() string {
	var s strings.Builder

	s.WriteString(fmt.Sprintf("Map to: %s\n\n", highlightStyle.Render(m.replacement)))
	s.WriteString("Enter glob pattern:\n")
	s.WriteString(dimStyle.Render("Use * for any characters, ? for single character\n"))
	s.WriteString(dimStyle.Render("Example: *AMAZON* matches 'AMAZON.COM' and 'AMAZON MARKETPLACE'\n\n"))

	s.WriteString("Pattern: ")
	s.WriteString(m.patternInput.View())

	s.WriteString(helpStyle.Render("\n\nEnter to confirm • Esc to go back"))

	return s.String()
}
