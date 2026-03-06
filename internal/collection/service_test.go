package collection

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gcaixeta/marginalia/internal/storage"
)

func TestListCollections(t *testing.T) {
	collections, err := ListCollections()
	if err != nil {
		t.Fatalf("ListCollections failed: %v", err)
	}

	if len(collections) == 0 {
		t.Log("No collections found (this is OK if none exist)")
		return
	}

	t.Logf("Found %d collections", len(collections))
	for _, c := range collections {
		t.Logf("  - %s (%d files)", c.Name, c.FileCount)
	}
}

func TestCreateCollection(t *testing.T) {
	testCollectionName := "test-collection-temp"
	
	// Clean up before test
	dataDir, err := storage.DataDir()
	if err != nil {
		t.Fatalf("Failed to get data dir: %v", err)
	}
	testPath := filepath.Join(dataDir, testCollectionName)
	os.RemoveAll(testPath)

	// Test creation
	err = CreateCollection(testCollectionName)
	if err != nil {
		t.Fatalf("CreateCollection failed: %v", err)
	}

	// Verify it exists
	if !CollectionExists(testCollectionName) {
		t.Fatalf("Collection was not created")
	}

	// Clean up after test
	os.RemoveAll(testPath)
}

func TestCollectionExists(t *testing.T) {
	// Test with a collection that should exist
	collections, err := ListCollections()
	if err != nil {
		t.Fatalf("ListCollections failed: %v", err)
	}

	if len(collections) > 0 {
		firstCollection := collections[0].Name
		if !CollectionExists(firstCollection) {
			t.Errorf("CollectionExists returned false for existing collection: %s", firstCollection)
		}
	}

	// Test with a collection that should not exist
	if CollectionExists("this-collection-should-not-exist-12345") {
		t.Error("CollectionExists returned true for non-existent collection")
	}
}

func TestGetCollectionStats(t *testing.T) {
	collections, err := ListCollections()
	if err != nil {
		t.Fatalf("ListCollections failed: %v", err)
	}

	if len(collections) == 0 {
		t.Skip("No collections available to test stats")
	}

	for _, c := range collections {
		count, err := GetCollectionStats(c.Name)
		if err != nil {
			t.Errorf("GetCollectionStats failed for %s: %v", c.Name, err)
			continue
		}

		if count != c.FileCount {
			t.Errorf("File count mismatch for %s: got %d, expected %d", c.Name, count, c.FileCount)
		}
	}
}
