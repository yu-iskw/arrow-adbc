# BigQuery Service Account Impersonation Example

This directory contains a Go example demonstrating how to use BigQuery service account impersonation with the ADBC driver.

## Prerequisites

1. **Go 1.23.0 or later**
2. **Google Cloud Project** with BigQuery enabled
3. **Service Account** with appropriate permissions
4. **Base Service Account** with Service Account Token Creator role on the target service account

## Setup

### 1. Environment Configuration

Copy the environment template and configure your settings:

```bash
cp env.template .env
```

Edit `.env` with your actual values:

```bash
# Required: Your Google Cloud Project ID
GOOGLE_CLOUD_PROJECT=your-project-id

# Required: Your BigQuery Dataset ID
ADBC_BIGQUERY_DATASET_ID=your-dataset-id

# Required: Service account to impersonate
ADBC_BIGQUERY_IMPERSONATE_TARGET=your-service-account@your-project-id.iam.gserviceaccount.com

# Optional: Path to service account key file (if not using Application Default Credentials)
# GOOGLE_APPLICATION_CREDENTIALS=/path/to/your/service-account-key.json
```

### 2. Authentication Setup

#### Option A: Application Default Credentials (Recommended)

```bash
# Set up Application Default Credentials
gcloud auth application-default login
```

#### Option B: Service Account Key File

```bash
# Download your service account key and set the path
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/your/service-account-key.json
```

### 3. Service Account Permissions

Ensure your base service account has the following roles on the target service account:

- **Service Account Token Creator** (`roles/iam.serviceAccountTokenCreator`)

## Running the Examples

### Quick Start

```bash
# Run with default environment variables
./test_example.sh
```

### Manual Execution

```bash
# Load environment variables
source .env

# Run the example
go run example.go
```

### Build and Run

```bash
# Build the executable
go build -o bigquery-example example.go

# Run the executable
./bigquery-example
```

## Examples Included

1. **Basic Query** - Simple query execution with impersonation
2. **Bulk Insert** - Insert data into BigQuery tables
3. **Metadata Queries** - Get table schema and metadata
4. **User Auth + Impersonation** - Using user authentication as base
5. **Error Handling** - Proper error handling for various scenarios

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `GOOGLE_CLOUD_PROJECT` | Yes | Your Google Cloud Project ID |
| `ADBC_BIGQUERY_DATASET_ID` | Yes | BigQuery dataset ID |
| `ADBC_BIGQUERY_IMPERSONATE_TARGET` | Yes | Service account email to impersonate |
| `GOOGLE_APPLICATION_CREDENTIALS` | No | Path to service account key file |

## Troubleshooting

### Common Issues

1. **Authentication Failed**
   - Ensure Application Default Credentials are set up correctly
   - Verify service account key file path (if using one)
   - Check that the base service account has the required permissions

2. **Permission Denied**
   - Verify the target service account has BigQuery permissions
   - Ensure the base service account has Service Account Token Creator role

3. **Project/Dataset Not Found**
   - Double-check your project ID and dataset ID
   - Ensure the dataset exists in the specified project

### Debug Mode

To see more detailed output, you can run with debug logging:

```bash
export ADBC_LOG_LEVEL=debug
go run example.go
```

**Note**: Sensitive environment variables (project ID, dataset ID, service account) are automatically masked in logs for security. Only the first and last few characters are displayed.

## Dependencies

The example uses the following main dependencies:

- `github.com/apache/arrow-adbc/go/adbc` - ADBC core library
- `github.com/apache/arrow-adbc/go/adbc/driver/bigquery` - BigQuery ADBC driver
- `github.com/apache/arrow-go/v18` - Arrow Go library

## License

Licensed under the Apache License, Version 2.0.
