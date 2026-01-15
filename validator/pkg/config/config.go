package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration from environment variables
type Config struct {
	// Output
	ResultsPath string // Default: /results/adapter-result.json

	// GCP Configuration
	ProjectID string // Required
	GCPRegion string // Optional, for regional checks

	// Validator Control
	DisabledValidators []string // Comma-separated list of validators to disable
	StopOnFirstFailure bool     // Default: false

	// API Validator Config
	RequiredAPIs []string // Default: compute.googleapis.com, iam.googleapis.com, etc.

	// Quota Validator Config (Post-MVP)
	RequiredVCPUs      int // Default: 0 (skip quota check)
	RequiredDiskGB     int
	RequiredIPAddresses int

	// Network Validator Config (Post-MVP)
	VPCName    string
	SubnetName string

	// Logging
	LogLevel string // debug, info, warn, error
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	cfg := &Config{
		ResultsPath:         getEnv("RESULTS_PATH", "/results/adapter-result.json"),
		ProjectID:           os.Getenv("PROJECT_ID"),
		GCPRegion:           getEnv("GCP_REGION", ""),
		StopOnFirstFailure:  getEnvBool("STOP_ON_FIRST_FAILURE", false),
		LogLevel:            getEnv("LOG_LEVEL", "info"),
		RequiredVCPUs:       getEnvInt("REQUIRED_VCPUS", 0),
		RequiredDiskGB:      getEnvInt("REQUIRED_DISK_GB", 0),
		RequiredIPAddresses: getEnvInt("REQUIRED_IP_ADDRESSES", 0),
		VPCName:             getEnv("VPC_NAME", ""),
		SubnetName:          getEnv("SUBNET_NAME", ""),
	}

	// Parse disabled validators
	if disabled := os.Getenv("DISABLED_VALIDATORS"); disabled != "" {
		cfg.DisabledValidators = strings.Split(disabled, ",")
		// Trim whitespace
		for i, v := range cfg.DisabledValidators {
			cfg.DisabledValidators[i] = strings.TrimSpace(v)
		}
	}

	// Parse required APIs
	defaultAPIs := []string{
		"compute.googleapis.com",
		"iam.googleapis.com",
		"cloudresourcemanager.googleapis.com",
	}
	if apis := os.Getenv("REQUIRED_APIS"); apis != "" {
		cfg.RequiredAPIs = strings.Split(apis, ",")
		// Trim whitespace
		for i, v := range cfg.RequiredAPIs {
			cfg.RequiredAPIs[i] = strings.TrimSpace(v)
		}
	} else {
		cfg.RequiredAPIs = defaultAPIs
	}

	// Validation
	if cfg.ProjectID == "" {
		return nil, fmt.Errorf("PROJECT_ID is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		b, err := strconv.ParseBool(value)
		if err == nil {
			return b
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		i, err := strconv.Atoi(value)
		if err == nil {
			return i
		}
	}
	return defaultValue
}

// IsValidatorEnabled checks if a validator should run
// All validators are enabled by default unless explicitly disabled
func (c *Config) IsValidatorEnabled(name string) bool {
	// Check if explicitly disabled
	for _, disabled := range c.DisabledValidators {
		if disabled == name {
			return false
		}
	}
	// Not disabled = enabled
	return true
}
