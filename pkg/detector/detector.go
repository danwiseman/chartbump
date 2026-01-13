package detector

import (
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
