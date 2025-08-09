package main

import (
	"strings"
	"testing"
)

func TestDetectDuplicates(t *testing.T) {
	tests := []struct {
		name         string
		kvMaps       []map[string]interface{}
		filenames    []string
		wantError    bool
		dupCount     int
		expectedKeys []string
	}{
		{
			name: "no duplicates",
			kvMaps: []map[string]interface{}{
				{"key1": "value1"},
				{"key2": "value2"},
			},
			filenames:    []string{"file1.yaml", "file2.yaml"},
			wantError:    false,
			dupCount:     0,
			expectedKeys: []string{},
		},
		{
			name: "duplicates across files",
			kvMaps: []map[string]interface{}{
				{
					"key1": "value1",
					"key2": "value2",
				},
				{
					"key1": "value3",
					"key3": "value4",
				},
			},
			filenames:    []string{"file1.yaml", "file2.yaml"},
			wantError:    false,
			dupCount:     1,
			expectedKeys: []string{"key1"},
		},
		{
			name: "nested duplicates",
			kvMaps: []map[string]interface{}{
				{
					"app": map[string]interface{}{
						"name":    "myapp",
						"version": "1.0",
					},
				},
				{
					"app": map[string]interface{}{
						"name": "myapp2",
					},
				},
			},
			filenames:    []string{"file1.yaml", "file2.yaml"},
			wantError:    false,
			dupCount:     1,
			expectedKeys: []string{"app/name"},
		},
		{
			name: "multiple duplicates",
			kvMaps: []map[string]interface{}{
				{
					"db": map[string]interface{}{
						"host": "localhost",
						"port": 5432,
					},
					"cache": map[string]interface{}{
						"ttl": 300,
					},
				},
				{
					"db": map[string]interface{}{
						"host": "remote",
					},
					"cache": map[string]interface{}{
						"ttl": 600,
					},
				},
			},
			filenames:    []string{"config1.yaml", "config2.yaml"},
			wantError:    false,
			dupCount:     2,
			expectedKeys: []string{"db/host", "cache/ttl"},
		},
		{
			name: "three files with duplicates",
			kvMaps: []map[string]interface{}{
				{"shared": "value1"},
				{"shared": "value2"},
				{"shared": "value3"},
			},
			filenames:    []string{"file1.yaml", "file2.yaml", "file3.yaml"},
			wantError:    false,
			dupCount:     1,
			expectedKeys: []string{"shared"},
		},
		{
			name:      "mismatched kvMaps and filenames",
			kvMaps:    []map[string]interface{}{{"key": "value"}},
			filenames: []string{"file1.yaml", "file2.yaml"},
			wantError: true,
			dupCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duplicates, err := detectDuplicates(tt.kvMaps, tt.filenames)

			if tt.wantError {
				if err == nil {
					t.Errorf("detectDuplicates() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("detectDuplicates() unexpected error: %v", err)
				return
			}

			if len(duplicates) != tt.dupCount {
				t.Errorf("detectDuplicates() returned %d duplicates, want %d", len(duplicates), tt.dupCount)
			}

			// Check if the expected keys are found
			for _, expectedKey := range tt.expectedKeys {
				found := false
				for _, dup := range duplicates {
					if dup.Key == expectedKey {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("detectDuplicates() did not find expected duplicate key: %s", expectedKey)
				}
			}

			// Verify that duplicates have at least 2 sources
			for _, dup := range duplicates {
				if len(dup.Files) < 2 {
					t.Errorf("duplicate key %s has only %d sources, expected at least 2", dup.Key, len(dup.Files))
				}
			}
		})
	}
}

func TestFormatDuplicateError(t *testing.T) {
	duplicates := []DuplicateInfo{
		{
			Key: "app/name",
			Files: []FileSource{
				{Filename: "config1.yaml", Value: "myapp"},
				{Filename: "config2.yaml", Value: "otherapp"},
			},
		},
		{
			Key: "db/host",
			Files: []FileSource{
				{Filename: "prod.yaml", Value: "prod.db.com"},
				{Filename: "staging.yaml", Value: "staging.db.com"},
			},
		},
	}

	result := formatDuplicateError(duplicates)

	// Check that the error message contains expected elements
	expectedStrings := []string{
		"ERROR: Duplicate keys detected",
		"app/name",
		"config1.yaml",
		"myapp",
		"config2.yaml",
		"otherapp",
		"db/host",
		"prod.yaml",
		"staging.yaml",
		"Total duplicate keys found: 2",
		"Aborting operation",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("formatDuplicateError() result missing expected string: %q", expected)
		}
	}
}
