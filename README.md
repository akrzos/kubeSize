# kubeSize

kubeSize is a kubernetes CLI plugin to easily aggregate sizing and capacity data for a Kubernetes cluster.

## Table of Contents

- [Install](#install)
  - [Download](#download)
  - [Compile](#compile)
- [Usage](#usage)
  - [Cluster](#cluster)
  - [Node-Role](#node-role)
  - [Node](#node)
  - [Namespace](#namespace)
  - [Output formats](#output-formats)
- [License](#license)

## Install

### Download

Linux

```console
curl -L https://github.com/akrzos/kubeSize/releases/download/v0.1.5/kubeSize_0.1.5_Linux_x86_64.tar.gz | tar xvz -C /usr/local/bin kubectl-capacity
```

Mac

```console
curl -L https://github.com/akrzos/kubeSize/releases/download/v0.1.5/kubeSize_0.1.5_macOS_x86_64.tar.gz | tar xvz - -C /usr/local/bin kubectl-capacity
```

### Compile

If you have a golang environment setup, you can compile the plugin and manually install it.

```console
$ git clone https://github.com/akrzos/kubeSize.git
Cloning into 'kubeSize'...
remote: Enumerating objects: 256, done.
remote: Counting objects: 100% (256/256), done.
remote: Compressing objects: 100% (130/130), done.
remote: Total 256 (delta 130), reused 203 (delta 85), pack-reused 0
Receiving objects: 100% (256/256), 110.97 KiB | 3.70 MiB/s, done.
Resolving deltas: 100% (130/130), done.
$ cd kubeSize/
$ make bin
go fmt ./cmd/...
go vet ./cmd/...
go build -o bin/kubectl-capacity github.com/akrzos/kubeSize/
$ mv bin/kubectl-capacity /usr/local/bin/
$ kubectl capacity
```

## Usage

kubeSize is used as a kubectl plugin and run from the kubectl CLI.

```console
kubectl capacity
```

Sub-commands aggregate cluster capacity data and display it for use. All sub-commands have shortened aliases:

```console
kubectl capacity c    # cluster
kubectl capacity nr   # node-role
kubectl capacity no   # node
kubectl capacity ns   # namespace
```

### Cluster

Aggregated cluster capacity data can easily be displayed with the `cluster` sub-command.

```console
$ kubectl capacity cluster
NODES                     PODS                                      CPU (cores)                                   MEMORY (GiB)
Total Ready Unready Unsch Capacity Allocatable Total Non-Term Avail Capacity    Allocatable Requests Limits Avail Capacity     Allocatable Requests Limits Avail
3     3     0       0     330      330         13    13       317   12.0        12.0        1.1      0.3    10.9  5.8          5.8         0.3      0.5    5.5
```

Flags:

- `-e, --ephemeral-storage` flag includes ephemeral storage capacity data in table output view.

### Node-Role

Capacity data aggregated and grouped by node-role can be displayed with the `node-role` sub-command. This is helpful to see the available space to deploy an application on your kubernetes cluster by looking at the worker/compute node-role available metrics (pods, cpu, memory, storage).

```console
$ kubectl capacity node-role
ROLE   NODES                     PODS                                      CPU (cores)                                   MEMORY (GiB)
       Total Ready Unready Unsch Capacity Allocatable Total Non-Term Avail Capacity    Allocatable Requests Limits Avail Capacity     Allocatable Requests Limits Avail
<none> 2     2     0       0     220      220         7     7        213   8.0         8.0         0.4      0.2    7.6   3.9          3.9         0.2      0.4    3.6
master 1     1     0       0     110      110         6     6        104   4.0         4.0         0.7      0.1    3.4   1.9          1.9         0.0      0.0    1.9
```

Flags:

- `-e, --ephemeral-storage` flag includes ephemeral storage capacity data in table output view.
- `-u, --unassigned` flag includes a row of data on non-terminated pods that have not been assigned a node. Total counts could be confusing if looking at cluster level capacity data compared to node-role data if there are unassigned pods.

### Node

Individual node capacity data can be displayed with the `node` sub-command.

```console
$ kubectl capacity node
NAME                STATUS ROLES  PODS                                      CPU (cores)                                   MEMORY (GiB)
                                  Capacity Allocatable Total Non-Term Avail Capacity    Allocatable Requests Limits Avail Capacity     Allocatable Requests Limits Avail
3node-control-plane Ready  master 110      110         6     6        104   4.0         4.0         0.7      0.1    3.4   1.9          1.9         0.0      0.0    1.9
3node-worker        Ready  <none> 110      110         3     3        107   4.0         4.0         0.2      0.1    3.8   1.9          1.9         0.1      0.2    1.8
3node-worker2       Ready  <none> 110      110         4     4        106   4.0         4.0         0.2      0.1    3.8   1.9          1.9         0.1      0.2    1.8
```

Flags:

- `-e, --ephemeral-storage` flag includes ephemeral storage capacity data in table output view.
- `-r, --sort-by-role` flag sorts table output by node-role rather than node name.
- `-t, --display-total` flag includes a row of data displaying totals for each column.
- `-u, --unassigned` flag includes a row of data on non-terminated pods that have not been assigned a node. Total counts could be confusing if looking at cluster level capacity data compared to node data if there are unassigned pods.

### Namespace

Individual namespace capacity usage can be viewed with the `namespace` sub-command.

```console
$ kubectl capacity namespace
NAMESPACE          PODS                      CPU (cores)        MEMORY (GiB)
                   Total Non-Term Unassigned Requests    Limits Requests     Limits
kube-system        12    12       0          1.1         0.3    0.3          0.5
local-path-storage 1     1        0          0.0         0.0    0.0          0.0
```

Flags:

- `-A, --all-namespaces` flag includes namespaces with 0 pods.
- `-e, --ephemeral-storage` flag includes ephemeral storage capacity data in table output view.
- `-n, --namespace string` flag selects a specific namespace.
- `-t, --display-total` flag includes a row of data displaying totals for each column.

### Output formats

kubeSize supports table, yaml, and json output formats. Table data is the default format and is designed to be read by humans. With table output, CPU metrics default to cores, Memory metrics into [GiB (gibibyte)](https://en.wikipedia.org/wiki/Byte#Multiple-byte_units) and Storage metrics into [GB (gigabyte)](https://en.wikipedia.org/wiki/Byte#Multiple-byte_units)

Flags:

- `-o, --output string` flag allows selecting of `table|json|yaml` output formats.
- `-d, --default-format` flag uses the default format of displaying resource quantities when in table format. (Json and Yaml already include this output format)

Examples:

```console
$ kubectl capacity c
NODES                     PODS                                      CPU (cores)                                   MEMORY (GiB)
Total Ready Unready Unsch Capacity Allocatable Total Non-Term Avail Capacity    Allocatable Requests Limits Avail Capacity     Allocatable Requests Limits Avail
1     1     0       0     110      110         11    11       99    4.0         4.0         11.4     0.1    -7.5  1.9          1.9         0.4      0.4    1.6
$ kubectl capacity c -d
NODES                     PODS                                      CPU                                         MEMORY
Total Ready Unready Unsch Capacity Allocatable Total Non-Term Avail Capacity Allocatable Requests Limits Avail  Capacity  Allocatable Requests Limits Avail
1     1     0       0     110      110         11    11       99    4        4           11450m   100m   -7450m 2036452Ki 2036452Ki   400Mi    390Mi  1626852Ki
$ kubectl capacity c -o yaml
TotalAllocatableCPU: "4"
TotalAllocatableCPUCores: 4
TotalAllocatableEphemeralStorage: 61255492Ki
TotalAllocatableEphemeralStorageGB: 62.725623807999995
TotalAllocatableMemory: 2036452Ki
TotalAllocatableMemoryGiB: 1.9421119689941406
TotalAllocatablePods: "110"
TotalAvailableCPU: -7450m
TotalAvailableCPUCores: -7.45
TotalAvailableEphemeralStorage: "59620766208"
TotalAvailableEphemeralStorageGB: 59.62076620799999
TotalAvailableMemory: 1626852Ki
TotalAvailableMemoryGiB: 1.5514869689941406
TotalAvailablePods: 99
TotalCapacityCPU: "4"
TotalCapacityCPUCores: 4
TotalCapacityEphemeralStorage: 61255492Ki
TotalCapacityEphemeralStorageGB: 62.725623807999995
TotalCapacityMemory: 2036452Ki
TotalCapacityMemoryGiB: 1.9421119689941406
TotalCapacityPods: "110"
TotalLimitsCPU: 100m
TotalLimitsCPUCores: 0.1
TotalLimitsEphemeralStorage: 3G
TotalLimitsEphemeralStorageGB: 3
TotalLimitsMemory: 390Mi
TotalLimitsMemoryGiB: 0.380859375
TotalNodeCount: 1
TotalNonTermPodCount: 11
TotalPodCount: 11
TotalReadyNodeCount: 1
TotalRequestsCPU: 11450m
TotalRequestsCPUCores: 11.45
TotalRequestsEphemeralStorage: "3104857600"
TotalRequestsEphemeralStorageGB: 3.1048576000000003
TotalRequestsMemory: 400Mi
TotalRequestsMemoryGiB: 0.390625
TotalUnreadyNodeCount: 0
TotalUnschedulableNodeCount: 0
$ kubectl capacity c -o json
{
  "TotalNodeCount": 1,
  "TotalReadyNodeCount": 1,
  "TotalUnreadyNodeCount": 0,
  "TotalUnschedulableNodeCount": 0,
  "TotalPodCount": 11,
  "TotalNonTermPodCount": 11,
  "TotalCapacityPods": "110",
  "TotalCapacityCPU": "4",
  "TotalCapacityCPUCores": 4,
  "TotalCapacityMemory": "2036452Ki",
  "TotalCapacityMemoryGiB": 1.9421119689941406,
  "TotalCapacityEphemeralStorage": "61255492Ki",
  "TotalCapacityEphemeralStorageGB": 62.725623807999995,
  "TotalAllocatablePods": "110",
  "TotalAllocatableCPU": "4",
  "TotalAllocatableCPUCores": 4,
  "TotalAllocatableMemory": "2036452Ki",
  "TotalAllocatableMemoryGiB": 1.9421119689941406,
  "TotalAllocatableEphemeralStorage": "61255492Ki",
  "TotalAllocatableEphemeralStorageGB": 62.725623807999995,
  "TotalAvailablePods": 99,
  "TotalRequestsCPU": "11450m",
  "TotalRequestsCPUCores": 11.45,
  "TotalLimitsCPU": "100m",
  "TotalLimitsCPUCores": 0.1,
  "TotalAvailableCPU": "-7450m",
  "TotalAvailableCPUCores": -7.45,
  "TotalRequestsMemory": "400Mi",
  "TotalRequestsMemoryGiB": 0.390625,
  "TotalLimitsMemory": "390Mi",
  "TotalLimitsMemoryGiB": 0.380859375,
  "TotalAvailableMemory": "1626852Ki",
  "TotalAvailableMemoryGiB": 1.5514869689941406,
  "TotalRequestsEphemeralStorage": "3104857600",
  "TotalRequestsEphemeralStorageGB": 3.1048576000000003,
  "TotalLimitsEphemeralStorage": "3G",
  "TotalLimitsEphemeralStorageGB": 3,
  "TotalAvailableEphemeralStorage": "59620766208",
  "TotalAvailableEphemeralStorageGB": 59.62076620799999
}
```

## License

This project has an [Apache 2.0 license](LICENSE).
