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

func TestExtractChartsNeedingBump(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected []string
	}{
		{
			name: "Single chart needs bump",
			output: ` mychart => (version: "0.1.0", path: "charts/mychart")
-----------------------------------------------------------
Linting chart "mychart"
Chart version not ok. Needs a version bump`,
			expected: []string{"charts/mychart"},
		},
		{
			name: "Multiple charts need bumps",
			output: ` chart1 => (version: "0.1.0", path: "charts/chart1")
Chart version not ok. Needs a version bump
 chart2 => (version: "0.2.0", path: "charts/chart2")
Chart version not ok. Needs a version bump`,
			expected: []string{"charts/chart1", "charts/chart2"},
		},
		{
			name: "No charts need bumps",
			output: ` mychart => (version: "0.1.0", path: "charts/mychart")
All charts linted successfully`,
			expected: []string(nil),
		},
		{
			name:     "Empty output",
			output:   "",
			expected: []string(nil),
		},
		{
			name: "Chart with other error",
			output: ` mychart => (version: "0.1.0", path: "charts/mychart")
Error: Invalid YAML syntax`,
			expected: []string(nil),
		},
		{
			name: "Duplicate chart path in output",
			output: `Charts to be processed:
 mychart => (version: "0.1.0", path: "charts/mychart")
Linting chart "mychart"
✖︎ mychart => (version: "0.1.0", path: "charts/mychart") > chart version not ok. Needs a version bump!`,
			expected: []string{"charts/mychart"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractChartsNeedingBump(tt.output)
			if len(result) != len(tt.expected) {
				t.Errorf("ExtractChartsNeedingBump() returned %d charts, expected %d", len(result), len(tt.expected))
				return
			}
			// Convert to map for order-independent comparison
			resultMap := make(map[string]bool)
			for _, chartPath := range result {
				resultMap[chartPath] = true
			}
			for _, expectedPath := range tt.expected {
				if !resultMap[expectedPath] {
					t.Errorf("ExtractChartsNeedingBump() missing expected chart: %q", expectedPath)
				}
			}
		})
	}
}
