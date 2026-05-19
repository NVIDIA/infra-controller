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

use carbide_uuid::spx::SpxPartitionId;
use chrono::{DateTime, Utc};
use config_version::ConfigVersion;
use rpc::errors::RpcDataConversionError;
use rpc::forge as rpc_forge;
use serde::{Deserialize, Serialize};
use sqlx::postgres::PgRow;
use sqlx::{FromRow, Row};

use crate::tenant::TenantOrganizationId;

#[derive(Clone, Debug, Default)]
pub struct SpxPartitionSearchFilter {
    pub name: Option<String>,
    pub tenant_org_id: Option<String>,
}

impl From<rpc_forge::SpxPartitionSearchFilter> for SpxPartitionSearchFilter {
    fn from(filter: rpc_forge::SpxPartitionSearchFilter) -> Self {
        SpxPartitionSearchFilter {
            name: filter.name,
            tenant_org_id: filter.tenant_org_id,
        }
    }
}

#[derive(Debug, Clone)]
pub struct NewSpxPartition {
    pub id: SpxPartitionId,
    pub name: String,
    pub description: String,
    pub tenant_organization_id: String,
    pub vni: Option<i32>,
}

impl TryFrom<rpc_forge::SpxPartitionCreationRequest> for NewSpxPartition {
    type Error = RpcDataConversionError;
    fn try_from(req: rpc_forge::SpxPartitionCreationRequest) -> Result<Self, Self::Error> {
        if req.tenant_organization_id.is_empty() {
            return Err(RpcDataConversionError::InvalidArgument(
                "tenant_organization_id is required".to_string(),
            ));
        }

        let id = req.id.unwrap_or_else(|| uuid::Uuid::new_v4().into());

        let (name, description) = req
            .metadata
            .map(|m| (m.name, m.description))
            .unwrap_or_default();

        Ok(NewSpxPartition {
            id,
            name,
            description,
            tenant_organization_id: req.tenant_organization_id,
            vni: req.vni.map(|v| v.try_into()).transpose().map_err(
                |e: std::num::TryFromIntError| {
                    RpcDataConversionError::InvalidValue(
                        format!(
                            "`{}` cannot be converted to VNI",
                            req.vni.unwrap_or_default()
                        ),
                        e.to_string(),
                    )
                },
            )?,
        })
    }
}

impl TryFrom<SpxPartition> for rpc_forge::SpxPartition {
    type Error = RpcDataConversionError;
    fn try_from(src: SpxPartition) -> Result<Self, Self::Error> {
        if src.vni.is_none() {
            return Err(RpcDataConversionError::InvalidValue(
                "VNI is required".to_string(),
                "VNI is required".to_string(),
            ));
        }
        let vni = src.vni.unwrap();
        Ok(rpc_forge::SpxPartition {
            id: Some(src.id),
            metadata: Some(rpc_forge::Metadata {
                name: src.name,
                description: src.description,
                ..Default::default()
            }),
            tenant_organization_id: src.tenant_organization_id.to_string(),
            vni: vni as u32,
        })
    }
}

#[derive(Debug, Clone)]
pub struct SpxPartition {
    pub id: SpxPartitionId,
    pub name: String,
    pub description: String,
    pub tenant_organization_id: TenantOrganizationId,
    pub config_version: ConfigVersion,
    pub vni: Option<i32>,
    pub created: DateTime<Utc>,
    pub updated: DateTime<Utc>,
    pub deleted: Option<DateTime<Utc>>,
}

/// Returns whether the SPX partition has been soft-deleted
pub fn is_marked_as_deleted(partition: &SpxPartition) -> bool {
    partition.deleted.is_some()
}

#[derive(Debug, Serialize, Deserialize)]
pub struct SpxPartitionSnapshotPgJson {
    pub id: SpxPartitionId,
    pub name: String,
    pub description: String,
    pub tenant_organization_id: String,
    pub config_version: ConfigVersion,
    pub vni: Option<i32>,
    pub created: DateTime<Utc>,
    pub updated: DateTime<Utc>,
    pub deleted: Option<DateTime<Utc>>,
}

impl TryFrom<SpxPartitionSnapshotPgJson> for SpxPartition {
    type Error = sqlx::Error;
    fn try_from(value: SpxPartitionSnapshotPgJson) -> sqlx::Result<Self> {
        let tenant_organization_id =
            TenantOrganizationId::try_from(value.tenant_organization_id.clone())
                .map_err(|e| sqlx::Error::Decode(Box::new(e)))?;

        Ok(Self {
            id: value.id,
            name: value.name,
            description: value.description,
            tenant_organization_id,
            config_version: value.config_version,
            vni: value.vni,
            created: value.created,
            updated: value.updated,
            deleted: value.deleted,
        })
    }
}

impl<'r> FromRow<'r, PgRow> for SpxPartitionSnapshotPgJson {
    fn from_row(row: &'r PgRow) -> Result<Self, sqlx::Error> {
        let json: serde_json::value::Value = row.try_get(0)?;
        SpxPartitionSnapshotPgJson::deserialize(json).map_err(|err| sqlx::Error::Decode(err.into()))
    }
}
