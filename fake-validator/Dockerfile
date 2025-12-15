FROM alpine:3.19

WORKDIR /app

# Copy the validator script
COPY validate.sh /app/validate.sh

# Make script executable
RUN chmod +x /app/validate.sh

ENTRYPOINT ["/app/validate.sh"]

LABEL name="fake-gcp-validator" \
      vendor="Red Hat" \
      version="0.0.1" \
      summary="Fake GCP Validator - Simulated GCP validation for testing" \
      description="A simulated GCP validator that mimics real validation behavior without making actual GCP API calls. Used for local development, testing, and CI/CD pipelines."
