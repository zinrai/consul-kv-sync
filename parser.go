package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// loadYAMLFile loads a single YAML file and returns its content as a map
func loadYAMLFile(filePath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	var content map[string]interface{}
	if err := yaml.Unmarshal(data, &content); err != nil {
		return nil, fmt.Errorf("failed to parse YAML file %s: %w", filePath, err)
	}

	return content, nil
}

// loadAllYAMLFiles loads all YAML files and returns their contents with filenames
func loadAllYAMLFiles(filePaths []string) ([]map[string]interface{}, []string, error) {
	kvMaps := make([]map[string]interface{}, 0, len(filePaths))
	filenames := make([]string, 0, len(filePaths))

	for _, filePath := range filePaths {
		content, err := loadYAMLFile(filePath)
		if err != nil {
			return nil, nil, err
		}

		kvMaps = append(kvMaps, content)
		filenames = append(filenames, filepath.Base(filePath))
	}

	return kvMaps, filenames, nil
}

// flattenKVPairs converts nested map structure to flat key-value pairs
func flattenKVPairs(data map[string]interface{}, prefix string) []KVPair {
	pairs := make([]KVPair, 0) // 空のスライスを初期化（nilではない）

	for key, value := range data {
		fullKey := buildKey(prefix, key)
		pairs = append(pairs, processValue(fullKey, value)...)
	}

	return pairs
}

// buildKey constructs a full key path
func buildKey(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "/" + key
}

// processValue handles different value types and returns KV pairs
func processValue(key string, value interface{}) []KVPair {
	switch v := value.(type) {
	case map[string]interface{}:
		return flattenKVPairs(v, key)
	case map[interface{}]interface{}:
		return processInterfaceMap(key, v)
	default:
		return []KVPair{{Key: key, Value: fmt.Sprintf("%v", value)}}
	}
}

// processInterfaceMap converts map[interface{}]interface{} and processes it
func processInterfaceMap(key string, m map[interface{}]interface{}) []KVPair {
	convertedMap := make(map[string]interface{})
	for k, v := range m {
		if strKey, ok := k.(string); ok {
			convertedMap[strKey] = v
		}
	}
	return flattenKVPairs(convertedMap, key)
}

// collectAllKVPairs collects all KV pairs from multiple YAML files
func collectAllKVPairs(kvMaps []map[string]interface{}) []KVPair {
	var allPairs []KVPair

	for _, kvMap := range kvMaps {
		pairs := flattenKVPairs(kvMap, "")
		allPairs = append(allPairs, pairs...)
	}

	return allPairs
}

// encodeValue encodes a string value to base64
func encodeValue(value string) string {
	return base64.StdEncoding.EncodeToString([]byte(value))
}

// createTransactionOps creates transaction operations from KV pairs
func createTransactionOps(pairs []KVPair) []TxnOp {
	ops := make([]TxnOp, len(pairs))

	for i, pair := range pairs {
		ops[i] = TxnOp{
			KV: &TxnKVOp{
				Verb:  "set",
				Key:   pair.Key,
				Value: encodeValue(pair.Value),
			},
		}
	}

	return ops
}

// chunkOps splits operations into chunks of specified size
func chunkOps(ops []TxnOp, chunkSize int) [][]TxnOp {
	var chunks [][]TxnOp

	for i := 0; i < len(ops); i += chunkSize {
		end := i + chunkSize
		if end > len(ops) {
			end = len(ops)
		}
		chunks = append(chunks, ops[i:end])
	}

	return chunks
}

// formatKVPairsForDisplay formats KV pairs for dry-run display
func formatKVPairsForDisplay(pairs []KVPair) string {
	var sb strings.Builder
	sb.WriteString("Key-Value pairs to be synced:\n")
	sb.WriteString("=" + strings.Repeat("=", 60) + "\n")

	for _, pair := range pairs {
		sb.WriteString(fmt.Sprintf("Key:   %s\n", pair.Key))
		sb.WriteString(fmt.Sprintf("Value: %s\n", pair.Value))
		sb.WriteString(strings.Repeat("-", 60) + "\n")
	}

	sb.WriteString(fmt.Sprintf("\nTotal: %d key-value pairs\n", len(pairs)))
	return sb.String()
}

// exportToJSON exports KV pairs in Consul JSON format to stdout
func exportToJSON(pairs []KVPair) error {
	type consulKV struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	kvData := make([]consulKV, len(pairs))
	for i, pair := range pairs {
		kvData[i] = consulKV{
			Key:   pair.Key,
			Value: encodeValue(pair.Value),
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "\t")
	if err := encoder.Encode(kvData); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}
