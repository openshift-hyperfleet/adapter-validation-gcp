package validator_test

import (
    "context"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"

    "validator/pkg/validator"
)

// Mock validator for testing
type MockValidator struct {
    name         string
    description  string
    runAfter     []string
    tags         []string
    enabled      bool
    validateFunc func(ctx context.Context, vctx *validator.Context) *validator.Result
}

func (m *MockValidator) Metadata() validator.ValidatorMetadata {
    return validator.ValidatorMetadata{
        Name:        m.name,
        Description: m.description,
        RunAfter:    m.runAfter,
        Tags:        m.tags,
    }
}

func (m *MockValidator) Enabled(ctx *validator.Context) bool {
    return m.enabled
}

func (m *MockValidator) Validate(ctx context.Context, vctx *validator.Context) *validator.Result {
    if m.validateFunc != nil {
        return m.validateFunc(ctx, vctx)
    }
    return &validator.Result{
        ValidatorName: m.name,
        Status:        validator.StatusSuccess,
        Reason:        "TestSuccess",
        Message:       "Test validation passed",
    }
}

var _ = Describe("Registry", func() {
    var (
        testRegistry *validator.Registry
        mockValidator1 *MockValidator
        mockValidator2 *MockValidator
    )

    BeforeEach(func() {
        testRegistry = validator.NewRegistry()
        mockValidator1 = &MockValidator{
            name:        "test-validator-1",
            description: "First test validator",
            runAfter:    []string{},
            tags:        []string{"test", "mock"},
            enabled:     true,
        }
        mockValidator2 = &MockValidator{
            name:        "test-validator-2",
            description: "Second test validator",
            runAfter:    []string{"test-validator-1"},
            tags:        []string{"test", "dependent"},
            enabled:     true,
        }
    })

    Describe("Register", func() {
        Context("when registering a new validator", func() {
            It("should add the validator to the registry", func() {
                testRegistry.Register(mockValidator1)
                validators := testRegistry.GetAll()
                Expect(validators).To(HaveLen(1))
                Expect(validators[0].Metadata().Name).To(Equal("test-validator-1"))
            })
        })

        Context("when registering multiple validators", func() {
            It("should add all validators to the registry", func() {
                testRegistry.Register(mockValidator1)
                testRegistry.Register(mockValidator2)
                validators := testRegistry.GetAll()
                Expect(validators).To(HaveLen(2))
            })
        })

        Context("when registering a validator with duplicate name", func() {
            It("should overwrite the existing validator", func() {
                testRegistry.Register(mockValidator1)
                duplicate := &MockValidator{
                    name:        "test-validator-1",
                    description: "Duplicate validator",
                    enabled:     true,
                }
                testRegistry.Register(duplicate)
                validators := testRegistry.GetAll()
                Expect(validators).To(HaveLen(1))
                Expect(validators[0].Metadata().Description).To(Equal("Duplicate validator"))
            })
        })
    })

    Describe("GetAll", func() {
        Context("when registry is empty", func() {
            It("should return an empty slice", func() {
                validators := testRegistry.GetAll()
                Expect(validators).To(BeEmpty())
            })
        })

        Context("when registry has validators", func() {
            It("should return all registered validators", func() {
                testRegistry.Register(mockValidator1)
                testRegistry.Register(mockValidator2)
                validators := testRegistry.GetAll()
                Expect(validators).To(HaveLen(2))
            })
        })
    })

    Describe("Get", func() {
        BeforeEach(func() {
            testRegistry.Register(mockValidator1)
            testRegistry.Register(mockValidator2)
        })

        Context("when getting a validator by name", func() {
            It("should return the validator if it exists", func() {
                v, exists := testRegistry.Get("test-validator-1")
                Expect(exists).To(BeTrue())
                Expect(v.Metadata().Name).To(Equal("test-validator-1"))
            })

            It("should return false if validator doesn't exist", func() {
                _, exists := testRegistry.Get("non-existent")
                Expect(exists).To(BeFalse())
            })
        })
    })
})
