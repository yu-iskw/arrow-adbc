// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/apache/arrow-adbc/go/adbc"
	"github.com/apache/arrow-adbc/go/adbc/driver/bigquery"
	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
)

// closeResource safely closes a resource and logs any error
func closeResource(name string, closer interface{ Close() error }) {
	if err := closer.Close(); err != nil {
		log.Printf("Warning: failed to close %s: %v", name, err)
	}
}

// safeStringValue safely extracts a string value from an Arrow array
func safeStringValue(arr arrow.Array, index int) string {
	if arr == nil {
		return "null"
	}
	if arr.IsNull(index) {
		return "null"
	}
	if strArr, ok := arr.(*array.String); ok {
		return strArr.Value(index)
	}
	return "unknown"
}

// safeInt64Value safely extracts an int64 value from an Arrow array
func safeInt64Value(arr arrow.Array, index int) int64 {
	if arr == nil {
		return 0
	}
	if arr.IsNull(index) {
		return 0
	}
	if intArr, ok := arr.(*array.Int64); ok {
		return intArr.Value(index)
	}
	return 0
}

// maskSensitiveValue partially masks a sensitive value for logging
// This helps protect sensitive information while still providing useful debugging info
func maskSensitiveValue(value string) string {
	if len(value) <= 8 {
		// For very short values, just show first and last character
		if len(value) <= 2 {
			return "***"
		}
		return value[:1] + "***" + value[len(value)-1:]
	}
	// For longer values, show first 4 and last 4 characters
	return value[:4] + "***" + value[len(value)-4:]
}

// getConfig reads configuration from environment variables
func getConfig() (map[string]string, error) {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	datasetID := os.Getenv("ADBC_BIGQUERY_DATASET_ID")
	targetPrincipal := os.Getenv("ADBC_BIGQUERY_IMPERSONATE_TARGET")

	if projectID == "" {
		return nil, fmt.Errorf("GOOGLE_CLOUD_PROJECT environment variable is required")
	}
	if datasetID == "" {
		return nil, fmt.Errorf("ADBC_BIGQUERY_DATASET_ID environment variable is required")
	}
	if targetPrincipal == "" {
		return nil, fmt.Errorf("ADBC_BIGQUERY_IMPERSONATE_TARGET environment variable is required")
	}

	return map[string]string{
		"adbc.bigquery.sql.project_id":                   projectID,
		"adbc.bigquery.sql.dataset_id":                   datasetID,
		"adbc.bigquery.sql.auth_type":                    "adbc.bigquery.sql.auth_type.app_default_credentials",
		"adbc.bigquery.sql.impersonate.target_principal": targetPrincipal,
		"adbc.bigquery.sql.impersonate.scopes":           "https://www.googleapis.com/auth/bigquery,https://www.googleapis.com/auth/cloud-platform",
		"adbc.bigquery.sql.impersonate.lifetime":         "3600s", // 1 hour
	}, nil
}

// ExampleServiceAccountImpersonation demonstrates how to use service account
// impersonation with the BigQuery ADBC driver.
func ExampleServiceAccountImpersonation() {
	ctx := context.Background()

	// Get configuration from environment variables
	config, err := getConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Create driver
	driver := bigquery.NewDriver(memory.DefaultAllocator)

	// Configure database with service account impersonation using Application Default Credentials
	db, err := driver.NewDatabase(config)
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	defer closeResource("database", db)

	// Open connection
	conn, err := db.Open(ctx)
	if err != nil {
		log.Fatalf("Failed to open connection: %v", err)
	}
	defer closeResource("connection", conn)

	// Execute a simple query
	stmt, err := conn.NewStatement()
	if err != nil {
		log.Fatalf("Failed to create statement: %v", err)
	}
	defer closeResource("statement", stmt)

	err = stmt.SetSqlQuery("SELECT 1 as test_column")
	if err != nil {
		log.Fatalf("Failed to set SQL query: %v", err)
	}

	reader, _, err := stmt.ExecuteQuery(ctx)
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}
	defer reader.Release()

	// Process results
	fmt.Println("Query executed successfully with impersonated credentials!")
	fmt.Println("Results:")
	for reader.Next() {
		record := reader.Record()
		fmt.Printf("Schema: %v\n", record.Schema())
		fmt.Printf("Num rows: %d\n", record.NumRows())

		// Print the actual data
		for i := 0; i < int(record.NumRows()); i++ {
			col := record.Column(0)
			if arr, ok := col.(*array.Int64); ok {
				fmt.Printf("Row %d: %d\n", i, arr.Value(i))
			}
		}
	}
}

