# Ethernet Isolation

This page explains how NICo provides Ethernet network isolation between
tenants and across VPCs, and how an operator configures and verifies it. It
is the Day-1 configuration guide; the architectural rationale lives in
[Networking Integrations](../../architecture/networking_integrations.md), and
the deep mechanics of VXLAN / EVPN, BGP route-targets, and routing profiles
live in [VPC Network Virtualization](../vpc/vpc_network_virtualization.md).

**Related pages**

- [Network Isolation Overview](../network_isolation.md)
- [Day 0 IP and Network Configuration](../../getting-started/installation-options/day0-ip-network-config.md)
  — the operator-facing Day 0 reference for IP pools, admin / underlay
  segments, DHCP, and DNS. The isolation guarantees on this page assume
  the Day 0 configuration is already in place.
- [VPC Network Virtualization](../vpc/vpc_network_virtualization.md) — the
  full VXLAN / EVPN, VRF, BGP, and routing-profile reference
- [VPC Routing Profiles](../vpc/vpc_routing_profiles.md)
- [VPC Peering](../vpc/vpc_peering_management.md)
- [Network Security Groups](network_security_groups.md)
- [IP Resource Pools](ip_resource_pools.md)
- [VNI Resource Pools](../vpc/vni_resource_pools.md)

---

## The Isolation Model

NICo's Ethernet isolation is built on three objects that compose into a
single chain. This page describes the **Native Networking (FNN)** model,
which is the official NICo network virtualization model.

```
Instance ──► Network Interface ──► VpcPrefix ──► VPC ──► VRF on DPU
                                  (/31 link-net is vended per interface)
```

Read this chain top to bottom:

1. A **tenant instance** owns one or more **network interfaces** on the host
   it is allocated to.
2. Each interface is allocated an IP address from one of the **VpcPrefixes**
   attached to a VPC. NICo carves a `/31` link-net from the VpcPrefix per
   interface — one address goes to the instance, the other goes to the
   DPU's SVI. The /31 is the operator-visible unit of consumption for the
   prefix; `VpcPrefixStatus` reports `total_31_segments` and
   `available_31_segments` so operators can size prefixes against expected
   instance counts.
3. Every VpcPrefix belongs to exactly one **VPC**. Attaching an interface
   to a VpcPrefix is what places the interface in the parent VPC. A VPC
   may have several VpcPrefixes; an instance may have interfaces drawing
   from prefixes in different VPCs.
4. On the DPU of the managed host backing the instance, each VPC the
   instance touches materialises as a **Linux VRF**. The DPU's SVI for
   each vended /31 lives in that VRF; everything beyond the SVI is
   routed.

> **Implementation detail.** Internally, NICo records each vended /31 as a
> `NetworkSegment` row attached to the VpcPrefix. Operators do not
> normally manipulate these directly — they are visible through the same
> RPCs but are tenant-managed only as a byproduct of instance
> configuration.

An operator configures three things independently to control what reaches
what:

| Concern | What it determines | Operator primitive |
|---|---|---|
| Subnet attachment | Which VPC's VRF an interface joins, and which CIDR pool its IP comes from | **VpcPrefix** attached to a VPC |
| Routing (L3) | Reachability between subnets and between VPCs across the fabric | VPC and its routing profile (type-5 EVPN imports / exports, route leaking) |
| Filtering (L3 / L4) | Which permitted flows reach an instance's prefixes and ports | Network Security Group |

An instance may have interfaces drawing from VpcPrefixes in **several VPCs
at once** (for example, a tenant-data VPC and a separate storage VPC). Each
VPC that the instance touches becomes its own VRF on the DPU; nothing
forwards between VRFs by default.

---

## VXLAN / EVPN Underlay (in Brief)

