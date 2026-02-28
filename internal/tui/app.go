// Package tui implements the terminal user interface for csvmap.
package tui

import (
	"github.com/erikburt/csvmap/internal/config"
	"github.com/erikburt/csvmap/internal/csv"
	"github.com/erikburt/csvmap/internal/dateformat"
	"github.com/erikburt/csvmap/internal/mapping"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// View represents the current TUI view/screen.
type View int

const (
	ViewPreview View = iota
	ViewColumnSelect
	ViewOperationSelect
	ViewDateConfig
	ViewMappingConfig
	ViewMappingInteractive
	ViewRowFilter
	ViewReview
	ViewComplete
	ViewError
)

// OperationType represents the type of column operation.
type OperationType int

const (
	OpNone OperationType = iota
	OpDateReformat
	OpStringMapping
	OpDropColumn
	OpFilterRows
	OpInvertSign
)

// ColumnOperation represents an operation to be applied to a column.
type ColumnOperation struct {
	ColumnIndex   int
	ColumnName    string
	Operation     OperationType
	// For date operations
	InputDateFormat  string
	OutputDateFormat string
	// For mapping operations
	MappingFile    *mapping.File
	MappedCount    int
	SkippedCount   int
	UnknownCount   int
	// Transformation results (value -> replacement)
	Transformations map[string]string
	// For row filtering
	RowsToDropCount int
	RowsToDrop      map[int]bool // row index -> should drop
}

// Model is the main application model.
type Model struct {
	// Configuration
	Config    *config.Config
	HasHeader *bool // nil if not yet determined

	// CSV data
	CSVPath string
	Data    *csv.Data

	// Current view state
	CurrentView View
	Error       error

	// Operations to apply
	Operations []ColumnOperation

	// Column reordering (maps display position to original column index)
	ColumnOrder []int

	// Current operation being configured
	CurrentOp *ColumnOperation

	// View-specific state
	PreviewModel            *PreviewModel
	ColumnSelectModel       *ColumnSelectModel
	OperationSelectModel    *OperationSelectModel
	DateConfigModel         *DateConfigModel
	MappingConfigModel      *MappingConfigModel
	MappingInteractiveModel *MappingInteractiveModel
	RowFilterModel          *RowFilterModel
	ReviewModel             *ReviewModel
	CompleteModel           *CompleteModel

	// Window size
	Width  int
	Height int
}

// Styles for the TUI
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginBottom(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)

	highlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)
)

// NewModel creates a new application model.
func NewModel(csvPath string, hasHeader *bool, cfg *config.Config) *Model {
	return &Model{
		Config:     cfg,
		HasHeader:  hasHeader,
		CSVPath:    csvPath,
		Operations: []ColumnOperation{},
	}
}

// Init initializes the model.
func (m *Model) Init() tea.Cmd {
	return m.loadCSV
}

// loadCSV loads the CSV file.
func (m *Model) loadCSV() tea.Msg {
	data, err := csv.ReadFile(m.CSVPath, m.HasHeader)
	if err != nil {
		return errMsg{err}
	}
	return csvLoadedMsg{data}
}

// Messages
type errMsg struct{ error }
type csvLoadedMsg struct{ data *csv.Data }
type viewChangeMsg struct{ view View }

// Update handles messages and updates the model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.CurrentView == ViewComplete || m.CurrentView == ViewError {
				return m, tea.Quit
			}
			// Allow q to quit from most views
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		}

	case errMsg:
		m.Error = msg.error
		m.CurrentView = ViewError
		return m, nil

	case csvLoadedMsg:
		m.Data = msg.data
		// Initialize column order to original order
		m.ColumnOrder = make([]int, m.Data.ColumnCount())
		for i := range m.ColumnOrder {
			m.ColumnOrder[i] = i
		}
		if m.HasHeader == nil {
			// Need to ask user about header
			m.CurrentView = ViewPreview
			m.PreviewModel = NewPreviewModel(m.Data, true)
		} else {
			// Header status known, go to column selection
			m.CurrentView = ViewPreview
			m.PreviewModel = NewPreviewModel(m.Data, false)
		}
		return m, nil

	case viewChangeMsg:
		m.CurrentView = msg.view
		return m, nil
	}

	// Delegate to current view
	return m.updateCurrentView(msg)
}

