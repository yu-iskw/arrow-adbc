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

//! BigQuery ADBC Driver

#[cfg(any(feature = "bundled", feature = "linked"))]
use std::ffi::{c_int, c_void};
use std::{fmt, sync::LazyLock};

#[cfg(any(feature = "bundled", feature = "linked"))]
use adbc_core::ffi::{FFI_AdbcDriverInitFunc, FFI_AdbcError, FFI_AdbcStatusCode};
use adbc_core::{
    driver_manager::ManagedDriver,
    error::Result,
    options::{AdbcVersion, OptionDatabase, OptionValue},
};

use crate::Database;

static DRIVER: LazyLock<Result<ManagedDriver>> = LazyLock::new(|| {
    ManagedDriver::load_dynamic_from_name(
        "adbc_driver_bigquery",
        Some(b"AdbcDriverBigQueryInit"),
        Default::default(),
    )
});

/// BigQuery ADBC Driver.
#[derive(Clone)]
pub struct Driver(ManagedDriver);

impl Default for Driver {
    fn default() -> Self {
        Self::try_load().expect("driver init")
    }
}

impl fmt::Debug for Driver {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.debug_struct("BigQueryDriver")
            .field("version", &self.0.version())
            .finish_non_exhaustive()
    }
}

#[cfg(any(feature = "bundled", feature = "linked"))]
extern "C" {
    #[link_name = "AdbcDriverBigQueryInit"]
    fn init(version: c_int, raw_driver: *mut c_void, err: *mut FFI_AdbcError) -> FFI_AdbcStatusCode;
}

impl Driver {
    /// Try to load the driver using the given ADBC version.
    pub fn try_load() -> Result<Self> {
        Self::try_new(Default::default())
    }

    fn try_new(version: AdbcVersion) -> Result<Self> {
        #[cfg(any(feature = "bundled", feature = "linked"))]
        {
            let driver_init: FFI_AdbcDriverInitFunc = init;
            ManagedDriver::load_static(&driver_init, version).map(Self)
        }
        #[cfg(not(any(feature = "bundled", feature = "linked")))]
        {
            let _ = version;
            Self::try_new_dynamic()
        }
    }

    fn try_new_dynamic() -> Result<Self> {
        DRIVER.clone().map(Self)
    }

    /// Dynamically load the driver library.
    pub fn try_load_dynamic() -> Result<Self> {
        Self::try_new_dynamic()
    }
}

impl adbc_core::Driver for Driver {
    type DatabaseType = Database;

    fn new_database(&mut self) -> Result<Self::DatabaseType> {
        self.0.new_database().map(Database)
    }

    fn new_database_with_opts(
        &mut self,
        opts: impl IntoIterator<Item = (OptionDatabase, OptionValue)>,
    ) -> Result<Self::DatabaseType> {
        self.0.new_database_with_opts(opts).map(Database)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn load_v1_1_0() {
        let result = ManagedDriver::load_dynamic_from_name(
            "adbc_driver_bigquery",
            Some(b"AdbcDriverBigQueryInit"),
            AdbcVersion::V110,
        );
        if let Err(err) = result {
            eprintln!("driver not available: {err}");
        }
    }
}
