package mapping

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMatch(t *testing.T) {
	// Create a temporary mapping file
	tmpDir, err := os.MkdirTemp("", "csvmap-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, "test.json")
	mf, err := Create(path)
	if err != nil {
		t.Fatal(err)
	}

	// Add some mappings
	mf.AddMapping("AMAZON.COM", "Amazon")
	mf.AddMapping("*WALMART*", "Walmart")
	mf.AddMapping("SPOTIFY*", "Spotify")

	tests := []struct {
		value       string
		expected    string
		shouldMatch bool
	}{
		{"AMAZON.COM", "Amazon", true},
		{"amazon.com", "Amazon", true}, // case insensitive
		{"WALMART STORE #1234", "Walmart", true},
		{"SOME WALMART THING", "Walmart", true},
		{"SPOTIFY USA", "Spotify", true},
		{"SPOTIFYMUSIC", "Spotify", true},
		{"NETFLIX", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		result, found := mf.Match(tt.value)
		if found != tt.shouldMatch {
			t.Errorf("Match(%q): expected found=%v, got %v", tt.value, tt.shouldMatch, found)
		}
		if tt.shouldMatch && result != tt.expected {
			t.Errorf("Match(%q) = %q, want %q", tt.value, result, tt.expected)
		}
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "csvmap-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, "test.json")

	// Create and save
	mf, err := Create(path)
	if err != nil {
		t.Fatal(err)
	}

	mf.AddMapping("TEST", "Replacement")
	mf.AddMapping("*PATTERN*", "PatternMatch")

	if err := mf.Save(); err != nil {
		t.Fatal(err)
	}

	// Load and verify
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if loaded.EntryCount() != 2 {
		t.Errorf("expected 2 entries, got %d", loaded.EntryCount())
	}

	// Test matching with loaded file
	if result, found := loaded.Match("TEST"); !found || result != "Replacement" {
		t.Errorf("loaded.Match(TEST) = %q, %v; want Replacement, true", result, found)
	}

	if result, found := loaded.Match("SOME PATTERN HERE"); !found || result != "PatternMatch" {
		t.Errorf("loaded.Match(SOME PATTERN HERE) = %q, %v; want PatternMatch, true", result, found)
	}
}