func (m *Model) updateCurrentView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.CurrentView {
	case ViewPreview:
		if m.PreviewModel == nil {
			return m, nil
		}
		result := m.PreviewModel.Update(msg)
		if result.Done {
			if result.HasHeader != nil {
				m.Data.SetHasHeader(*result.HasHeader)
			}
			m.CurrentView = ViewColumnSelect
			m.ColumnSelectModel = NewColumnSelectModel(m.Data, m.Operations, m.ColumnOrder)
		}
		return m, nil

	case ViewColumnSelect:
		if m.ColumnSelectModel == nil {
			return m, nil
		}
		result := m.ColumnSelectModel.Update(msg)
		if result.Done {
			// Always update column order from result
			if result.ColumnOrder != nil {
				m.ColumnOrder = result.ColumnOrder
			}
			if result.RemoveOpIndex >= 0 {
				// Remove the operation at this index
				m.Operations = append(m.Operations[:result.RemoveOpIndex], m.Operations[result.RemoveOpIndex+1:]...)
				m.ColumnSelectModel = NewColumnSelectModel(m.Data, m.Operations, m.ColumnOrder)
			} else if result.QuickDrop && result.SelectedColumn >= 0 {
				// Quick drop column with x key
				m.CurrentOp = &ColumnOperation{
					ColumnIndex: result.SelectedColumn,
					ColumnName:  m.Data.Headers[result.SelectedColumn],
					Operation:   OpDropColumn,
				}
				m.Operations = append(m.Operations, *m.CurrentOp)
				m.ColumnSelectModel = NewColumnSelectModel(m.Data, m.Operations, m.ColumnOrder)
			} else if result.SelectedColumn >= 0 {
				m.CurrentOp = &ColumnOperation{
					ColumnIndex: result.SelectedColumn,
					ColumnName:  m.Data.Headers[result.SelectedColumn],
				}
				m.CurrentView = ViewOperationSelect
				m.OperationSelectModel = NewOperationSelectModel(m.CurrentOp.ColumnName)
			} else {
				// User is done adding operations, go to review
				m.CurrentView = ViewReview
				m.ReviewModel = NewReviewModel(m.Data, m.Operations, m.CSVPath, m.ColumnOrder)
			}
		}
		return m, nil

	case ViewOperationSelect:
		if m.OperationSelectModel == nil {
			return m, nil
		}
		result := m.OperationSelectModel.Update(msg)
		if result.Done {
			m.CurrentOp.Operation = result.Operation
			if result.Operation == OpDateReformat {
				m.CurrentView = ViewDateConfig
				values := m.Data.ColumnValues(m.CurrentOp.ColumnIndex)
				m.DateConfigModel = NewDateConfigModel(values)
			} else if result.Operation == OpStringMapping {
				m.CurrentView = ViewMappingConfig
				m.MappingConfigModel = NewMappingConfigModel(m.Config)
			} else if result.Operation == OpDropColumn {
				// Drop column requires no configuration, add directly
				m.Operations = append(m.Operations, *m.CurrentOp)
				m.CurrentView = ViewColumnSelect
				m.ColumnSelectModel = NewColumnSelectModel(m.Data, m.Operations, m.ColumnOrder)
			} else if result.Operation == OpFilterRows {
				// Start row filtering
				m.CurrentView = ViewRowFilter
				m.CurrentOp.RowsToDrop = make(map[int]bool)
				m.RowFilterModel = NewRowFilterModel(m.Data, m.CurrentOp.ColumnIndex)
			} else if result.Operation == OpInvertSign {
				// Invert sign requires no configuration, add directly
				m.Operations = append(m.Operations, *m.CurrentOp)
				m.CurrentView = ViewColumnSelect
				m.ColumnSelectModel = NewColumnSelectModel(m.Data, m.Operations, m.ColumnOrder)
			} else {
				// Cancelled, go back to column select
				m.CurrentView = ViewColumnSelect
				m.ColumnSelectModel = NewColumnSelectModel(m.Data, m.Operations, m.ColumnOrder)
			}
		}
		return m, nil

	case ViewDateConfig:
		if m.DateConfigModel == nil {
			return m, nil
		}
		result := m.DateConfigModel.Update(msg)
		if result.Done {
			if result.Cancelled {
				m.CurrentView = ViewColumnSelect
				m.ColumnSelectModel = NewColumnSelectModel(m.Data, m.Operations, m.ColumnOrder)
			} else {
				m.CurrentOp.InputDateFormat = result.InputFormat
				m.CurrentOp.OutputDateFormat = result.OutputFormat
				m.Operations = append(m.Operations, *m.CurrentOp)
				m.CurrentView = ViewColumnSelect
				m.ColumnSelectModel = NewColumnSelectModel(m.Data, m.Operations, m.ColumnOrder)
			}
		}
		return m, nil

	case ViewMappingConfig:
		if m.MappingConfigModel == nil {
			return m, nil
		}
		result := m.MappingConfigModel.Update(msg)
		if result.Done {
			if result.Cancelled {
				m.CurrentView = ViewColumnSelect
				m.ColumnSelectModel = NewColumnSelectModel(m.Data, m.Operations, m.ColumnOrder)
			} else {
				m.CurrentOp.MappingFile = result.MappingFile
				m.CurrentOp.Transformations = make(map[string]string)
				// Start interactive mapping
				m.CurrentView = ViewMappingInteractive
				values := m.Data.UniqueColumnValues(m.CurrentOp.ColumnIndex)
				m.MappingInteractiveModel = NewMappingInteractiveModel(
					m.Data,
					m.CurrentOp.ColumnIndex,
					m.CurrentOp.MappingFile,
					values,
				)
			}
		}
		return m, nil

	case ViewMappingInteractive:
		if m.MappingInteractiveModel == nil {
			return m, nil
		}
		result := m.MappingInteractiveModel.Update(msg)
		if result.Done {
			m.CurrentOp.Transformations = result.Transformations
			m.CurrentOp.MappedCount = result.MappedCount
			m.CurrentOp.SkippedCount = result.SkippedCount
			m.CurrentOp.UnknownCount = result.UnknownCount
			m.Operations = append(m.Operations, *m.CurrentOp)
			m.CurrentView = ViewColumnSelect
			m.ColumnSelectModel = NewColumnSelectModel(m.Data, m.Operations, m.ColumnOrder)
		}
		return m, nil

	case ViewRowFilter:
		if m.RowFilterModel == nil {
			return m, nil
		}
		result := m.RowFilterModel.Update(msg)
		if result.Done {
			m.CurrentOp.RowsToDrop = result.RowsToDrop
			m.CurrentOp.RowsToDropCount = len(result.RowsToDrop)
			m.Operations = append(m.Operations, *m.CurrentOp)
			m.CurrentView = ViewColumnSelect
			m.ColumnSelectModel = NewColumnSelectModel(m.Data, m.Operations, m.ColumnOrder)
		}
		return m, nil

	case ViewReview:
		if m.ReviewModel == nil {
			return m, nil
		}
		result := m.ReviewModel.Update(msg)
		if result.Done {
			if result.Confirmed {
				// Apply transformations and write output
				err := m.applyAndWrite()
				if err != nil {
					m.Error = err
					m.CurrentView = ViewError
				} else {
					m.CurrentView = ViewComplete
					m.CompleteModel = NewCompleteModel(config.OutputFilePath(m.CSVPath))
				}
			} else {
				// Go back to column selection
				m.CurrentView = ViewColumnSelect
				m.ColumnSelectModel = NewColumnSelectModel(m.Data, m.Operations, m.ColumnOrder)
			}
		}
		return m, nil

	case ViewComplete:
		if m.CompleteModel == nil {
			return m, nil
		}
		result := m.CompleteModel.Update(msg)
		if result.Done {
			return m, tea.Quit
		}
		return m, nil
	}

	return m, nil
}

