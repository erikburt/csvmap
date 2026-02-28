package tui

import (
	"fmt"
	"strings"

	"csvmap/internal/config"
	"csvmap/internal/mapping"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// MappingConfigPhase represents the current phase of mapping configuration.
type MappingConfigPhase int

const (
	MappingPhaseSelect MappingConfigPhase = iota
	MappingPhaseCreate
)

// MappingConfigModel handles mapping file selection/creation.
type MappingConfigModel struct {
	cfg          *config.Config
	mappingFiles []string
	phase        MappingConfigPhase
	cursor       int
	nameInput    textinput.Model
	mappingFile  *mapping.File
	done         bool
	cancelled    bool
	err          error
}

// MappingConfigResult contains the result of mapping configuration.
type MappingConfigResult struct {
	Done        bool
	Cancelled   bool
	MappingFile *mapping.File
}

// NewMappingConfigModel creates a new mapping configuration model.
func NewMappingConfigModel(cfg *config.Config) *MappingConfigModel {
	files, _ := cfg.ListMappingFiles()

	nameInput := textinput.New()
	nameInput.Placeholder = "e.g., visa-merchants"
	nameInput.CharLimit = 50

	return &MappingConfigModel{
		cfg:          cfg,
		mappingFiles: files,
		phase:        MappingPhaseSelect,
		nameInput:    nameInput,
	}
}

// Update handles input for mapping configuration.
func (m *MappingConfigModel) Update(msg tea.Msg) MappingConfigResult {
	if m.done {
		return MappingConfigResult{
			Done:        true,
			Cancelled:   m.cancelled,
			MappingFile: m.mappingFile,
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.phase == MappingPhaseCreate {
				m.phase = MappingPhaseSelect
				return MappingConfigResult{Done: false}
			}
			m.done = true
			m.cancelled = true
			return MappingConfigResult{Done: true, Cancelled: true}
		}
	}

	switch m.phase {
	case MappingPhaseSelect:
		return m.updateSelect(msg)
	case MappingPhaseCreate:
		return m.updateCreate(msg)
	}

	return MappingConfigResult{Done: false}
}

func (m *MappingConfigModel) updateSelect(msg tea.Msg) MappingConfigResult {
	// Options: existing files + "Create new"
	maxCursor := len(m.mappingFiles)

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
			if m.cursor < len(m.mappingFiles) {
				// Load existing file
				path := m.cfg.MappingFilePath(m.mappingFiles[m.cursor])
				mf, err := mapping.Load(path)
				if err != nil {
					m.err = err
					return MappingConfigResult{Done: false}
				}
				m.mappingFile = mf
				m.done = true
				return MappingConfigResult{
					Done:        true,
					MappingFile: m.mappingFile,
				}
			} else {
				// Create new
				m.phase = MappingPhaseCreate
				m.nameInput.Focus()
			}
		}
	}
	return MappingConfigResult{Done: false}
}

func (m *MappingConfigModel) updateCreate(msg tea.Msg) MappingConfigResult {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			name := strings.TrimSpace(m.nameInput.Value())
			if name == "" {
				return MappingConfigResult{Done: false}
			}

			// Ensure config dirs exist
			if err := m.cfg.EnsureDirs(); err != nil {
				m.err = err
				return MappingConfigResult{Done: false}
			}

			// Create new mapping file
			path := m.cfg.MappingFilePath(name)
			mf, err := mapping.Create(path)
			if err != nil {
				m.err = err
				return MappingConfigResult{Done: false}
			}

			m.mappingFile = mf
			m.done = true
			return MappingConfigResult{
				Done:        true,
				MappingFile: m.mappingFile,
			}
		}
	}

	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)
	_ = cmd

	return MappingConfigResult{Done: false}
}

// View renders the mapping configuration view.
func (m *MappingConfigModel) View() string {
	switch m.phase {
	case MappingPhaseSelect:
		return m.viewSelect()
	case MappingPhaseCreate:
		return m.viewCreate()
	}
	return ""
}

func (m *MappingConfigModel) viewSelect() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Select Mapping File"))
	s.WriteString("\n\n")

	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())))
		s.WriteString("\n\n")
	}

	s.WriteString(subtitleStyle.Render(fmt.Sprintf("Mappings directory: %s", m.cfg.MappingsDir)))
	s.WriteString("\n\n")

	if len(m.mappingFiles) == 0 {
		s.WriteString(dimStyle.Render("No existing mapping files found."))
		s.WriteString("\n\n")
	} else {
		s.WriteString(subtitleStyle.Render("Existing mapping files:"))
		s.WriteString("\n")

		for i, file := range m.mappingFiles {
			cursor := "  "
			style := dimStyle
			if m.cursor == i {
				cursor = selectedStyle.Render("▸ ")
				style = selectedStyle
			}

			// Try to load and show entry count
			path := m.cfg.MappingFilePath(file)
			mf, err := mapping.Load(path)
			entryInfo := ""
			if err == nil {
				entryInfo = dimStyle.Render(fmt.Sprintf(" (%d entries)", mf.EntryCount()))
			}

			s.WriteString(fmt.Sprintf("%s%s%s\n", cursor, style.Render(file), entryInfo))
		}
		s.WriteString("\n")
	}

	// Create new option
	cursor := "  "
	style := successStyle
	if m.cursor == len(m.mappingFiles) {
		cursor = selectedStyle.Render("▸ ")
		style = selectedStyle
	}
	s.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render("+ Create new mapping file")))

	s.WriteString(helpStyle.Render("\n↑/↓ to navigate • Enter to select • Esc to cancel"))

	return s.String()
}

func (m *MappingConfigModel) viewCreate() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Create New Mapping File"))
	s.WriteString("\n\n")

	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())))
		s.WriteString("\n\n")
	}

	s.WriteString("Enter a name for the mapping file:\n")
	s.WriteString(dimStyle.Render("(e.g., visa-merchants, amex-transactions)"))
	s.WriteString("\n\n")

	s.WriteString("Name: ")
	s.WriteString(m.nameInput.View())
	s.WriteString("\n")

	if m.nameInput.Value() != "" {
		s.WriteString(dimStyle.Render(fmt.Sprintf("\nWill create: %s", m.cfg.MappingFilePath(m.nameInput.Value()))))
	}

	s.WriteString(helpStyle.Render("\n\nEnter to create • Esc to go back"))

	return s.String()
}
