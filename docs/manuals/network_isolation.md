# Network Isolation

NICo enforces tenant network isolation across three independent fabrics. Each
fabric uses a different mechanism, is configured through a different operator
API, and is verified separately. This page summarises the model so an operator
can choose the right guide; it is not a replacement for the per-fabric
configuration guides linked below.

| Fabric | Operator-facing primitive | Isolation enforced by |
|---|---|---|
| Ethernet | VPC + VpcPrefix (+ optional Network Security Group) | DPU VRF per VPC (HBN / NVUE) over a pure type-5 EVPN overlay |
| InfiniBand | InfiniBand partition | UFM P_Key partition membership; `IbFabricMonitor` reconciler |
| NVLink | NVLink logical partition | NMX-M / NMX-C partition lifecycle; `NvlPartitionMonitor` reconciler |

---

## Ethernet

A tenant's instance reaches a VPC by drawing addresses from one of the
**VpcPrefixes** attached to that VPC. NICo carves a /31 link-net per
interface from the prefix — one address to the instance, one to the
DPU's SVI in the VPC's VRF. An instance may participate in several VPCs
at once by having interfaces drawing from prefixes in different VPCs.
On the DPU of the managed host backing the instance, each related VPC
materialises as a Linux VRF; every host interface drawing from a prefix
in that VPC lives in that VRF.

VRFs are isolated by default. Cross-VPC reachability requires explicit VPC
peering or controlled route leaking via the VPC's routing profile. Layer 3 / 4
filtering within or across VPCs is provided by Network Security Groups,
attached at VPC or instance scope.

See [Ethernet Isolation](networking/ethernet_isolation.md) for the operator
configuration guide.

---

## InfiniBand

Each tenant InfiniBand partition maps to a UFM P_Key. Membership is enforced
by the subnet manager at the fabric level: hosts that are not full members of
a P_Key cannot exchange traffic with other members of that P_Key, regardless
of physical connectivity. NICo reconciles desired partition membership against
UFM via the `IbFabricMonitor` background task and surfaces the synchronisation
status to operators and to tenants.

See [InfiniBand Isolation](networking/infiniband_isolation.md) for the operator
configuration guide.

---

## NVLink

NVLink logical partitions group GPUs across hosts into a single isolated
NVLink domain. NICo drives partition lifecycle against the NMX-M REST API and
the NMX-C gRPC API and reconciles desired partitions periodically. Each tenant
instance that requests NVLink connectivity is placed into the partition
corresponding to its allocation; a host whose GPUs are not in a partition
cannot reach any other host's GPUs over NVLink.

See [NVLink Partitioning](nvlink_partitioning.md) for the operator
configuration guide.

---

## Cross-cutting behaviour

The following invariants apply to every fabric.

- **Per-fabric synchronisation status.** Each instance's `InstanceStatus`
  exposes a per-fabric `configs_synced` field that is `true` only when the
  observed fabric state matches the desired configuration. The aggregate
  `configs_synced` field is the logical AND of all per-fabric fields and gates
  the instance's `Ready` state.
- **Provisioning blocks on isolation convergence.** During initial
  provisioning, the instance state machine waits until every requested fabric
  has applied the desired configuration before the instance is marked `Ready`.
  Tenants observe this as the `Configuring` tenant state, and the machine
  remains in `WaitingForNetworkConfig` until the DPU reports back.
- **Termination blocks on isolation convergence.** During termination, the
  state machine waits until every fabric reports that the host has been
  removed from all tenant partitions before the instance is reported as
  deleted. This guarantees a terminated instance cannot continue to exchange
  traffic on any fabric.
- **Force-delete still tears down fabric state.** Force-deleting a managed
  host explicitly detaches it from every fabric through the same external
  APIs the normal lifecycle uses, so external fabric managers do not retain
  stale tenant references.
- **External fabric reachability is monitored.** Each external fabric service
  (UFM, NMX-M, NMX-C) is monitored from NICo with request-success and latency
  metrics so that fabric-side outages can be distinguished from NICo-side
  configuration errors.

For the architectural rationale and the patterns shared across all three
fabrics, see
[Networking Integrations](../architecture/networking_integrations.md).

For the Day 0 IP, DHCP, DNS, and admin-network configuration that every
isolation guarantee on this page rests on, see
[Day 0 IP and Network Configuration](../getting-started/installation-options/day0-ip-network-config.md).
