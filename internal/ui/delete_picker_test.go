package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gcaixeta/marginalia/internal/storage"
)

func newTestDeleteModel() DeletePickerModel {
	files := []storage.FileItem{
		{Name: "note-one", Collection: "journal", Path: "/tmp/note-one.md", ModTime: time.Now()},
		{Name: "note-two", Collection: "journal", Path: "/tmp/note-two.md", ModTime: time.Now()},
	}
	m := DeletePickerModel{
		allFiles:      files,
		filteredFiles: files,
	}
	return m
}

func TestDeletePickerInitialMode(t *testing.T) {
	m := newTestDeleteModel()
	if m.filterMode {
		t.Error("Expected filterMode to be false on init")
	}
}

func TestDeletePickerEnterFilterMode(t *testing.T) {
	m := newTestDeleteModel()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	updated, _ := m.Update(msg)
	m = updated.(DeletePickerModel)

	if !m.filterMode {
		t.Error("Expected filterMode to be true after pressing /")
	}
}

func TestDeletePickerCharInNavMode(t *testing.T) {
	m := newTestDeleteModel()

	// j should move cursor down, not type into input
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updated, _ := m.Update(msg)
	m = updated.(DeletePickerModel)

	if m.cursor != 1 {
		t.Errorf("Expected cursor to be 1 after pressing j, got %d", m.cursor)
	}
	if m.input != "" {
		t.Errorf("Expected input to be empty after pressing j in nav mode, got %q", m.input)
	}
}

func TestDeletePickerCharInFilterMode(t *testing.T) {
	m := newTestDeleteModel()
	m.filterMode = true

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updated, _ := m.Update(msg)
	m = updated.(DeletePickerModel)

	if m.input != "j" {
		t.Errorf("Expected input to be \"j\" after pressing j in filter mode, got %q", m.input)
	}
	if m.cursor != 0 {
		t.Errorf("Expected cursor to reset to 0 after typing in filter mode, got %d", m.cursor)
	}
}

func TestDeletePickerEscExitsFilter(t *testing.T) {
	m := newTestDeleteModel()
	m.filterMode = true
	m.input = "hello"

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ := m.Update(msg)
	m = updated.(DeletePickerModel)

	if m.filterMode {
		t.Error("Expected filterMode to be false after Esc in filter mode")
	}
	if m.input != "hello" {
		t.Errorf("Expected input to be preserved after Esc, got %q", m.input)
	}
	if m.cancelled {
		t.Error("Expected model not to be cancelled after Esc in filter mode")
	}
}

func TestDeletePickerEnterSelectsInFilterMode(t *testing.T) {
	m := newTestDeleteModel()
	m.filterMode = true

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := m.Update(msg)
	m = updated.(DeletePickerModel)

	if m.selected == nil {
		t.Error("Expected selected to be set after Enter in filter mode")
	}
	if cmd == nil {
		t.Error("Expected quit command after Enter in filter mode")
	}
}

func TestDeletePickerViewShowsMode(t *testing.T) {
	m := newTestDeleteModel()

	// Nav mode: should show -- NORMAL -- indicator
	view := m.View()
	if !strings.Contains(view, "-- NORMAL --") {
		t.Error("Expected view to show '-- NORMAL --' in nav mode")
	}
	if strings.Contains(view, "-- FILTER --") {
		t.Error("Expected view to NOT show '-- FILTER --' in nav mode")
	}
	if strings.Contains(view, "█") {
		t.Error("Expected view to NOT show cursor █ in nav mode")
	}

	// Filter mode: should show -- FILTER -- indicator and cursor █
	m.filterMode = true
	view = m.View()
	if !strings.Contains(view, "-- FILTER --") {
		t.Error("Expected view to show '-- FILTER --' in filter mode")
	}
	if strings.Contains(view, "-- NORMAL --") {
		t.Error("Expected view to NOT show '-- NORMAL --' in filter mode")
	}
	if !strings.Contains(view, "█") {
		t.Error("Expected view to show cursor █ in filter mode")
	}
}
