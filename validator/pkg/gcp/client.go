package gcp

import (
    "context"
    "fmt"
    "log/slog"
    "net/http"
    "time"

    "golang.org/x/oauth2/google"
    "google.golang.org/api/cloudresourcemanager/v1"
    "google.golang.org/api/compute/v1"
    "google.golang.org/api/googleapi"
    "google.golang.org/api/iam/v1"
    "google.golang.org/api/monitoring/v3"
    "google.golang.org/api/option"
    "google.golang.org/api/serviceusage/v1"
)

const (
    // Retry configuration
    initialBackoff = 100 * time.Millisecond
    maxBackoff     = 30 * time.Second
    maxRetries     = 5

    // Retryable HTTP status codes
    statusRateLimited    = 429
    statusServiceUnavail = 503
    statusInternalError  = 500
)

// getDefaultClient creates an HTTP client with WIF authentication
// Creates a new client for each call with the specified scopes
// google.DefaultClient handles connection pooling and credential caching internally
func getDefaultClient(ctx context.Context, scopes ...string) (*http.Client, error) {
    return google.DefaultClient(ctx, scopes...)
}

// retryWithBackoff wraps GCP API calls with exponential backoff retry logic
func retryWithBackoff(ctx context.Context, operation func() error) error {
    var lastErr error
    backoff := initialBackoff

    for attempt := 0; attempt < maxRetries; attempt++ {
        if attempt > 0 {
            // Calculate exponential backoff with jitter
            if backoff < maxBackoff {
                backoff = backoff * 2
                if backoff > maxBackoff {
                    backoff = maxBackoff
                }
            }
            slog.Debug("Retrying GCP API call", "attempt", attempt, "backoff", backoff)

            select {
            case <-time.After(backoff):
            case <-ctx.Done():
                return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
            }
        }

        lastErr = operation()
        if lastErr == nil {
            return nil // Success
        }

        // Check if error is retryable
        if apiErr, ok := lastErr.(*googleapi.Error); ok {
            // Retry on rate limit, service unavailable, and internal errors
            if apiErr.Code == statusRateLimited ||
               apiErr.Code == statusServiceUnavail ||
               apiErr.Code == statusInternalError {
                continue
            }
            // Don't retry on other errors (4xx client errors, etc.)
            return lastErr
        }

        // Retry on network/context errors
        if ctx.Err() != nil {
            return fmt.Errorf("context error: %w", ctx.Err())
        }
    }

    return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// ClientFactory creates GCP service clients with WIF authentication
type ClientFactory struct {
    projectID string
    logger    *slog.Logger
}

// NewClientFactory creates a new GCP client factory
func NewClientFactory(projectID string, logger *slog.Logger) *ClientFactory {
    return &ClientFactory{
        projectID: projectID,
        logger:    logger,
    }
}

// CreateComputeService creates a Compute Engine service client with minimal scopes
func (f *ClientFactory) CreateComputeService(ctx context.Context) (*compute.Service, error) {
    f.logger.Debug("Creating Compute Engine service client with WIF")

    // Use readonly scope for read-only operations (quota checks, list instances, etc.)
    client, err := getDefaultClient(ctx, compute.ComputeReadonlyScope)
    if err != nil {
        return nil, fmt.Errorf("failed to create default client: %w", err)
    }

    var svc *compute.Service
    err = retryWithBackoff(ctx, func() error {
        var createErr error
        svc, createErr = compute.NewService(ctx, option.WithHTTPClient(client))
        return createErr
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create compute service: %w", err)
    }

    return svc, nil
}

// CreateIAMService creates an IAM service client with minimal scopes
func (f *ClientFactory) CreateIAMService(ctx context.Context) (*iam.Service, error) {
    f.logger.Debug("Creating IAM service client with WIF")

    // Use readonly scope for validation (checking service accounts, roles, etc.)
    client, err := getDefaultClient(ctx, "https://www.googleapis.com/auth/cloud-platform.read-only")
    if err != nil {
        return nil, fmt.Errorf("failed to create default client: %w", err)
    }

    var svc *iam.Service
    err = retryWithBackoff(ctx, func() error {
        var createErr error
        svc, createErr = iam.NewService(ctx, option.WithHTTPClient(client))
        return createErr
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create IAM service: %w", err)
    }

    return svc, nil
}

// CreateCloudResourceManagerService creates a Cloud Resource Manager service client with minimal scopes
func (f *ClientFactory) CreateCloudResourceManagerService(ctx context.Context) (*cloudresourcemanager.Service, error) {
    f.logger.Debug("Creating Cloud Resource Manager service client with WIF")

    // Use readonly scope for read-only project operations
    client, err := getDefaultClient(ctx, cloudresourcemanager.CloudPlatformReadOnlyScope)
    if err != nil {
        return nil, fmt.Errorf("failed to create default client: %w", err)
    }

    var svc *cloudresourcemanager.Service
    err = retryWithBackoff(ctx, func() error {
        var createErr error
        svc, createErr = cloudresourcemanager.NewService(ctx, option.WithHTTPClient(client))
        return createErr
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create cloud resource manager service: %w", err)
    }

    return svc, nil
}

// CreateServiceUsageService creates a Service Usage service client with minimal scopes
func (f *ClientFactory) CreateServiceUsageService(ctx context.Context) (*serviceusage.Service, error) {
    f.logger.Debug("Creating Service Usage service client with WIF")

    // Use readonly scope for checking API enablement status
    client, err := getDefaultClient(ctx, serviceusage.CloudPlatformReadOnlyScope)
    if err != nil {
        return nil, fmt.Errorf("failed to create default client: %w", err)
    }

    var svc *serviceusage.Service
    err = retryWithBackoff(ctx, func() error {
        var createErr error
        svc, createErr = serviceusage.NewService(ctx, option.WithHTTPClient(client))
        return createErr
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create service usage service: %w", err)
    }

    return svc, nil
}

// CreateMonitoringService creates a Monitoring service client with minimal scopes
func (f *ClientFactory) CreateMonitoringService(ctx context.Context) (*monitoring.Service, error) {
    f.logger.Debug("Creating Monitoring service client with WIF")

    // Use readonly scope for reading metrics/alerts
    client, err := getDefaultClient(ctx, monitoring.MonitoringReadScope)
    if err != nil {
        return nil, fmt.Errorf("failed to create default client: %w", err)
    }

    var svc *monitoring.Service
    err = retryWithBackoff(ctx, func() error {
        var createErr error
        svc, createErr = monitoring.NewService(ctx, option.WithHTTPClient(client))
        return createErr
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create monitoring service: %w", err)
    }

    return svc, nil
}

// Test helpers - exported for testing purposes only

// GetDefaultClientForTesting exposes getDefaultClient for testing
func GetDefaultClientForTesting(ctx context.Context, scopes ...string) (*http.Client, error) {
    return getDefaultClient(ctx, scopes...)
}

// RetryWithBackoffForTesting exposes retryWithBackoff for testing
func RetryWithBackoffForTesting(ctx context.Context, operation func() error) error {
    return retryWithBackoff(ctx, operation)
}