// applyAndWrite applies all transformations and writes the output file.
func (m *Model) applyAndWrite() error {
	// Clone the data to avoid modifying the original
	outputData := m.Data.Clone()

	// First, apply transformations (before dropping columns, as indices would change)
	for _, op := range m.Operations {
		switch op.Operation {
		case OpDateReformat:
			converter := dateformat.NewConverter(op.InputDateFormat, op.OutputDateFormat)
			outputData.TransformColumn(op.ColumnIndex, converter.TransformFunc())

		case OpStringMapping:
			outputData.TransformColumn(op.ColumnIndex, func(value string) string {
				if replacement, ok := op.Transformations[value]; ok {
					return replacement
				}
				return value
			})

		case OpInvertSign:
			outputData.TransformColumn(op.ColumnIndex, invertSign)
		}
	}

	// Collect rows to drop from all filter operations
	rowsToDrop := make(map[int]bool)
	for _, op := range m.Operations {
		if op.Operation == OpFilterRows {
			for rowIdx := range op.RowsToDrop {
				rowsToDrop[rowIdx] = true
			}
		}
	}

	// Drop rows if any
	if len(rowsToDrop) > 0 {
		outputData.DropRows(rowsToDrop)
	}

	// Collect columns to drop
	dropSet := make(map[int]bool)
	for _, op := range m.Operations {
		if op.Operation == OpDropColumn {
			dropSet[op.ColumnIndex] = true
		}
	}

	// Build final column order: reordered columns minus dropped columns
	var finalOrder []int
	for _, origIdx := range m.ColumnOrder {
		if !dropSet[origIdx] {
			finalOrder = append(finalOrder, origIdx)
		}
	}

	// Apply reordering (this also handles drops since we only include non-dropped columns)
	if len(finalOrder) > 0 && len(finalOrder) < len(m.ColumnOrder) || !isIdentityOrder(m.ColumnOrder) {
		// Reorder first using full order
		outputData.ReorderColumns(m.ColumnOrder)
		// Then drop the columns that need dropping (now at their reordered positions)
		var dropIndices []int
		for i, origIdx := range m.ColumnOrder {
			if dropSet[origIdx] {
				dropIndices = append(dropIndices, i)
			}
		}
		if len(dropIndices) > 0 {
			outputData.DropColumns(dropIndices)
		}
	} else if len(dropSet) > 0 {
		// No reordering, just drop columns at original positions
		var dropIndices []int
		for idx := range dropSet {
			dropIndices = append(dropIndices, idx)
		}
		outputData.DropColumns(dropIndices)
	}

	// Check if output file exists
	outputPath := config.OutputFilePath(m.CSVPath)
	if csv.FileExists(outputPath) {
		// The review screen should have already confirmed overwrite
	}

	// Write the output file
	return csv.WriteFile(outputData, outputPath)
}

