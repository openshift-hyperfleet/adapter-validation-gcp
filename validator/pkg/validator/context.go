package validator

import (
	compute "google.golang.org/api/compute/v1"
	iam "google.golang.org/api/iam/v1"
	cloudresourcemanager "google.golang.org/api/cloudresourcemanager/v1"
	serviceusage "google.golang.org/api/serviceusage/v1"
	monitoring "google.golang.org/api/monitoring/v3"

	"validator/pkg/config"
)

// Context provides shared resources and configuration to all validators
type Context struct {
	// Configuration
	Config *config.Config

	// GCP Clients (lazily initialized, shared across validators)
	ComputeService          *compute.Service
	IAMService              *iam.Service
	CloudResourceManagerSvc *cloudresourcemanager.Service
	ServiceUsageService     *serviceusage.Service
	MonitoringService       *monitoring.Service

	// Shared state between validators
	ProjectNumber int64

	// Results from previous validators (for dependency checking)
	Results map[string]*Result
}
