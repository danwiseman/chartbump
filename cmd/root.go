package cmd

import (
	"fmt"
	"os"
	"path/filepath"

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
bumps the patch version in Chart.yaml if version-related issues are detected.`,
	Args: cobra.ExactArgs(1),
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
	chartPath := args[0]

	// Validate that Chart.yaml exists
	chartFile := filepath.Join(chartPath, "Chart.yaml")
	if _, err := os.Stat(chartFile); os.IsNotExist(err) {
		return fmt.Errorf("Chart.yaml not found in directory: %s", chartPath)
	}

	fmt.Printf("Running chartbump on: %s\n\n", chartPath)

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
	fmt.Println("\nChecking for version-related issues...")
	if !detector.DetectVersionIssue(lintOutput) {
		fmt.Println("✗ Lint failed, but no version-related issues detected")
		fmt.Println("No version bump will be performed")
		return fmt.Errorf("ct lint failed with non-version issues")
	}

	fmt.Println("✓ Version-related issue detected")

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
