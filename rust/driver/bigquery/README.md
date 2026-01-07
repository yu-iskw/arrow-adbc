<!---
  Licensed to the Apache Software Foundation (ASF) under one
  or more contributor license agreements.  See the NOTICE file
  distributed with this work for additional information
  regarding copyright ownership.  The ASF licenses this file
  to you under the Apache License, Version 2.0 (the
  "License"); you may not use this file except in compliance
  with the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing,
  software distributed under the License is distributed on an
  "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
  KIND, either express or implied.  See the License for the
  specific language governing permissions and limitations
  under the License.
-->

# BigQuery driver for Arrow Database Connectivity (ADBC)

This crate provides Rust bindings for the experimental
[BigQuery ADBC driver](https://arrow.apache.org/adbc/current/driver/bigquery.html).
It is a thin wrapper around the native implementation and is loaded via the
ADBC driver manager. Builders are provided to configure the driver using
environment variables.

## Example

```rust,no_run
use adbc_bigquery::{connection, database, Driver};
use adbc_core::{Connection, Statement};

let mut driver = Driver::try_load()?;
let mut database = database::Builder::from_env()?.build(&mut driver)?;
let mut connection = connection::Builder::from_env()?.build(&database)?;
let mut statement = connection.new_statement().unwrap();
statement.set_sql_query("SELECT 1")?;
let _ = statement.execute()?;
# Ok::<(), Box<dyn std::error::Error>>(())
```
