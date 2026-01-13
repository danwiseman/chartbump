package detector

import (
	"regexp"
	"strings"
)

// versionBumpPhrases are specific phrases that indicate a version bump is needed
var versionBumpPhrases = []string{
	"chart version not ok",
	"needs a version bump",
	"version bump",
	"chart version \"not ok\"",
}

// DetectVersionIssue checks if the ct lint output contains the specific error
// indicating that a version bump is needed. This is more precise than checking
// for generic keywords to avoid false positives from missing tools or other errors.
func DetectVersionIssue(lintOutput string) bool {
	lowerOutput := strings.ToLower(lintOutput)

	// Check for specific version bump phrases
	for _, phrase := range versionBumpPhrases {
		if strings.Contains(lowerOutput, strings.ToLower(phrase)) {
			return true
		}
	}

	return false
}

// ExtractChartsNeedingBump parses ct lint output to extract the paths of charts
// that need version bumps. Returns a slice of chart directory paths.
func ExtractChartsNeedingBump(lintOutput string) []string {
	var charts []string
	lines := strings.Split(lintOutput, "\n")

	// Track the current chart being processed
	var currentChartPath string

	// Regex to extract chart path from: mychart => (version: "0.1.0", path: "charts/mychart")
	pathRegex := regexp.MustCompile(`path:\s*"([^"]+)"`)

	for _, line := range lines {
		// Check if this line declares a chart being processed
		if matches := pathRegex.FindStringSubmatch(line); len(matches) > 1 {
			currentChartPath = matches[1]
		}

		// Check if this line indicates a version bump is needed
		lowerLine := strings.ToLower(line)
		for _, phrase := range versionBumpPhrases {
			if strings.Contains(lowerLine, strings.ToLower(phrase)) && currentChartPath != "" {
				charts = append(charts, currentChartPath)
				currentChartPath = "" // Reset to avoid duplicates
				break
			}
		}
	}

	return charts
}
