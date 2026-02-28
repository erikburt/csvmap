// Package config handles configuration directory management for csvmap.
package config

import (
	"os"
	"path/filepath"
)

// Config holds application configuration paths.
type Config struct {
	BaseDir     string
	MappingsDir string
}

// DefaultConfigDir returns the default configuration directory path.
func DefaultConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to ~/.config on Unix
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "csvmap"), nil
}

// New creates a new Config with the given base directory.
// If baseDir is empty, it uses the default config directory.
func New(baseDir string) (*Config, error) {
	var err error
	if baseDir == "" {
		baseDir, err = DefaultConfigDir()
		if err != nil {
			return nil, err
		}
	}

	cfg := &Config{
		BaseDir:     baseDir,
		MappingsDir: filepath.Join(baseDir, "mappings"),
	}

	return cfg, nil
}

// EnsureDirs creates the configuration directories if they don't exist.
func (c *Config) EnsureDirs() error {
	if err := os.MkdirAll(c.MappingsDir, 0755); err != nil {
		return err
	}
	return nil
}

// MappingFilePath returns the full path for a mapping file by name.
func (c *Config) MappingFilePath(name string) string {
	if !hasJSONExtension(name) {
		name = name + ".json"
	}
	return filepath.Join(c.MappingsDir, name)
}

// ListMappingFiles returns a list of mapping file names in the mappings directory.
func (c *Config) ListMappingFiles() ([]string, error) {
	entries, err := os.ReadDir(c.MappingsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && hasJSONExtension(entry.Name()) {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

// CopyMappingToDir copies or symlinks a mapping file from an external path
// into the central mappings directory.
func (c *Config) CopyMappingToDir(externalPath string) (string, error) {
	if err := c.EnsureDirs(); err != nil {
		return "", err
	}

	filename := filepath.Base(externalPath)
	destPath := filepath.Join(c.MappingsDir, filename)

	// If the file is already in the mappings directory, return as-is
	absExternal, err := filepath.Abs(externalPath)
	if err != nil {
		return "", err
	}
	absDest, err := filepath.Abs(destPath)
	if err != nil {
		return "", err
	}
	if absExternal == absDest {
		return destPath, nil
	}

	// Read and copy the file
	data, err := os.ReadFile(externalPath)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return "", err
	}

	return destPath, nil
}

// OutputFilePath generates the output file path by inserting "_mapped" before the extension.
func OutputFilePath(inputPath string) string {
	dir := filepath.Dir(inputPath)
	base := filepath.Base(inputPath)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]
	return filepath.Join(dir, name+"_mapped"+ext)
}

func hasJSONExtension(name string) bool {
	return len(name) > 5 && name[len(name)-5:] == ".json"
}
