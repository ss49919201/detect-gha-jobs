# detect-gha-jobs

A CLI tool to extract job information from GitHub Actions workflow files

## Installation

```bash
go install github.com/ss49919201/detect-gha-jobs@latest
```

## Usage

```bash
detect-gha-jobs <workflow-file-or-directory>
```

### Examples

```bash
# Process a single workflow file
detect-gha-jobs .github/workflows/ci.yml

# Process workflow files recursively in a directory
detect-gha-jobs .
```

## Output Format

- For a single file:

  - Workflow name
  - Job information:
    - Job ID
    - Job name (if specified)

- For a directory:
  - Processes all .yml/.yaml files in .github/workflows directory
  - For each file:
    - File path
    - Workflow name
    - Job information:
      - Job ID
      - Job name (if specified)
