# Firmware Updates

This page covers when NICo updates firmware, how automated updates are
scheduled, how to control update windows, and where NICo hands off to
platform-specific firmware procedures.

Firmware updates run in two operating moments:

- **Initial ingestion and pre-ingestion.** NICo can update firmware before a
  host is fully ingested when the current firmware is too old for safe
  discovery, host-to-DPU pairing, or the configured minimum firmware policy.
- **Approved service windows.** After NICo manages a machine, firmware updates
  are scheduled during approved maintenance windows and are limited by health
  and concurrency policy.

For related background, see:

- [Architecture Overview](../architecture/overview.md#machine-update-manager)
- [Managed Host State Diagrams](../architecture/state_machines/managedhost.md)
- [DPU Lifecycle Management](../dpu-management/dpu-lifecycle-management.md#firmware-upgrades)
- [Rack State Machine](../architecture/state_machines/rack.md)
- [Redfish Workflow](../architecture/redfish_workflow.md)
- [Core Metrics](../manuals/metrics/core_metrics.md)

## Workflow Boundaries

NICo has several firmware paths. Use the path that matches the component being
updated.

| Scope | NICo path | Use |
|---|---|---|
| Host BMC, BIOS, UEFI, and platform firmware listed in host firmware metadata | Machine Update Manager and the managed-host state machine | Fleet host firmware drift and host firmware service windows. |
| DPU NIC firmware drift | Machine Update Manager and DPU reprovisioning | Automatic DPU reprovisioning when DPU NIC firmware is outside the configured baseline. |
| DPU BMC, UEFI, CEC, ERoT, and related DPU firmware | DPU reprovisioning | Verification and update as part of the DPU lifecycle. |
| NVIDIA platform procedures | Manual service procedure plus NICo completion gate | NVIDIA-managed platforms, such as GB200, that require a field procedure before NICo resumes automatic checks. |
| OEM platform procedures | OEM instructions plus NICo scheduling or script handoff | Vendor-specific packages, such as SMC/Supermicro platforms, where the OEM procedure is the authority for platform-specific steps. |
| Rack, compute tray, switch, and power shelf firmware | Rack maintenance, RMS, and component-manager workflows | Rack-level and component-level firmware outside host Machine Update Manager scheduling. |

Use `carbide-admin-cli` for the examples below. These commands talk to the NICo
Core gRPC API:

```bash
carbide-admin-cli -c <core-api-url> <command>
```

## Machine Update Manager

Machine Update Manager is a scheduler. It does not flash firmware directly. On
each run it:

1. Clears completed host and DPU update markers.
2. Counts machines already in maintenance or firmware update states.
3. Counts healthy and unhealthy hosts.
4. Computes how many additional updates can start.
5. Asks each enabled update module to start work until the site limit is
   reached.

Automated updates are selected only when the target is eligible. The normal
eligibility gates are:

- The machine is known to NICo and is in a state that can be updated.
- The relevant update module is enabled.
- The machine needs firmware according to the configured baseline.
- The machine is not already in maintenance or another update flow.
- Site health and concurrency policy allow another host to leave service.
- A required update window is active when the firmware entry requires explicit
  start.

The actual firmware work runs inside the managed-host lifecycle. This keeps
host power, BMC operations, DPU reprovisioning, health reports, and state
transitions under one lifecycle controller.

## Configuration

Firmware behavior is controlled by site configuration and firmware metadata.

### Machine Update Scheduling

| Field | Default | Meaning |
|---|---:|---|
| `machine_update_run_interval` | `300s` when unset | How often Machine Update Manager runs. |
| `machine_updater.instance_autoreboot_period.start` | unset | Start of the UTC window when machines may automatically reboot for updates. |
| `machine_updater.instance_autoreboot_period.end` | unset | End of the UTC window when machines may automatically reboot for updates. |
| `machine_updater.max_concurrent_machine_updates_absolute` | unset | Hard cap on concurrent machine updates. |
| `machine_updater.max_concurrent_machine_updates_percent` | unset | Percentage cap on concurrent machine updates. When both caps are set, NICo uses the lower effective limit. |

Example:

```toml
machine_update_run_interval = 60

[machine_updater]
instance_autoreboot_period.start = "2026-05-22T01:00:00Z"
instance_autoreboot_period.end = "2026-05-22T05:00:00Z"
max_concurrent_machine_updates_absolute = 5
max_concurrent_machine_updates_percent = 10
```

### Host Firmware Settings

| Field | Default | Meaning |
|---|---:|---|
| `firmware_global.autoupdate` | `false` | Enables automatic host firmware updates. |
| `firmware_global.host_enable_autoupdate` | `[]` | Host models or machine IDs to force-enable for host firmware autoupdate. |
| `firmware_global.host_disable_autoupdate` | `[]` | Host models or machine IDs to force-disable for host firmware autoupdate. |
| `firmware_global.run_interval` | `30s` | Firmware manager polling interval. |
| `firmware_global.max_uploads` | `4` | Maximum concurrent firmware uploads. |
| `firmware_global.concurrency_limit` | `16` | Maximum concurrent firmware flashing operations. |
| `firmware_global.firmware_directory` | `/opt/carbide/firmware` | Directory for firmware binaries and `metadata.toml` files. |
| `firmware_global.host_firmware_upgrade_retry_interval` | `60m` | Delay before retrying a failed host firmware upgrade. |
| `firmware_global.instance_updates_manual_tagging` | `true` | Requires manual tagging before host firmware updates are applied. |
| `firmware_global.no_reset_retries` | `false` | Disables retry logic after BMC resets during firmware operations. |
| `firmware_global.hgx_bmc_gpu_reboot_delay` | `30s` | Delay after GPU reboot before the HGX BMC can be accessed. |
| `firmware_global.requires_manual_upgrade` | `false` | Forces firmware upgrades through a manual completion gate. |
| `firmware_global.max_concurrent_bfb_copies` | `10` | Maximum concurrent BFB copies. |

Example:

```toml
[firmware_global]
autoupdate = true
host_enable_autoupdate = ["PowerEdge R750"]
host_disable_autoupdate = []
run_interval = "30s"
max_uploads = 4
concurrency_limit = 16
firmware_directory = "/opt/carbide/firmware"
host_firmware_upgrade_retry_interval = "60m"
requires_manual_upgrade = false
```

### DPU Firmware Settings

| Field | Default | Meaning |
|---|---:|---|
| `dpu_config.dpu_nic_firmware_initial_update_enabled` | `false` | Enables DPU NIC firmware updates on initial discovery. |
| `dpu_config.dpu_nic_firmware_reprovision_update_enabled` | `true` | Enables DPU NIC firmware updates during reprovisioning. |
| `dpu_config.dpu_models` | BF2/BF3 defaults | DPU firmware definitions. |
| `dpu_config.dpu_nic_firmware_update_versions` | BF2/BF3 NIC versions | Accepted DPU NIC firmware versions. Firmware drift outside this set can trigger automated DPU reprovisioning. |
| `dpu_config.dpu_enable_secure_boot` | `false` | Enables the secure boot flow for DPU provisioning through Redfish. |

## Firmware Baselines

NICo compares observed firmware inventory with configured firmware baselines.
Host baselines are loaded from embedded configuration and from
`metadata.toml` files under `firmware_global.firmware_directory`. Metadata can
define vendor and model matching, component ordering, known firmware versions,
default versions, minimum pre-ingestion versions, and whether explicit start is
required.

List host firmware entries known to NICo:

```bash
carbide-admin-cli -c <core-api-url> firmware show
```

The output includes vendor, model, component type, inventory-name match,
version, and whether the update needs explicit start.

Keep live "latest firmware version" tables in a single baseline source and link
to that source from site runbooks. Do not copy live version tables into this
operations page; version tables change independently from the update workflow.
If a site mirrors firmware versions into a wiki or dashboard, make the mirror
point back to the same source-controlled baseline or approved version catalog.

To update a host firmware baseline:

1. Add or update the firmware metadata for the target vendor and model.
2. Place the firmware binary where `firmware_global.firmware_directory` points,
   or provide the approved URL/script metadata used by the site.
3. Mark the intended version as the default for the component.
4. Verify the new baseline with `firmware show`.
5. Apply the update through a service window or machine-specific update window.

Example host firmware metadata:

```toml
[host_models.dell_r750]
vendor = "Dell"
model = "PowerEdge R750"
ordering = ["uefi", "bmc"]

[host_models.dell_r750.components.bmc]
current_version_reported_as = "^Installed-.*__iDRAC."
preingest_upgrade_when_below = "0.5"
known_firmware = [
  { version = "1.1", filename = "/opt/carbide/firmware/dell/r750_bmc_1.1.fw", checksum = "<md5>", default = true },
]

[host_models.dell_r750.components.uefi]
current_version_reported_as = "^Installed-.*__BIOS.Setup."
preingest_upgrade_when_below = "0.5"
known_firmware = [
  { version = "2.0", filename = "/opt/carbide/firmware/dell/r750_uefi_2.0.fw", checksum = "<md5>", default = true },
]
```

## Host Firmware Updates

Host firmware updates use Redfish inventory and firmware metadata. Common host
components include BMC, BIOS, UEFI, HGX BMC, GPU, NIC, and platform-specific
firmware, depending on the platform metadata.

During a host firmware update, NICo can:

- Upload firmware through Redfish or run an approved firmware script.
- Poll Redfish tasks and firmware inventory.
- Reset the BMC or host when required by the platform.
- Re-check firmware versions after activation.
- Apply a `HostUpdateInProgress` health report that prevents allocation while
  update work is active.

Enable, disable, or clear machine-specific auto-update behavior:

```bash
carbide-admin-cli -c <core-api-url> machine auto-update --machine <machine-id> --enable
carbide-admin-cli -c <core-api-url> machine auto-update --machine <machine-id> --disable
carbide-admin-cli -c <core-api-url> machine auto-update --machine <machine-id> --clear
```

Set an explicit firmware update window for one or more machines:

```bash
carbide-admin-cli -c <core-api-url> managed-host start-updates \
  --machines <machine-id-1> <machine-id-2> \
  --start 2026-05-22T01:00:00-0700 \
  --end 2026-05-22T05:00:00-0700
```

Cancel pending start windows:

```bash
carbide-admin-cli -c <core-api-url> managed-host start-updates \
  --machines <machine-id> \
  --cancel
```

Request host reprovisioning when the host must be put back through the
managed-host firmware path:

```bash
carbide-admin-cli -c <core-api-url> host reprovision set \
  --id <machine-id> \
  --update-firmware \
  --update-message "<ticket-or-maintenance-reference>"
```

## DPU Firmware Updates

DPU firmware is managed as part of the managed-host lifecycle. NICo tracks:

| Component | Inventory name | Meaning |
|---|---|---|
| DPU NIC firmware | `DPU_NIC` | Primary NIC firmware on the BlueField. |
| DPU BMC firmware | `BMC_Firmware` | DPU management controller firmware. |
| DPU UEFI firmware | `DPU_UEFI` | DPU boot firmware. |
| ATF / ERoT firmware | `Bluefield_FW_ERoT` | Arm Trusted Firmware or External Root of Trust. |

Machine Update Manager uses DPU NIC firmware drift as the automatic trigger.
During DPU reprovisioning, NICo verifies and updates the DPU firmware set
against the configured baseline.

Inspect DPU firmware status:

```bash
carbide-admin-cli -c <core-api-url> dpu versions
```

For the full DPU firmware flow, see
[DPU Lifecycle Management](../dpu-management/dpu-lifecycle-management.md#firmware-upgrades).

## Manual Platform Updates

Some platforms require a manual field procedure before NICo can complete the
firmware workflow.

### NVIDIA Platforms

For NVIDIA-managed platforms, follow the approved NVIDIA service procedure for
the exact platform and firmware package. GB200 is the common example for this
flow.

When manual firmware upgrade is required, NICo moves the managed host to a
manual waiting state. Complete the approved GB200 firmware procedure first.
After the field procedure is complete, mark the manual firmware upgrade
complete so NICo can resume automatic checks:

```bash
carbide-admin-cli -c <core-api-url> host reprovision mark-manual-upgrade-complete --id <machine-id>
```

### OEM Platforms

For OEM platforms, consult the OEM support site for the exact platform and
firmware package before starting work. SMC/Supermicro is one example of this
path. NICo can still schedule, track, or hand off the update when the site has
integrated an approved script or manual workflow. Treat the OEM procedure as
the authority for vendor-specific prerequisites, activation steps, and recovery.

## Rack and Component Firmware

Rack, compute tray, switch, and power shelf firmware are handled separately
from Machine Update Manager.

Use rack firmware records when applying a firmware bundle to a rack:

```bash
carbide-admin-cli -c <core-api-url> rack-firmware list
carbide-admin-cli -c <core-api-url> rack-firmware apply <rack-id> <firmware-id> prod
carbide-admin-cli -c <core-api-url> rack-firmware status <job-id>
carbide-admin-cli -c <core-api-url> rack-firmware history
```

Use component-manager firmware commands for targeted compute tray, switch,
power shelf, or rack updates:

```bash
carbide-admin-cli -c <core-api-url> component-manager update-firmware compute-tray \
  --machine-id <machine-id> \
  --target-version <target-version> \
  --component bmc,bios

carbide-admin-cli -c <core-api-url> component-manager update-firmware switch \
  --switch-id <switch-id> \
  --target-version <target-version> \
  --component bmc,cpld,bios,nvos

carbide-admin-cli -c <core-api-url> component-manager update-firmware power-shelf \
  --power-shelf-id <power-shelf-id> \
  --target-version <target-version> \
  --component pmc,psu
```

Check component firmware status:

```bash
carbide-admin-cli -c <core-api-url> component-manager get-firmware-update-status compute-tray --machine-id <machine-id>
carbide-admin-cli -c <core-api-url> component-manager get-firmware-update-status switch --switch-id <switch-id>
carbide-admin-cli -c <core-api-url> component-manager get-firmware-update-status power-shelf --power-shelf-id <power-shelf-id>
```

Rack maintenance can also run a firmware-upgrade activity for a full rack or a
selected subset of machines, switches, or power shelves:

```bash
carbide-admin-cli -c <core-api-url> rack maintenance start \
  --rack <rack-id> \
  --activities firmware-upgrade \
  --firmware-version <target-version>

carbide-admin-cli -c <core-api-url> rack maintenance start \
  --rack <rack-id> \
  --machine-ids <machine-id-1>,<machine-id-2> \
  --activities firmware-upgrade \
  --firmware-version <target-version> \
  --components BMC,BIOS
```

## Monitor Progress

Start with the object being updated, then move to logs and metrics when progress
is unclear.

| Task | Command |
|---|---|
| Show managed-host state and handler outcome | `carbide-admin-cli -c <core-api-url> managed-host show <machine-id>` |
| Show machine details and health reports | `carbide-admin-cli -c <core-api-url> machine show <machine-id>` |
| Show host firmware baseline entries | `carbide-admin-cli -c <core-api-url> firmware show` |
| Show DPU firmware status | `carbide-admin-cli -c <core-api-url> dpu versions` |
| Show rack firmware job status | `carbide-admin-cli -c <core-api-url> rack-firmware status <job-id>` |
| Show component firmware status | `carbide-admin-cli -c <core-api-url> component-manager get-firmware-update-status <target> ...` |

Useful metrics:

| Metric | Meaning |
|---|---|
| `carbide_pending_host_firmware_update_count` | Host machines that need a host firmware update. |
| `carbide_active_host_firmware_update_count` | Host machines actively updating firmware. |
| `carbide_pending_dpu_nic_firmware_update_count` | Machines with DPU NIC firmware drift. |
| `carbide_unavailable_dpu_nic_firmware_update_count` | Machines with DPU NIC firmware drift that are not currently available for update. |
| `carbide_running_dpu_updates_count` | Machines running DPU firmware update work. |
| `carbide_dpu_firmware_version_count` | DPUs reporting a firmware version. |
| `carbide_preingestion_waiting_download` | Pre-ingestion hosts waiting for firmware downloads on other machines to complete. |
| `carbide_preingestion_waiting_installation` | Pre-ingestion hosts with firmware uploaded and installation in progress. |

## Recovery

If a firmware update does not progress:

1. Inspect the managed-host, rack, or component status for the failing object.
2. Check the health reports and handler outcome for the reason NICo is waiting.
3. Check `carbide-api` logs for Redfish task failures, BMC reachability errors,
   script failures, or RMS/component-manager job errors.
4. Follow the platform or OEM recovery procedure before retrying a failed
   firmware operation.
5. Retry through the same NICo workflow after the underlying platform condition
   is corrected.

A host firmware failure can place the host in a failed firmware-upgrade state.
NICo retries according to `firmware_global.host_firmware_upgrade_retry_interval`
where retry is supported. A rack or component firmware failure should be
tracked through the rack firmware job, rack state, or component-manager status
for that target.

After the platform condition is corrected, reset a failed host reprovisioning
flow only when the site runbook calls for it:

```bash
carbide-admin-cli -c <core-api-url> managed-host reset-host-reprovisioning --machine <machine-id>
```
