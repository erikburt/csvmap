// Package mapping handles string mapping with glob pattern support for csvmap.
package mapping

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
)

// Entry represents a single mapping entry.
type Entry struct {
	Pattern     string `json:"pattern"`
	Replacement string `json:"replacement"`
}

// File represents a mapping file with its entries.
type File struct {
	Mappings []Entry `json:"mappings"`
	FilePath string  `json:"-"`

	// Compiled globs for matching (not persisted)
	compiledGlobs []compiledEntry
}

type compiledEntry struct {
	entry Entry
	glob  glob.Glob
}

// Load reads a mapping file from disk.
func Load(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty mapping file
			return &File{
				FilePath: path,
				Mappings: []Entry{},
			}, nil
		}
		return nil, fmt.Errorf("failed to read mapping file: %w", err)
	}

	var mf File
	if err := json.Unmarshal(data, &mf); err != nil {
		return nil, fmt.Errorf("failed to parse mapping file: %w", err)
	}

	mf.FilePath = path
	if err := mf.compile(); err != nil {
		return nil, err
	}

	return &mf, nil
}

// Save writes the mapping file to disk using atomic write.
func (mf *File) Save() error {
	if mf.FilePath == "" {
		return fmt.Errorf("no file path set")
	}

	data, err := json.MarshalIndent(mf, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize mapping file: %w", err)
	}

	// Atomic write: temp file + rename
	dir := filepath.Dir(mf.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	tempFile, err := os.CreateTemp(dir, "mapping-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()

	success := false
	defer func() {
		if !success {
			os.Remove(tempPath)
		}
	}()

	if _, err := tempFile.Write(data); err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Rename(tempPath, mf.FilePath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	success = true
	return nil
}

// compile compiles all glob patterns for efficient matching.
func (mf *File) compile() error {
	mf.compiledGlobs = make([]compiledEntry, len(mf.Mappings))

	for i, entry := range mf.Mappings {
		// Case-insensitive matching by lowercasing the pattern
		g, err := glob.Compile(strings.ToLower(entry.Pattern))
		if err != nil {
			return fmt.Errorf("invalid glob pattern %q: %w", entry.Pattern, err)
		}
		mf.compiledGlobs[i] = compiledEntry{
			entry: entry,
			glob:  g,
		}
	}

	return nil
}

// Match finds a matching entry for the given value.
// Returns the replacement and true if found, empty string and false otherwise.
// Exact matches are checked first, then glob patterns in file order.
func (mf *File) Match(value string) (string, bool) {
	if value == "" {
		return "", false
	}

	valueLower := strings.ToLower(value)

	// First pass: exact matches (case-insensitive)
	for _, ce := range mf.compiledGlobs {
		patternLower := strings.ToLower(ce.entry.Pattern)
		// Check if pattern has no glob characters (exact match)
		if !containsGlobChars(ce.entry.Pattern) {
			if valueLower == patternLower {
				return ce.entry.Replacement, true
			}
		}
	}

	// Second pass: glob patterns in file order
	for _, ce := range mf.compiledGlobs {
		if containsGlobChars(ce.entry.Pattern) {
			if ce.glob.Match(valueLower) {
				return ce.entry.Replacement, true
			}
		}
	}

	return "", false
}

// AddMapping adds a new mapping entry and recompiles.
func (mf *File) AddMapping(pattern, replacement string) error {
	entry := Entry{
		Pattern:     pattern,
		Replacement: replacement,
	}

	mf.Mappings = append(mf.Mappings, entry)

	// Compile the new entry
	g, err := glob.Compile(strings.ToLower(pattern))
	if err != nil {
		// Remove the invalid entry
		mf.Mappings = mf.Mappings[:len(mf.Mappings)-1]
		return fmt.Errorf("invalid glob pattern %q: %w", pattern, err)
	}

	mf.compiledGlobs = append(mf.compiledGlobs, compiledEntry{
		entry: entry,
		glob:  g,
	})

	return nil
}

// containsGlobChars checks if a pattern contains glob special characters.
func containsGlobChars(pattern string) bool {
	return strings.ContainsAny(pattern, "*?[")
}

// Name returns the base name of the mapping file without extension.
func (mf *File) Name() string {
	base := filepath.Base(mf.FilePath)
	if len(base) > 5 && base[len(base)-5:] == ".json" {
		return base[:len(base)-5]
	}
	return base
}

// EntryCount returns the number of mapping entries.
func (mf *File) EntryCount() int {
	return len(mf.Mappings)
}

// Create creates a new empty mapping file at the specified path.
func Create(path string) (*File, error) {
	mf := &File{
		FilePath: path,
		Mappings: []Entry{},
	}

	if err := mf.Save(); err != nil {
		return nil, err
	}

	return mf, nil
}