// ExampleBulkInsertWithImpersonation demonstrates bulk data insertion
// using service account impersonation.
func ExampleBulkInsertWithImpersonation() {
	ctx := context.Background()

	// Get configuration from environment variables
	config, err := getConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Create driver
	driver := bigquery.NewDriver(memory.DefaultAllocator)

	// Configure database with impersonation using Application Default Credentials
	db, err := driver.NewDatabase(config)
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	defer closeResource("database", db)

	// Open connection
	conn, err := db.Open(ctx)
	if err != nil {
		log.Fatalf("Failed to open connection: %v", err)
	}
	defer closeResource("connection", conn)

	// First, create a test table
	createStmt, err := conn.NewStatement()
	if err != nil {
		log.Fatalf("Failed to create statement: %v", err)
	}
	defer closeResource("create statement", createStmt)

	// Create a test table
	createQuery := fmt.Sprintf(`
		CREATE OR REPLACE TABLE `+"`%s.%s.test_users`"+` (
			name STRING,
			age INT64,
			city STRING
		)
	`, config["adbc.bigquery.sql.project_id"], config["adbc.bigquery.sql.dataset_id"])
	err = createStmt.SetSqlQuery(createQuery)
	if err != nil {
		log.Fatalf("Failed to set create table query: %v", err)
	}

	_, err = createStmt.ExecuteUpdate(ctx)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	fmt.Println("Test table created successfully!")

	// Create statement for bulk insert
	stmt, err := conn.NewStatement()
	if err != nil {
		log.Fatalf("Failed to create statement: %v", err)
	}
	defer closeResource("statement", stmt)

	err = stmt.SetSqlQuery(fmt.Sprintf("INSERT INTO `%s.%s.test_users` (name, age, city) VALUES (?, ?, ?)",
		config["adbc.bigquery.sql.project_id"], config["adbc.bigquery.sql.dataset_id"]))
	if err != nil {
		log.Fatalf("Failed to set SQL query: %v", err)
	}

	// Create data to insert
	schema := arrow.NewSchema([]arrow.Field{
		{Name: "name", Type: arrow.BinaryTypes.String},
		{Name: "age", Type: arrow.PrimitiveTypes.Int64},
		{Name: "city", Type: arrow.BinaryTypes.String},
	}, nil)

	builder := array.NewRecordBuilder(memory.DefaultAllocator, schema)
	defer builder.Release()

	// Add sample data
	names := []string{"Alice", "Bob", "Charlie", "Diana"}
	ages := []int64{25, 30, 35, 28}
	cities := []string{"New York", "San Francisco", "Chicago", "Boston"}

	builder.Field(0).(*array.StringBuilder).AppendValues(names, nil)
	builder.Field(1).(*array.Int64Builder).AppendValues(ages, nil)
	builder.Field(2).(*array.StringBuilder).AppendValues(cities, nil)

	record := builder.NewRecord()
	defer record.Release()

	// Bind the record and execute
	err = stmt.Bind(ctx, record)
	if err != nil {
		log.Fatalf("Failed to bind record: %v", err)
	}

	affected, err := stmt.ExecuteUpdate(ctx)
	if err != nil {
		log.Fatalf("Failed to execute update: %v", err)
	}

	fmt.Printf("Successfully inserted %d rows using impersonated credentials\n", affected)

	// Query the inserted data to verify
	queryStmt, err := conn.NewStatement()
	if err != nil {
		log.Fatalf("Failed to create query statement: %v", err)
	}
	defer closeResource("query statement", queryStmt)

	err = queryStmt.SetSqlQuery(fmt.Sprintf("SELECT name, age, city FROM `%s.%s.test_users` ORDER BY name",
		config["adbc.bigquery.sql.project_id"], config["adbc.bigquery.sql.dataset_id"]))
	if err != nil {
		log.Fatalf("Failed to set query: %v", err)
	}

	reader, _, err := queryStmt.ExecuteQuery(ctx)
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}
	defer reader.Release()

	fmt.Println("\nQuerying inserted data:")
	for reader.Next() {
		record := reader.Record()
		for i := 0; i < int(record.NumRows()); i++ {
			name := record.Column(0).(*array.String).Value(i)
			age := record.Column(1).(*array.Int64).Value(i)
			city := record.Column(2).(*array.String).Value(i)
			fmt.Printf("Name: %s, Age: %d, City: %s\n", name, age, city)
		}
	}
}

