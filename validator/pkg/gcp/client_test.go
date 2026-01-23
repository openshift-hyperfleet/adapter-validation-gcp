package gcp_test

import (
    "context"
    "errors"
    "log/slog"
    "time"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    "google.golang.org/api/googleapi"

    "validator/pkg/gcp"
)

var _ = Describe("GCP Client", func() {
    Describe("getDefaultClient", func() {
        Context("with different scopes", func() {
            It("should create new clients for each scope", func() {
                ctx := context.Background()
                scopes1 := []string{"https://www.googleapis.com/auth/cloud-platform.read-only"}
                scopes2 := []string{"https://www.googleapis.com/auth/compute.readonly"}

                // First call with scopes1
                client1, err1 := gcp.GetDefaultClientForTesting(ctx, scopes1...)
                Expect(err1).NotTo(HaveOccurred())
                Expect(client1).NotTo(BeNil())

                // Second call with scopes2 should return a different instance
                client2, err2 := gcp.GetDefaultClientForTesting(ctx, scopes2...)
                Expect(err2).NotTo(HaveOccurred())
                Expect(client2).NotTo(BeNil())
                Expect(client2).NotTo(BeIdenticalTo(client1), "Expected different client instances for different scopes")
            })

            It("should create valid clients", func() {
                ctx := context.Background()
                scopes := []string{"https://www.googleapis.com/auth/cloud-platform.read-only"}

                client, err := gcp.GetDefaultClientForTesting(ctx, scopes...)
                Expect(err).NotTo(HaveOccurred())
                Expect(client).NotTo(BeNil())
                Expect(client.Transport).NotTo(BeNil())
            })
        })
    })

    Describe("retryWithBackoff", func() {
        var ctx context.Context

        BeforeEach(func() {
            ctx = context.Background()
        })

        Context("when operation succeeds on first attempt", func() {
            It("should return success without retrying", func() {
                callCount := 0
                operation := func() error {
                    callCount++
                    return nil
                }

                err := gcp.RetryWithBackoffForTesting(ctx, operation)
                Expect(err).NotTo(HaveOccurred())
                Expect(callCount).To(Equal(1), "Should only call once on success")
            })
        })

        Context("with retryable errors", func() {
            DescribeTable("should retry based on error code",
                func(errorCode int, shouldRetry bool, expectedAttempts int) {
                    callCount := 0
                    operation := func() error {
                        callCount++
                        return &googleapi.Error{Code: errorCode}
                    }

                    err := gcp.RetryWithBackoffForTesting(ctx, operation)
                    Expect(err).To(HaveOccurred(), "Should return error")
                    Expect(callCount).To(Equal(expectedAttempts))
                },
                Entry("429 Rate Limit - should retry", 429, true, 5),
                Entry("503 Service Unavailable - should retry", 503, true, 5),
                Entry("500 Internal Error - should retry", 500, true, 5),
                Entry("404 Not Found - should not retry", 404, false, 1),
                Entry("403 Forbidden - should not retry", 403, false, 1),
            )
        })

        Context("when context is cancelled during retry", func() {
            It("should stop retrying and return context error", func() {
                ctx, cancel := context.WithCancel(context.Background())
                callCount := 0

                operation := func() error {
                    callCount++
                    if callCount == 2 {
                        cancel() // Cancel on second attempt
                    }
                    return &googleapi.Error{Code: 503} // Retryable error
                }

                err := gcp.RetryWithBackoffForTesting(ctx, operation)
                Expect(err).To(HaveOccurred())
                Expect(errors.Is(err, context.Canceled)).To(BeTrue(), "Should return context.Canceled error")
                Expect(callCount).To(Equal(2), "Should have attempted twice before cancellation")
            })
        })

        Context("when context times out", func() {
            It("should return deadline exceeded error", func() {
                ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
                defer cancel()

                operation := func() error {
                    return &googleapi.Error{Code: 503} // Keep retrying
                }

                err := gcp.RetryWithBackoffForTesting(ctx, operation)
                Expect(err).To(HaveOccurred())
                Expect(errors.Is(err, context.DeadlineExceeded)).To(BeTrue(), "Should return deadline exceeded error")
            })
        })

        Context("when max retries are exceeded", func() {
            It("should return error after 5 attempts", func() {
                callCount := 0
                operation := func() error {
                    callCount++
                    return &googleapi.Error{Code: 503} // Always fail with retryable error
                }

                err := gcp.RetryWithBackoffForTesting(ctx, operation)
                Expect(err).To(HaveOccurred())
                Expect(err.Error()).To(ContainSubstring("max retries exceeded"))
                Expect(callCount).To(Equal(5), "Should attempt 5 times (initial + 4 retries)")
            })
        })

        Context("with non-googleapi errors", func() {
            It("should retry generic errors until max retries", func() {
                callCount := 0
                operation := func() error {
                    callCount++
                    return errors.New("generic error")
                }

                err := gcp.RetryWithBackoffForTesting(ctx, operation)
                Expect(err).To(HaveOccurred())
                Expect(callCount).To(Equal(5), "Should retry generic errors until max retries")
            })
        })
    })

    Describe("ClientFactory", func() {
        var (
            projectID string
            logger    *slog.Logger
        )

        BeforeEach(func() {
            projectID = "test-project"
            logger = slog.Default()
        })

        Describe("NewClientFactory", func() {
            It("should create a new factory with correct values", func() {
                factory := gcp.NewClientFactory(projectID, logger)
                Expect(factory).NotTo(BeNil())

                // Note: We can't directly test private fields, but we can test behavior
                // by using the factory to create services (which would fail if projectID is wrong)
            })

            It("should accept different project IDs", func() {
                factory := gcp.NewClientFactory("my-test-project", logger)
                Expect(factory).NotTo(BeNil())
            })
        })

        // Note: Testing actual GCP service creation requires either:
        // 1. Mocking google.DefaultClient (complex, requires dependency injection)
        // 2. Integration tests with real GCP credentials
        // 3. Using interfaces and dependency injection (architectural change)
        //
        // For now, we test the factory creation and leave service creation for integration tests.
        // The CreateXXXService methods follow the same pattern, so testing one validates the pattern.
    })
})
