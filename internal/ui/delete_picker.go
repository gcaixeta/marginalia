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

	modeNormalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Bold(true)

	modeFilterStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205")).
				Bold(true)
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
	filterMode    bool
	width         int
	height        int
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
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		if m.filterMode {
			switch msg.String() {
			case "ctrl+c":
				m.cancelled = true
				return m, tea.Quit

			case "esc":
				m.filterMode = false

			case "enter":
				if len(m.filteredFiles) > 0 {
					m.selected = &m.filteredFiles[m.cursor]
					return m, tea.Quit
				}

			case "backspace":
				if len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
					m.updateFilteredFiles()
					if m.cursor >= len(m.filteredFiles) {
						m.cursor = 0
					}
				}

			default:
				if len(msg.String()) == 1 {
					m.input += msg.String()
					m.updateFilteredFiles()
					m.cursor = 0
				}
			}
		} else {
			switch msg.String() {
			case "ctrl+c", "esc", "q":
				m.cancelled = true
				return m, tea.Quit

			case "enter":
				if len(m.filteredFiles) > 0 {
					m.selected = &m.filteredFiles[m.cursor]
					return m, tea.Quit
				}

			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}

			case "down", "j":
				if m.cursor < len(m.filteredFiles)-1 {
					m.cursor++
				}

			case "/":
				m.filterMode = true
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

	// 1. Build statusline (pinned to bottom)
	var statusline string
	if m.filterMode {
		statusline = modeFilterStyle.Render("-- FILTER --") +
			helpStyle.Render(fmt.Sprintf("  %d note(s)  [Esc] normal • [Enter] select", len(m.filteredFiles)))
	} else {
		statusline = modeNormalStyle.Render("-- NORMAL --") +
			helpStyle.Render(fmt.Sprintf("  %d note(s)  [↑↓/jk] navigate • [/] filter • [Enter] select • [Q] cancel", len(m.filteredFiles)))
	}

	// 2. Determine maxVisible based on available height
	statuslineH := lipgloss.Height(statusline)
	maxVisible := 15
	if m.height > 0 {
		contentH := m.height - statuslineH
		headerLines := 6 // title(2) + blank + filter + blank + list-start-blank
		maxVisible = contentH - headerLines
		if maxVisible < 1 {
			maxVisible = 1
		}
	}

	// 3. Build content area
	var b strings.Builder

	// Title
	b.WriteString(deleteTitleStyle.Render("Select Note to Delete"))
	b.WriteString("\n\n")

	// Input field
	b.WriteString(helpStyle.Render("Filter: "))
	b.WriteString(inputFieldStyle.Render(m.input))
	if m.filterMode {
		b.WriteString(inputFieldStyle.Render("█"))
	}
	b.WriteString("\n")

	// Show error if any
	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n")
	}

	// List files grouped by collection
	if len(m.filteredFiles) == 0 {
		b.WriteString("\n")
		b.WriteString(emptyMessageStyle.Render("No notes found"))
		b.WriteString("\n")
	} else {
		b.WriteString("\n")

		type flatItem struct {
			isHeader bool
			label    string
			fileIdx  int
		}

		// Phase 1: build flat list
		var flatList []flatItem
		cursorFlatIdx := 0
		currentColl := ""

		for i, file := range m.filteredFiles {
			if file.Collection != currentColl {
				currentColl = file.Collection
				flatList = append(flatList, flatItem{fileIdx: -1, label: ""}) // blank spacer
				flatList = append(flatList, flatItem{isHeader: true, label: currentColl, fileIdx: -1})
			}
			if i == m.cursor {
				cursorFlatIdx = len(flatList)
			}
			dateStr := formatDate(file.ModTime)
			cursor := "  "
			style := normalFileStyle
			if i == m.cursor {
				cursor = "▸ "
				style = selectedFileStyle
			}
			line := style.Render(fmt.Sprintf("%s%s %s", cursor, file.Name, fileDateStyle.Render(fmt.Sprintf("(%s)", dateStr))))
			flatList = append(flatList, flatItem{fileIdx: i, label: line})
		}

		// Phase 2: viewport over flatList using cursorFlatIdx
		start := 0
		end := len(flatList)
		if len(flatList) > maxVisible {
			if cursorFlatIdx >= maxVisible-2 {
				start = cursorFlatIdx - maxVisible + 3
				end = cursorFlatIdx + 3
				if end > len(flatList) {
					end = len(flatList)
					start = end - maxVisible
					if start < 0 {
						start = 0
					}
				}
			} else {
				end = maxVisible
			}
		}

		// Phase 3: render
		for i := start; i < end; i++ {
			item := flatList[i]
			switch {
			case item.fileIdx == -1 && !item.isHeader:
				b.WriteString("\n")
			case item.isHeader:
				b.WriteString(collectionHeaderStyle.Render(fmt.Sprintf("📁 %s", item.label)))
				b.WriteString("\n")
			default:
				b.WriteString(item.label)
				b.WriteString("\n")
			}
		}

		// Scroll indicator
		if len(flatList) > maxVisible {
			firstFile, lastFile := -1, -1
			for i := start; i < end; i++ {
				if flatList[i].fileIdx >= 0 {
					if firstFile == -1 {
						firstFile = flatList[i].fileIdx + 1
					}
					lastFile = flatList[i].fileIdx + 1
				}
			}
			b.WriteString("\n")
			b.WriteString(helpStyle.Render(fmt.Sprintf("notas %d–%d de %d", firstFile, lastFile, len(m.filteredFiles))))
		}
	}

	content := b.String()

	// 4. Compose: constrain content and pin statusline to bottom
	if m.height > 0 {
		contentArea := lipgloss.NewStyle().Height(m.height - statuslineH).Render(content)
		return lipgloss.JoinVertical(lipgloss.Top, contentArea, statusline)
	}
	return content + "\n" + statusline
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

	finalModel, err := runProgram(model)
	if err != nil {
		return nil, err
	}

	m := finalModel.(DeletePickerModel)

	if m.cancelled {
		return nil, fmt.Errorf("cancelled by user")
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
		return fmt.Sprintf("today %s", t.Format("15:04"))
	}

	// If yesterday
	yesterday := now.AddDate(0, 0, -1)
	if t.Year() == yesterday.Year() && t.Month() == yesterday.Month() && t.Day() == yesterday.Day() {
		return fmt.Sprintf("yesterday %s", t.Format("15:04"))
	}
	
	// If this year, show day and month
	if t.Year() == now.Year() {
		return t.Format("02 Jan")
	}
	
	// Otherwise show full date
	return t.Format("02 Jan 2006")
}
