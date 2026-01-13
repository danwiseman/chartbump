package detector

import (
	"strings"
)

// versionKeywords are keywords that indicate a version-related issue in helm lint output
var versionKeywords = []string{
	"version",
	"already exists",
	"duplicate version",
	"duplicate",
	"chart version",
	"bump",
}

// DetectVersionIssue checks if the helm lint output contains version-related errors
func DetectVersionIssue(lintOutput string) bool {
	lowerOutput := strings.ToLower(lintOutput)

	for _, keyword := range versionKeywords {
		if strings.Contains(lowerOutput, strings.ToLower(keyword)) {
			return true
		}
	}

	return false
}
