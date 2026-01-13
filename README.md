# ChartBump

A Go CLI tool that automatically bumps Helm chart versions when version-related issues are detected by `ct lint` (chart-testing).

## Features

- Runs `helm dep update` to ensure dependencies are current
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

Basic usage:

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

Combine flags:

```bash
./chartbump --target-branch main --dry-run /path/to/helm/chart
```

Show help:

```bash
./chartbump --help
```

## How It Works

1. **Validates** that Chart.yaml exists in the target directory
2. **Runs** `helm dep update` to update chart dependencies
3. **Executes** `ct lint` (chart-testing) and captures the output
   - Uses `--target-branch` flag if specified for git-based comparison
4. **Analyzes** the lint output for version-related keywords:
   - "version"
   - "already exists"
   - "duplicate"
   - "chart version"
   - "bump"
5. **Bumps** the patch version if a version issue is detected
6. **Updates** Chart.yaml with the new version
7. **Verifies** the fix by running `ct lint` again

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

## License

MIT
