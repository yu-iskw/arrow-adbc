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

//! A builder for a [`Connection`]

use std::fmt;
#[cfg(feature = "env")]
use std::env;

use adbc_core::{
    error::Result,
    options::{OptionConnection, OptionValue},
    Database as _,
};

use crate::{builder::BuilderIter, Connection, Database};

/// A builder for [`Connection`].
#[derive(Clone, Default)]
#[non_exhaustive]
pub struct Builder {
    /// Result buffer size ([`Self::RESULT_BUFFER_SIZE`]).
    pub result_buffer_size: Option<i64>,
    /// Prefetch concurrency ([`Self::PREFETCH_CONCURRENCY`]).
    pub prefetch_concurrency: Option<i64>,
    /// Other options.
    pub other: Vec<(OptionConnection, OptionValue)>,
}

impl fmt::Debug for Builder {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.debug_struct("Builder").field("...", &self.other).finish()
    }
}

#[cfg(feature = "env")]
impl Builder {
    /// Environment variable for [`Self::result_buffer_size`].
    pub const RESULT_BUFFER_SIZE_ENV: &str = "ADBC_BIGQUERY_RESULT_BUFFER_SIZE";
    /// Environment variable for [`Self::prefetch_concurrency`].
    pub const PREFETCH_CONCURRENCY_ENV: &str = "ADBC_BIGQUERY_PREFETCH_CONCURRENCY";

    /// Construct a builder from environment variables.
    pub fn from_env() -> Result<Self> {
        #[cfg(feature = "dotenv")]
        let _ = dotenvy::dotenv();

        Ok(Self {
            result_buffer_size: env::var(Self::RESULT_BUFFER_SIZE_ENV)
                .ok()
                .and_then(|v| v.parse::<i64>().ok()),
            prefetch_concurrency: env::var(Self::PREFETCH_CONCURRENCY_ENV)
                .ok()
                .and_then(|v| v.parse::<i64>().ok()),
            ..Default::default()
        })
    }
}

impl Builder {
    const COUNT: usize = 2;

    pub const RESULT_BUFFER_SIZE: &str = "adbc.bigquery.sql.query.result_buffer_size";
    pub const PREFETCH_CONCURRENCY: &str = "adbc.bigquery.sql.query.prefetch_concurrency";

    /// Build a [`Connection`] using the provided [`Database`].
    pub fn build(self, database: &Database) -> Result<Connection> {
        database.new_connection_with_opts(self)
    }
}

impl IntoIterator for Builder {
    type Item = (OptionConnection, OptionValue);
    type IntoIter = BuilderIter<OptionConnection, { Builder::COUNT }>;

    fn into_iter(self) -> Self::IntoIter {
        BuilderIter::new(
            [
                self.result_buffer_size
                    .map(OptionValue::Int)
                    .map(|value| (Builder::RESULT_BUFFER_SIZE.into(), value)),
                self.prefetch_concurrency
                    .map(OptionValue::Int)
                    .map(|value| (Builder::PREFETCH_CONCURRENCY.into(), value)),
            ],
            self.other,
        )
    }
}
