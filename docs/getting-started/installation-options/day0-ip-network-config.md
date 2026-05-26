# Day 0 IP and Network Configuration

This guide is the single Day 0 reference for IP address management and in-band/out-of-band network configuration for a NICo deployment. It walks an operator through every IP pool that must exist, how DHCP and DNS are served, and how to verify each piece end-to-end. Follow this page once during initial site bring-up; subsequent host ingestion and tenant operations rely on the configuration described here.

The values here are entered into the `siteConfig` TOML block of `helm-prereqs/values/ncx-core.yaml` (see [Quick Start Guide, Step 3c](../quick-start.md#3c-configure-ncx-core-site-deployment)) and into your site DNS and DHCP relay infrastructure. This page does not replace per-topic references — it consolidates them in the order an operator needs them. Sizing formulas, switch configuration, and BMC ingestion details are linked rather than duplicated.

---

## Prerequisites

Before configuring the items on this page, complete:

- [Hardware](../prerequisites/hardware.md) — server, DPU, and BMC inventory.
- [Network Prerequisites](../prerequisites/network.md) — VNI/ASN allocation, BGP/EVPN, route targets, and switch configuration.
- [BMC and Out-of-Band Setup](../prerequisites/bmc-oob-setup.md) — physical OOB connectivity and BMC credentials.

This page assumes the underlay and overlay routing decisions described in those pages have already been made.

---

## 1. IP Pool Allocation

NICo consumes IP addresses from three distinct sources:

| Source | Owner | Used for |
|---|---|---|
| The parent datacenter | External infrastructure (not NICo) | Control-plane management network IPs (site controller node BMCs, K8s node IPs) |
| Operator-supplied subnets configured at install time | NICo `siteConfig` | Admin network, DPU loopbacks, VPC loopbacks, tenant networks |
| The OS install image | Site operator (static assignment) | Site controller host OS ↔ DPU PF representor `/31` |

Plan capacity for all pools **before** running `setup.sh`. Pools can be grown at runtime without a restart, but an exhausted pool blocks the next provisioning operation immediately. A 20–25% headroom margin over expected peak allocation is a reasonable default.

For per-pool sizing formulas (servers, DPUs, VPCs, instance types), see [Network Prerequisites — IP Address Pools](../prerequisites/network.md#ip-address-pools).

### 1.1 Loopback Address Pools

NICo defines two named loopback pools in the `[pools.<name>]` section of `siteConfig`:

| Pool name | Allocation unit | Required | Purpose |
|---|---|---|---|
| `lo-ip` | One IP per managed DPU | Yes | DPU loopback advertised over BGP; VTEP address for the VXLAN underlay; BGP peer identifier |
| `vpc-dpu-lo` | One IP per (VPC, DPU) pair | Yes | Per-VPC VTEP used in the VPC overlay; allocated on demand when an instance is first placed on a DPU for a given VPC |

Define each pool with either a `prefix` or one or more `ranges`. Providing both, or neither, prevents the API server from starting.

```toml
[pools.lo-ip]
type = "ipv4"
prefix = "10.180.62.0/26"

[pools.vpc-dpu-lo]
type = "ipv4"
ranges = [
  { start = "10.180.63.1", end = "10.180.63.254" },
]
```

For pool semantics, runtime inspection (`admin-cli resource-pool list`), and the `grow` operation, see [IP Resource Pools](../../manuals/networking/ip_resource_pools.md).

> **Note:** Tenant workload IPs (the addresses an instance sees on its NICs) are not managed through these pools.

### 1.2 Host OS IP Assignment

The host OS on each site controller node receives its primary IP from the parent datacenter (typically via DHCP). NICo does not assign or manage the site controller's host OS IP.

For each site controller node that participates in DPU mode, a single `/31` point-to-point subnet is allocated between the host OS and the DPU PF representor. These IPs are **statically assigned at OS install time** — not by NICo and not by the parent datacenter DHCP server. One `/31` per site controller node (with DPUs) is required; nodes without DPUs need only one host IP.

For managed hosts (ingested machines), the host OS IP comes from the **admin network** until the host is assigned to a tenant:

- The admin network is a NICo-managed pool. Allocations are made by `carbide-api` and pushed to the host via DHCP.
- Size the admin network for at least one usable IP per managed server, plus network and broadcast addresses. Multiple admin segments may be declared in `[networks.<name>]`; each managed host sources its admin IP from whichever segment matches.
- When a tenant is assigned, the host's interfaces leave the admin network and join the relevant tenant networks (see [Network Prerequisites — Tenant Networks](../prerequisites/network.md#tenant-networks)).

The admin network is defined in the `[networks.admin]` block of `siteConfig`:

```toml
[networks.admin]
prefix = "10.180.64.0/24"
gateway = "10.180.64.1"
mtu = 1500
```

> **Warning:** `[networks.admin]` `prefix` and `gateway` must be non-empty. `carbide-api` panics at startup if either field is the empty string.

### 1.3 OOB/BMC IP Addresses (Static vs. Dynamic)

Every host BMC, DPU BMC, and DPU OOB interface needs an IP on the OOB management network. NICo supports two modes for each BMC interface:

| Mode | When IP is fixed | Configured by |
|---|---|---|
| **Dynamic** (default) | At first DHCP discovery, NICo allocates from the management network pool | `carbide-dhcp` + the relevant `[networks.<name>]` block in `siteConfig` |
| **Static** (predefined) | Set in `expected_machines.json` per host; `carbide-dhcp` serves that exact address on first contact | `bmc_ip_address` field per machine |

Mixing modes within the same site is supported — each host can use whichever mode is convenient.

**Per node**, expect to allocate:
- 1 IP for the host BMC.
- For hosts with DPUs: 1 IP for the DPU ARM OS + 1 IP for the DPU BMC, per DPU.

So a host with one DPU consumes three OOB addresses; a host without DPUs consumes one.

The OOB management network is declared as one or more NICo-managed network segments in `siteConfig` `[networks.<name>]` (block names are operator-chosen). Each segment carries its own prefix, gateway, and MTU. The OOB switches **must run a DHCP relay** pointed at the `carbide-dhcp` LoadBalancer VIP — they must not assign addresses themselves. See [BMC and Out-of-Band Setup](../prerequisites/bmc-oob-setup.md) for switch-side relay configuration.

### 1.4 Predefined BMC IP Allocation for Expected Machines

For sites that require a stable, pre-known BMC IP per host (for example, to wire DNS records or firewall rules before ingestion), set `bmc_ip_address` in the `expected_machines.json` manifest:

```json
{
  "expected_machines": [
    {
      "bmc_mac_address": "C4:5A:B1:C8:38:0D",
      "bmc_username": "root",
      "bmc_password": "default-password1",
      "chassis_serial_number": "SERIAL-1",
      "bmc_ip_address": "10.180.70.11"
    }
  ]
}
```

When `bmc_ip_address` is present:

- The address pre-allocates a machine interface in `carbide-api` at manifest upload time.
- The first DHCP DISCOVER from that BMC's MAC is answered with the pre-allocated address — `carbide-dhcp` does not draw from the dynamic OOB pool for that host.
- The pre-allocated address must fall within a network segment prefix declared in `siteConfig` `[networks.<name>]`, and it must not overlap any range used by the dynamic pool for that segment.

For the full `expected_machines.json` schema and upload command, see [Ingesting Hosts](../../provisioning/ingesting-hosts.md).

---

## 2. DHCP Configuration

### 2.1 How `carbide-dhcp` Works

`carbide-dhcp` is **not** a standalone DHCP daemon. It is a [Kea DHCP](https://www.isc.org/kea/) hooks library (`cdylib`) loaded into the upstream Kea v4 server inside the `carbide-dhcp` container. Every DHCPDISCOVER/REQUEST is intercepted by the hooks library and forwarded to `carbide-api` over mTLS gRPC (the `discover_dhcp` RPC). `carbide-api` decides what address to lease based on:

- Whether the source MAC matches an entry in `expected_machines` with a `bmc_ip_address` (predefined allocation).
- Otherwise, whether the source MAC is a known host/DPU BMC or DPU OOB interface — `carbide-api` consults the corresponding network segment pool and allocates the next free address.
- Vendor class (option 60) determines whether the client is a PXE/iPXE/BlueField boot client, which influences the boot options returned.

The hook callouts (`lease4_select` and `lease4_renew`) overwrite the lease that Kea would have selected — `yiaddr`, valid lifetime, and DHCP options are replaced with the values `carbide-api` produced, and the hook can return `SKIP` to cancel Kea's own lease assignment and database write. The result is written to Kea's memfile (`kea-leases4.csv`), but the authoritative record lives in `carbide-api`. From an operator perspective this means:

- The state-of-truth for every lease lives in `carbide-api`'s database, not in Kea's lease file.
- There is no standalone DHCP configuration file to populate with reservations — reservations come from `expected_machines.json` and `siteConfig` network segments.
- If `carbide-api` is unreachable, the hooks library serves cached negative responses (negative cache TTL: 5 minutes); this is a degraded-mode safety net, not a fallback pool.

### 2.2 DHCP Configuration for Host BMCs, DPU BMCs, and DPU OOB Addresses

All three interface types are served by the same `carbide-dhcp` instance. What distinguishes them at the wire level is which network segment the relayed request lands in — `carbide-api` selects the segment by matching the relay's `giaddr` against the `gateway` field of each `[networks.<name>]` block in `siteConfig`:

| Interface | DHCP request originates on | Served from |
|---|---|---|
| Host BMC | OOB management network | The `[networks.<name>]` segment whose `gateway` matches the OOB relay's `giaddr` |
| DPU BMC | OOB management network | Same as host BMC — both attach to whichever management segment matches `giaddr` |
| DPU OOB (ARM OS) | OOB management network | A management segment matched the same way; may share the BMC segment or be a distinct segment, depending on how `[networks.<name>]` blocks are declared |

Each `[networks.<name>]` block declares:

| Field | Purpose |
|---|---|
| `type` | Segment classification: `admin` for the admin segment, `underlay` for routed/per-TOR segments. NICo uses this to decide which segment is eligible for which interface role. |
| `prefix` | The IPv4 CIDR for the segment. |
| `gateway` | The address the OOB DHCP relay sets as `giaddr`; `carbide-api` matches the inbound request to this segment by comparing `giaddr` against this field. |
| `mtu` | MTU advertised to clients on this segment. |
| `reserve_first` | Number of leading addresses in the prefix to hold back from the dynamic pool (typically 5 — covers the network address, gateway, broadcast, plus headroom). |

A real site declares one `admin` segment and one `underlay` segment per OOB-facing TOR; the OOB management network is fragmented across as many `underlay` blocks as there are TORs.

To configure these flows:

1. **Declare the management network segments in `siteConfig`.** Use the schema above. The admin segment is not a singleton — a site may declare multiple admin/management segments, and the host's IP is sourced from whichever segment's `gateway` matches the relay's `giaddr` on the inbound DHCP request.
2. **Configure the DHCP relay on every OOB switch** to forward DHCP traffic to the `carbide-dhcp` LoadBalancer VIP (the IP assigned to the `carbide-dhcp` service by MetalLB in [Quick Start Step 3h](../quick-start.md#3h-assign-service-vips)). The relay must be on the same L2 broadcast domain as the BMCs and DPUs it serves.
3. **For predefined IPs**, upload `expected_machines.json` with `bmc_ip_address` populated **before** the host first powers on. Uploading after the BMC has already received a dynamic lease will not retroactively change its IP — release the lease (`carbide-admin-cli ... em ...`) and power-cycle the BMC.
4. **Set `dhcp_servers`** in `siteConfig` to the list of DHCP server IPs reachable from bare-metal hosts. This list is informational and is passed through to agents; it does not change how `carbide-dhcp` itself serves leases. May be left as `[]`.

The values that `carbide-dhcp` returns in DHCP options (nameservers, NTP servers, next-server, boot file, etc.) are sourced from:

- The Kea hook parameters in the `carbide-dhcp` Helm chart (`carbide-nameserver`, `carbide-ntpserver`, etc.) — set these to the `unbound.forge` (or `unbound.nico`, see [section 3](#3-dns-configuration)) recursive resolver VIP and your enterprise NTP server addresses.
- The per-segment definitions in `siteConfig` `[networks.<name>]` blocks — gateway, MTU, additional routes.

> **NTP note:** NICo does not run a standalone NTP service. NTP server addresses are distributed to managed hosts via DHCP option 42, configured in the `carbide-dhcp` chart Kea hook parameters (`carbide-ntpserver`). Point this to your enterprise NTP servers.

### 2.3 How to Verify DHCP Is Working

After deployment, validate the DHCP path end-to-end:

**Confirm the `carbide-dhcp` service is reachable on its LoadBalancer VIP:**

```bash
kubectl get svc carbide-dhcp -n forge-system
```

Both EXTERNAL-IP and TYPE=`LoadBalancer` must be populated. A `<pending>` IP indicates a MetalLB issue — see [Reference Installation — MetalLB troubleshooting](reference-install.md#metallb-loadbalancer-services-stuck-in-pending).

**Tail `carbide-dhcp` logs while a BMC powers on:**

```bash
kubectl logs -n forge-system -l app.kubernetes.io/name=carbide-dhcp --tail=20 -f
```

Each DISCOVER should produce a log line showing the source MAC, the resolved segment, and either a leased address or a `discover_dhcp` gRPC error from `carbide-api`. A `DeadlineExceeded` or `Unavailable` error means the hook cannot reach `carbide-api`; check the `carbide-api` LoadBalancer and TLS material.

**Inspect Kea's lease file** to confirm a lease was committed:

```bash
kubectl exec -n forge-system deploy/carbide-dhcp -- \
    cat /var/lib/kea/kea-leases4.csv | head
```

The lease IP and MAC should match what `carbide-api` allocated. The lease file is authoritative for Kea only — `carbide-api` is the system of record.

**From the OOB relay's vantage point**, verify packets are being forwarded by checking the switch's relay statistics (`show ip dhcp relay statistics` on Cumulus / SONiC). DISCOVER packets sent should match OFFER packets received.

For DHCP-related stuck states during ingestion, see the [WaitingForNetworkConfig playbook](../../playbooks/stuck_objects/waiting_for_network_config.md).

---

## 3. DNS Configuration

NICo's DNS layer has two distinct pieces:

| Piece | Backed by | Serves |
|---|---|---|
| `carbide-dns` | Either a PowerDNS Authoritative Server bridged to `carbide-api`, or a standalone DNS server inside the `carbide-dns` binary itself (see [section 3.1](#31-carbide-dns-zones-and-what-they-serve)) — both modes call `carbide-api` for record data | The site's authoritative zones — generated from machine, instance, and tenant records in the `carbide-api` database |
| `unbound` (recursive resolver) | Unbound | The resolver that managed machines (host BMCs, host OS, DPU OS, DPU BMCs) use for *all* DNS lookups |

These two roles are independent. Managed machines never query `carbide-dns` directly — they query the recursive resolver, which forwards or recurses as needed.

### 3.1 `carbide-dns` Zones and What They Serve

`carbide-dns` serves the site's authoritative zones from `carbide-api`'s database. It can run in either of two modes — pick one at deploy time:

| Mode | How DNS reaches `carbide-dns` | Selected by |
|---|---|---|
| **PowerDNS remote backend** (default) | PowerDNS Authoritative Server listens on UDP/TCP 53 and connects to `carbide-dns` over a Unix domain socket using PowerDNS's JSON remote-backend protocol. `carbide-dns` translates each request into a gRPC call to `carbide-api`. | Default; no `--listen` flag set on the `carbide-dns` binary. |
| **Standalone DNS server** | `carbide-dns` itself listens on UDP/TCP 53 and answers queries directly, calling `carbide-api` over gRPC for record data. PowerDNS is not in the path. | Pass `--listen=[::]:53` (or another address) to the `carbide-dns` binary. |

Both modes resolve the same records from the same source — the difference is only whether PowerDNS sits in front. Production deployments use both: choose based on whether you already operate PowerDNS at your site or prefer one fewer process to manage. The rest of this page applies to either mode.

The zones served are seeded by the `initial_domain_name` field in `siteConfig` (for example, `mysite.example.com`). On first start, `carbide-api` creates the corresponding domain record; `carbide-dns` then exposes whatever records exist in that zone in `carbide-api`'s database.

UFM endpoints under `default.ufm.<initial_domain_name>` are one example of records served this way when InfiniBand is configured (see [InfiniBand Setup](../../playbooks/ib_runbook.md)).

Operators do not edit `carbide-dns` zone files directly. Zone content is a function of `carbide-api`'s database state.

To configure `carbide-dns`:

1. Set `initial_domain_name` in `siteConfig` to your site's DNS domain.
2. Choose the mode (PowerDNS remote backend or standalone) and configure the `carbide-dns` Deployment / StatefulSet accordingly — set or omit `--listen` on the binary's command line, and include or exclude PowerDNS from the pod spec.
3. Assign a stable LoadBalancer VIP to the front-end service that listens on UDP/TCP 53 (one per replica via `perPodAnnotations`; see [Quick Start Step 3h](../quick-start.md#3h-assign-service-vips)). In PowerDNS mode this is the PowerDNS service; in standalone mode it is the `carbide-dns` service.
4. Delegate the `initial_domain_name` zone from your upstream DNS to those VIPs, or configure your recursive resolver to forward queries for the zone to them.

### 3.2 `unbound` Recursive Resolver for Managed Machines

Managed machines (host OS, DPU OS, host BMCs, DPU BMCs) need a recursive resolver that can resolve **both** the site-internal NICo service zone and external names. NICo deploys an `unbound` instance for this purpose.

The resolver address is distributed to managed machines via **DHCP option 6**, set in the `carbide-dhcp` Kea hook parameter `carbide-nameserver`. Managed machines have no compiled-in resolver address — changing the resolver is a DHCP configuration change, not a rebuild.

The resolver is responsible for:

- Recursive resolution of external (public-internet) names — needed for package fetches, NTP, etc.
- Authoritative resolution of the NICo service zone (`.forge`, `.nico`, or whichever convention your deployment uses; see below).
- Forwarding to `carbide-dns` for the site domain configured in `initial_domain_name`.

To configure `unbound`:

1. Populate the `local_data.conf` ConfigMap consumed by the `unbound` Helm chart with one A record per service VIP (see [section 3.3](#33-forge-dns-service-endpoints)).
2. Add a forward zone entry for `initial_domain_name` pointing at the `carbide-dns` VIPs.
3. Allow public-internet recursion (the default for the upstream `unbound` image) unless your site is fully air-gapped.

The `unbound` pod auto-reloads when the ConfigMap changes.

### 3.3 `.forge` DNS Service Endpoints

A fixed set of NICo service hostnames are resolved by DPU agents, host PXE loaders, and other in-band management components at runtime. Several of these names are **compiled into binaries or embedded shell scripts** and cannot be overridden via config — DNS is the only way to redirect them.

Two TLD conventions exist:

- **`.forge`** is the compiled default in `crates/agent/src/util.rs` and the host PXE loader scripts. The agent resolves `carbide-pxe.forge`, `carbide-ntp.forge`, etc. at startup. This is the TLD used by deployments built from the current binaries.
- **`.nico`** is the rebranded TLD documented in [`deploy/DNS.md`](https://github.com/NVIDIA/infra-controller-core/blob/main/deploy/DNS.md). New deployments may use this convention, but only if the agent and PXE images have been rebuilt with the new TLD.

Choose the convention that matches your binaries — do not mix. Verify by checking what the agent actually resolves at startup (`kubectl exec -n forge-system <agent-pod> -- getent hosts carbide-pxe.forge` or the `.nico` equivalent).

The required A records (shown for `.forge`; substitute `.nico` if your binaries use it) are:

| Hostname | Port | Resolves to | Purpose | Configurable at runtime? |
|---|---|---|---|---|
| `carbide-api.forge` | 443 | `carbide-api` external LoadBalancer VIP | NICo gRPC API | Yes — `NICO_API_URL` env var on most clients |
| `carbide-pxe.forge` | 80 | `carbide-pxe` LoadBalancer VIP | iPXE scripts, cloud-init, internal APT, TLS root CA | **No** — hardcoded in the compiled DPU agent |
| `carbide-static-pxe.forge` | 80 | Static PXE asset server VIP | `scout.cpio.zst`, `scout.efi`, BFB images, and other static boot artifacts | **No** — hardcoded in the host boot scripts that ship inside boot images |
| `carbide-ntp.forge` | 123 | Operator-supplied NTP server IP(s) — the record points at your existing NTP infrastructure, not a NICo-deployed service | NTP time sync; agent reads this and re-advertises via DHCP option 42 | **No** — hostname is hardcoded in the compiled DPU agent; multiple A records recommended |
| `unbound.forge` | 53 | `unbound` LoadBalancer VIP | Recursive DNS resolver | Yes — the resolver address itself is distributed via DHCP option 6 |
| `otel-receiver.forge` | 443 | OTel receiver VIP on the site controller | OTLP ingestion endpoint for DPU otel-collector sidecars | Yes — set in the otel-collector configuration YAML and re-deployed |

One additional `.forge` hostname, `socks.forge`, is hardcoded into the DPU agent as the SOCKS5 outbound proxy for DPU extension-service pods. Add a corresponding A record only if your environment runs a SOCKS5 proxy for that purpose; it is not part of every NICo deployment. For per-endpoint detail (consumers, in-cluster addresses, hardcode locations, and the `unbound`-vs-other-resolver guidance), see [`deploy/DNS.md`](https://github.com/NVIDIA/infra-controller-core/blob/main/deploy/DNS.md). That file is the canonical endpoint reference; the table above is the operator-facing summary.

> **Note:** Neither `.forge` nor `.nico` is a publicly registered TLD. Both are used exclusively on the isolated OOB management network. Configure the recursive resolver to treat the chosen TLD as locally authoritative and **not** forward queries to upstream public resolvers.

### 3.4 How to Verify DNS Is Working

**From the site controller**, confirm `carbide-dns` is responding:

```bash
kubectl get svc carbide-dns -n forge-system
dig +short @<carbide-dns-vip> <initial_domain_name>
```

**From an OOB-network vantage point** (a host BMC console, a managed host's BMC web UI shell, or any client on the OOB network), confirm the service zone resolves:

```bash
for name in carbide-api.forge carbide-pxe.forge carbide-static-pxe.forge \
            carbide-ntp.forge unbound.forge otel-receiver.forge; do
    printf "%-30s -> %s\n" "$name" "$(dig +short "$name" @<UNBOUND_VIP> || echo 'FAILED')"
done
```

Substitute `.nico` if that is the TLD baked into your binaries. Every name must return a non-empty A record set; a `FAILED` or empty result means the `local_data.conf` ConfigMap is missing that record. If your environment also runs a SOCKS5 proxy, extend the loop with `socks.forge`.

**Confirm reachability on the expected ports:**

```bash
# carbide-api gRPC (TLS handshake)
openssl s_client -connect carbide-api.forge:443 </dev/null 2>/dev/null | grep -E '^(subject|Verify)'

# carbide-pxe
curl -sf --max-time 5 http://carbide-pxe.forge/ -o /dev/null && echo OK || echo FAILED

# unbound recursing externally
dig +short +timeout=3 example.com @unbound.forge
```

A successful external recursion via `unbound.forge` confirms both DHCP option 6 (clients learn the resolver) and `unbound`'s recursion policy are correct.

---

## 4. End-to-end Day 0 Checklist

Use this checklist as the final gate before powering on the first host BMC:

- [ ] `siteConfig` `[pools.lo-ip]` and `[pools.vpc-dpu-lo]` populated with non-empty ranges.
- [ ] `siteConfig` `[networks.admin]` has non-empty `prefix` and `gateway`.
- [ ] One or more OOB network segments declared in `[networks.<name>]`, sized for `1 + 2 × DPU_count` IPs per host.
- [ ] `initial_domain_name` set in `siteConfig`.
- [ ] `dhcp_servers` set in `siteConfig` (or left as `[]`).
- [ ] `expected_machines.json` uploaded for every host; `bmc_ip_address` populated for any host that needs a predefined BMC IP.
- [ ] OOB switches configured with a DHCP relay pointing to the `carbide-dhcp` LoadBalancer VIP.
- [ ] LoadBalancer VIPs assigned for `carbide-api`, `carbide-dhcp`, `carbide-pxe`, `carbide-dns` (one per replica), `carbide-ssh-console-rs`, and `unbound`.
- [ ] `unbound`'s `local_data.conf` ConfigMap contains A records for `carbide-api`, `carbide-pxe`, `carbide-static-pxe`, `carbide-ntp`, `unbound`, and `otel-receiver` in the `.forge` (or `.nico`) zone; the `carbide-ntp` record points to your operator-supplied NTP server.
- [ ] `carbide-dns` zone for `initial_domain_name` is delegated from upstream DNS, or `unbound` forwards the zone to the `carbide-dns` VIPs.
- [ ] `unbound.forge` resolves every NICo service hostname (verified with the `dig` loop in [section 3.4](#34-how-to-verify-dns-is-working)).
- [ ] `carbide-dhcp` logs show DISCOVER → OFFER for a test BMC power-on.

When every item is checked, proceed to [Ingesting Hosts](../../provisioning/ingesting-hosts.md).

---

## Related Pages

- [Network Prerequisites](../prerequisites/network.md) — VNI/ASN/IPv4 sizing, BGP/EVPN, route targets, switch configuration.
- [BMC and Out-of-Band Setup](../prerequisites/bmc-oob-setup.md) — OOB physical network, DHCP relay setup, BMC credentials.
- [IP Resource Pools](../../manuals/networking/ip_resource_pools.md) — `lo-ip` / `vpc-dpu-lo` semantics, sizing, `admin-cli resource-pool grow`.
- [Quick Start Guide](../quick-start.md) — the install flow that consumes the configuration described here.
- [Reference Installation](reference-install.md) — phase-by-phase install with troubleshooting for MetalLB, DNS, and `carbide-api`.
- [Ingesting Hosts](../../provisioning/ingesting-hosts.md) — `expected_machines.json` schema and upload commands.
- [`deploy/DNS.md`](https://github.com/NVIDIA/infra-controller-core/blob/main/deploy/DNS.md) — canonical reference for NICo service hostnames, ports, and hardcoded-vs-configurable status.
