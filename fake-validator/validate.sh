#!/bin/sh
set -e

# Default configuration
RESULTS_PATH="${RESULTS_PATH:-/results/adapter-result.json}"
SIMULATE_RESULT="${SIMULATE_RESULT:-success}"

echo "Fake GCP Validator starting..."
echo "Simulating result: ${SIMULATE_RESULT}"
echo "Results path: ${RESULTS_PATH}"

# Ensure results directory exists
RESULTS_DIR=$(dirname "${RESULTS_PATH}")
mkdir -p "${RESULTS_DIR}"

case "${SIMULATE_RESULT}" in
  success)
    echo "Writing success result..."
    cat > "${RESULTS_PATH}" <<EOF
{
  "status": "success",
  "reason": "ValidationPassed",
  "message": "GCP environment validated successfully (simulated)",
  "details": {
    "simulation": true,
    "checks_run": 5,
    "checks_passed": 5,
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  }
}
EOF
    echo "Success result written to ${RESULTS_PATH}"
    exit 0
    ;;

  failure)
    echo "Writing failure result..."
    cat > "${RESULTS_PATH}" <<EOF
{
  "status": "failure",
  "reason": "MissingPermissions",
  "message": "Service account lacks required IAM permissions (simulated)",
  "details": {
    "simulation": true,
    "missing_permissions": [
      "compute.instances.list",
      "iam.serviceAccounts.get"
    ],
    "service_account": "fake-sa@project.iam.gserviceaccount.com",
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  }
}
EOF
    echo "Failure result written to ${RESULTS_PATH}"
    exit 1
    ;;

  hang)
    echo "Simulating hang (sleeping indefinitely)..."
    sleep 9999999
    ;;

  crash)
    echo "Simulating crash (exiting without writing results)..."
    exit 1
    ;;

  invalid-json)
    echo "Writing invalid JSON result..."
    echo "{ this is not valid json }" > "${RESULTS_PATH}"
    exit 0
    ;;

  missing-status)
    echo "Writing result with missing status field..."
    cat > "${RESULTS_PATH}" <<EOF
{
  "reason": "TestReason",
  "message": "This result is missing the status field"
}
EOF
    exit 0
    ;;

  *)
    echo "Unknown SIMULATE_RESULT value: ${SIMULATE_RESULT}"
    echo "Valid values: success, failure, hang, crash, invalid-json, missing-status"
    echo "Defaulting to success..."
    cat > "${RESULTS_PATH}" <<EOF
{
  "status": "success",
  "reason": "ValidationPassed",
  "message": "GCP environment validated successfully (default simulation)",
  "details": {
    "simulation": true,
    "note": "Unknown SIMULATE_RESULT value, defaulted to success"
  }
}
EOF
    exit 0
    ;;
esac
