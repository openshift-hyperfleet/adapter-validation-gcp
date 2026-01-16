package validators_test

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"validator/pkg/config"
	"validator/pkg/validator"
	"validator/pkg/validators"
)

var _ = Describe("QuotaCheckValidator", func() {
	var (
		v    *validators.QuotaCheckValidator
		vctx *validator.Context
	)

	BeforeEach(func() {
		v = &validators.QuotaCheckValidator{}

		// Set up minimal config
		Expect(os.Setenv("PROJECT_ID", "test-project")).To(Succeed())
		cfg, err := config.LoadFromEnv()
		Expect(err).NotTo(HaveOccurred())

		vctx = &validator.Context{
			Config:  cfg,
			Results: make(map[string]*validator.Result),
		}
	})

	AfterEach(func() {
		Expect(os.Unsetenv("PROJECT_ID")).To(Succeed())
	})

	Describe("Metadata", func() {
		It("should return correct metadata", func() {
			meta := v.Metadata()
			Expect(meta.Name).To(Equal("quota-check"))
			Expect(meta.Description).To(ContainSubstring("quota"))
			Expect(meta.Description).To(ContainSubstring("stub"))
			Expect(meta.RunAfter).To(ConsistOf("api-enabled")) // Depends on api-enabled
			Expect(meta.Tags).To(ContainElement("post-mvp"))
			Expect(meta.Tags).To(ContainElement("quota"))
			Expect(meta.Tags).To(ContainElement("stub"))
		})

		It("should depend on api-enabled (Level 1)", func() {
			meta := v.Metadata()
			Expect(meta.RunAfter).To(ConsistOf("api-enabled"))
		})
	})

	Describe("Enabled", func() {
		Context("when validator is not explicitly disabled", func() {
			It("should be enabled by default", func() {
				enabled := v.Enabled(vctx)
				Expect(enabled).To(BeTrue())
			})
		})

		Context("when validator is explicitly disabled", func() {
			BeforeEach(func() {
				Expect(os.Setenv("DISABLED_VALIDATORS", "quota-check")).To(Succeed())
				cfg, err := config.LoadFromEnv()
				Expect(err).NotTo(HaveOccurred())
				vctx.Config = cfg
			})

			AfterEach(func() {
				Expect(os.Unsetenv("DISABLED_VALIDATORS")).To(Succeed())
			})

			It("should be disabled", func() {
				enabled := v.Enabled(vctx)
				Expect(enabled).To(BeFalse())
			})
		})

	})

	Describe("Validate", func() {
		It("should return success with stub message", func() {
			ctx := context.Background()
			result := v.Validate(ctx, vctx)
			Expect(result).NotTo(BeNil())
			Expect(result.Status).To(Equal(validator.StatusSuccess))
			Expect(result.Reason).To(Equal("QuotaCheckStub"))
			Expect(result.Message).To(ContainSubstring("not yet implemented"))
		})

		It("should include stub metadata in details", func() {
			ctx := context.Background()
			result := v.Validate(ctx, vctx)
			Expect(result.Details).To(HaveKey("stub"))
			Expect(result.Details["stub"]).To(BeTrue())
			Expect(result.Details).To(HaveKey("implemented"))
			Expect(result.Details["implemented"]).To(BeFalse())
		})
	})
})