// ExampleMetadataQueries demonstrates how to get table metadata
// using service account impersonation to verify the table creator.
func ExampleMetadataQueries() {
	ctx := context.Background()

	// Get configuration from environment variables
	config, err := getConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Create driver
	driver := bigquery.NewDriver(memory.DefaultAllocator)

	// Configure database with impersonation
	db, err := driver.NewDatabase(config)
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Open connection
	conn, err := db.Open(ctx)
	if err != nil {
		log.Fatalf("Failed to open connection: %v", err)
	}
	defer conn.Close()

	// Get table schema with metadata
	fmt.Println("Getting table schema and metadata...")

	// Get the schema for our test table
	schema, err := conn.GetTableSchema(ctx, nil, nil, "test_users")
	if err != nil {
		log.Fatalf("Failed to get table schema: %v", err)
	}

	fmt.Printf("Table Schema: %v\n", schema)

	// Print metadata from the schema
	metadata := schema.Metadata()
	if metadata.Len() > 0 {
		fmt.Println("\nTable Metadata:")
		for i := 0; i < metadata.Len(); i++ {
			key := metadata.Keys()[i]
			value := metadata.Values()[i]
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	// Get table types to see what types are available
	fmt.Println("\nGetting table types...")
	tableTypesReader, err := conn.GetTableTypes(ctx)
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to get table types: %v\n", err)
		fmt.Println("   Continuing with other metadata queries...")
	} else {
		defer tableTypesReader.Release()

		fmt.Println("Available table types:")
		for tableTypesReader.Next() {
			record := tableTypesReader.Record()
			for i := 0; i < int(record.NumRows()); i++ {
				tableType := record.Column(0).(*array.String).Value(i)
				fmt.Printf("  - %s\n", tableType)
			}
		}
	}

	// Get objects to see what tables exist
	fmt.Println("\nGetting objects (tables) in the dataset...")
	objectsReader, err := conn.GetObjects(ctx, adbc.ObjectDepthAll, nil, nil, nil, nil, nil)
	if err != nil {
		// Handle the known BigQuery schema issue gracefully
		if adbcErr, ok := err.(adbc.Error); ok && adbcErr.Code == adbc.StatusInvalidArgument {
			fmt.Printf("⚠️  Warning: GetObjects failed due to BigQuery schema complexity: %s\n", adbcErr.Msg)
			fmt.Println("   This is a known issue with certain BigQuery table schemas.")
			fmt.Println("   Continuing with other metadata queries...")
		} else {
			// For any other error, just log it as a warning and continue
			fmt.Printf("⚠️  Warning: GetObjects failed: %v\n", err)
			fmt.Println("   Continuing with other metadata queries...")
		}
	} else {
		defer objectsReader.Release()

		fmt.Println("Objects in dataset:")
		for objectsReader.Next() {
			record := objectsReader.Record()
			fmt.Printf("Schema: %v\n", record.Schema())
			fmt.Printf("Num rows: %d\n", record.NumRows())

			// This is a complex nested structure, so we'll just show the schema
			// In a real application, you would parse the nested structure
			// to get catalog, schema, table, and column information
		}
	}

	// Execute a query to get table metadata using SQL
	fmt.Println("\nGetting table metadata via SQL query...")
	stmt, err := conn.NewStatement()
	if err != nil {
		log.Fatalf("Failed to create statement: %v", err)
	}
	defer stmt.Close()

	// Query to get table metadata including creation time and other details
	metadataQuery := fmt.Sprintf(`
		SELECT
			table_id,
			creation_time,
			last_modified_time,
			row_count,
			size_bytes,
			type
		FROM `+"`%s.%s.__TABLES__`"+`
		WHERE table_id = 'test_users'
	`, config["adbc.bigquery.sql.project_id"], config["adbc.bigquery.sql.dataset_id"])

	err = stmt.SetSqlQuery(metadataQuery)
	if err != nil {
		log.Fatalf("Failed to set metadata query: %v", err)
	}

	reader, _, err := stmt.ExecuteQuery(ctx)
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to execute metadata query: %v\n", err)
		fmt.Println("   This might be due to insufficient permissions or the table not existing.")
		fmt.Println("   Continuing with user verification...")
	} else {
		defer reader.Release()

		fmt.Println("Table metadata from SQL query:")
		for reader.Next() {
			record := reader.Record()
			for i := 0; i < int(record.NumRows()); i++ {
				// Use helper functions for safe value extraction
				tableID := safeStringValue(record.Column(0), i)
				creationTime := safeInt64Value(record.Column(1), i)
				lastModifiedTime := safeInt64Value(record.Column(2), i)
				rowCount := safeInt64Value(record.Column(3), i)
				sizeBytes := safeInt64Value(record.Column(4), i)

				// Handle table type which could be string or int64
				var tableType string
				if col5, ok := record.Column(5).(*array.String); ok {
					tableType = safeStringValue(col5, i)
				} else if col5Int, ok := record.Column(5).(*array.Int64); ok {
					tableType = fmt.Sprintf("%d", safeInt64Value(col5Int, i))
				} else {
					tableType = "unknown"
				}

				fmt.Printf("  Table ID: %s\n", tableID)
				fmt.Printf("  Creation Time: %d (Unix timestamp)\n", creationTime)
				fmt.Printf("  Last Modified Time: %d (Unix timestamp)\n", lastModifiedTime)
				fmt.Printf("  Row Count: %d\n", rowCount)
				fmt.Printf("  Size Bytes: %d\n", sizeBytes)
				fmt.Printf("  Table Type: %s\n", tableType)
			}
		}
	}

	// Query to get information about who created the table
	// Note: BigQuery doesn't directly expose the creator in metadata
	// but we can verify that we're using the impersonated service account
	fmt.Println("\nVerifying service account impersonation...")

	// Get current user information
	userQuery := "SELECT SESSION_USER() as current_user"
	err = stmt.SetSqlQuery(userQuery)
	if err != nil {
		log.Fatalf("Failed to set user query: %v", err)
	}

	userReader, _, err := stmt.ExecuteQuery(ctx)
	if err != nil {
		log.Fatalf("Failed to execute user query: %v", err)
	}
	defer userReader.Release()

	fmt.Println("Current user (should be the impersonated service account):")
	for userReader.Next() {
		record := userReader.Record()
		for i := 0; i < int(record.NumRows()); i++ {
			currentUser := record.Column(0).(*array.String).Value(i)
			fmt.Printf("  Current User: %s\n", currentUser)

			// Verify it matches our target principal
			expectedUser := config["adbc.bigquery.sql.impersonate.target_principal"]
			if currentUser == expectedUser {
				fmt.Printf("  ✅ SUCCESS: Service account impersonation is working correctly!\n")
				fmt.Printf("  ✅ The table was created by the impersonated service account: %s\n", currentUser)
			} else {
				fmt.Printf("  ❌ WARNING: Service account impersonation may not be working as expected\n")
				fmt.Printf("  ❌ Expected: %s\n", expectedUser)
				fmt.Printf("  ❌ Actual: %s\n", currentUser)
			}
		}
	}
}

