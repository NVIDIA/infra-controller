#!/bin/bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


# This script tweaks the original protodefs from nico because they're messy in a way that keeps us from building.
# Try and keep this script safe to rerun on the protodefs multiple times.

# Cross-platform sed -i (macOS requires '', Linux doesn't)
sedi() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "$@"
    else
        sed -i "$@"
    fi
}

# dpa_rpc.proto has a duplicate message "Metadata", we don't need any of it so just remove it
rm -f nicoproto/dpa_rpc.proto
sedi -e '/^import.*dpa_rpc/d' nicoproto/nico.proto

# nmx_c.proto and rack_manager.proto have duplicate ReturnCode/Metadata types, not needed
rm -f nicoproto/nmx_c.proto nicoproto/rack_manager.proto

# dns.proto has duplicate Domain/DomainList/Metadata types that collide with nico.proto
# when generated into the same Go package. Flow doesn't use DNS RPCs, so remove dns.proto
# and strip all dns references from nico.proto.
rm -f nicoproto/dns.proto
sedi -e '/^import.*dns\.proto/d' nicoproto/nico.proto
# Remove RPCs that reference dns.* types (CreateDomain through GetAllDomainMetadata)
sedi -e '/rpc CreateDomain(dns\./d' \
     -e '/rpc UpdateDomain(dns\./d' \
     -e '/rpc DeleteDomain(dns\./d' \
     -e '/returns (dns\./d' \
     -e '/rpc FindDomain(dns\./d' \
     -e '/rpc LookupRecord(dns\./d' \
     -e '/rpc GetAllDomains(dns\./d' \
     -e '/rpc GetAllDomainMetadata(dns\./d' \
     nicoproto/nico.proto
# Fix PxeDomain oneof: remove dns.Domain field, keep only legacy
sedi -e '/dns\.Domain/d' nicoproto/nico.proto
# Remove comments referencing dns.proto
sedi -e '/dns\.proto/d' nicoproto/nico.proto

# Both site explorer and main nico have a PowerState enum
sedi -e 's/ PowerState/ ComputerSystemPowerState/g' nicoproto/site_explorer.proto

# This one I'm blaming on the code generator.
sedi -e 's/MachineValidationStarted started/MachineValidationStarted oneof_started/' \
     -e 's/MachineValidationInProgress in_progress/MachineValidationInProgress oneof_in_progress/g' \
     -e 's/MachineValidationCompleted completed/MachineValidationCompleted oneof_completed/g' \
     nicoproto/nico.proto

# Prepend SPDX license header to all proto files so protoc-gen-go carries it into generated .pb.go
LICENSE_HEADER="// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the \"License\");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an \"AS IS\" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
"
for f in nicoproto/*.proto; do
    if ! head -1 "$f" | grep -q "SPDX"; then
        tmp="$(mktemp)"
        {
            printf '%s\n' "$LICENSE_HEADER"
            cat "$f"
        } > "$tmp" && mv "$tmp" "$f"
    fi
done
