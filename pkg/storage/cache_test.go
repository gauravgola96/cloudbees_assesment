package storage

import (
	"os"
	"testing"
)

func TestNewFileCache(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := NewDiskCache(tempDir)
	if err != nil {
		t.Fatalf("NewFileCache failed: %v", err)
	}
	if cache.filePath != tempDir {
		t.Errorf("Expected cacheDir %s, got %s", tempDir, cache.filePath)
	}
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Errorf("Cache directory was not created: %v", err)
	}
}

func TestLogExists(t *testing.T) {
	tempDir := t.TempDir()
	cache, _ := NewDiskCache(tempDir)
	buildID := "test-build-exists"
	nonExistentBuildID := "test-build-nonexistent"

	filePath := cache.GetPath(buildID)
	_ = os.WriteFile(filePath, []byte("dummy content"), 0644)

	if !cache.LogExists(buildID) {
		t.Errorf("LogExists should return true for existing file")
	}
	if cache.LogExists(nonExistentBuildID) {
		t.Errorf("LogExists should return false for non-existent file")
	}
}
