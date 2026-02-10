package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gcaixeta/marginalia/internal/storage"
)

// Styles for the delete picker UI
var (
	deleteTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("196")).
				PaddingBottom(1)

	collectionHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")).
				Bold(true).
				PaddingTop(1).
				PaddingLeft(1)

	selectedFileStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Bold(true).
				PaddingLeft(2)

	normalFileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			PaddingLeft(4)

	fileDateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	inputFieldStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	emptyMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Italic(true).
				PaddingLeft(2)
)

// DeletePickerModel holds the state of the file deletion picker
type DeletePickerModel struct {
	allFiles      []storage.FileItem
	filteredFiles []storage.FileItem
	input         string
	cursor        int
	selected      *storage.FileItem
	cancelled     bool
	err           error
}

// NewDeletePickerModel creates a new delete picker model
func NewDeletePickerModel(initialFilter string) (DeletePickerModel, error) {
	files, err := storage.ListAllFiles()
	if err != nil {
		return DeletePickerModel{}, err
	}

	// Sort files by collection, then by name
	sort.Slice(files, func(i, j int) bool {
		if files[i].Collection != files[j].Collection {
			return files[i].Collection < files[j].Collection
		}
		return files[i].Name < files[j].Name
	})

	model := DeletePickerModel{
		allFiles: files,
		input:    initialFilter,
		cursor:   0,
	}

	model.updateFilteredFiles()
	return model, nil
}

// Init initializes the model
func (m DeletePickerModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m DeletePickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.cancelled = true
			return m, tea.Quit

		case "enter":
			if len(m.filteredFiles) > 0 {
				m.selected = &m.filteredFiles[m.cursor]
				return m, tea.Quit
			}
			return m, nil

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.filteredFiles)-1 {
				m.cursor++
			}

		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
				m.updateFilteredFiles()
				// Reset cursor if out of bounds
				if m.cursor >= len(m.filteredFiles) {
					m.cursor = 0
				}
			}

		default:
			// Handle regular character input
			if len(msg.String()) == 1 {
				m.input += msg.String()
				m.updateFilteredFiles()
				// Reset cursor to top when filtering
				m.cursor = 0
			}
		}
	}

	return m, nil
}

// View renders the UI
func (m DeletePickerModel) View() string {
	if m.cancelled {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(deleteTitleStyle.Render("Selecionar Nota para Excluir"))
	b.WriteString("\n\n")

	// Input field
	b.WriteString("Filtrar: ")
	b.WriteString(inputFieldStyle.Render(m.input))
	b.WriteString(inputFieldStyle.Render("█"))
	b.WriteString("\n")

	// Show error if any
	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(fmt.Sprintf("Erro: %v", m.err)))
		b.WriteString("\n")
	}

	// List files grouped by collection
	if len(m.filteredFiles) == 0 {
		b.WriteString("\n")
		b.WriteString(emptyMessageStyle.Render("Nenhuma nota encontrada"))
		b.WriteString("\n")
	} else {
		b.WriteString("\n")
		
		// Group files by collection for display
		currentCollection := ""
		visibleIndex := 0
		maxVisible := 15

		for i, file := range m.filteredFiles {
			// Show collection header when collection changes
			if file.Collection != currentCollection {
				currentCollection = file.Collection
				b.WriteString("\n")
				b.WriteString(collectionHeaderStyle.Render(fmt.Sprintf("📁 %s", currentCollection)))
				b.WriteString("\n")
			}

			// Only show a window of items for long lists
			if visibleIndex >= maxVisible {
				remaining := len(m.filteredFiles) - i
				b.WriteString("\n")
				b.WriteString(helpStyle.Render(fmt.Sprintf("... e mais %d nota(s)", remaining)))
				break
			}

			cursor := "  "
			style := normalFileStyle

			if i == m.cursor {
				cursor = "▸ "
				style = selectedFileStyle
			}

			// Format: filename (date)
			dateStr := formatDate(file.ModTime)
			line := fmt.Sprintf("%s%s %s", cursor, file.Name, fileDateStyle.Render(fmt.Sprintf("(%s)", dateStr)))
			b.WriteString(style.Render(line))
			b.WriteString("\n")

			visibleIndex++
		}

		// Show count
		b.WriteString("\n")
		b.WriteString(helpStyle.Render(fmt.Sprintf("Total: %d nota(s)", len(m.filteredFiles))))
		b.WriteString("\n")
	}

	// Help text
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("[↑↓/jk] navegar • [Enter] selecionar • [Esc/Q] cancelar"))

	return b.String()
}

// updateFilteredFiles filters the files based on the input
func (m *DeletePickerModel) updateFilteredFiles() {
	if m.input == "" {
		m.filteredFiles = m.allFiles
		return
	}

	m.filteredFiles = []storage.FileItem{}
	inputLower := strings.ToLower(m.input)

	for _, file := range m.allFiles {
		nameLower := strings.ToLower(file.Name)
		collectionLower := strings.ToLower(file.Collection)

		if strings.Contains(nameLower, inputLower) || strings.Contains(collectionLower, inputLower) {
			m.filteredFiles = append(m.filteredFiles, file)
		}
	}
}

// RunDeletePicker runs the delete picker and returns the selected file or nil if cancelled
func RunDeletePicker(initialFilter string) (*storage.FileItem, error) {
	model, err := NewDeletePickerModel(initialFilter)
	if err != nil {
		return nil, err
	}

	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	m := finalModel.(DeletePickerModel)

	if m.cancelled {
		return nil, fmt.Errorf("cancelado pelo usuário")
	}

	if m.err != nil {
		return nil, m.err
	}

	return m.selected, nil
}

// formatDate formats a time.Time into a readable string
func formatDate(t time.Time) string {
	now := time.Now()
	
	// If today, show time
	if t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day() {
		return fmt.Sprintf("hoje %s", t.Format("15:04"))
	}
	
	// If yesterday
	yesterday := now.AddDate(0, 0, -1)
	if t.Year() == yesterday.Year() && t.Month() == yesterday.Month() && t.Day() == yesterday.Day() {
		return fmt.Sprintf("ontem %s", t.Format("15:04"))
	}
	
	// If this year, show day and month
	if t.Year() == now.Year() {
		return t.Format("02 Jan")
	}
	
	// Otherwise show full date
	return t.Format("02 Jan 2006")
}
