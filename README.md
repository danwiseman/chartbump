# ChartBump

A Go CLI tool that automatically bumps Helm chart versions when version-related issues are detected by `ct lint` (chart-testing).

## Features

- **Two modes of operation:**
  - Single chart mode: Target a specific chart directory
  - Auto-detect mode: Automatically detect and bump all changed charts in a repository
- Runs `helm dep update` to ensure dependencies are current (single chart mode)
- Executes `ct lint` (chart-testing) to check for issues
- Supports git branch comparison for chart-testing
- Detects version-related problems in lint output
- Automatically bumps the patch version (e.g., 0.1.2 → 0.1.3)
- Preserves Chart.yaml formatting and comments
- Supports dry-run mode to preview changes

## Prerequisites

- Go 1.16 or higher
- Helm 3.x installed and available in PATH
- chart-testing (ct) installed and available in PATH ([Installation Guide](https://github.com/helm/chart-testing))

## Installation

```bash
go build -o chartbump
```

Or install directly:

```bash
go install
```

## Usage

### Single Chart Mode

Target a specific chart directory:

```bash
./chartbump /path/to/helm/chart
```

With git branch comparison:

```bash
./chartbump --target-branch main /path/to/helm/chart
```

Dry run (preview changes without modifying files):

```bash
./chartbump --dry-run /path/to/helm/chart
```

### Auto-Detect Mode

Automatically detect and bump all changed charts in the repository:

```bash
./chartbump --target-branch main
```

This mode runs `ct lint --target-branch main` without specifying a chart directory, allowing ct to automatically detect changed charts based on git history. All charts that need version bumps will be processed automatically.

Dry run in auto-detect mode:

```bash
./chartbump --target-branch main --dry-run
```

**Note:** `--target-branch` is required when no chart directory is specified.

### Help

Show help:

```bash
./chartbump --help
```

## How It Works

### Single Chart Mode

1. **Validates** that Chart.yaml exists in the target directory
2. **Runs** `helm dep update` to update chart dependencies
3. **Executes** `ct lint` (chart-testing) and captures the output
   - Uses `--target-branch` flag if specified for git-based comparison
4. **Analyzes** the lint output for specific version bump requirement:
   - Only bumps if output contains: "chart version not ok" or "Needs a version bump"
   - Will NOT bump for other errors (missing yamllint/yamale, validation errors, etc.)
5. **Bumps** the patch version if version bump is specifically required
6. **Updates** Chart.yaml with the new version
7. **Verifies** the fix by running `ct lint` again

### Auto-Detect Mode

1. **Executes** `ct lint --target-branch <branch>` without specifying charts
2. **Parses** the output to identify all charts that need version bumps
3. **For each chart** that needs a version bump:
   - Reads the current version from Chart.yaml
   - Bumps the patch version
   - Updates Chart.yaml with the new version
4. **Reports** summary of successful and failed bumps

## Example

Given a Chart.yaml with version `0.1.2`:

```yaml
apiVersion: v2
name: my-chart
description: A Helm chart
version: 0.1.2
```

If `ct lint` reports a version-related issue, chartbump will update it to `0.1.3`:

```yaml
apiVersion: v2
name: my-chart
description: A Helm chart
version: 0.1.3
```

## Project Structure

```
chartbump/
├── main.go              # Entry point
├── cmd/
│   └── root.go          # Cobra CLI command
├── pkg/
│   ├── helm/
│   │   └── client.go    # Helm command execution
│   ├── chart/
│   │   ├── parser.go    # Chart.yaml parsing
│   │   └── version.go   # Version bumping logic
│   └── detector/
│       └── detector.go  # Version issue detection
└── test-chart/          # Example chart for testing
```

## Testing

A test chart is included in the `test-chart/` directory. To test the tool:

```bash
./chartbump test-chart
```

Note: This will only bump the version if `ct lint` detects a version-related issue.

## Behavior

### When Version Bump is Needed
If `ct lint` outputs an error containing "chart version not ok" or "Needs a version bump", the tool will:
- Bump the patch version in Chart.yaml
- Update the file
- Re-run ct lint to verify the fix

### When Other Errors Occur
If `ct lint` fails for other reasons (missing tools, validation errors, etc.), the tool will:
- Display the lint output
- Show common reasons for lint failure
- Exit without modifying Chart.yaml

Example output when yamllint is missing:
```
✗ ct lint failed, but does not require a version bump

Common reasons for lint failure:
  - Missing required tools (yamllint, yamale)
  - Chart validation errors
  - YAML syntax issues

No version bump will be performed.
```

## License

MIT
