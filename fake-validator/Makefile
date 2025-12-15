# Makefile for fake-gcp-validator

# Project metadata
IMAGE_NAME := fake-gcp-validator
VERSION ?= 0.0.1
IMAGE_REGISTRY ?= quay.io/openshift-hyperfleet
IMAGE_TAG ?= latest

# Build metadata
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# Dev image configuration
QUAY_USER ?=
DEV_TAG ?= dev-$(GIT_COMMIT)

# Test configuration
TEST_RESULTS_PATH ?= /tmp/test-result.json

# Container runtime detection
DOCKER_AVAILABLE := $(shell if docker info >/dev/null 2>&1; then echo "true"; else echo "false"; fi)
PODMAN_AVAILABLE := $(shell if podman info >/dev/null 2>&1; then echo "true"; else echo "false"; fi)

ifeq ($(DOCKER_AVAILABLE),true)
    CONTAINER_RUNTIME := docker
    CONTAINER_CMD := docker
else ifeq ($(PODMAN_AVAILABLE),true)
    CONTAINER_RUNTIME := podman
    CONTAINER_CMD := podman
else
    CONTAINER_RUNTIME := none
    CONTAINER_CMD := sh -c 'echo "No container runtime found. Please install Docker or Podman." && exit 1'
endif

.PHONY: help
help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: test
test: ## Test the validator script locally
	@echo "Testing validator script..."
	@if [ ! -f ./validate.sh ]; then \
		echo "❌ ERROR: ./validate.sh not found"; \
		exit 1; \
	fi
	@if [ ! -x ./validate.sh ]; then \
		echo "❌ ERROR: ./validate.sh is not executable"; \
		echo "Run: chmod +x ./validate.sh"; \
		exit 1; \
	fi
	@echo ""
	@echo "=== Testing success scenario ==="
	@SIMULATE_RESULT=success RESULTS_PATH=$(TEST_RESULTS_PATH) ./validate.sh
	@cat $(TEST_RESULTS_PATH)
	@echo ""
	@echo "=== Testing failure scenario ==="
	-@SIMULATE_RESULT=failure RESULTS_PATH=$(TEST_RESULTS_PATH) ./validate.sh
	@cat $(TEST_RESULTS_PATH)
	@echo ""
	@echo "=== Testing invalid-json scenario ==="
	@SIMULATE_RESULT=invalid-json RESULTS_PATH=$(TEST_RESULTS_PATH) ./validate.sh
	@cat $(TEST_RESULTS_PATH)
	@echo ""
	@echo "✅ All scenarios tested"
	@rm -f $(TEST_RESULTS_PATH)

.PHONY: image
image: ## Build container image with Docker or Podman
ifeq ($(CONTAINER_RUNTIME),none)
	@echo "❌ ERROR: No container runtime found"
	@echo "Please install Docker or Podman"
	@exit 1
else
	@echo "Building container image with $(CONTAINER_RUNTIME)..."
	$(CONTAINER_CMD) build -t $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) .
	@echo "✅ Image built: $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)"
endif

.PHONY: image-push
image-push: image ## Build and push container image to registry
ifeq ($(CONTAINER_RUNTIME),none)
	@echo "❌ ERROR: No container runtime found"
	@echo "Please install Docker or Podman"
	@exit 1
else
	@echo "Pushing image $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)..."
	$(CONTAINER_CMD) push $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
	@echo "✅ Image pushed: $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)"
endif

.PHONY: image-dev
image-dev: ## Build and push to personal Quay registry (requires QUAY_USER)
ifndef QUAY_USER
	@echo "❌ ERROR: QUAY_USER is not set"
	@echo ""
	@echo "Usage: QUAY_USER=myuser make image-dev"
	@echo ""
	@echo "This will build and push to: quay.io/$$QUAY_USER/$(IMAGE_NAME):$(DEV_TAG)"
	@exit 1
endif
ifeq ($(CONTAINER_RUNTIME),none)
	@echo "❌ ERROR: No container runtime found"
	@echo "Please install Docker or Podman"
	@exit 1
else
	@echo "Building dev image quay.io/$(QUAY_USER)/$(IMAGE_NAME):$(DEV_TAG)..."
	$(CONTAINER_CMD) build -t quay.io/$(QUAY_USER)/$(IMAGE_NAME):$(DEV_TAG) .
	@echo "Pushing dev image..."
	$(CONTAINER_CMD) push quay.io/$(QUAY_USER)/$(IMAGE_NAME):$(DEV_TAG)
	@echo ""
	@echo "✅ Dev image pushed: quay.io/$(QUAY_USER)/$(IMAGE_NAME):$(DEV_TAG)"
endif

.PHONY: clean
clean: ## Clean test artifacts
	@echo "Cleaning..."
	@rm -f $(TEST_RESULTS_PATH)
