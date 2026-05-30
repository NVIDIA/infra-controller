# DPF Setup for NICo Integration

## Introduction

NICo supports two ways of provisioning DPUs:

1. iPXE based
2. DPF based

This manual covers the **DPF based** flow and is written as a deployment guide
for DPF when it is going to be used by NICo. It assumes that a working
Kubernetes cluster is already available, and is intentionally agnostic to the
specific cluster implementation (kubeadm, k3s, RKE2, managed clouds, etc.) —
any conformant cluster that satisfies the DPF prerequisites is acceptable.

This guide is **not a replacement** for the official DPF documentation. The
authoritative source for installing and configuring DPF is the upstream guide:

- <https://docs.nvidia.com/networking/display/dpf25101>

NICo is designed to follow the Zero-Trust use case detailed in the DPF documentation: [DPF Zero-Trust Mode](https://docs.nvidia.com/networking/display/dpf25101/hbn-in-dpf-zero-trust).

You should follow that guide as the base. The instructions below only describe
the **deltas, additions, and tweaks** that need to be applied on top of the
official DPF flow so that NICo can integrate with the resulting DPF
installation. This manual is based on **DPF 26.04**; minor adjustments may be
necessary on other versions and on environments other than a development
setup.

The guide is organized into the following sections:

1. **Prerequisites** — work that must be done before installing DPF.
2. **DPF Installation** — NICo-relevant notes when installing the DPF operator.
3. **Post-Installation Configuration** — the cluster state that must be in
   place after DPF is installed and before NICo starts.

> **Note**: NICo expects DPF to be installed and configured on the same
> Kubernetes cluster where NICo (the controller) runs.

---

## 1. Prerequisites

The official DPF guide lists a set of cluster-level prerequisites (Argo CD,
cert-manager, Kamaji etc.). Follow that guide for those
components.

NICo reuses several of those same components (notably Argo CD and
cert-manager). If they are already installed for NICo, **do not reinstall
them** — only configure the missing pieces and adapt the existing
installations so DPF can use them. The subsections below cover the prerequisite
configuration that is specific to a NICo + DPF deployment.

### 1.1. Create the DPF operator namespace

All DPF operator workloads, secrets, ConfigMaps, and CRs live in the
`dpf-operator-system` namespace. Create it idempotently:

```bash
kubectl get namespace dpf-operator-system &>/dev/null \
  || kubectl create namespace dpf-operator-system
```

### 1.2. Image pull and helm repository credentials

Access to the DPF staging Helm chart and related container images requires authentication through NVIDIA NGC. Both the DPF operator and the workloads it deploys will need credentials for pulling Helm charts and container images from private registries. For detailed instructions, see: https://docs.nvidia.com/networking/display/dpf25101/using-private-registries.

#### 1.2.a. `hbn-user-password` Secret

A random local credential pair used by the HBN (Host-Based Networking) DPUService,
which runs FRR on the DPU. The DPF operator picks this Secret up by label.

```bash
kubectl -n dpf-operator-system create secret generic hbn-user-password \
  --from-literal=password=`tr -dc 'a-z0-9' < /dev/urandom | head -c 10` \
  || kubectl get secret hbn-user-password -n dpf-operator-system

kubectl -n dpf-operator-system label secret hbn-user-password \
  dpu.nvidia.com/image-pull-secret=""
```

The `dpu.nvidia.com/image-pull-secret=""` label is a DPF convention that tells
the operator *"propagate this Secret into DPUService image-pull secrets."* The
label is reused even though this is not strictly an image-pull Secret — DPF's
controllers selector-match on this label to mirror Secrets onto the DPU
cluster.

#### 1.2.b. `dpf-pull-secret` docker-registry Secret

Credentials for `nvcr.io`, used by the DPF operator and by the operands it
deploys to pull staging images.

```bash
kubectl -n dpf-operator-system create secret docker-registry dpf-pull-secret \
  --docker-server=nvcr.io \
  --docker-username='$oauthtoken' \
  --docker-password="$NGC_API_KEY" \
  || kubectl get secret dpf-pull-secret -n dpf-operator-system

kubectl -n dpf-operator-system label secret dpf-pull-secret \
  dpu.nvidia.com/image-pull-secret=""
```

#### 1.2.c. Argo CD repository Secrets for Helm charts

DPF pulls several Helm charts via Argo CD. Apply the following Secrets so that
Argo CD can authenticate to the NGC Helm repositories:

```yaml
---
apiVersion: v1
kind: Secret
metadata:
  name: ngc-doca-oci-helm
  namespace: argocd
  labels:
    argocd.argoproj.io/secret-type: repository
stringData:
  name: nvstaging-doca-oci
  url: nvcr.io/nvstaging/doca
  type: helm
  password: $NGC_API_KEY
data:
  # $oauthtoken base64 encoded. This prevents envsubst from substituting the value.
  username: JG9hdXRodG9rZW4=
    ## true
  enableOCI: dHJ1ZQ==
---
apiVersion: v1
kind: Secret
metadata:
  name: ngc-doca-https-helm
  namespace: argocd
  labels:
    argocd.argoproj.io/secret-type: repository
stringData:
  name: nvstaging-doca-https
  url: https://helm.ngc.nvidia.com/nvstaging/doca
  type: helm
  password: $NGC_API_KEY
data:
  username: JG9hdXRodG9rZW4=
---
apiVersion: v1
kind: Secret
metadata:
  name: ngc-carbide-https-helm
  namespace: argocd
  labels:
    argocd.argoproj.io/secret-type: repository
stringData:
  name: nvstaging-carbide-https
  url: https://helm.ngc.nvidia.com/0837451325059433/carbide-dev
  type: helm
  password: $NGC_API_KEY
data:
  username: JG9hdXRodG9rZW4=
```

Each Secret is labelled `argocd.argoproj.io/secret-type: repository`, which is
how Argo CD discovers Helm repositories.

| Secret name | Repo URL | Type | Used by |
|---|---|---|---|
| `ngc-doca-oci-helm` | `nvcr.io/nvstaging/doca` | OCI helm | DPF operator chart pulls |
| `ngc-doca-https-helm` | `https://helm.ngc.nvidia.com/nvstaging/doca` | HTTPS helm | Some DPUService charts |
| `ngc-carbide-https-helm` | `https://helm.ngc.nvidia.com/0837451325059433/carbide-dev` | HTTPS helm | Carbide-private DPUService charts |

### 1.3. Cert-manager policy and RBAC for DPF

DPF relies on cert-manager to mint short-lived certificates. If the cluster
runs `approver-policy` (CRD `policy.cert-manager.io/CertificateRequestPolicy`),
**no CSR will be approved unless a matching policy whitelists it**, and DPF's
CSRs will hang in `Pending` indefinitely.

Two objects must therefore be installed:

1. A `CertificateRequestPolicy` that is permissive for the
   `dpf-operator-system` namespace.
2. A `ClusterRole` + `ClusterRoleBinding` granting cert-manager itself the
   `use` verb on that policy.

> **Note**: The policy and role below use wildcard (`*`) values for
> convenience. In production, the exact set of allowed names, SANs, and usages
> should be tightened with help from the DPF team.

#### `policy.yaml`

```yaml
---
apiVersion: policy.cert-manager.io/v1alpha1
kind: CertificateRequestPolicy
metadata:
  labels:
    argocd.argoproj.io/instance: dpf-pki-policies
  name: dpf-approval-policy
spec:
  selector:
    namespace:
      matchNames: [dpf-operator-system]
    issuerRef:
      name: '*'
      kind: '*'
      group: '*'
  allowed:
    commonName:
      value: '*'
    dnsNames:
      values: ['*']
    ipAddresses:
      values: ['*']
    uris:
      values: ['*']
    emailAddresses:
      values: ['*']
    isCA: true
    usages:
      - server auth
      - client auth
      - digital signature
      - key encipherment
```

This allows any CertificateRequest in the `dpf-operator-system` namespace,
against any issuer, with any SAN (DNS / IP / URI / email), CA or not, with the
listed usages.

#### `rbac-role.yaml`

```yaml
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cert-manager-policy:dpf-approval-policy
rules:
  - apiGroups: [policy.cert-manager.io]
    resources: [certificaterequestpolicies]
    verbs: [use]
    resourceNames: [dpf-approval-policy]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cert-manager-policy:dpf-approval-policy
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cert-manager-policy:dpf-approval-policy
subjects:
  - kind: ServiceAccount
    name: cert-manager
    namespace: cert-manager
```

Without this binding cert-manager's controller cannot reference the policy and
**all DPF CSRs will hang in pending**.

---

## 2. DPF Installation

Follow the upstream DPF installation guide for the actual install procedure:

- <https://docs.nvidia.com/networking/display/dpf25101>

When installing the DPF operator chart, two parameter overrides are required
for a NICo-integrated deployment. The example command below illustrates how to
set them:

```bash
REGISTRY="oci://path/to/doca"
TAG="v26.4.0-rc.3"
helm upgrade --install -n dpf-operator-system \
  --set "enableNodeFeatureRules=false" \
  --set "imagePullSecrets[0].name=dpf-pull-secret" \
  dpf-operator $REGISTRY/dpf-operator --version=$TAG
```

NICo-specific notes on the parameters:

- `enableNodeFeatureRules=false` — the chart's bundled `NodeFeatureRule`
  resources are disabled because nodes are labeled via NFD's own configuration
  (relying on PCI class `0200`).
- `imagePullSecrets[0].name=dpf-pull-secret` — ties the operator's pods to the
  pull Secret created in step 1.2.b so that staging images can be pulled.

Adjust `REGISTRY` and `TAG` to the version of DPF you are deploying.

---

## 3. Post-Installation Configuration (before NICo starts)

Once the DPF operator is running, the following objects must be applied
**before NICo is started**. They configure the DPF operator for NICo's
provisioning model and grant the orchestrator the access it needs.

### 3.1. Cluster-wide RBAC for the NICo orchestrator

The NICo orchestrator (the `carbide-api` ServiceAccount in NICo's default
deployment) needs to read and write across namespaces, including
`dpf-operator-system` and the per-DPU namespaces. Grant it `cluster-admin` via
a `ClusterRoleBinding`:

```yaml
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: carbide-api-dpf
  namespace: dpf-operator-system
rules:
  - apiGroups: ["provisioning.dpu.nvidia.com"]
    resources: ["bfbs"]
    verbs: ["get", "list", "create", "patch", "delete"]
  - apiGroups: ["provisioning.dpu.nvidia.com"]
    resources: ["dpus"]
    verbs: ["get", "list", "watch", "patch", "delete"]
  - apiGroups: ["provisioning.dpu.nvidia.com"]
    resources: ["dpudevices"]
    verbs: ["get", "list", "create", "patch", "delete"]
  - apiGroups: ["provisioning.dpu.nvidia.com"]
    resources: ["dpunodes"]
    verbs: ["get", "list", "create", "patch", "delete"]
  - apiGroups: ["provisioning.dpu.nvidia.com"]
    resources: ["dpunodemaintenances"]
    verbs: ["get", "patch"]
  - apiGroups: ["provisioning.dpu.nvidia.com"]
    resources: ["dpuflavors"]
    verbs: ["get", "create"]
  - apiGroups: ["provisioning.dpu.nvidia.com"]
    resources: ["dpusets"]
    verbs: ["get", "patch"]
  - apiGroups: ["provisioning.dpu.nvidia.com"]
    resources: ["dpuclusters"]
    verbs: ["get", "list"]
  - apiGroups: ["svc.dpu.nvidia.com"]
    resources: ["dpudeployments"]
    verbs: ["get", "list", "create", "patch", "delete"]
  - apiGroups: ["svc.dpu.nvidia.com"]
    resources: ["dpuservices", "dpuservicechains"]
    verbs: ["get", "list", "create", "patch", "delete"]
  - apiGroups: ["svc.dpu.nvidia.com"]
    resources: ["dpuserviceinterfaces", "dpuservicetemplates", "dpuserviceconfigurations", "dpuservicenads"]
    verbs: ["get", "list", "create", "patch", "delete"]
  - apiGroups: ["operator.dpu.nvidia.com"]
    resources: ["dpfoperatorconfigs"]
    verbs: ["get", "patch"]
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "create"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: carbide-api-dpf
  namespace: dpf-operator-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: carbide-api-dpf
subjects:
  - kind: ServiceAccount
    name: carbide-api
    namespace: forge-system
```

If you are integrating a different orchestrator or running NICo under a
different ServiceAccount or namespace, replace `carbide-api` / `forge-system`
accordingly. The principle is the same: the orchestrator SA must be able to
manage DPF objects cluster-wide.

### 3.2. `DPFOperatorConfig`

This is the operator-level CR that tells DPF how to behave in a NICo environment. For more information about the available fields and their details, refer to the official DPF guide.

```yaml
---
apiVersion: operator.dpu.nvidia.com/v1alpha1
kind: DPFOperatorConfig
metadata:
  name: dpfoperatorconfig
  namespace: dpf-operator-system
spec:
  dpuDetector:
    disable: true
  provisioningController:
    osInstallTimeout: "60m"
    installInterface:
      installViaRedfish:
        skipDPUNodeDiscovery: true
  overrides:
    # Replace with the IP of the KubeAPI server where DPF control plane is running
    kubernetesAPIServerVIP: "REPLACE_ME"
    # Replace with the port of the KubeAPI server where DPF control plane is running
    kubernetesAPIServerPort: "REPLACE_ME"
    argoCDNamespace: argocd
  kamajiClusterManager:
    disable: false
  networking:
    highSpeedMTU: 9000
  imagePullSecrets:
    - dpf-pull-secret
```

Field-by-field:

| Field | Meaning |
|---|---|
| `dpuDetector.disable: true` | DPF normally polls hosts to discover new DPUs. NICo disables auto-discovery because DPUs are fed in via `DPUSet` CRs from the orchestrator. |
| `provisioningController.osInstallTimeout: "60m"` | Total budget for the OS install flow per DPU. |
| `installViaRedfish` | Provision DPUs by talking Redfish to the host BMC (vs. PXE-based). |
| `skipDPUNodeDiscovery: true` | Do not auto-detect DPUs as Kubernetes nodes — DPF is told about them explicitly by NICo. |
| `overrides.kubernetesAPIServerVIP` | Replace `CONTROL_PLANE_IP` with the host-cluster API-server VIP that DPUs should reach. |
| `overrides.kubernetesAPIServerPort` | Host-cluster API-server port (`6443` by default). |
| `overrides.argoCDNamespace` | Namespace where Argo CD is installed. |
| `kamajiClusterManager.disable: false` | Use Kamaji as the DPU control plane. |
| `networking.highSpeedMTU: 9000` | Jumbo frames on the high-speed fabric. |
| `imagePullSecrets: dpf-pull-secret` | Pull Secret inserted into every DPUService spawned by the operator. |

### 3.3. `DPUCluster`

The `DPUCluster` CR defines the Kubernetes control plane that DPU nodes will join. The `interface` and `vip` fields must be customized for the environment. For more information about the available fields and their details, refer to the official DPF guide.

```yaml
---
apiVersion: provisioning.dpu.nvidia.com/v1alpha1
kind: DPUCluster
metadata:
  name: carbide-dpf-cluster
  namespace: dpf-operator-system
spec:
  type: kamaji
  maxNodes: 1000
  clusterEndpoint:
    keepalived:
      # Controller interface where the Kamaji cluster IP is configured
      interface: "REPLACE_ME"
      # External IP used by the Kamaji cluster that needs to be accessible from the DPUs
      vip: "REPLACE_ME"
      virtualRouterID: 126
      nodeSelector:
        node-role.kubernetes.io/control-plane: 'true'
```

Field-by-field:

| Field | Meaning |
|---|---|
| `type: kamaji` | Use the Kamaji cluster manager; the DPU control plane runs as a Kamaji `TenantControlPlane` in the host cluster. |
| `maxNodes: 1000` | Hard cap on DPU nodes that can join. |
| `clusterEndpoint.keepalived.interface` | Host network interface on which keepalived advertises the VIP. |
| `clusterEndpoint.keepalived.vip` | Floating IP that DPU nodes use to reach their control plane. |
| `clusterEndpoint.keepalived.virtualRouterID: 126` | VRRP ID; **must be unique per host** if multiple keepalived instances run there. |
| `nodeSelector` | Schedule keepalived only on control-plane nodes. |

### 3.4. VIP LoadBalancer Service and Endpoints

This step exposes the Kamaji cluster IP so it is routable from the DPUs. It may not be required in environments where routing to the VIP is already in place; in that case skip it.

The Service uses a fixed `loadBalancerIP` matching the VIP set in the `DPUCluster` above. Replace the `loadBalancerIP` value before applying.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: dpu-cluster-vip-loadbalancer
  namespace: dpf-operator-system
  annotations:
    metallb.io/address-pool: carbide
spec:
  allocateLoadBalancerNodePorts: true
  loadBalancerIP: "External IP used by the Kamaji cluster"
  ports:
  - port: 80
    targetPort: 80
    protocol: TCP
  type: LoadBalancer
---
apiVersion: v1
kind: Endpoints
metadata:
  name: dpu-cluster-vip-loadbalancer
  namespace: dpf-operator-system
subsets:
- addresses:
  - ip: 192.0.2.10     # dummy/test IP (RFC 5737 range)
  ports:
  - port: 80
```

What this does and why it looks unusual:

- The `Service` is type `LoadBalancer` with a fixed `loadBalancerIP` (the same VIP used by the `DPUCluster` keepalived). The `metallb.io/address-pool: carbide` annotation tells MetalLB to pull the IP from the `carbide` pool defined elsewhere.
- A **manually-created `Endpoints`** object with a single dummy RFC 5737 IP (`192.0.2.10`) is created **with the same name** as the Service. This is a Kubernetes idiom: when an `Endpoints` resource has the same name as a Service that has **no selector**, the kubelet uses those Endpoints verbatim.  Putting a dummy IP here means: *"reserve the VIP via MetalLB, but route nothing — keepalived is the actual front-end."*
- Net effect: MetalLB advertises the VIP to the network so external machines (DPUs, BMCs) can reach it, while keepalived handles the actual TCP termination.

If your environment uses a different LoadBalancer mechanism (kube-vip, a cloud-provider LB, etc.), use it to expose the VIP and point the `DPUCluster`'s `keepalived.vip` at the same address.

---

After all post-installation configuration objects are applied and reconciled
successfully, the DPF stack is ready and NICo can be started.
