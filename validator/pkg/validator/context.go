package validator

import (
    "context"
    "fmt"
    "log/slog"
    "sync"

    "google.golang.org/api/cloudresourcemanager/v1"
    "google.golang.org/api/compute/v1"
    "google.golang.org/api/iam/v1"
    "google.golang.org/api/monitoring/v3"
    "google.golang.org/api/serviceusage/v1"

    "validator/pkg/config"
    "validator/pkg/gcp"
)

// Context provides shared resources and configuration to all validators
// Implements least-privilege principle through lazy initialization:
// - Services are only created when first requested by validators
// - OAuth scopes are only requested for services that are actually used
// - Disabled validators never trigger authentication for their services
// Thread-safe: Uses sync.Once to ensure services are initialized exactly once
type Context struct {
    // Configuration
    Config *config.Config

    // Client factory for creating GCP service clients
    clientFactory *gcp.ClientFactory

    // GCP Clients (lazily initialized, shared across validators)
    // These are private to enforce use of getter methods
    computeService          *compute.Service
    iamService              *iam.Service
    cloudResourceManagerSvc *cloudresourcemanager.Service
    serviceUsageService     *serviceusage.Service
    monitoringService       *monitoring.Service

    // Thread-safe lazy initialization guards
    // Each sync.Once ensures its corresponding service is created exactly once,
    // even when called concurrently from multiple validators
    computeOnce          sync.Once
    iamOnce              sync.Once
    cloudResourceMgrOnce sync.Once
    serviceUsageOnce     sync.Once
    monitoringOnce       sync.Once

    // Shared state between validators
    ProjectNumber int64

    // Results from previous validators (for dependency checking)
    Results map[string]*Result
}

// NewContext creates a new validation context with a client factory
func NewContext(cfg *config.Config, logger *slog.Logger) *Context {
    return &Context{
        Config:        cfg,
        clientFactory: gcp.NewClientFactory(cfg.ProjectID, logger),
        Results:       make(map[string]*Result),
    }
}

// GetComputeService returns the Compute Engine service, creating it lazily on first use
// Only requests compute.readonly scope when a validator actually needs it
// Thread-safe: Uses sync.Once to ensure the service is created exactly once
func (c *Context) GetComputeService(ctx context.Context) (*compute.Service, error) {
    var err error
    c.computeOnce.Do(func() {
        c.computeService, err = c.clientFactory.CreateComputeService(ctx)
        if err != nil {
            err = fmt.Errorf("failed to create compute service: %w", err)
        }
    })
    if err != nil {
        return nil, err
    }
    return c.computeService, nil
}

// GetIAMService returns the IAM service, creating it lazily on first use
// Only requests cloud-platform.read-only scope when a validator actually needs it
// Thread-safe: Uses sync.Once to ensure the service is created exactly once
func (c *Context) GetIAMService(ctx context.Context) (*iam.Service, error) {
    var err error
    c.iamOnce.Do(func() {
        c.iamService, err = c.clientFactory.CreateIAMService(ctx)
        if err != nil {
            err = fmt.Errorf("failed to create IAM service: %w", err)
        }
    })
    if err != nil {
        return nil, err
    }
    return c.iamService, nil
}

// GetCloudResourceManagerService returns the Cloud Resource Manager service, creating it lazily on first use
// Only requests cloudresourcemanager.readonly scope when a validator actually needs it
// Thread-safe: Uses sync.Once to ensure the service is created exactly once
func (c *Context) GetCloudResourceManagerService(ctx context.Context) (*cloudresourcemanager.Service, error) {
    var err error
    c.cloudResourceMgrOnce.Do(func() {
        c.cloudResourceManagerSvc, err = c.clientFactory.CreateCloudResourceManagerService(ctx)
        if err != nil {
            err = fmt.Errorf("failed to create cloud resource manager service: %w", err)
        }
    })
    if err != nil {
        return nil, err
    }
    return c.cloudResourceManagerSvc, nil
}

// GetServiceUsageService returns the Service Usage service, creating it lazily on first use
// Only requests serviceusage.readonly scope when a validator actually needs it
// Thread-safe: Uses sync.Once to ensure the service is created exactly once
func (c *Context) GetServiceUsageService(ctx context.Context) (*serviceusage.Service, error) {
    var err error
    c.serviceUsageOnce.Do(func() {
        c.serviceUsageService, err = c.clientFactory.CreateServiceUsageService(ctx)
        if err != nil {
            err = fmt.Errorf("failed to create service usage service: %w", err)
        }
    })
    if err != nil {
        return nil, err
    }
    return c.serviceUsageService, nil
}

// GetMonitoringService returns the Monitoring service, creating it lazily on first use
// Only requests monitoring.read scope when a validator actually needs it
// Thread-safe: Uses sync.Once to ensure the service is created exactly once
func (c *Context) GetMonitoringService(ctx context.Context) (*monitoring.Service, error) {
    var err error
    c.monitoringOnce.Do(func() {
        c.monitoringService, err = c.clientFactory.CreateMonitoringService(ctx)
        if err != nil {
            err = fmt.Errorf("failed to create monitoring service: %w", err)
        }
    })
    if err != nil {
        return nil, err
    }
    return c.monitoringService, nil
}
