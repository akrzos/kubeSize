# kubeSize

kubeSize is a kubernetes CLI plugin to easily aggregate sizing and capacity data for a Kubernetes cluster.

## Table of Contents

- [Install](#install)
- [Usage](#usage)
- [License](#license)

## Install

### Krew install

If you have [krew](https://krew.sigs.k8s.io/) installed, you can install this plugin with the krew plugin yaml.

```console
$ git clone https://github.com/akrzos/kubeSize.git
Cloning into 'kubeSize'...
remote: Enumerating objects: 256, done.
remote: Counting objects: 100% (256/256), done.
remote: Compressing objects: 100% (130/130), done.
remote: Total 256 (delta 130), reused 203 (delta 85), pack-reused 0
Receiving objects: 100% (256/256), 110.97 KiB | 3.83 MiB/s, done.
Resolving deltas: 100% (130/130), done.
$ kubectl krew install --manifest=kubeSize/deploy/krew/capacity.yaml
Installing plugin: capacity
Installed plugin: capacity
\
 | Use this plugin:
 | 	kubectl capacity
 | Documentation:
 | 	https://github.com/akrzos/kubeSize
/
$ kubectl capacity
```

### Compile and install

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
$ kubectl capacity
```

Sub-commands aggregate cluster capacity data and display it for use. Flags can modify what data is collected or the format of the output.

### Cluster Capacity

Aggregated cluster capacity data can easily be displayed with the `cluster` sub-command.

```console
$ kubectl capacity cluster
NODES                     PODS                                      CPU (cores)                                   MEMORY (GiB)
Total Ready Unready Unsch Capacity Allocatable Total Non-Term Avail Capacity    Allocatable Requests Limits Avail Capacity     Allocatable Requests Limits Avail
3     3     0       0     330      330         13    13       317   12.0        12.0        1.1      0.3    10.9  5.8          5.8         0.3      0.5    5.5
```

### Node-Role Capacity

Capacity data aggregated and grouped by node-role can be displayed with the `node-role` sub-command. This is helpful to see the available space to deploy an application on your kubernetes cluster by looking at the worker or compute node's available metrics (pods, cpu, memory).

```console
$ kubectl capacity node-role
ROLE   NODES                     PODS                                      CPU (cores)                                   MEMORY (GiB)
       Total Ready Unready Unsch Capacity Allocatable Total Non-Term Avail Capacity    Allocatable Requests Limits Avail Capacity     Allocatable Requests Limits Avail
<none> 2     2     0       0     220      220         7     7        213   8.0         8.0         0.4      0.2    7.6   3.9          3.9         0.2      0.4    3.6
master 1     1     0       0     110      110         6     6        104   4.0         4.0         0.7      0.1    3.4   1.9          1.9         0.0      0.0    1.9
```

Use the `--unassigned` or `-u` flag to include data on non-terminated pods that have not been assigned to a node that could be confused if looking at cluster level capacity data.

### Node Capacity

Individual node capacity data can be displayed with the `node` sub-command.

```console
$ kubectl capacity node
NAME                STATUS ROLES  PODS                                      CPU (cores)                                   MEMORY (GiB)
                                  Capacity Allocatable Total Non-Term Avail Capacity    Allocatable Requests Limits Avail Capacity     Allocatable Requests Limits Avail
3node-control-plane Ready  master 110      110         6     6        104   4.0         4.0         0.7      0.1    3.4   1.9          1.9         0.0      0.0    1.9
3node-worker        Ready  <none> 110      110         3     3        107   4.0         4.0         0.2      0.1    3.8   1.9          1.9         0.1      0.2    1.8
3node-worker2       Ready  <none> 110      110         4     4        106   4.0         4.0         0.2      0.1    3.8   1.9          1.9         0.1      0.2    1.8
```

Use the `--sort-by-role` or `-r` flag to sort by node-role rather than node name in the event your nodes are named randomly. Use the `--unassigned` or `-u` flag to include data on non-terminated pods that have not been assigned to a node that could be confused if looking at cluster level capacity data.

### Namespace Capacity

Individual namespace capacity usage can be viewed with the `namespace` sub-command.

```console
$ kubectl capacity namespace
NAMESPACE          PODS                      CPU (cores)        MEMORY (GiB)
                   Total Non-Term Unassigned Requests    Limits Requests     Limits
kube-system        12    12       0          1.1         0.3    0.3          0.5
local-path-storage 1     1        0          0.0         0.0    0.0          0.0
```

Use the `--all-namespaces` or `-A` to view all namespaces including those without any pods. Use the `--namespace` or `-n` flag to select a specific namespace.

### Shortened command names

All sub-commands have shortened aliases:

```console
$ kubectl capacity c    # cluster
$ kubectl capacity nr   # node-role
$ kubectl capacity no   # node
$ kubectl capacity ns   # namespace
```

### Output formats

kubeSize supports table, yaml, and json output formats. Table data is the default format and is designed to be read by humans. With table output, CPU metrics default to cores and Memory metrics into [GiB (gibibyte)](https://en.wikipedia.org/wiki/Byte#Multiple-byte_units).

Use the output flag to change output formats:

```console
$ kubectl capacity c
NODES                     PODS                                      CPU (cores)                                   MEMORY (GiB)
Total Ready Unready Unsch Capacity Allocatable Total Non-Term Avail Capacity    Allocatable Requests Limits Avail Capacity     Allocatable Requests Limits Avail
1     1     0       0     110      110         9     9        101   4.0         4.0         0.8      0.1    3.1   1.9          1.9         0.2      0.4    1.8
$ kubectl capacity c -o yaml
TotalAllocatableCPU: "4"
TotalAllocatableCPUCores: 4
TotalAllocatableMemory: 2036452Ki
TotalAllocatableMemoryGiB: 1.9421119689941406
TotalAllocatablePods: "110"
TotalAvailableCPU: 3150m
TotalAvailableCPUCores: 3.15
TotalAvailableMemory: 1841892Ki
TotalAvailableMemoryGiB: 1.7565650939941406
TotalAvailablePods: 101
TotalCapacityCPU: "4"
TotalCapacityCPUCores: 4
TotalCapacityMemory: 2036452Ki
TotalCapacityMemoryGiB: 1.9421119689941406
TotalCapacityPods: "110"
TotalLimitsCPU: 100m
TotalLimitsCPUCores: 0.1
TotalLimitsMemory: 390Mi
TotalLimitsMemoryGiB: 0.380859375
TotalNodeCount: 1
TotalNonTermPodCount: 9
TotalPodCount: 9
TotalReadyNodeCount: 1
TotalRequestsCPU: 850m
TotalRequestsCPUCores: 0.85
TotalRequestsMemory: 190Mi
TotalRequestsMemoryGiB: 0.185546875
TotalUnreadyNodeCount: 0
TotalUnschedulableNodeCount: 0
$ kubectl capacity c -o json
{
  "TotalNodeCount": 1,
  "TotalReadyNodeCount": 1,
  "TotalUnreadyNodeCount": 0,
  "TotalUnschedulableNodeCount": 0,
  "TotalPodCount": 9,
  "TotalNonTermPodCount": 9,
  "TotalCapacityPods": "110",
  "TotalCapacityCPU": "4",
  "TotalCapacityCPUCores": 4,
  "TotalCapacityMemory": "2036452Ki",
  "TotalCapacityMemoryGiB": 1.9421119689941406,
  "TotalAllocatablePods": "110",
  "TotalAllocatableCPU": "4",
  "TotalAllocatableCPUCores": 4,
  "TotalAllocatableMemory": "2036452Ki",
  "TotalAllocatableMemoryGiB": 1.9421119689941406,
  "TotalAvailablePods": 101,
  "TotalRequestsCPU": "850m",
  "TotalRequestsCPUCores": 0.85,
  "TotalLimitsCPU": "100m",
  "TotalLimitsCPUCores": 0.1,
  "TotalAvailableCPU": "3150m",
  "TotalAvailableCPUCores": 3.15,
  "TotalRequestsMemory": "190Mi",
  "TotalRequestsMemoryGiB": 0.185546875,
  "TotalLimitsMemory": "390Mi",
  "TotalLimitsMemoryGiB": 0.380859375,
  "TotalAvailableMemory": "1841892Ki",
  "TotalAvailableMemoryGiB": 1.7565650939941406
}
```

Note that yaml and json formats automatically include the "human" Cores and GiB formats for each metric.

## License

This project has an [Apache 2.0 license](LICENSE).
