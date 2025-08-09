package main

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
)

func TestFlattenKVPairs(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		prefix   string
		expected []KVPair
	}{
		{
			name: "simple key-value",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			prefix: "",
			expected: []KVPair{
				{Key: "key1", Value: "value1"},
				{Key: "key2", Value: "value2"},
			},
		},
		{
			name: "nested structure",
			input: map[string]interface{}{
				"hoge": map[string]interface{}{
					"fuga": map[string]interface{}{
						"moge": "hogehoge",
					},
				},
			},
			prefix: "",
			expected: []KVPair{
				{Key: "hoge/fuga/moge", Value: "hogehoge"},
			},
		},
		{
			name: "mixed nested and flat",
			input: map[string]interface{}{
				"flat": "value",
				"nested": map[string]interface{}{
					"level1": map[string]interface{}{
						"level2": "deep_value",
					},
					"another": "value2",
				},
			},
			prefix: "",
			expected: []KVPair{
				{Key: "flat", Value: "value"},
				{Key: "nested/level1/level2", Value: "deep_value"},
				{Key: "nested/another", Value: "value2"},
			},
		},
		{
			name: "with prefix",
			input: map[string]interface{}{
				"key": "value",
			},
			prefix: "prefix",
			expected: []KVPair{
				{Key: "prefix/key", Value: "value"},
			},
		},
		{
			name: "numeric values",
			input: map[string]interface{}{
				"int":   123,
				"float": 45.67,
				"bool":  true,
			},
			prefix: "",
			expected: []KVPair{
				{Key: "int", Value: "123"},
				{Key: "float", Value: "45.67"},
				{Key: "bool", Value: "true"},
			},
		},
		{
			name: "empty map",
			input: map[string]interface{}{
				"empty": map[string]interface{}{},
			},
			prefix:   "",
			expected: []KVPair{},
		},
		{
			name: "nil value",
			input: map[string]interface{}{
				"null_key": nil,
			},
			prefix: "",
			expected: []KVPair{
				{Key: "null_key", Value: "<nil>"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenKVPairs(tt.input, tt.prefix)

			// Sort both slices for consistent comparison
			sort.Slice(result, func(i, j int) bool {
				return result[i].Key < result[j].Key
			})
			sort.Slice(tt.expected, func(i, j int) bool {
				return tt.expected[i].Key < tt.expected[j].Key
			})

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("flattenKVPairs() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestChunkOps(t *testing.T) {
	// Create test operations
	ops := make([]TxnOp, 150)
	for i := range ops {
		ops[i] = TxnOp{
			KV: &TxnKVOp{
				Verb:  "set",
				Key:   fmt.Sprintf("key%d", i),
				Value: "value",
			},
		}
	}

	tests := []struct {
		name           string
		opsCount       int
		chunkSize      int
		expectedChunks int
		lastChunkSize  int
	}{
		{
			name:           "exact chunks",
			opsCount:       128,
			chunkSize:      64,
			expectedChunks: 2,
			lastChunkSize:  64,
		},
		{
			name:           "partial last chunk",
			opsCount:       150,
			chunkSize:      64,
			expectedChunks: 3,
			lastChunkSize:  22,
		},
		{
			name:           "single chunk",
			opsCount:       50,
			chunkSize:      64,
			expectedChunks: 1,
			lastChunkSize:  50,
		},
		{
			name:           "empty ops",
			opsCount:       0,
			chunkSize:      64,
			expectedChunks: 0,
			lastChunkSize:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testOps := ops[:tt.opsCount]
			chunks := chunkOps(testOps, tt.chunkSize)

			if len(chunks) != tt.expectedChunks {
				t.Errorf("chunkOps() returned %d chunks, want %d", len(chunks), tt.expectedChunks)
			}

			if tt.expectedChunks > 0 {
				lastChunk := chunks[len(chunks)-1]
				if len(lastChunk) != tt.lastChunkSize {
					t.Errorf("last chunk size = %d, want %d", len(lastChunk), tt.lastChunkSize)
				}
			}
		})
	}
}

func TestEncodeValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "hogehoge",
			expected: "aG9nZWhvZ2U=",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "unicode string",
			input:    "こんにちは",
			expected: "44GT44KT44Gr44Gh44Gv",
		},
		{
			name:     "special characters",
			input:    "key=value&foo=bar",
			expected: "a2V5PXZhbHVlJmZvbz1iYXI=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := encodeValue(tt.input)
			if result != tt.expected {
				t.Errorf("encodeValue(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
