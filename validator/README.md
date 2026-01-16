# GCP Validator

Extensible Go-based validation framework for GCP prerequisites before cluster provisioning.

## Features

- **Parallel Execution**: Validators run concurrently when dependencies allow
- **Dependency Management**: DAG-based scheduling with automatic cycle detection
- **Auto-discovery**: Validators self-register via `init()`
- **WIF Authentication**: Workload Identity Federation for secure GCP access

## Current Validators

1. **api-enabled**: Verifies required GCP APIs are enabled
2. **quota-check**: Placeholder stub for future quota validation

## Quick Start

### Build

```bash
make build  # Build binary
make test   # Run tests
make image  # Build container image
```

### Run Locally

```bash
export PROJECT_ID=my-gcp-project
export RESULTS_PATH=/tmp/results.json

./bin/validator
cat /tmp/results.json
```

### Run in Docker

```bash
docker run --rm \
  -e PROJECT_ID=my-project \
  -v /tmp/results:/results \
  gcp-validator
```

## Configuration

### Required
- `PROJECT_ID` - GCP project ID to validate

### Optional
- `RESULTS_PATH` - Output file path (default: `/results/adapter-result.json`)
- `DISABLED_VALIDATORS` - Comma-separated list to disable (e.g., `quota-check`)
- `STOP_ON_FIRST_FAILURE` - Stop on first failure (default: `false`)
- `REQUIRED_APIS` - APIs to check (default: `compute.googleapis.com,iam.googleapis.com,cloudresourcemanager.googleapis.com`)
- `LOG_LEVEL` - Log level: `debug`, `info`, `warn`, `error` (default: `info`)

## Output Format

### Success
```json
{
  "status": "success",
  "reason": "ValidationPassed",
  "message": "All GCP validation checks passed successfully",
  "details": {
    "checks_run": 1,
    "checks_passed": 1,
    "timestamp": "2026-01-15T10:30:00Z",
    "validators": [
      {
        "validator_name": "api-enabled",
        "status": "success",
        "reason": "AllAPIsEnabled",
        "message": "All 3 required APIs are enabled",
        "duration_ns": 234000000,
        "timestamp": "2026-01-15T10:30:00Z"
      }
    ]
  }
}
```

### Failure
```json
{
  "status": "failure",
  "reason": "ValidationFailed",
  "message": "1 validation check(s) failed: api-enabled (forbidden). Passed: 0/1",
  "details": {
    "checks_run": 1,
    "checks_passed": 0,
    "failed_checks": ["api-enabled"],
    "timestamp": "2026-01-15T10:30:00Z",
    "validators": [
      {
        "validator_name": "api-enabled",
        "status": "failure",
        "reason": "forbidden",
        "message": "Failed to check API compute.googleapis.com: ...",
        "duration_ns": 123000000,
        "timestamp": "2026-01-15T10:30:00Z"
      }
    ]
  }
}
```

## Adding a New Validator

Create a file in `pkg/validators/` implementing the `Validator` interface:

```go
package validators

import (
    "context"
    "validator/pkg/validator"
)

type MyValidator struct{}

func init() {
    validator.Register(&MyValidator{})
}

func (v *MyValidator) Metadata() validator.ValidatorMetadata {
    return validator.ValidatorMetadata{
        Name:        "my-validator",
        Description: "Validates something important",
        RunAfter:    []string{"api-enabled"},  // Dependencies
        Tags:        []string{"custom"},
    }
}

func (v *MyValidator) Enabled(ctx *validator.Context) bool {
    return ctx.Config.IsValidatorEnabled("my-validator")
}

func (v *MyValidator) Validate(ctx context.Context, vctx *validator.Context) *validator.Result {
    // Validation logic here
    return &validator.Result{
        Status:  validator.StatusSuccess,
        Reason:  "CheckPassed",
        Message: "Validation successful",
    }
}
```

The validator is automatically discovered, ordered by dependencies, and executed in parallel.
- Register validator via `init()`
- Define dependency via `RunAfter` in `Metadata`

## Testing

```bash
make test          # Run all tests
make lint          # Run linter
```

Tests use Ginkgo/Gomega BDD framework.

## Architecture

### Execution Flow
1. Load configuration from environment variables
2. Discover and register all validators via `init()`
3. Build dependency graph (DAG) and detect cycles
4. Execute validators in parallel by dependency level
5. Aggregate results and write to output file

### Security
- Uses GCP Application Default Credentials (ADC)
- Supports Workload Identity Federation in Kubernetes
- Minimal read-only scopes per service
- Each validator gets only the permissions it needs
