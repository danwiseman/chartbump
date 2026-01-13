package chart

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Chart represents the structure of a Helm Chart.yaml file
type Chart struct {
	APIVersion  string `yaml:"apiVersion"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Type        string `yaml:"type,omitempty"`
	Version     string `yaml:"version"`
	AppVersion  string `yaml:"appVersion,omitempty"`
}

// ReadChart reads and parses a Chart.yaml file from the given directory
func ReadChart(chartPath string) (*Chart, error) {
	chartFile := filepath.Join(chartPath, "Chart.yaml")

	data, err := os.ReadFile(chartFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read Chart.yaml: %w", err)
	}

	var chart Chart
	if err := yaml.Unmarshal(data, &chart); err != nil {
		return nil, fmt.Errorf("failed to parse Chart.yaml: %w", err)
	}

	if chart.Version == "" {
		return nil, fmt.Errorf("Chart.yaml does not contain a version field")
	}

	return &chart, nil
}

// UpdateChartVersion updates the version in Chart.yaml while preserving formatting
func UpdateChartVersion(chartPath, newVersion string) error {
	chartFile := filepath.Join(chartPath, "Chart.yaml")

	// Read the file
	data, err := os.ReadFile(chartFile)
	if err != nil {
		return fmt.Errorf("failed to read Chart.yaml: %w", err)
	}

	// Parse with yaml.v3 to preserve structure
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return fmt.Errorf("failed to parse Chart.yaml: %w", err)
	}

	// Find and update the version field
	if err := updateVersionInNode(&node, newVersion); err != nil {
		return err
	}

	// Marshal back to YAML
	newData, err := yaml.Marshal(&node)
	if err != nil {
		return fmt.Errorf("failed to marshal updated Chart.yaml: %w", err)
	}

	// Write back to file
	if err := os.WriteFile(chartFile, newData, 0644); err != nil {
		return fmt.Errorf("failed to write Chart.yaml: %w", err)
	}

	return nil
}

// updateVersionInNode recursively finds and updates the version field in the YAML node
func updateVersionInNode(node *yaml.Node, newVersion string) error {
	if node.Kind == yaml.DocumentNode {
		if len(node.Content) > 0 {
			return updateVersionInNode(node.Content[0], newVersion)
		}
		return fmt.Errorf("empty document")
	}

	if node.Kind == yaml.MappingNode {
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			if keyNode.Value == "version" {
				valueNode.Value = newVersion
				return nil
			}
		}
	}

	return fmt.Errorf("version field not found in Chart.yaml")
}
