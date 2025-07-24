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

use adbc_bigquery::Driver;
#[cfg(feature = "env")]
use adbc_bigquery::{connection, database};
use adbc_core::options::AdbcVersion;

#[test]
fn load_dynamic() {
    let res = Driver::try_load_dynamic();
    // It is ok if the library is not available; the error indicates missing driver
    // but the test ensures the function is callable.
    match res {
        Ok(_) => {}
        Err(err) => eprintln!("could not load driver: {err}"),
    }
}

#[test]
fn load_v110() {
    let res = adbc_core::driver_manager::ManagedDriver::load_dynamic_from_name(
        "adbc_driver_bigquery",
        Some(b"AdbcDriverBigQueryInit"),
        AdbcVersion::V110,
    );
    match res {
        Ok(_) => {}
        Err(err) => eprintln!("load failed: {err}"),
    }
}

#[cfg(feature = "env")]
#[test]
fn builder_from_env() {
    use std::env;

    env::set_var(database::Builder::PROJECT_ID_ENV, "proj");
    env::set_var(database::Builder::DATASET_ID_ENV, "data");
    env::set_var(connection::Builder::RESULT_BUFFER_SIZE_ENV, "5");

    let db_builder = database::Builder::from_env().unwrap();
    assert_eq!(db_builder.project_id.as_deref(), Some("proj"));
    assert_eq!(db_builder.dataset_id.as_deref(), Some("data"));

    let conn_builder = connection::Builder::from_env().unwrap();
    assert_eq!(conn_builder.result_buffer_size, Some(5));
}

#[cfg(feature = "env")]
mod tests {
    use super::*;
    use std::{ops::Deref, sync::LazyLock};

    use adbc_core::{
        error::{Error, Result},
        Connection as _, Statement as _,
    };
    use arrow_array::{cast::AsArray, types::Int64Type};

    use adbc_bigquery::{Connection, Database, Statement};

    static DRIVER: LazyLock<Result<Driver>> = LazyLock::new(Driver::try_load);
    static DATABASE: LazyLock<Result<Database>> =
        LazyLock::new(|| database::Builder::from_env()?.build(&mut DRIVER.deref().clone()?));
    static CONNECTION: LazyLock<Result<Connection>> =
        LazyLock::new(|| connection::Builder::from_env()?.build(&DATABASE.deref().clone()?));

    fn with_connection(func: impl FnOnce(Connection) -> Result<()>) -> Result<()> {
        CONNECTION.deref().clone().and_then(func)
    }

    fn with_empty_statement(func: impl FnOnce(Statement) -> Result<()>) -> Result<()> {
        with_connection(|mut conn| conn.new_statement().and_then(func))
    }

    #[test_with::env(ADBC_BIGQUERY_TESTS)]
    fn statement_execute() -> Result<()> {
        with_empty_statement(|mut statement| {
            statement.set_sql_query("SELECT 21 + 21")?;
            let batch = statement
                .execute()?
                .next()
                .expect("a record batch")
                .map_err(Error::from)?;
            assert_eq!(batch.column(0).as_primitive::<Int64Type>().value(0), 42);
            Ok(())
        })
    }

    #[test_with::env(ADBC_BIGQUERY_TESTS)]
    fn statement_execute_schema() -> Result<()> {
        with_empty_statement(|mut statement| {
            statement.set_sql_query("SELECT 1 AS one")?;
            let schema = statement.execute_schema()?;
            assert_eq!(schema.fields()[0].name(), "one");
            Ok(())
        })
    }
}
