package chart

import (
	"fmt"
	"strconv"
	"strings"
)

// BumpPatch increments the patch version of a semantic version string.
// Example: "0.1.2" becomes "0.1.3"
func BumpPatch(version string) (string, error) {
	version = strings.TrimSpace(version)

	// Handle version strings with 'v' prefix
	hasPrefix := false
	if strings.HasPrefix(version, "v") {
		hasPrefix = true
		version = strings.TrimPrefix(version, "v")
	}

	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid semantic version format: %s (expected X.Y.Z)", version)
	}

	// Parse major, minor, patch
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid minor version: %s", parts[1])
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", fmt.Errorf("invalid patch version: %s", parts[2])
	}

	// Increment patch
	patch++

	newVersion := fmt.Sprintf("%d.%d.%d", major, minor, patch)
	if hasPrefix {
		newVersion = "v" + newVersion
	}

	return newVersion, nil
}
