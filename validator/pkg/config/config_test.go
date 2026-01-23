package config_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"validator/pkg/config"
)

var _ = Describe("Config", func() {
	BeforeEach(func() {
		// Clear environment variables - GinkgoT().Setenv automatically restores them
		envVars := []string{
			"RESULTS_PATH", "PROJECT_ID", "GCP_REGION",
			"DISABLED_VALIDATORS", "STOP_ON_FIRST_FAILURE",
			"REQUIRED_APIS", "LOG_LEVEL",
			"REQUIRED_VCPUS", "REQUIRED_DISK_GB", "REQUIRED_IP_ADDRESSES",
			"VPC_NAME", "SUBNET_NAME", "MAX_WAIT_TIME_SECONDS",
		}
		for _, key := range envVars {
			GinkgoT().Setenv(key, "")
		}
	})

	Describe("LoadFromEnv", func() {
		Context("with minimal required configuration", func() {
			BeforeEach(func() {
				GinkgoT().Setenv("PROJECT_ID", "test-project-123")
			})

			It("should load config with defaults", func() {
				cfg, err := config.LoadFromEnv()
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg.ProjectID).To(Equal("test-project-123"))
				Expect(cfg.ResultsPath).To(Equal("/results/adapter-result.json"))
				Expect(cfg.LogLevel).To(Equal("info"))
				Expect(cfg.StopOnFirstFailure).To(BeFalse())
			})

			It("should set default required APIs", func() {
				cfg, err := config.LoadFromEnv()
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg.RequiredAPIs).To(ConsistOf(
					"compute.googleapis.com",
					"iam.googleapis.com",
					"cloudresourcemanager.googleapis.com",
				))
			})
		})

		Context("without required PROJECT_ID", func() {
			It("should return an error", func() {
				_, err := config.LoadFromEnv()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("PROJECT_ID is required"))
			})
		})

		Context("with custom configuration", func() {
			BeforeEach(func() {
				GinkgoT().Setenv("PROJECT_ID", "custom-project")
				GinkgoT().Setenv("RESULTS_PATH", "/custom/path/results.json")
				GinkgoT().Setenv("GCP_REGION", "us-central1")
				GinkgoT().Setenv("LOG_LEVEL", "debug")
				GinkgoT().Setenv("STOP_ON_FIRST_FAILURE", "true")
			})

			It("should load all custom values", func() {
				cfg, err := config.LoadFromEnv()
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg.ProjectID).To(Equal("custom-project"))
				Expect(cfg.ResultsPath).To(Equal("/custom/path/results.json"))
				Expect(cfg.GCPRegion).To(Equal("us-central1"))
				Expect(cfg.LogLevel).To(Equal("debug"))
				Expect(cfg.StopOnFirstFailure).To(BeTrue())
			})
		})

		Context("with disabled validators", func() {
			BeforeEach(func() {
				GinkgoT().Setenv("PROJECT_ID", "test-project")
				GinkgoT().Setenv("DISABLED_VALIDATORS", "quota-check,network-check")
			})

			It("should parse the disabled validators list", func() {
				cfg, err := config.LoadFromEnv()
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg.DisabledValidators).To(ConsistOf("quota-check", "network-check"))
			})
		})

		Context("with disabled validators containing whitespace", func() {
			BeforeEach(func() {
				GinkgoT().Setenv("PROJECT_ID", "test-project")
				GinkgoT().Setenv("DISABLED_VALIDATORS", " quota-check , network-check ")
			})

			It("should trim whitespace from validator names", func() {
				cfg, err := config.LoadFromEnv()
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg.DisabledValidators).To(ConsistOf("quota-check", "network-check"))
			})
		})

		Context("with custom required APIs", func() {
			BeforeEach(func() {
				GinkgoT().Setenv("PROJECT_ID", "test-project")
				GinkgoT().Setenv("REQUIRED_APIS", "compute.googleapis.com,storage.googleapis.com")
			})

			It("should parse the required APIs list", func() {
				cfg, err := config.LoadFromEnv()
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg.RequiredAPIs).To(ConsistOf("compute.googleapis.com", "storage.googleapis.com"))
			})
		})

		Context("with integer configurations", func() {
			BeforeEach(func() {
				GinkgoT().Setenv("PROJECT_ID", "test-project")
				GinkgoT().Setenv("REQUIRED_VCPUS", "100")
				GinkgoT().Setenv("REQUIRED_DISK_GB", "500")
				GinkgoT().Setenv("REQUIRED_IP_ADDRESSES", "10")
			})

			It("should parse integer values", func() {
				cfg, err := config.LoadFromEnv()
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg.RequiredVCPUs).To(Equal(100))
				Expect(cfg.RequiredDiskGB).To(Equal(500))
				Expect(cfg.RequiredIPAddresses).To(Equal(10))
			})
		})

		Context("with invalid integer values", func() {
			BeforeEach(func() {
				GinkgoT().Setenv("PROJECT_ID", "test-project")
				GinkgoT().Setenv("REQUIRED_VCPUS", "not-a-number")
			})

			It("should use default value for invalid integers", func() {
				cfg, err := config.LoadFromEnv()
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg.RequiredVCPUs).To(Equal(0))
			})
		})

		Context("with invalid boolean values", func() {
			BeforeEach(func() {
				GinkgoT().Setenv("PROJECT_ID", "test-project")
				GinkgoT().Setenv("STOP_ON_FIRST_FAILURE", "not-a-bool")
			})

			It("should use default value for invalid booleans", func() {
				cfg, err := config.LoadFromEnv()
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg.StopOnFirstFailure).To(BeFalse())
			})
		})

		Context("with network validator config", func() {
			BeforeEach(func() {
				GinkgoT().Setenv("PROJECT_ID", "test-project")
				GinkgoT().Setenv("VPC_NAME", "my-vpc")
				GinkgoT().Setenv("SUBNET_NAME", "my-subnet")
			})

			It("should load network configuration", func() {
				cfg, err := config.LoadFromEnv()
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg.VPCName).To(Equal("my-vpc"))
				Expect(cfg.SubnetName).To(Equal("my-subnet"))
			})
		})
	})

	Describe("IsValidatorEnabled", func() {
		var cfg *config.Config

		BeforeEach(func() {
			GinkgoT().Setenv("PROJECT_ID", "test-project")
		})

		Context("with no disabled list", func() {
			BeforeEach(func() {
				var err error
				cfg, err = config.LoadFromEnv()
				Expect(err).NotTo(HaveOccurred())
			})

			It("should enable all validators by default", func() {
				Expect(cfg.IsValidatorEnabled("api-enabled")).To(BeTrue())
				Expect(cfg.IsValidatorEnabled("quota-check")).To(BeTrue())
				Expect(cfg.IsValidatorEnabled("any-validator")).To(BeTrue())
			})
		})

		Context("with disabled validators list", func() {
			BeforeEach(func() {
				GinkgoT().Setenv("DISABLED_VALIDATORS", "quota-check")
				var err error
				cfg, err = config.LoadFromEnv()
				Expect(err).NotTo(HaveOccurred())
			})

			It("should disable validators in the list", func() {
				Expect(cfg.IsValidatorEnabled("quota-check")).To(BeFalse())
				Expect(cfg.IsValidatorEnabled("api-enabled")).To(BeTrue())
				Expect(cfg.IsValidatorEnabled("network-check")).To(BeTrue())
			})
		})

		Context("with multiple disabled validators", func() {
			BeforeEach(func() {
				GinkgoT().Setenv("DISABLED_VALIDATORS", "quota-check,network-check")
				var err error
				cfg, err = config.LoadFromEnv()
				Expect(err).NotTo(HaveOccurred())
			})

			It("should disable all validators in the list", func() {
				Expect(cfg.IsValidatorEnabled("quota-check")).To(BeFalse())
				Expect(cfg.IsValidatorEnabled("network-check")).To(BeFalse())
				Expect(cfg.IsValidatorEnabled("api-enabled")).To(BeTrue())
			})
		})
	})
})
