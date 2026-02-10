package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gcaixeta/marginalia/internal/collection"
	"github.com/gcaixeta/marginalia/internal/slug"
)

// Styles for the picker UI
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			PaddingBottom(1)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Bold(true).
				PaddingLeft(2)

	normalItemStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	newCollectionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true).
				PaddingLeft(2)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			PaddingTop(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
)

// PickerModel holds the state of the collection picker
type PickerModel struct {
	collections      []collection.Collection
	filteredItems    []pickerItem
	input            string
	cursor           int
	selected         string
	cancelled        bool
	err              error
	width            int
	height           int
}

type pickerItem struct {
	name        string
	fileCount   int
	isNewItem   bool
}

// PickerResult contains the result of the picker interaction
type PickerResult struct {
	Selected  string
	Cancelled bool
	Error     error
}

// NewPickerModel creates a new collection picker model
func NewPickerModel() (PickerModel, error) {
	collections, err := collection.ListCollections()
	if err != nil {
		return PickerModel{}, err
	}

	model := PickerModel{
		collections: collections,
		cursor:      0,
	}

	model.updateFilteredItems()
	return model, nil
}

// Init initializes the model
func (m PickerModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m PickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit

		case "enter":
			if len(m.filteredItems) > 0 {
				selectedItem := m.filteredItems[m.cursor]
				
				if selectedItem.isNewItem {
					// Normalize the collection name
					normalizedName := slug.MakeSlug(m.input)
					if normalizedName == "" {
						m.err = fmt.Errorf("nome de collection inválido")
						return m, nil
					}
					
					// Create the new collection
					err := collection.CreateCollection(normalizedName)
					if err != nil {
						m.err = err
						return m, nil
					}
					m.selected = normalizedName
				} else {
					m.selected = selectedItem.name
				}
				return m, tea.Quit
			}
			return m, nil

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.filteredItems)-1 {
				m.cursor++
			}

		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
				m.updateFilteredItems()
				// Reset cursor if it's out of bounds
				if m.cursor >= len(m.filteredItems) {
					m.cursor = 0
				}
			}

		default:
			// Handle regular character input
			if len(msg.String()) == 1 {
				m.input += msg.String()
				m.updateFilteredItems()
				// Reset cursor to top when filtering
				m.cursor = 0
			}
		}
	}

	return m, nil
}

// View renders the UI
func (m PickerModel) View() string {
	if m.cancelled {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("Selecionar Collection"))
	b.WriteString("\n\n")

	// Input field
	b.WriteString("Buscar: ")
	b.WriteString(inputStyle.Render(m.input))
	b.WriteString(inputStyle.Render("█"))
	b.WriteString("\n\n")

	// Show error if any
	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Erro: %v", m.err)))
		b.WriteString("\n\n")
	}

	// List items
	if len(m.filteredItems) == 0 {
		b.WriteString(normalItemStyle.Render("Nenhuma collection encontrada"))
		b.WriteString("\n")
	} else {
		maxVisible := 10
		start := 0
		end := len(m.filteredItems)

		// Handle scrolling for long lists
		if len(m.filteredItems) > maxVisible {
			if m.cursor >= maxVisible-2 {
				start = m.cursor - maxVisible + 3
				end = m.cursor + 3
				if end > len(m.filteredItems) {
					end = len(m.filteredItems)
					start = end - maxVisible
					if start < 0 {
						start = 0
					}
				}
			} else {
				end = maxVisible
			}
		}

		for i := start; i < end; i++ {
			item := m.filteredItems[i]
			cursor := "  "
			style := normalItemStyle

			if i == m.cursor {
				cursor = "▸ "
				if item.isNewItem {
					style = newCollectionStyle
				} else {
					style = selectedItemStyle
				}
			}

			if item.isNewItem {
				line := fmt.Sprintf("%s✨ Criar nova: \"%s\"", cursor, slug.MakeSlug(m.input))
				b.WriteString(style.Render(line))
			} else {
				line := fmt.Sprintf("%s%s (%d nota%s)", cursor, item.name, item.fileCount, pluralize(item.fileCount))
				b.WriteString(style.Render(line))
			}
			b.WriteString("\n")
		}

		// Show scroll indicator
		if len(m.filteredItems) > maxVisible {
			shown := end - start
			b.WriteString("\n")
			b.WriteString(helpStyle.Render(fmt.Sprintf("Mostrando %d de %d", shown, len(m.filteredItems))))
			b.WriteString("\n")
		}
	}

	// Help text
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("[↑↓/jk] navegar • [Enter] selecionar • [Esc] cancelar"))

	return b.String()
}

// updateFilteredItems filters the collections based on the input
func (m *PickerModel) updateFilteredItems() {
	m.filteredItems = []pickerItem{}

	if m.input == "" {
		// Show all collections when input is empty
		for _, c := range m.collections {
			m.filteredItems = append(m.filteredItems, pickerItem{
				name:      c.Name,
				fileCount: c.FileCount,
				isNewItem: false,
			})
		}
		return
	}

	inputLower := strings.ToLower(m.input)
	exactMatch := false

	// Filter collections by name (case-insensitive)
	for _, c := range m.collections {
		nameLower := strings.ToLower(c.Name)
		
		if nameLower == inputLower {
			exactMatch = true
		}
		
		if strings.Contains(nameLower, inputLower) {
			m.filteredItems = append(m.filteredItems, pickerItem{
				name:      c.Name,
				fileCount: c.FileCount,
				isNewItem: false,
			})
		}
	}

	// If no exact match exists, add option to create new collection
	if !exactMatch && m.input != "" {
		m.filteredItems = append(m.filteredItems, pickerItem{
			name:      m.input,
			fileCount: 0,
			isNewItem: true,
		})
	}
}

// RunPicker runs the collection picker and returns the selected collection
func RunPicker() (string, error) {
	model, err := NewPickerModel()
	if err != nil {
		return "", err
	}

	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	m := finalModel.(PickerModel)
	
	if m.cancelled {
		return "", fmt.Errorf("cancelado pelo usuário")
	}

	if m.err != nil {
		return "", m.err
	}

	return m.selected, nil
}

// pluralize returns "s" if count is not 1, otherwise empty string
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
