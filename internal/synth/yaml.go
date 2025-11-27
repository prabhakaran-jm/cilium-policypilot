package synth

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// WritePoliciesToFile writes policies to a YAML file
func WritePoliciesToFile(policies []*Policy, filePath string) error {
	if len(policies) == 0 {
		return fmt.Errorf("no policies to write")
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate YAML content
	var yamlContent strings.Builder

	// Write each policy separated by "---"
	for i, policy := range policies {
		if i > 0 {
			yamlContent.WriteString("---\n")
		}

		data, err := yaml.Marshal(policy)
		if err != nil {
			return fmt.Errorf("failed to marshal policy to YAML: %w", err)
		}

		yamlContent.Write(data)
	}

	// Write to file
	if err := os.WriteFile(filePath, []byte(yamlContent.String()), 0644); err != nil {
		return fmt.Errorf("failed to write policies file: %w", err)
	}

	return nil
}

// PolicyToYAML converts a single policy to YAML string
func PolicyToYAML(policy *Policy) (string, error) {
	data, err := yaml.Marshal(policy)
	if err != nil {
		return "", fmt.Errorf("failed to marshal policy to YAML: %w", err)
	}
	return string(data), nil
}
