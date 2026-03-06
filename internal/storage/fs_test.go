package storage

import (
	"testing"
)

func TestListAllFiles(t *testing.T) {
	files, err := ListAllFiles()
	if err != nil {
		t.Fatalf("ListAllFiles() error = %v", err)
	}

	// Should have some files (or at least return empty slice, not nil)
	if files == nil {
		t.Error("ListAllFiles() returned nil instead of empty slice")
	}

	// Verify file structure if files exist
	for _, file := range files {
		if file.Path == "" {
			t.Error("File has empty Path")
		}
		if file.Name == "" {
			t.Error("File has empty Name")
		}
		if file.Collection == "" {
			t.Error("File has empty Collection")
		}
	}
}

func TestFindFiles(t *testing.T) {
	tests := []struct {
		name       string
		searchTerm string
		wantError  bool
	}{
		{
			name:       "empty search term returns all files",
			searchTerm: "",
			wantError:  false,
		},
		{
			name:       "search by term",
			searchTerm: "test",
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := FindFiles(tt.searchTerm)
			if (err != nil) != tt.wantError {
				t.Errorf("FindFiles() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if files == nil {
				t.Error("FindFiles() returned nil instead of empty slice")
			}
		})
	}
}
