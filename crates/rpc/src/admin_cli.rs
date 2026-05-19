/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
/*

/// admin_cli.rs
///
/// General utility code for working with and displaying data
/// with the admin CLI.

*/

use std::env;
use std::sync::atomic::{AtomicBool, Ordering};

pub use output::{Destination, OutputFormat};
#[cfg(feature = "sqlx")]
use sqlx::{Pool, Postgres};

/// SUMMARY is a global variable that is being used by a few structs which
/// implement serde::Serialize with skip_serialization_if.
///
/// I had wanted the ability to have summarized or extended versions of
/// serialized output, and decided I could use skip_serialization_if along with
/// a function that looks at a global variable.
///
/// You set --extended on the CLI, which controls whether or not to summarized
/// (default is summarized).
static SUMMARY: AtomicBool = AtomicBool::new(false);

pub fn serde_just_print_summary<T>(_: &T) -> bool {
    SUMMARY.load(Ordering::SeqCst)
}

pub fn just_print_summary() -> bool {
    SUMMARY.load(Ordering::SeqCst)
}

pub fn set_summary(val: bool) {
    SUMMARY.store(val, Ordering::SeqCst);
}

/// get_db_url returns the full DB URL to use for connecting (and resetting,
/// if requested).
pub fn get_db_url(db_url: &str, db_name: &str) -> String {
    // Attempt to grab the DATABASE_URL first.
    // If it doesn't exist, fall back to args.db_url.
    let db_base = match env::var("DATABASE_URL") {
        Ok(val) => val,
        Err(_) => db_url.to_string(),
    };
    db_base + "/" + db_name
}

#[cfg(feature = "sqlx")]
/// connect connects to the database for the provided db_url, which probably
/// comes from get_db_url.
pub async fn connect(db_url: &str) -> eyre::Result<Pool<Postgres>> {
    let pool = sqlx::Pool::<sqlx::postgres::Postgres>::connect(db_url).await?;
    Ok(pool)
}

/// ToTable is a trait which is used alongside the cli_output command
/// and being able to prettytable print results.
pub trait ToTable {
    fn into_table(self) -> eyre::Result<String>
    where
        Self: Sized,
    {
        Ok("not implemented".to_string())
    }
}

pub mod output {
    use std::fs::File;
    use std::io::{Write, stdout};

    use clap::ValueEnum;
    use serde::Serialize;
    pub use table::IntoTable;
    use table::{render_ascii_table, render_csv_table};

    /// Destination is an enum used to determine whether CLI output is going
    /// to a file path or stdout.
    pub enum Destination {
        Path(String),
        Stdout(),
    }

    #[derive(Default, PartialEq, Eq, ValueEnum, Clone, Copy, Debug)]
    #[clap(rename_all = "kebab_case")]
    pub enum OutputFormat {
        #[default]
        AsciiTable,
        Csv,
        Json,
        Yaml,
    }

    /// The FormattedOutput trait allows you to handle CLI output for a data
    /// structure. It has no required methods, however you do need to implement
    /// serde::Serialize and our own IntoTable trait, as this is a supertrait of
    /// those two.
    pub trait FormattedOutput: Serialize + IntoTable {
        /// Format the output data as bytes (probably UTF-8 text).
        fn format_output(&self, format: OutputFormat) -> Vec<u8> {
            match format {
                OutputFormat::Json => {
                    serde_json::to_vec_pretty(self).expect("Could not serialize as JSON")
                }
                OutputFormat::Yaml => {
                    let mut out = Vec::new();
                    serde_yaml::to_writer(&mut out, self).expect("Could not serialize as YAML");
                    out
                }
                OutputFormat::AsciiTable => render_ascii_table(self),
                OutputFormat::Csv => render_csv_table(self),
            }
        }

        /// Format the output data and write it to the specified destination.
        fn write_output(
            &self,
            format: OutputFormat,
            destination: Destination,
        ) -> std::io::Result<()> {
            let output = self.format_output(format);
            match destination {
                Destination::Stdout() => {
                    let mut stdout_guard = stdout().lock();
                    stdout_guard.write_all(output.as_slice())
                }
                Destination::Path(path) => {
                    File::create(path).and_then(|mut file| file.write_all(output.as_slice()))
                }
            }
        }
    }

    pub mod table {
        use std::borrow::Cow;

        use prettytable::{Row, Table};

        /// The IntoTable trait is used to help the AsciiTable and CSV
        /// formatters turn a data structure into tabular data.
        pub trait IntoTable: Sized {
            type Row;

            /// Return the header with the titles for each column. These should
            /// be ordered the same as in the `.row_values()` implementation.
            fn header(&self) -> &[&str];

            /// Return a slice spanning all of the rows.
            fn all_rows(&self) -> &[Self::Row];

            /// Return a Vec with the text values for this row, in the same
            /// order as `.header()`. If the Row's internal fields contain
            /// values represented as strings, these values can be returned as
            /// Cow::Borrowed, and if not you will need to string-format them
            /// and return these as Cow::Owned.
            fn row_values(row: &'_ Self::Row) -> Vec<Cow<'_, str>>;

            // fn render_text_table(&self) -> String {
            //     let table = make_table(self);
            //     format!("{table}")
            // }
        }

        // This is not a trait method in order to keep the `prettytable` types
        // out of the public API.
        fn make_table<T>(data_source: &T) -> Table
        where
            T: IntoTable,
        {
            let mut table = Table::new();
            let header = Row::from(data_source.header());
            table.set_titles(header);
            let rows = data_source.all_rows();
            rows.iter().for_each(|row| {
                let values = T::row_values(row);
                let row = Row::from(values);
                table.add_row(row);
            });

            table
        }

        pub fn render_ascii_table<T>(data_source: &T) -> Vec<u8>
        where
            T: IntoTable,
        {
            let mut out = Vec::new();
            let table = make_table(data_source);
            table.print(&mut out).expect("Couldn't render ASCII table");
            out
        }

        pub fn render_csv_table<T>(data_source: &T) -> Vec<u8>
        where
            T: IntoTable,
        {
            let mut out = Vec::new();
            let table = make_table(data_source);
            table.to_csv(&mut out).expect("Couldn't render CSV table");
            out
        }
    }
}
