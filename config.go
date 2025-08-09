package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// loadEnvironments loads the environment configuration from the specified file
func loadEnvironments(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config.Environments); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if len(config.Environments) == 0 {
		return nil, fmt.Errorf("no environments defined in config file")
	}

	return &config, nil
}

// getEnvironmentFiles returns the list of files for the specified environment
func getEnvironmentFiles(config *Config, environment string) ([]string, error) {
	files, exists := config.Environments[environment]
	if !exists {
		return nil, buildEnvironmentNotFoundError(config, environment)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files defined for environment '%s'", environment)
	}

	return files, nil
}

// buildEnvironmentNotFoundError creates a helpful error message with available environments
func buildEnvironmentNotFoundError(config *Config, environment string) error {
	availableEnvs := make([]string, 0, len(config.Environments))
	for env := range config.Environments {
		availableEnvs = append(availableEnvs, env)
	}
	return fmt.Errorf("environment '%s' not found. Available environments: %v", environment, availableEnvs)
}

// resolveFilePaths resolves the file paths relative to the config directory
func resolveFilePaths(configPath string, files []string) []string {
	configDir := filepath.Dir(configPath)
	resolvedPaths := make([]string, len(files))

	for i, file := range files {
		if filepath.IsAbs(file) {
			resolvedPaths[i] = file
		} else {
			resolvedPaths[i] = filepath.Join(configDir, "kv-files", file)
		}
	}

	return resolvedPaths
}
