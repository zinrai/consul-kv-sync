package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// yamlFiles returns the sorted list of YAML file paths directly inside dir.
// Every .yaml / .yml file is included; subdirectories are ignored.
func yamlFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	var paths []string
	for _, entry := range entries {
		if entry.IsDir() || !isYAMLFile(entry.Name()) {
			continue
		}
		paths = append(paths, filepath.Join(dir, entry.Name()))
	}

	if len(paths) == 0 {
		return nil, fmt.Errorf("no YAML files found in %s", dir)
	}

	sort.Strings(paths)
	return paths, nil
}

// isYAMLFile reports whether name has a YAML extension.
func isYAMLFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".yaml" || ext == ".yml"
}
