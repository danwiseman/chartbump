package helm

import (
	"bytes"
	"fmt"
	"os/exec"
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

// RunCTLint runs 'ct lint' (chart-testing) on the specified chart directory and returns the output
func RunCTLint(chartPath, targetBranch string) (string, error) {
	args := []string{"lint"}

	if targetBranch != "" {
		args = append(args, "--target-branch", targetBranch)
	}

	args = append(args, "--charts", chartPath)

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
