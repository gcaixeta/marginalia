package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gcaixeta/marginalia/internal/storage"
)

var (
	browseTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("75")).
				PaddingBottom(1)

	browseSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Bold(true).
				PaddingLeft(2)

	browseNormalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				PaddingLeft(4)
)

type BrowsePickerModel struct {
	allFiles      []storage.FileItem
	filteredFiles []storage.FileItem
	input         string
	cursor        int
	selected      *storage.FileItem
	cancelled     bool
	filterMode    bool
	width         int
	height        int
}

func NewBrowsePickerModel() (BrowsePickerModel, error) {
	files, err := storage.ListAllFiles()
	if err != nil {
		return BrowsePickerModel{}, err
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].Collection != files[j].Collection {
			return files[i].Collection < files[j].Collection
		}
		return files[i].Name < files[j].Name
	})

	model := BrowsePickerModel{
		allFiles: files,
		cursor:   0,
	}
	model.updateFilteredFiles()
	return model, nil
}

func (m BrowsePickerModel) Init() tea.Cmd {
	return nil
}

func (m BrowsePickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			case "ctrl+c", "esc", "q", "Q":
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

func (m BrowsePickerModel) View() string {
	if m.cancelled {
		return ""
	}

	// 1. Build statusline (pinned to bottom)
	var statusline string
	if m.filterMode {
		statusline = modeFilterStyle.Render("-- FILTER --") +
			helpStyle.Render(fmt.Sprintf("  %d note(s)  [Esc] normal • [Enter] open", len(m.filteredFiles)))
	} else {
		statusline = modeNormalStyle.Render("-- NORMAL --") +
			helpStyle.Render(fmt.Sprintf("  %d note(s)  [↑↓/jk] navigate • [/] filter • [Enter] open • [Q] quit", len(m.filteredFiles)))
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

	b.WriteString(browseTitleStyle.Render("Your Notes"))
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("Filter: "))
	b.WriteString(inputFieldStyle.Render(m.input))
	if m.filterMode {
		b.WriteString(inputFieldStyle.Render("█"))
	}
	b.WriteString("\n")

	if len(m.filteredFiles) == 0 {
		b.WriteString("\n")
		b.WriteString(emptyMessageStyle.Render("No notes found"))
		b.WriteString("\n")
	} else {
		b.WriteString("\n")

		currentCollection := ""
		visibleIndex := 0

		for i, file := range m.filteredFiles {
			if file.Collection != currentCollection {
				currentCollection = file.Collection
				b.WriteString("\n")
				b.WriteString(collectionHeaderStyle.Render(fmt.Sprintf("📁 %s", currentCollection)))
				b.WriteString("\n")
			}

			if visibleIndex >= maxVisible {
				remaining := len(m.filteredFiles) - i
				b.WriteString("\n")
				b.WriteString(helpStyle.Render(fmt.Sprintf("... and %d more note(s)", remaining)))
				break
			}

			cursor := "  "
			style := browseNormalStyle

			if i == m.cursor {
				cursor = "▸ "
				style = browseSelectedStyle
			}

			dateStr := formatDate(file.ModTime)
			line := fmt.Sprintf("%s%s %s", cursor, file.Name, fileDateStyle.Render(fmt.Sprintf("(%s)", dateStr)))
			b.WriteString(style.Render(line))
			b.WriteString("\n")

			visibleIndex++
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

func (m *BrowsePickerModel) updateFilteredFiles() {
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

func RunBrowsePicker() (*storage.FileItem, error) {
	model, err := NewBrowsePickerModel()
	if err != nil {
		return nil, err
	}

	finalModel, err := runProgram(model)
	if err != nil {
		return nil, err
	}

	m := finalModel.(BrowsePickerModel)

	if m.cancelled || m.selected == nil {
		return nil, nil
	}

	return m.selected, nil
}
