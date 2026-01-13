package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/danwiseman/chartbump/pkg/chart"
	"github.com/danwiseman/chartbump/pkg/detector"
	"github.com/danwiseman/chartbump/pkg/helm"
	"github.com/spf13/cobra"
)

var (
	dryRun       bool
	targetBranch string
)

var rootCmd = &cobra.Command{
	Use:   "chartbump [chart-directory]",
	Short: "Automatically bump Helm chart versions based on ct lint output",
	Long: `chartbump runs ct lint (chart-testing) on a target directory and automatically
bumps the patch version in Chart.yaml if version-related issues are detected.

If no chart directory is specified, ct lint will run with --target-branch to
auto-detect changed charts in the repository.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runChartBump,
}

func init() {
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would change without modifying files")
	rootCmd.Flags().StringVar(&targetBranch, "target-branch", "", "Git branch to compare against (for ct lint)")
}

func Execute() error {
	return rootCmd.Execute()
}

func runChartBump(cmd *cobra.Command, args []string) error {
	// If no directory provided, run in auto-detect mode
	if len(args) == 0 {
		if targetBranch == "" {
			return fmt.Errorf("--target-branch is required when no chart directory is specified")
		}
		return runAutoDetectMode()
	}

	// Single chart mode
	chartPath := args[0]

	// Validate that Chart.yaml exists
	chartFile := filepath.Join(chartPath, "Chart.yaml")
	if _, err := os.Stat(chartFile); os.IsNotExist(err) {
		return fmt.Errorf("Chart.yaml not found in directory: %s", chartPath)
	}

	fmt.Printf("Running chartbump on: %s\n\n", chartPath)
	return runSingleChartMode(chartPath)
}

func runSingleChartMode(chartPath string) error {
	// Step 1: Run helm dep update
	fmt.Println("Running helm dep update...")
	if err := helm.RunDepUpdate(chartPath); err != nil {
		fmt.Printf("Warning: %v\n", err)
		// Continue even if dep update fails - chart might not have dependencies
	} else {
		fmt.Println("✓ Dependencies updated")
	}

	// Step 2: Run ct lint
	fmt.Println("\nRunning ct lint...")
	if targetBranch != "" {
		fmt.Printf("Target branch: %s\n", targetBranch)
	}
	lintOutput, lintErr := helm.RunCTLint(chartPath, targetBranch)

	if lintErr == nil {
		fmt.Println("✓ ct lint passed - no version bump needed")
		return nil
	}

	fmt.Printf("ct lint output:\n%s\n", lintOutput)

	// Step 3: Detect version issue
	fmt.Println("\nChecking for version bump requirement...")
	if !detector.DetectVersionIssue(lintOutput) {
		fmt.Println("✗ ct lint failed, but does not require a version bump")
		fmt.Println("\nCommon reasons for lint failure:")
		fmt.Println("  - Missing required tools (yamllint, yamale)")
		fmt.Println("  - Chart validation errors")
		fmt.Println("  - YAML syntax issues")
		fmt.Println("\nNo version bump will be performed.")
		return fmt.Errorf("ct lint failed with non-version issues")
	}

	fmt.Println("✓ Version bump required - proceeding with patch version bump")

	// Step 4: Read current chart
	fmt.Println("\nReading Chart.yaml...")
	currentChart, err := chart.ReadChart(chartPath)
	if err != nil {
		return fmt.Errorf("failed to read chart: %w", err)
	}

	fmt.Printf("Current version: %s\n", currentChart.Version)

	// Step 5: Bump version
	newVersion, err := chart.BumpPatch(currentChart.Version)
	if err != nil {
		return fmt.Errorf("failed to bump version: %w", err)
	}

	fmt.Printf("New version: %s\n", newVersion)

	// Step 6: Update Chart.yaml (if not dry-run)
	if dryRun {
		fmt.Println("\n[DRY RUN] Would update Chart.yaml with new version")
		return nil
	}

	fmt.Println("\nUpdating Chart.yaml...")
	if err := chart.UpdateChartVersion(chartPath, newVersion); err != nil {
		return fmt.Errorf("failed to update chart version: %w", err)
	}

	fmt.Println("✓ Chart.yaml updated successfully")

	// Step 7: Verify with ct lint
	fmt.Println("\nVerifying with ct lint...")
	verifyOutput, verifyErr := helm.RunCTLint(chartPath, targetBranch)
	if verifyErr != nil {
		fmt.Printf("Warning: Lint still shows issues:\n%s\n", verifyOutput)
		return fmt.Errorf("version bump completed but lint still fails")
	}

	fmt.Println("✓ Verification successful - chart now passes ct lint")
	fmt.Printf("\n✨ Successfully bumped version from %s to %s\n", currentChart.Version, newVersion)

	return nil
}

func runAutoDetectMode() error {
	fmt.Println("Running chartbump in auto-detect mode")
	fmt.Printf("Target branch: %s\n\n", targetBranch)

	// Run ct lint without specifying charts (auto-detect changed charts)
	fmt.Println("Running ct lint to detect changed charts...")
	lintOutput, lintErr := helm.RunCTLint("", targetBranch)

	if lintErr == nil {
		fmt.Println("✓ All charts passed ct lint - no version bumps needed")
		return nil
	}

	fmt.Printf("ct lint output:\n%s\n", lintOutput)

	// Extract charts that need version bumps
	fmt.Println("\nAnalyzing charts for version bump requirements...")
	chartsNeedingBump := detector.ExtractChartsNeedingBump(lintOutput)

	if len(chartsNeedingBump) == 0 {
		fmt.Println("✗ ct lint failed, but no charts require version bumps")
		fmt.Println("\nCommon reasons for lint failure:")
		fmt.Println("  - Missing required tools (yamllint, yamale)")
		fmt.Println("  - Chart validation errors")
		fmt.Println("  - YAML syntax issues")
		fmt.Println("\nNo version bumps will be performed.")
		return fmt.Errorf("ct lint failed with non-version issues")
	}

	fmt.Printf("Found %d chart(s) requiring version bumps:\n", len(chartsNeedingBump))
	for _, chartPath := range chartsNeedingBump {
		fmt.Printf("  - %s\n", chartPath)
	}

	// Process each chart
	var successCount, failCount int
	for _, chartPath := range chartsNeedingBump {
		fmt.Printf("\n%s Processing %s %s\n", strings.Repeat("=", 20), chartPath, strings.Repeat("=", 20))

		if err := bumpChartVersion(chartPath); err != nil {
			fmt.Printf("✗ Failed to bump %s: %v\n", chartPath, err)
			failCount++
		} else {
			fmt.Printf("✓ Successfully bumped %s\n", chartPath)
			successCount++
		}
	}

	// Summary
	fmt.Printf("\n%s Summary %s\n", strings.Repeat("=", 30), strings.Repeat("=", 30))
	fmt.Printf("Successfully bumped: %d\n", successCount)
	fmt.Printf("Failed: %d\n", failCount)

	if failCount > 0 {
		return fmt.Errorf("failed to bump %d chart(s)", failCount)
	}

	return nil
}

func bumpChartVersion(chartPath string) error {
	// Read current chart
	currentChart, err := chart.ReadChart(chartPath)
	if err != nil {
		return fmt.Errorf("failed to read chart: %w", err)
	}

	fmt.Printf("  Current version: %s\n", currentChart.Version)

	// Bump version
	newVersion, err := chart.BumpPatch(currentChart.Version)
	if err != nil {
		return fmt.Errorf("failed to bump version: %w", err)
	}

	fmt.Printf("  New version: %s\n", newVersion)

	// Update Chart.yaml (if not dry-run)
	if dryRun {
		fmt.Println("  [DRY RUN] Would update Chart.yaml with new version")
		return nil
	}

	if err := chart.UpdateChartVersion(chartPath, newVersion); err != nil {
		return fmt.Errorf("failed to update chart version: %w", err)
	}

	return nil
}
