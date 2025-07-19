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

//! A builder for a [`Database`]

use std::fmt;
#[cfg(feature = "env")]
use std::env;

use adbc_core::{
    error::Result,
    options::{OptionDatabase, OptionValue},
    Driver as _,
};

use crate::{builder::BuilderIter, Database, Driver};

/// A builder for [`Database`].
#[derive(Clone, Default)]
#[non_exhaustive]
pub struct Builder {
    /// Authentication type ([`Self::AUTH_TYPE`]).
    pub auth_type: Option<String>,
    /// Credentials ([`Self::AUTH_CREDENTIALS`]).
    pub credentials: Option<String>,
    /// OAuth client ID ([`Self::AUTH_CLIENT_ID`]).
    pub client_id: Option<String>,
    /// OAuth client secret ([`Self::AUTH_CLIENT_SECRET`]).
    pub client_secret: Option<String>,
    /// OAuth refresh token ([`Self::AUTH_REFRESH_TOKEN`]).
    pub refresh_token: Option<String>,
    /// Project ID ([`Self::PROJECT_ID`]).
    pub project_id: Option<String>,
    /// Dataset ID ([`Self::DATASET_ID`]).
    pub dataset_id: Option<String>,
    /// Table ID ([`Self::TABLE_ID`]).
    pub table_id: Option<String>,
    /// Other options.
    pub other: Vec<(OptionDatabase, OptionValue)>,
}

impl fmt::Debug for Builder {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.debug_struct("Builder").field("...", &self.other).finish()
    }
}

#[cfg(feature = "env")]
impl Builder {
    /// Environment variable for [`Self::auth_type`].
    pub const AUTH_TYPE_ENV: &str = "ADBC_BIGQUERY_AUTH_TYPE";
    /// Environment variable for [`Self::credentials`].
    pub const AUTH_CREDENTIALS_ENV: &str = "ADBC_BIGQUERY_AUTH_CREDENTIALS";
    /// Environment variable for [`Self::client_id`].
    pub const AUTH_CLIENT_ID_ENV: &str = "ADBC_BIGQUERY_AUTH_CLIENT_ID";
    /// Environment variable for [`Self::client_secret`].
    pub const AUTH_CLIENT_SECRET_ENV: &str = "ADBC_BIGQUERY_AUTH_CLIENT_SECRET";
    /// Environment variable for [`Self::refresh_token`].
    pub const AUTH_REFRESH_TOKEN_ENV: &str = "ADBC_BIGQUERY_AUTH_REFRESH_TOKEN";
    /// Environment variable for [`Self::project_id`].
    pub const PROJECT_ID_ENV: &str = "ADBC_BIGQUERY_PROJECT_ID";
    /// Environment variable for [`Self::dataset_id`].
    pub const DATASET_ID_ENV: &str = "ADBC_BIGQUERY_DATASET_ID";
    /// Environment variable for [`Self::table_id`].
    pub const TABLE_ID_ENV: &str = "ADBC_BIGQUERY_TABLE_ID";

    /// Construct a builder from environment variables.
    pub fn from_env() -> Result<Self> {
        #[cfg(feature = "dotenv")]
        let _ = dotenvy::dotenv();

        Ok(Self {
            auth_type: env::var(Self::AUTH_TYPE_ENV).ok(),
            credentials: env::var(Self::AUTH_CREDENTIALS_ENV).ok(),
            client_id: env::var(Self::AUTH_CLIENT_ID_ENV).ok(),
            client_secret: env::var(Self::AUTH_CLIENT_SECRET_ENV).ok(),
            refresh_token: env::var(Self::AUTH_REFRESH_TOKEN_ENV).ok(),
            project_id: env::var(Self::PROJECT_ID_ENV).ok(),
            dataset_id: env::var(Self::DATASET_ID_ENV).ok(),
            table_id: env::var(Self::TABLE_ID_ENV).ok(),
            ..Default::default()
        })
    }
}

impl Builder {
    /// Number of fields in the builder (except other).
    const COUNT: usize = 8;

    pub const AUTH_TYPE: &str = "adbc.bigquery.sql.auth_type";
    pub const AUTH_CREDENTIALS: &str = "adbc.bigquery.sql.auth_credentials";
    pub const AUTH_CLIENT_ID: &str = "adbc.bigquery.sql.auth.client_id";
    pub const AUTH_CLIENT_SECRET: &str = "adbc.bigquery.sql.auth.client_secret";
    pub const AUTH_REFRESH_TOKEN: &str = "adbc.bigquery.sql.auth.refresh_token";
    pub const PROJECT_ID: &str = "adbc.bigquery.sql.project_id";
    pub const DATASET_ID: &str = "adbc.bigquery.sql.dataset_id";
    pub const TABLE_ID: &str = "adbc.bigquery.sql.table_id";

    /// Build a [`Database`] using the provided [`Driver`].
    pub fn build(self, driver: &mut Driver) -> Result<Database> {
        driver.new_database_with_opts(self)
    }
}

impl IntoIterator for Builder {
    type Item = (OptionDatabase, OptionValue);
    type IntoIter = BuilderIter<OptionDatabase, { Builder::COUNT }>;

    fn into_iter(self) -> Self::IntoIter {
        BuilderIter::new(
            [
                self.auth_type
                    .map(OptionValue::String)
                    .map(|value| (Builder::AUTH_TYPE.into(), value)),
                self.credentials
                    .map(OptionValue::String)
                    .map(|value| (Builder::AUTH_CREDENTIALS.into(), value)),
                self.client_id
                    .map(OptionValue::String)
                    .map(|value| (Builder::AUTH_CLIENT_ID.into(), value)),
                self.client_secret
                    .map(OptionValue::String)
                    .map(|value| (Builder::AUTH_CLIENT_SECRET.into(), value)),
                self.refresh_token
                    .map(OptionValue::String)
                    .map(|value| (Builder::AUTH_REFRESH_TOKEN.into(), value)),
                self.project_id
                    .map(OptionValue::String)
                    .map(|value| (Builder::PROJECT_ID.into(), value)),
                self.dataset_id
                    .map(OptionValue::String)
                    .map(|value| (Builder::DATASET_ID.into(), value)),
                self.table_id
                    .map(OptionValue::String)
                    .map(|value| (Builder::TABLE_ID.into(), value)),
            ],
            self.other,
        )
    }
}