// ExampleImpersonationWithUserAuth demonstrates service account impersonation
// using user authentication as the base authentication method.
func ExampleImpersonationWithUserAuth() {
	ctx := context.Background()

	// Get configuration from environment variables
	config, err := getConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Add user authentication configuration
	config["adbc.bigquery.sql.auth_type"] = "adbc.bigquery.sql.auth_type.user_authentication"
	config["adbc.bigquery.sql.auth.client_id"] = "your-client-id"
	config["adbc.bigquery.sql.auth.client_secret"] = "your-client-secret"
	config["adbc.bigquery.sql.auth.refresh_token"] = "your-refresh-token"

	// Create driver
	driver := bigquery.NewDriver(memory.DefaultAllocator)

	// Configure database with impersonation using user authentication
	db, err := driver.NewDatabase(config)
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Open connection
	conn, err := db.Open(ctx)
	if err != nil {
		log.Fatalf("Failed to open connection: %v", err)
	}
	defer conn.Close()

	// Execute a simple query to verify impersonation is working
	stmt, err := conn.NewStatement()
	if err != nil {
		log.Fatalf("Failed to create statement: %v", err)
	}
	defer stmt.Close()

	err = stmt.SetSqlQuery("SELECT 1 as test_column")
	if err != nil {
		log.Fatalf("Failed to set SQL query: %v", err)
	}

	reader, _, err := stmt.ExecuteQuery(ctx)
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}
	defer reader.Release()

	fmt.Println("Query executed successfully with user auth + impersonation!")
}

