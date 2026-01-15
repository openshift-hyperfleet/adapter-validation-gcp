package config_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"validator/pkg/config"
)

var _ = Describe("Config", func() {
	var originalEnv map[string]string

	BeforeEach(func() {
		// Save original environment
		originalEnv = make(map[string]string)
		envVars := []string{
			"RESULTS_PATH", "PROJECT_ID", "GCP_REGION",
			"DISABLED_VALIDATORS", "STOP_ON_FIRST_FAILURE",
			"REQUIRED_APIS", "LOG_LEVEL",
			"REQUIRED_VCPUS", "REQUIRED_DISK_GB", "REQUIRED_IP_ADDRESSES",
			"VPC_NAME", "SUBNET_NAME",
		}
		for _, v := range envVars {
			originalEnv[v] = os.Getenv(v)
			Expect(os.Unsetenv(v)).To(Succeed())
		}
	})

	AfterEach(func() {
		// Restore original environment
		for k, v := range originalEnv {
			if v != "" {
				Expect(os.Setenv(k, v)).To(Succeed())
			} else {
				Expect(os.Unsetenv(k)).To(Succeed())
			}
		}
	})

	Describe("LoadFromEnv", func() {
		Context("with minimal required configuration", func() {
			BeforeEach(func() {
				Expect(os.Setenv("PROJECT_ID", "test-project-123")).To(Succeed())
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
				Expect(os.Setenv("PROJECT_ID", "custom-project")).To(Succeed())
				Expect(os.Setenv("RESULTS_PATH", "/custom/path/results.json")).To(Succeed())
				Expect(os.Setenv("GCP_REGION", "us-central1")).To(Succeed())
				Expect(os.Setenv("LOG_LEVEL", "debug")).To(Succeed())
				Expect(os.Setenv("STOP_ON_FIRST_FAILURE", "true")).To(Succeed())
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
				Expect(os.Setenv("PROJECT_ID", "test-project")).To(Succeed())
				Expect(os.Setenv("DISABLED_VALIDATORS", "quota-check,network-check")).To(Succeed())
			})

			It("should parse the disabled validators list", func() {
				cfg, err := config.LoadFromEnv()
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg.DisabledValidators).To(ConsistOf("quota-check", "network-check"))
			})
		})

		Context("with disabled validators containing whitespace", func() {
			BeforeEach(func() {
				Expect(os.Setenv("PROJECT_ID", "test-project")).To(Succeed())
				Expect(os.Setenv("DISABLED_VALIDATORS", " quota-check , network-check ")).To(Succeed())
			})

			It("should trim whitespace from validator names", func() {
				cfg, err := config.LoadFromEnv()
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg.DisabledValidators).To(ConsistOf("quota-check", "network-check"))
			})
		})

		Context("with custom required APIs", func() {
			BeforeEach(func() {
				Expect(os.Setenv("PROJECT_ID", "test-project")).To(Succeed())
				Expect(os.Setenv("REQUIRED_APIS", "compute.googleapis.com,storage.googleapis.com")).To(Succeed())
			})

			It("should parse the required APIs list", func() {
				cfg, err := config.LoadFromEnv()
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg.RequiredAPIs).To(ConsistOf("compute.googleapis.com", "storage.googleapis.com"))
			})
		})

		Context("with integer configurations", func() {
			BeforeEach(func() {
				Expect(os.Setenv("PROJECT_ID", "test-project")).To(Succeed())
				Expect(os.Setenv("REQUIRED_VCPUS", "100")).To(Succeed())
				Expect(os.Setenv("REQUIRED_DISK_GB", "500")).To(Succeed())
				Expect(os.Setenv("REQUIRED_IP_ADDRESSES", "10")).To(Succeed())
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
				Expect(os.Setenv("PROJECT_ID", "test-project")).To(Succeed())
				Expect(os.Setenv("REQUIRED_VCPUS", "not-a-number")).To(Succeed())
			})

			It("should use default value for invalid integers", func() {
				cfg, err := config.LoadFromEnv()
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg.RequiredVCPUs).To(Equal(0))
			})
		})

		Context("with invalid boolean values", func() {
			BeforeEach(func() {
				Expect(os.Setenv("PROJECT_ID", "test-project")).To(Succeed())
				Expect(os.Setenv("STOP_ON_FIRST_FAILURE", "not-a-bool")).To(Succeed())
			})

			It("should use default value for invalid booleans", func() {
				cfg, err := config.LoadFromEnv()
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg.StopOnFirstFailure).To(BeFalse())
			})
		})

		Context("with network validator config", func() {
			BeforeEach(func() {
				Expect(os.Setenv("PROJECT_ID", "test-project")).To(Succeed())
				Expect(os.Setenv("VPC_NAME", "my-vpc")).To(Succeed())
				Expect(os.Setenv("SUBNET_NAME", "my-subnet")).To(Succeed())
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
			Expect(os.Setenv("PROJECT_ID", "test-project")).To(Succeed())
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
				Expect(os.Setenv("DISABLED_VALIDATORS", "quota-check")).To(Succeed())
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
				Expect(os.Setenv("DISABLED_VALIDATORS", "quota-check,network-check")).To(Succeed())
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
