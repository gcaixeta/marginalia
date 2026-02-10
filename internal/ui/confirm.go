package ui

import (
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles for the confirmation dialog
var (
	confirmTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("196")).
				PaddingBottom(1)

	fileInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			PaddingLeft(2)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true).
			PaddingTop(1).
			PaddingBottom(1)

	confirmButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true).
				PaddingLeft(2)

	cancelButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				PaddingLeft(2)

	selectedButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true).
				PaddingLeft(2)
)

// ConfirmModel holds the state of the confirmation dialog
type ConfirmModel struct {
	filePath    string
	relPath     string
	confirmed   bool
	cancelled   bool
	cursor      int // 0 = No, 1 = Yes
}

// NewConfirmModel creates a new confirmation dialog model
func NewConfirmModel(filePath string, dataDir string) ConfirmModel {
	relPath, err := filepath.Rel(dataDir, filePath)
	if err != nil {
		relPath = filePath
	}

	return ConfirmModel{
		filePath: filePath,
		relPath:  relPath,
		cursor:   0, // Default to "No"
	}
}

// Init initializes the model
func (m ConfirmModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m ConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q", "n":
			m.cancelled = true
			return m, tea.Quit

		case "enter", "y":
			if msg.String() == "y" || m.cursor == 1 {
				m.confirmed = true
			} else {
				m.cancelled = true
			}
			return m, tea.Quit

		case "left", "h":
			m.cursor = 0

		case "right", "l":
			m.cursor = 1

		case "tab":
			m.cursor = (m.cursor + 1) % 2
		}
	}

	return m, nil
}

// View renders the UI
func (m ConfirmModel) View() string {
	if m.cancelled || m.confirmed {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(confirmTitleStyle.Render("⚠  CONFIRMAR EXCLUSÃO"))
	b.WriteString("\n\n")

	// File info
	b.WriteString("Arquivo: ")
	b.WriteString(fileInfoStyle.Render(m.relPath))
	b.WriteString("\n\n")

	// Warning
	b.WriteString(warningStyle.Render("Esta ação não pode ser desfeita!"))
	b.WriteString("\n\n")

	// Buttons
	b.WriteString("Tem certeza que deseja excluir este arquivo?\n\n")

	// No button (default)
	if m.cursor == 0 {
		b.WriteString(selectedButtonStyle.Render("▸ [N] Não"))
	} else {
		b.WriteString(cancelButtonStyle.Render("  [N] Não"))
	}

	b.WriteString("    ")

	// Yes button
	if m.cursor == 1 {
		b.WriteString(selectedButtonStyle.Render("▸ [Y] Sim, excluir"))
	} else {
		b.WriteString(confirmButtonStyle.Render("  [Y] Sim, excluir"))
	}

	b.WriteString("\n\n")

	// Help text
	b.WriteString(helpStyle.Render("[←→/hl/Tab] navegar • [Enter] confirmar • [Esc/Q] cancelar"))

	return b.String()
}

// RunConfirmDialog runs the confirmation dialog and returns true if confirmed
func RunConfirmDialog(filePath string, dataDir string) (bool, error) {
	model := NewConfirmModel(filePath, dataDir)

	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return false, err
	}

	m := finalModel.(ConfirmModel)

	return m.confirmed, nil
}
