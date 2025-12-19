package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadConfig reads the configuration file at the given path,
// processes imports recursively, and returns the merged configuration.
func LoadConfig(path string) (*Config, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	visited := make(map[string]bool)
	return loadConfigRecursive(absPath, visited)
}

func loadConfigRecursive(path string, visited map[string]bool) (*Config, error) {
	if visited[path] {
		return nil, fmt.Errorf("circular import detected at %s", path)
	}
	visited[path] = true

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	baseDir := filepath.Dir(path)
	mergedCfg := &Config{
		Vars:     make(map[string]string),
		Commands: []CommandConfig{},
	}

	// 1. Process imports first (files imported earlier in the list are processed first,
	// but usually we want local config to override imports, or imports to extend?
	// Common pattern: Imports are "bases".
	// Let's load imports and merge them INTO the current config.
	for _, importPath := range cfg.Imports {
		// Resolve relative path
		fullImportPath := importPath
		if !filepath.IsAbs(importPath) {
			fullImportPath = filepath.Join(baseDir, importPath)
		}

		importedCfg, err := loadConfigRecursive(fullImportPath, visited)
		if err != nil {
			return nil, err
		}

		// Merge imported config into mergedCfg
		mergeConfigs(mergedCfg, importedCfg)
	}

	// 2. Merge the current file's content ON TOP of the imports
	// However, for commands, we probably want to accumulate them.
	// For vars, local overrides imports? Yes.
	mergeConfigs(mergedCfg, &cfg)

	return mergedCfg, nil
}

func mergeConfigs(base, override *Config) {
	// Merge Vars
	for k, v := range override.Vars {
		base.Vars[k] = v
	}

	// Append Commands
	// We might want to deduplicate by name, but for now just appending allows overrides?
	// Cobra will handle duplicate names by crashing or ignoring.
	// Let's just append for now, but in root.go we handle duplication.
	base.Commands = append(base.Commands, override.Commands...)
}
