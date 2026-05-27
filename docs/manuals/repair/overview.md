# Repair Workflows Overview

NICo supports two tenant-facing repair paths and provider-facing repair workflows. Start here to choose the right workflow.

## Choosing a Repair Path

| Situation | Use | Result |
|---|---|---|
| The tenant should keep the assigned instance while repair is attempted. | [Online Repair](online_repair.md) | The assigned instance moves from `Ready` to `Repairing`, and then back to `Ready` when the repair override is cleared. |
| The issue requires disruptive repair, or online repair did not resolve it. | [Release Instance for Full Repair](release_instance_for_repair.md) | The tenant releases the instance. NICo cleans and quarantines the machine before repair handling. |
| A platform admin or repair tenant needs to claim and repair a released machine. | [Repair Tenant Workflow](repair_tenant_workflow.md) | The repair tenant creates a targeted repair instance, sets repair outcome labels, and releases the machine back to ready or failed handling. |
| Provider automation, repair tenants, or operators need to understand repair signals and completion behavior. | [Repair System Integration](repair_integration.md) | Provider-side workflows use health overrides and repair completion signals to return machines safely to the allocation pool. |

## Workflow Summary

Online repair is the least disruptive path. It is requested through the Machine update API, keeps the tenant assignment in place, and uses a repair health override to keep the instance in `Repairing`.

Full repair is the disruptive path. It is requested through the Instance delete/release API with a machine health issue. The tenant gives up the instance, and NICo prevents the machine from returning to normal allocation until repair and validation are complete.

An instance cannot be released for full repair while it is still in online repair. If online repair is active and the issue now requires full repair, clear online repair first, wait for the instance to return to `Ready`, and then release the instance for full repair.

When clearing online repair because the repair failed, update the instance labels before releasing it for full repair. A label such as `onlineRepair.status: Failed` keeps the failure visible to tenant and operator tooling while the instance is being escalated.

Repair system integration is the provider-side behavior behind full repair. It explains how `tenant-reported-issue`, `repair-request`, repair tenants, and manual provider actions fit together.

The repair tenant workflow is the operator runbook for the middle of the full repair process: targeted instance creation into a dedicated repair tenant, machine repair, repair outcome labeling, and repair tenant release.

## API Surface

| Workflow | REST operation | Restish operation |
|---|---|---|
| Online repair | `PATCH /v2/org/{org}/nico/machine/{machineId}` | `update-machine` |
| Mark failed online repair | `PATCH /v2/org/{org}/nico/instance/{instanceId}` | `update-instance` |
| Release for full repair | `DELETE /v2/org/{org}/nico/instance/{instanceId}` | `delete-instance` |
| Repair tenant machine pickup | `POST /v2/org/{org}/nico/instance` | `create-instance` |
| Repair status and outcome labeling | `PATCH /v2/org/{org}/nico/machine/{machineId}` | `update-machine` |

Use the individual runbooks for exact payloads and validation rules.

## Reader Guide

Tenant admins should read:

1. [Online Repair](online_repair.md)
2. [Release Instance for Full Repair](release_instance_for_repair.md)

Provider operators and repair automation owners should also read:

1. [Repair Tenant Workflow](repair_tenant_workflow.md)
2. [Repair System Integration](repair_integration.md)
