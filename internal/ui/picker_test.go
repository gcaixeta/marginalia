package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewPickerModel(t *testing.T) {
	model, err := NewPickerModel()
	if err != nil {
		t.Fatalf("NewPickerModel failed: %v", err)
	}

	if model.cursor != 0 {
		t.Errorf("Expected cursor to be 0, got %d", model.cursor)
	}

	if model.input != "" {
		t.Errorf("Expected empty input, got %s", model.input)
	}

	t.Logf("Picker initialized with %d collections", len(model.collections))
}

func TestFilteredItems(t *testing.T) {
	model, err := NewPickerModel()
	if err != nil {
		t.Fatalf("NewPickerModel failed: %v", err)
	}

	// Test with empty input - should show all collections
	if len(model.filteredItems) != len(model.collections) {
		t.Errorf("Expected %d filtered items with empty input, got %d",
			len(model.collections), len(model.filteredItems))
	}

	// Test filtering
	if len(model.collections) > 0 {
		// Use first letter of first collection
		firstLetter := string(model.collections[0].Name[0])
		model.input = strings.ToLower(firstLetter)
		model.updateFilteredItems()

		t.Logf("After filtering by '%s': %d items", model.input, len(model.filteredItems))

		// Should have at least one match or create option
		if len(model.filteredItems) == 0 {
			t.Error("Expected at least one filtered item")
		}
	}
}

func TestCreateNewOption(t *testing.T) {
	model, err := NewPickerModel()
	if err != nil {
		t.Fatalf("NewPickerModel failed: %v", err)
	}

	// Test with input that doesn't match any collection
	model.input = "completely-new-collection-name-xyz"
	model.updateFilteredItems()

	// Should show the "create new" option
	hasCreateOption := false
	for _, item := range model.filteredItems {
		if item.isNewItem {
			hasCreateOption = true
			break
		}
	}

	if !hasCreateOption {
		t.Error("Expected 'create new' option for non-matching input")
	}
}

func TestNavigation(t *testing.T) {
	model, err := NewPickerModel()
	if err != nil {
		t.Fatalf("NewPickerModel failed: %v", err)
	}

	if len(model.filteredItems) < 2 {
		t.Skip("Need at least 2 items to test navigation")
	}

	// Test down navigation
	initialCursor := model.cursor
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updatedModel, _ := model.Update(msg)
	model = updatedModel.(PickerModel)

	if model.cursor != initialCursor+1 {
		t.Errorf("Expected cursor to move down, got %d", model.cursor)
	}

	// Test up navigation
	msg = tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ = model.Update(msg)
	model = updatedModel.(PickerModel)

	if model.cursor != initialCursor {
		t.Errorf("Expected cursor to move back up, got %d", model.cursor)
	}
}

func TestView(t *testing.T) {
	model, err := NewPickerModel()
	if err != nil {
		t.Fatalf("NewPickerModel failed: %v", err)
	}

	view := model.View()
	
	// Check that view contains expected elements
	if !strings.Contains(view, "Selecionar Collection") {
		t.Error("View should contain title")
	}

	if !strings.Contains(view, "Buscar:") {
		t.Error("View should contain search field")
	}

	if !strings.Contains(view, "navegar") {
		t.Error("View should contain help text")
	}
}

func TestCancellation(t *testing.T) {
	model, err := NewPickerModel()
	if err != nil {
		t.Fatalf("NewPickerModel failed: %v", err)
	}

	// Send Esc key
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, cmd := model.Update(msg)
	model = updatedModel.(PickerModel)

	if !model.cancelled {
		t.Error("Expected model to be cancelled after Esc")
	}

	if cmd == nil {
		t.Error("Expected quit command after cancellation")
	}
}