NICo carries tenant traffic over a VXLAN / EVPN overlay. Each DPU is a VTEP
that peers with the site fabric (route servers or top-of-rack switches) using
BGP EVPN. Each VPC is identified on the overlay by a VNI; the per-VPC VRF on
the DPU imports and exports BGP routes tagged with route-targets derived from
the VPC's VNI and the site's `datacenter_asn`. Isolation between VPCs is a
direct consequence: a VRF imports only the route-targets its routing profile
declares, so a route advertised in one VPC does not appear in another VPC's
forwarding table.

The tenant overlay is a **pure type-5 EVPN (IP-prefix) overlay**. NICo does
not stretch any tenant L2 segment across the fabric. The host-to-DPU link is
layer-2 (the segment's VLAN ID is the multiplexer), the DPU acts as the L3
gateway via the segment's SVI, and the DPU re-advertises the host's
instance route into the fabric as a type-5 EVPN prefix tagged with the VPC's
route-target. Cross-host reachability inside a tenant VPC is therefore an
L3 routing decision, never an L2 bridging decision; the segment's VNI is an
L3VNI identifying the parent VPC's VRF, not an L2VNI extending a broadcast
domain.

The admin overlay (see [Default Isolation](#default-isolation-the-admin-overlay))
is the one exception: admin segments carry both an L2VNI and an L3VNI to
support admin-side workflows that occasionally require L2 reachability.
Tenant segments never do.

For the full BGP / route-target / VTEP picture (loopback pools, per-DPU ASN,
internal vs. external VNI pools, default-route mechanisms, and the
configuration checklist for a new site), follow
[VPC Network Virtualization](../vpc/vpc_network_virtualization.md). This page
does not duplicate that material.

---

## Routing Isolation: VPCs and VRFs

A VPC is the unit of routing isolation:

- Every VPC has its own VRF on every DPU that hosts an instance with an
  interface in that VPC. The VRF holds only the routes the VPC's routing
  profile permits.
- Routes inside one VPC's VRF are not visible to another VPC's VRF. This is
  enforced by BGP route-target import / export, not by ACLs, and applies
  identically to instances of the same tenant and instances of different
  tenants.
- Reachability between two VPCs is opt-in. An operator can establish it
  through:
  - **VPC peering** — see
    [VPC Peering](../vpc/vpc_peering_management.md). Peered VPCs install
    each other's host routes via additional route-target imports.
  - **A shared external route-target** declared in both VPCs' routing
    profiles, when the network team operates a transit VRF.
  - **Controlled route leaking** between a VPC VRF and the underlay default
    VRF via the `leak_*` fields on the routing profile. This is intended for
    internet access or for narrow injection of underlay prefixes, and is
    described in detail under
    [VPC Network Virtualization → Controlled Route Leaking](../vpc/vpc_network_virtualization.md#controlled-route-leaking).
- The set of prefixes a tenant must never reach (for example, the management
  plane) is declared once site-wide in `deny_prefixes` in the API server
  configuration. The DPU enforces this as an ACL on every tenant VRF.

---

## Default Isolation: The Admin Overlay

NICo guarantees that a managed host is never permitted to carry tenant
traffic unless an explicit tenant configuration places it into a VPC. This
guarantee is upheld by an **admin overlay** that is separate from every
tenant VPC.

- During site initialisation NICo creates an admin VPC and a set of admin
  network segments. These are not tenant-visible and exist only to give the
  DPU somewhere safe to attach a host before, between, and after tenant
  allocations.
- When the API server assembles the per-host network configuration for a
  DPU, it sets `use_admin_network = true` in `ManagedHostNetworkConfig`
  whenever the host has no instance allocated, the instance has no
  interfaces configured for this DPU, or the host is in a transient
  lifecycle state in which tenant traffic must not flow. The DPU agent
  then places every host interface on the admin overlay instead of any
  tenant VPC's VRF.
- A DPU that cannot retrieve its configuration at all — for example,
  because the host is unknown to NICo and `GetManagedHostNetworkConfig`
  returns `NOT_FOUND` — places itself into **isolated mode**: every host
  interface is detached from any tenant overlay until NICo issues an
  explicit configuration. This is the fail-closed default; nothing in the
  data path will silently fall back to a tenant network.
- The same admin-overlay path is used to enforce isolation during
  instance termination. The instance state machine blocks the
  termination flow until the DPU confirms that all tenant interfaces have
  been moved off tenant VPCs and onto the admin overlay (or that the
  machine has been tagged with a health alert preventing reuse). This is
  what guarantees that a tenant whose instance has been released cannot
  remain on the wire as a "ghost instance".

A "default VPC" in the cloud-provider sense — a system-created VPC that
every tenant inherits — does not exist in NICo. Tenants get the VPCs an
operator (or the tenant API) creates for them; absent any such VPC, an
instance has nowhere to send tenant traffic.

For the deeper handler-level account, see
[VPC Network Virtualization → How a DPU Gets Its Configuration](../vpc/vpc_network_virtualization.md#how-a-dpu-gets-its-configuration).

---

## Subnet Attachment: VPC Prefixes

A **VpcPrefix** is the tenant-facing primitive for declaring an IPv4 or
IPv6 CIDR pool that a VPC may draw instance-interface addresses from.
Creating a VpcPrefix is how a tenant says "instances allocated into this
VPC should get IPs from this CIDR."

A VpcPrefix has:

- A **parent VPC**, set at creation time. The prefix cannot be moved
  between VPCs.
- A **CIDR** (`config.prefix`), IPv4 or IPv6, that NICo carves /31
  link-nets out of when instance interfaces are allocated.
- A **status** field reporting `total_31_segments` and
  `available_31_segments`. A prefix is exhausted when available reaches
  zero; further interface allocations against this prefix fail until
  either another VpcPrefix with capacity is attached to the same VPC, or
  the exhausted prefix is replaced.
- A **metadata** block (name, labels) for operator and tenant
  bookkeeping.

A VPC may have any number of VpcPrefixes attached. When an instance
interface is created in a VPC, NICo selects one of that VPC's prefixes
with available capacity and vends the next free /31 from it: one address
to the instance, the other to the DPU's SVI in the VPC's VRF.

The fabric carries the prefix as a type-5 EVPN route in the parent VPC's
VRF, tagged with the route-targets the VPC's routing profile declares.
There is no per-prefix VNI; the L3VNI is the VPC's VNI and applies
uniformly to every prefix attached to that VPC.

### How a Tenant Creates a VpcPrefix

In the FNN model, prefixes are created via the gRPC `CreateVpcPrefix` RPC
or its REST equivalent (`POST /v2/org/{org}/carbide/vpc-prefix`).
Operators can drive the same flow via `admin-cli`:

```
nico-admin-cli vpc-prefix create --vpc-id <vpc-id> --prefix <cidr> [--name <label>]
nico-admin-cli vpc-prefix show
```

Update and delete subcommands follow the same pattern. `DeleteVpcPrefix`
succeeds only when no /31 from the prefix is currently in use; in
practice that means every instance interface drawing from the prefix
must be released first.

### Implementation Note: NetworkSegments

Internally, NICo records each vended /31 as a `NetworkSegment` row
attached to the parent VpcPrefix. The `NetworkSegment` is the unit of
on-the-wire configuration the DPU receives — one VLAN ID, one /31, one
SVI in the VPC's VRF — but operators rarely manipulate segments
directly. The VpcPrefix's `available_31_segments` counter is the
operator's view of pool consumption, and the per-instance
`configs_synced.ethernet` field is the operator's view of convergence.
Segments come and go with instance-interface lifecycle; the VpcPrefix
is what the tenant declares and what persists.

### Multi-Prefix and Multi-VPC Instances

The VpcPrefix-attached-to-a-VPC model supports two common patterns
without any extra tenant configuration:

- **Multiple prefixes in the same VPC.** A VPC may host several
  VpcPrefixes — for example, a primary tenant subnet and a separate
  storage subnet. All of them land in the same VRF on the DPU. This is
  the right approach when a workload needs distinct CIDRs but a single
  routing policy.
- **Prefixes in different VPCs.** An instance's interfaces may draw
  from prefixes attached to different VPCs. Each VPC materialises as
  its own VRF on the DPU; the instance straddles them at the OS layer.
  This is how a workload joins, for example, a tenant-data VPC and a
  shared-services VPC simultaneously without giving up VPC-level
  routing isolation.

Per-interface attachment is configured through `UpdateInstanceConfig`.
The state machine guards `Ready` until the DPU reports that every
interface has converged on its target VPC / prefix / VRF; tenants
observe the in-flight state as `Configuring`.

---

## Layer 3 / 4 Isolation: Network Security Groups

Network Security Groups (NSGs) provide stateful or stateless rule-based
filtering on top of the VRF model. A NSG is a tenant-owned object that
carries a prioritised list of `(direction, protocol, src/dst CIDR,
src/dst port range, action)` rules and is attached at one of two scopes:

- **VPC scope.** Every instance in the VPC inherits the NSG's rules.
- **Instance scope.** The NSG applies to that instance only and overrides
  the VPC-scope NSG (if any).

NSG rules are pushed to the DPU as part of the same per-interface
configuration response that carries the segment and VRF information, and are
materialised into NVUE ACLs by the DPU agent. Propagation status is reported
back to NICo so that the per-instance `configs_synced` signal reflects NSG
convergence as well as VRF / segment convergence.

NSGs are the right tool for tenant-controlled segmentation **inside** a VPC
(for example, blocking East-West traffic between application tiers) and for
restricting which underlay-leaked prefixes a tenant is permitted to reach.
They are not a substitute for the VPC's routing isolation: an NSG cannot
make two VPCs reachable that the routing profile keeps apart.

See [Network Security Groups](network_security_groups.md) for the full rule
syntax, RPC reference, attachment workflow, and the current limitations.

---

## Configuration Workflow

Configure Ethernet isolation in the following order. Each step builds on
the previous one.

### 1. Confirm site-level prerequisites

Before any tenant configuration, the site must already have:

- The full Day 0 IP and network configuration in place. See
  [Day 0 IP and Network Configuration](../../getting-started/installation-options/day0-ip-network-config.md)
  for the canonical reference — admin network segment(s) declared in
  `[networks.admin]`, OOB management segments under `[networks.<name>]`,
  the `lo-ip` and `vpc-dpu-lo` loopback pools, DHCP relays, and DNS.
- Configured `fnn-asn`, `vpc-vni`, and (if external VPCs are needed)
  `external-vpc-vni` resource pools.
- A `datacenter_asn`, a `deny_prefixes` list, and at least one routing
  profile defined under `fnn.routing_profiles`.
- BGP / EVPN agreement with the network team on route-target convention.

If any of these are missing, follow the Day 0 page first, then the
[VPC Network Virtualization → Configuration Checklist for a New Site](../vpc/vpc_network_virtualization.md#configuration-checklist-for-a-new-site).
None of the steps below will succeed otherwise.

### 2. Create the VPC

Use `admin-cli vpc create` (or `CreateVpc` directly) and supply the routing
profile name. The profile determines whether the VPC is internal or external
and which route-targets it imports and exports. Creation fails fast if the
named profile is not defined site-wide.

### 3. Attach VpcPrefixes to the VPC

Use `admin-cli vpc-prefix create` (or `CreateVpcPrefix`). Supply the
parent VPC and the CIDR. Repeat for as many prefixes as the VPC needs.
Per-instance /31 link-nets and their underlying NetworkSegment rows are
vended automatically as instances are allocated; the operator does not
create segments directly.

### 4. (Optional) Create and attach Network Security Groups

For the tenant traffic patterns that require L3 / L4 filtering, create
NSGs and attach them at VPC or instance scope. See
[Network Security Groups](network_security_groups.md).

### 5. Allocate instances and attach interfaces

When an instance is allocated, declare each interface's parent VPC in
the instance configuration. NICo selects a VpcPrefix in that VPC with
capacity and vends a /31 to the interface. The instance is blocked in
`WaitingForNetworkConfig` until the DPU reports every interface
converged on the requested VPC, prefix, NSG, and VRF; only then does
the instance reach `Ready`.

---

## Verification

For each tenant that an operator has configured, confirm the following:

1. **VPC presence and status.** `admin-cli vpc list` shows the VPC in
   `Ready` state; `admin-cli vpc show <id>` shows the expected routing
   profile and (if applicable) the attached NSG.
2. **VpcPrefix presence and capacity.** `admin-cli vpc-prefix show`
   (optionally filtered by VPC) lists every VpcPrefix attached to the
   VPC and reports `available_31_segments`. Every VPC that needs to host
   instances must have at least one VpcPrefix with available capacity.
3. **Instance configuration convergence.** `admin-cli instance show <id>`
   reports `tenant.state = Ready` and `configs_synced.ethernet = true`.
   While provisioning is in flight, the tenant state is `Configuring` and
   the per-fabric `configs_synced.ethernet` is `false`.
4. **DPU view.** `admin-cli dpu show <machine_id>` (or the equivalent
   per-host inspection) reports a separate VRF for each VPC the instance
   touches, and the host interfaces are members of the expected VRFs.
5. **Underlay reachability.** All BGP sessions reported by the DPU are
   `established`, and the DPU's loopback addresses are reachable from
   every other DPU and from the route servers.

The negative case is equally important: an unallocated host should report
its DPU on the admin overlay, with no tenant VRF present. A host whose
`GetManagedHostNetworkConfig` returns `NOT_FOUND` should report itself in
isolated mode and refuse to carry tenant traffic.

---

## Troubleshooting

Most Ethernet-isolation symptoms reduce to one of the following classes.
Diagnose in order; each pointer leads to a focused troubleshooting section
elsewhere in the docs.

| Symptom | Likely class | Where to look |
|---|---|---|
| Instance never leaves `Configuring`; `configs_synced.ethernet = false` | DPU has not converged on the requested VPC / prefix / VRF | [Stuck in WaitingForNetworkConfig and DPU Health](../../playbooks/stuck_objects/waiting_for_network_config.md) |
| `CreateVpc` returns `NOT_FOUND` for the routing profile | Routing profile name not defined in `fnn.routing_profiles` | [VPC Routing Profiles](../vpc/vpc_routing_profiles.md) |
| Instance allocation fails with "no prefix capacity" | The instance's VPC has no VpcPrefix with `available_31_segments > 0`; attach another prefix | This page, [Subnet Attachment: VPC Prefixes](#subnet-attachment-vpc-prefixes) |
| Intra-VPC reachability works, internet access fails | Routing-profile `route_target_imports` or default-route mechanism misconfigured | [VPC Network Virtualization → Internet Connectivity](../vpc/vpc_network_virtualization.md#internet-connectivity) |
| Egress works, no return traffic | Network device VRFs not importing the VPC's `route_targets_on_exports` | [VPC Network Virtualization → Export Route-Targets and Return-Path Reachability](../vpc/vpc_network_virtualization.md#export-route-targets-and-return-path-reachability) |
| Two instances in the same VPC cannot reach each other on a permitted port | NSG rule denies the flow | [Network Security Groups](network_security_groups.md) |
| Tenant traffic appears on an unallocated host | `use_admin_network` not asserted; treat as a security incident | [Force Deleting and Rebuilding Hosts](../../playbooks/force_delete.md) and the architecture reference |
| `DeleteVpcPrefix` fails | At least one /31 from the prefix is still in use by an instance interface; release those instances first | This page, [How a Tenant Creates a VpcPrefix](#how-a-tenant-creates-a-vpcprefix) |
| New VPC has nowhere to put instances | No VpcPrefix attached yet; create one with `admin-cli vpc-prefix create --vpc-id <id> --prefix <cidr>` | This page |