// isIdentityOrder checks if the order is [0, 1, 2, ...] (no reordering)
func isIdentityOrder(order []int) bool {
	for i, idx := range order {
		if i != idx {
			return false
		}
	}
	return true
}

// invertSign flips the sign of a numeric value.
// Handles formats like: -12.34, 12.34, $-12.34, -$12.34, $12.34, etc.
func invertSign(value string) string {
	if value == "" {
		return value
	}

	// Check if there's a minus sign anywhere in the value
	minusIdx := -1
	for i, c := range value {
		if c == '-' {
			minusIdx = i
			break
		}
	}

	if minusIdx >= 0 {
		// Remove the minus sign
		return value[:minusIdx] + value[minusIdx+1:]
	}

	// No minus sign - need to add one
	// Find the position to insert: after any leading currency symbols/spaces, before digits
	insertPos := 0
	for i, c := range value {
		if c >= '0' && c <= '9' {
			insertPos = i
			break
		}
		insertPos = i + 1
	}

	return value[:insertPos] + "-" + value[insertPos:]
}

// View renders the current view.
func (m *Model) View() string {
	switch m.CurrentView {
	case ViewPreview:
		if m.PreviewModel != nil {
			return m.PreviewModel.View()
		}
	case ViewColumnSelect:
		if m.ColumnSelectModel != nil {
			return m.ColumnSelectModel.View()
		}
	case ViewOperationSelect:
		if m.OperationSelectModel != nil {
			return m.OperationSelectModel.View()
		}
	case ViewDateConfig:
		if m.DateConfigModel != nil {
			return m.DateConfigModel.View()
		}
	case ViewMappingConfig:
		if m.MappingConfigModel != nil {
			return m.MappingConfigModel.View()
		}
	case ViewMappingInteractive:
		if m.MappingInteractiveModel != nil {
			return m.MappingInteractiveModel.View()
		}
	case ViewRowFilter:
		if m.RowFilterModel != nil {
			return m.RowFilterModel.View()
		}
	case ViewReview:
		if m.ReviewModel != nil {
			return m.ReviewModel.View()
		}
	case ViewComplete:
		if m.CompleteModel != nil {
			return m.CompleteModel.View()
		}
	case ViewError:
		return m.viewError()
	}

	return "Loading..."
}

func (m *Model) viewError() string {
	s := titleStyle.Render("Error") + "\n\n"
	s += errorStyle.Render(m.Error.Error()) + "\n\n"
	s += helpStyle.Render("Press q or Ctrl+C to exit")
	return s
}
