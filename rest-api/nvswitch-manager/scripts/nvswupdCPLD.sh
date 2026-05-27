#!/usr/bin/env bash
#
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
#
# Usage:
#   ./update_cpld.sh <SWITCH_IP> <USER> <PASSWORD> <LOCAL_IMAGE_PATH>
#
# Example:
#   ./update_cpld.sh 7.243.155.195 'USER' 'PASSWORD' \
#     ./YTL_131/SwitchTray/CPLD_Prod_000370_REV0600_000377_REV1500_000373_REV1200_000390_REV0400_image.bin

set -uo pipefail

SWITCH_IP="${1:-}"
USER="${2:-}"
PASS="${3:-}"
LOCAL_IMG="${4:-}"

if [[ -z "$SWITCH_IP" || -z "$USER" || -z "$PASS" || -z "$LOCAL_IMG" ]]; then
  echo "Usage: $0 <SWITCH_IP> <USER> <PASSWORD> <LOCAL_IMAGE_PATH>"
  exit 1
fi

if [[ ! -f "$LOCAL_IMG" ]]; then
  echo "Image file '$LOCAL_IMG' not found"
  exit 1
fi

REMOTE_IMG="$(basename "$LOCAL_IMG")"
REMOTE_PATH="/home/${USER}/${REMOTE_IMG}"

SSHPASS_BASE=(sshpass -p "$PASS")
SSH_OPTS=(-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null)

echo "Copying image to switch..."
"${SSHPASS_BASE[@]}" scp "${SSH_OPTS[@]}" "$LOCAL_IMG" "${USER}@${SWITCH_IP}:/home/${USER}" || {
  echo "SCP failed"
  exit 1
}

sleep 1

echo "Listing remote directory (sanity check)..."
"${SSHPASS_BASE[@]}" ssh "${SSH_OPTS[@]}" "${USER}@${SWITCH_IP}" "ls -l \"$REMOTE_PATH\"" || {
  echo "Remote file not found after copy"
  exit 1
}

sleep 1

echo "Fetching CPLD firmware into NVOS..."
"${SSHPASS_BASE[@]}" ssh "${SSH_OPTS[@]}" "${USER}@${SWITCH_IP}" \
  "nv action fetch platform firmware CPLD1 file://$REMOTE_PATH" || {
  echo "nv fetch failed"
  exit 1
}

sleep 1

echo "Installing CPLD firmware..."
"${SSHPASS_BASE[@]}" ssh "${SSH_OPTS[@]}" "${USER}@${SWITCH_IP}" \
  "nv action install platform firmware CPLD1 files \"$REMOTE_IMG\"" || {
  echo "nv install failed"
  exit 1
}

sleep 1

echo "Deleting fetched CPLD firmware file from NVOS..."
"${SSHPASS_BASE[@]}" ssh "${SSH_OPTS[@]}" "${USER}@${SWITCH_IP}" \
  "nv action delete platform firmware CPLD1 files \"$REMOTE_IMG\"" || {
  echo "nv delete failed"
  exit 1
}

echo "Done."
