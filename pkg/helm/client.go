package helm

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// RunDepUpdate runs 'helm dep update' on the specified chart directory
func RunDepUpdate(chartPath string) error {
	cmd := exec.Command("helm", "dep", "update", chartPath)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("helm dep update failed: %w\nOutput: %s", err, stderr.String())
	}

	return nil
}

// RunDepUpdateAll iterates over all immediate subdirectories in basePath
// and runs 'helm dep update' on each. This is useful when ct lint fails
// due to missing dependencies across multiple charts.
func RunDepUpdateAll(basePath string) error {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", basePath, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		chartDir := filepath.Join(basePath, entry.Name())
		// Only run dep update if the directory contains a Chart.yaml
		if _, err := os.Stat(filepath.Join(chartDir, "Chart.yaml")); os.IsNotExist(err) {
			continue
		}
		if err := RunDepUpdate(chartDir); err != nil {
			fmt.Printf("Warning: dep update failed for %s: %v\n", chartDir, err)
			// Continue with other charts even if one fails
		}
	}

	return nil
}

// RunCTLint runs 'ct lint' (chart-testing) on the specified chart directory and returns the output.
// If chartPath is empty, ct lint will run without --charts flag to auto-detect changed charts.
func RunCTLint(chartPath, targetBranch string) (string, error) {
	args := []string{"lint"}

	if targetBranch != "" {
		args = append(args, "--target-branch", targetBranch)
	}

	if chartPath != "" {
		args = append(args, "--charts", chartPath)
	}

	cmd := exec.Command("ct", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Combine stdout and stderr for complete output
	output := stdout.String() + stderr.String()

	// ct lint returns non-zero exit code if there are errors,
	// but we still want to capture and return the output
	if err != nil {
		// Return output even if there's an error, as we need to parse it
		return output, fmt.Errorf("ct lint encountered issues: %w", err)
	}

	return output, nil
}

// RunHelmDocs runs 'helm-docs' on the specified chart directory
func RunHelmDocs(chartPath string) error {
	cmd := exec.Command("helm-docs", chartPath)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("helm-docs failed: %w\nOutput: %s", err, stderr.String())
	}

	return nil
}
