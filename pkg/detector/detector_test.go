package detector

import "testing"

func TestDetectVersionIssue(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected bool
	}{
		{
			name:     "Version bump required",
			output:   "Chart version not ok. Needs a version bump",
			expected: true,
		},
		{
			name:     "Chart version not ok lowercase",
			output:   "chart version not ok",
			expected: true,
		},
		{
			name:     "Needs a version bump",
			output:   "Some other text. Needs a version bump.",
			expected: true,
		},
		{
			name:     "Version bump mentioned",
			output:   "Please apply a version bump to continue",
			expected: true,
		},
		{
			name:     "Missing yamllint",
			output:   "Error: yamllint is not installed",
			expected: false,
		},
		{
			name:     "Missing yamale",
			output:   "Error: yamale is not installed",
			expected: false,
		},
		{
			name:     "Generic error with version word",
			output:   "Error checking version compatibility",
			expected: false,
		},
		{
			name:     "YAML syntax error",
			output:   "Error: Invalid YAML syntax at line 5",
			expected: false,
		},
		{
			name:     "Lint passed",
			output:   "All charts linted successfully",
			expected: false,
		},
		{
			name:     "Empty output",
			output:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectVersionIssue(tt.output)
			if result != tt.expected {
				t.Errorf("DetectVersionIssue(%q) = %v, expected %v", tt.output, result, tt.expected)
			}
		})
	}
}
