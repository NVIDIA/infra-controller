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
#[allow(non_snake_case, unknown_lints, clippy::all)]
#[rustfmt::skip]
pub mod common;
#[allow(non_snake_case, unknown_lints, clippy::all)]
#[rustfmt::skip]
pub mod dns;
#[allow(non_snake_case, unknown_lints, clippy::all)]
#[rustfmt::skip]
pub mod core;

// Backward-compat alias module. See crates/rpc/src/protos/mod.rs for rationale —
// proto file was renamed forge.proto → core.proto with `package forge → package
// core` and `service Forge → service Core`. This alias keeps existing
// `crate::generated::forge::*` callsites compiling unchanged.
#[allow(unused_imports)]
#[rustfmt::skip]
pub mod forge {
    pub use super::core::*;
    pub mod forge_server {
        pub use super::super::core::core_server::Core as Forge;
        pub use super::super::core::core_server::CoreServer as ForgeServer;
    }
}
#[allow(non_snake_case, unknown_lints, clippy::all)]
#[rustfmt::skip]
pub mod health;
#[allow(non_snake_case, unknown_lints, clippy::all)]
#[rustfmt::skip]
pub mod machine_discovery;
#[allow(non_snake_case, unknown_lints, clippy::all)]
#[rustfmt::skip]
pub mod measured_boot;
#[allow(non_snake_case, unknown_lints, clippy::all)]
#[rustfmt::skip]
pub mod mlx_device;
#[allow(non_snake_case, unknown_lints, clippy::all)]
#[rustfmt::skip]
pub mod scout_firmware_upgrade;
#[allow(non_snake_case, unknown_lints, clippy::all)]
#[rustfmt::skip]
pub mod site_explorer;