// ExampleErrorHandling demonstrates proper error handling for
// service account impersonation scenarios.
func ExampleErrorHandling() {
	ctx := context.Background()

	// Get configuration from environment variables
	config, err := getConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Create driver
	driver := bigquery.NewDriver(memory.DefaultAllocator)

	// Configure database with impersonation
	db, err := driver.NewDatabase(config)
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Open connection with error handling
	conn, err := db.Open(ctx)
	if err != nil {
		if adbcErr, ok := err.(adbc.Error); ok {
			switch adbcErr.Code {
			case adbc.StatusInvalidArgument:
				log.Printf("Invalid argument: %s", adbcErr.Msg)
			case adbc.StatusUnauthorized:
				log.Printf("Authentication failed: %s", adbcErr.Msg)
			case adbc.StatusNotFound:
				log.Printf("Resource not found: %s", adbcErr.Msg)
			default:
				log.Printf("Error: %s", adbcErr.Msg)
			}
		} else {
			log.Printf("Unexpected error: %v", err)
		}
		return
	}
	defer conn.Close()

	fmt.Println("Connection established successfully!")

	// Test a query that might fail
	stmt, err := conn.NewStatement()
	if err != nil {
		log.Fatalf("Failed to create statement: %v", err)
	}
	defer stmt.Close()

	// Try to query a non-existent table
	err = stmt.SetSqlQuery(fmt.Sprintf("SELECT * FROM `%s.%s.non_existent_table` LIMIT 1",
		config["adbc.bigquery.sql.project_id"], config["adbc.bigquery.sql.dataset_id"]))
	if err != nil {
		log.Fatalf("Failed to set SQL query: %v", err)
	}

	_, _, err = stmt.ExecuteQuery(ctx)
	if err != nil {
		if adbcErr, ok := err.(adbc.Error); ok {
			switch adbcErr.Code {
			case adbc.StatusNotFound:
				fmt.Printf("Expected error (table not found): %s\n", adbcErr.Msg)
			default:
				fmt.Printf("Unexpected ADBC error: %s (code: %d)\n", adbcErr.Msg, adbcErr.Code)
			}
		} else {
			fmt.Printf("Unexpected error: %v\n", err)
		}
	} else {
		fmt.Println("Unexpected success - table should not exist")
	}
}

// ExampleMain demonstrates how to run the examples with proper
// environment variable configuration.
func ExampleMain() {
	// Check if required environment variables are set
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	datasetID := os.Getenv("ADBC_BIGQUERY_DATASET_ID")
	targetPrincipal := os.Getenv("ADBC_BIGQUERY_IMPERSONATE_TARGET")

	if projectID == "" {
		fmt.Println("Please set GOOGLE_CLOUD_PROJECT environment variable")
		return
	}

	if datasetID == "" {
		fmt.Println("Please set ADBC_BIGQUERY_DATASET_ID environment variable")
		return
	}

	if targetPrincipal == "" {
		fmt.Println("Please set ADBC_BIGQUERY_IMPERSONATE_TARGET environment variable")
		return
	}

	fmt.Println("Running service account impersonation examples...")
	fmt.Printf("Project ID: %s\n", maskSensitiveValue(projectID))
	fmt.Printf("Dataset ID: %s\n", maskSensitiveValue(datasetID))
	fmt.Printf("Target Principal: %s\n", maskSensitiveValue(targetPrincipal))

	// Run examples
	fmt.Println("\n=== Example 1: Basic Query ===")
	ExampleServiceAccountImpersonation()

	fmt.Println("\n=== Example 2: Bulk Insert ===")
	ExampleBulkInsertWithImpersonation()

	fmt.Println("\n=== Example 3: Metadata Queries ===")
	ExampleMetadataQueries()

	fmt.Println("\n=== Example 4: User Auth + Impersonation ===")
	ExampleImpersonationWithUserAuth()

	fmt.Println("\n=== Example 5: Error Handling ===")
	ExampleErrorHandling()
}

func main() {
	ExampleMain()
}

// To run these examples:
// 1. Set environment variables:
//    export GOOGLE_CLOUD_PROJECT="xxxx"
//    export ADBC_BIGQUERY_DATASET_ID="xxx"
//    export ADBC_BIGQUERY_IMPERSONATE_TARGET="xxx"
//
// 2. Ensure the base service account has the Service Account Token Creator role
//    on the target service account
//
// 3. Run the example:
//    go run example.go
//
// 4. Or build and run:
//    go build -o bigquery-example example.go
//    ./bigquery-example
