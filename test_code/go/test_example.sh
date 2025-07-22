#!/bin/bash

# Test script for BigQuery Service Account Impersonation Example
# This script sets up the environment and runs the example
#
# Before running this script, you can:
# 1. Copy env.template to .env and configure your settings
# 2. Or set the environment variables directly

set -e  # Exit on any error

echo "=== BigQuery Service Account Impersonation Example Test ==="
echo

# Function to mask sensitive values for logging
mask_value() {
    local value="$1"
    local len=${#value}
    if [ $len -le 8 ]; then
        if [ $len -le 2 ]; then
            echo "***"
        else
            echo "${value:0:1}***${value: -1}"
        fi
    else
        echo "${value:0:4}***${value: -4}"
    fi
}

# Load .env file if it exists
if [ -f ".env" ]; then
    echo "Loading environment variables from .env file..."
    export $(cat .env | grep -v '^#' | xargs)
fi

# Check if required environment variables are set
if [ -z "$GOOGLE_CLOUD_PROJECT" ]; then
    echo "Setting GOOGLE_CLOUD_PROJECT to default value..."
    export GOOGLE_CLOUD_PROJECT="xxxx"
fi

if [ -z "$ADBC_BIGQUERY_DATASET_ID" ]; then
    echo "Setting ADBC_BIGQUERY_DATASET_ID to default value..."
    export ADBC_BIGQUERY_DATASET_ID="xxx"
fi

if [ -z "$ADBC_BIGQUERY_IMPERSONATE_TARGET" ]; then
    echo "Setting ADBC_BIGQUERY_IMPERSONATE_TARGET to default value..."
    export ADBC_BIGQUERY_IMPERSONATE_TARGET="xxxx"
fi

echo "Environment variables:"
echo "  GOOGLE_CLOUD_PROJECT: $(mask_value "$GOOGLE_CLOUD_PROJECT")"
echo "  ADBC_BIGQUERY_DATASET_ID: $(mask_value "$ADBC_BIGQUERY_DATASET_ID")"
echo "  ADBC_BIGQUERY_IMPERSONATE_TARGET: $(mask_value "$ADBC_BIGQUERY_IMPERSONATE_TARGET")"
echo

echo "Building the example..."
go build -o bigquery-example example.go

echo "Running the example..."
echo "Note: This will fail if the service account impersonation is not properly configured."
echo "Make sure the base service account has the 'Service Account Token Creator' role on the target service account."
echo

# Run the example
./bigquery-example

echo
echo "=== Test completed ==="
