package validator

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ValidatorMetadata contains all validator configuration
// This is the single source of truth for validator properties
type ValidatorMetadata struct {
	Name        string   // Unique identifier (e.g., "wif-check")
	Description string   // Human-readable description
	RunAfter    []string // Validators this should run after (dependencies)
	Tags        []string // For grouping/filtering (e.g., "mvp", "network", "quota")
}

// Validator is the core interface all validators must implement
type Validator interface {
	// Metadata returns validator configuration (name, dependencies, etc.)
	Metadata() ValidatorMetadata

	// Enabled determines if this validator should run based on context/config
	Enabled(ctx *Context) bool

	// Validate performs the actual validation logic
	Validate(ctx context.Context, vctx *Context) *Result
}

// Status represents the validation outcome
type Status string

const (
	StatusSuccess Status = "success"
	StatusFailure Status = "failure"
	StatusSkipped Status = "skipped"
)

// Result represents the outcome of a single validator
type Result struct {
	ValidatorName string                 `json:"validator_name"`
	Status        Status                 `json:"status"`
	Reason        string                 `json:"reason"`
	Message       string                 `json:"message"`
	Details       map[string]interface{} `json:"details,omitempty"`
	Duration      time.Duration          `json:"duration_ns"`
	Timestamp     time.Time              `json:"timestamp"`
}

// AggregatedResult combines all validator results into the expected output format
type AggregatedResult struct {
	Status  Status                 `json:"status"`
	Reason  string                 `json:"reason"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details"`
}

// Aggregate combines multiple validator results into final output
func Aggregate(results []*Result) *AggregatedResult {
	checksRun := len(results)
	checksPassed := 0
	var failedChecks []string
	var failureDescriptions []string

	// Single pass to collect all failure information
	for _, r := range results {
		switch r.Status {
		case StatusSuccess:
			checksPassed++
		case StatusFailure:
			failedChecks = append(failedChecks, r.ValidatorName)
			failureDescriptions = append(failureDescriptions, fmt.Sprintf("%s (%s)", r.ValidatorName, r.Reason))
		}
	}

	details := map[string]interface{}{
		"checks_run":    checksRun,
		"checks_passed": checksPassed,
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
		"validators":    results,
	}

	if checksPassed == checksRun {
		return &AggregatedResult{
			Status:  StatusSuccess,
			Reason:  "ValidationPassed",
			Message: "All GCP validation checks passed successfully",
			Details: details,
		}
	}

	details["failed_checks"] = failedChecks

	// Build informative failure message with pass ratio and reasons
	message := fmt.Sprintf("%d validation check(s) failed: %s. Passed: %d/%d",
		len(failureDescriptions),
		strings.Join(failureDescriptions, ", "),
		checksPassed,
		checksRun)

	return &AggregatedResult{
		Status:  StatusFailure,
		Reason:  "ValidationFailed",
		Message: message,
		Details: details,
	}
}
