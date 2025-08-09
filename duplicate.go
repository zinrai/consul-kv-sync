package main

import (
	"fmt"
	"sort"
	"strings"
)

// detectDuplicates checks for duplicate keys across all YAML files
func detectDuplicates(kvMaps []map[string]interface{}, filenames []string) ([]DuplicateInfo, error) {
	if len(kvMaps) != len(filenames) {
		return nil, fmt.Errorf("mismatch between number of KV maps (%d) and filenames (%d)", len(kvMaps), len(filenames))
	}

	keyTracker := make(map[string][]FileSource)

	// Collect all keys from all files
	for i, kvMap := range kvMaps {
		// Flatten the map to get all keys
		pairs := flattenKVPairs(kvMap, "")
		for _, pair := range pairs {
			keyTracker[pair.Key] = append(keyTracker[pair.Key], FileSource{
				Filename: filenames[i],
				Value:    pair.Value,
			})
		}
	}

	// Find duplicates
	var duplicates []DuplicateInfo
	for key, sources := range keyTracker {
		if len(sources) > 1 {
			duplicates = append(duplicates, DuplicateInfo{
				Key:   key,
				Files: sources,
			})
		}
	}

	// Sort duplicates by key for consistent output
	sort.Slice(duplicates, func(i, j int) bool {
		return duplicates[i].Key < duplicates[j].Key
	})

	return duplicates, nil
}

// formatDuplicateError formats duplicate information into an error message
func formatDuplicateError(duplicates []DuplicateInfo) string {
	var sb strings.Builder
	sb.WriteString("ERROR: Duplicate keys detected across YAML files:\n\n")

	for _, dup := range duplicates {
		sb.WriteString(fmt.Sprintf("Key: \"%s\"\n", dup.Key))
		for _, file := range dup.Files {
			sb.WriteString(fmt.Sprintf("  - File: %s, Value: \"%v\"\n", file.Filename, file.Value))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("Total duplicate keys found: %d\n", len(duplicates)))
	sb.WriteString("Aborting operation to prevent data inconsistency.")

	return sb.String()
}
