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
pub mod bundle;
pub mod journal;
pub mod machine;
pub mod pcr;
pub mod profile;
pub mod records;
pub mod report;
pub mod site;

#[derive(thiserror::Error, Debug)]
pub enum Error {
    #[error("{0}")]
    Parse(String),
    #[error("{0}")]
    RpcConversion(String),
}

pub type Result<T> = std::result::Result<T, Error>;

pub trait DisplayName {
    fn display_name() -> &'static str;
}

pub trait FromGrpc<M>: TryFrom<M> + DisplayName
where
    <Self as std::convert::TryFrom<M>>::Error: std::fmt::Display,
{
    fn from_grpc(msg: M) -> Result<Self> {
        Self::try_from(msg).map_err(|e| {
            Error::RpcConversion(format!("bad message: {}: {e}", Self::display_name()))
        })
    }
}

pub trait FromGrpcOpt<M>: FromGrpc<M>
where
    <Self as std::convert::TryFrom<M>>::Error: std::fmt::Display,
{
    fn from_grpc_opt(msg: Option<M>) -> Result<Self> {
        msg.ok_or_else(|| {
            Error::RpcConversion(format!("{} is unexpectedly empty", Self::display_name()))
        })
        .and_then(Self::from_grpc)
    }
}

pub trait FromPbVec<M: Clone>: FromGrpc<M>
where
    <Self as std::convert::TryFrom<M>>::Error: std::fmt::Display,
{
    fn from_pb_vec(pbs: &[M]) -> Result<Vec<Self>> {
        pbs.iter()
            .map(|record| Self::from_grpc(record.clone()))
            .collect()
    }
}
