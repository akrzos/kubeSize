/*
Copyright © 2021 Alex Krzos akrzos@redhat.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/yaml"
)

const (
	tableDisplay string = "table"
	jsonDisplay  string = "json"
	yamlDisplay  string = "yaml"
)

// Available = allocatable - (scheduled aka non-term pod or requests.cpu/memory)
type ClusterCapacityData struct {
	TotalNodeCount                     int
	TotalReadyNodeCount                int
	TotalUnreadyNodeCount              int
	TotalUnschedulableNodeCount        int
	TotalPodCount                      int
	TotalNonTermPodCount               int
	TotalCapacityPods                  resource.Quantity
	TotalCapacityCPU                   resource.Quantity
	TotalCapacityCPUCores              float64
	TotalCapacityMemory                resource.Quantity
	TotalCapacityMemoryGiB             float64
	TotalCapacityEphemeralStorage      resource.Quantity
	TotalCapacityEphemeralStorageGB    float64
	TotalAllocatablePods               resource.Quantity
	TotalAllocatableCPU                resource.Quantity
	TotalAllocatableCPUCores           float64
	TotalAllocatableMemory             resource.Quantity
	TotalAllocatableMemoryGiB          float64
	TotalAllocatableEphemeralStorage   resource.Quantity
	TotalAllocatableEphemeralStorageGB float64
	TotalAvailablePods                 int
	TotalRequestsCPU                   resource.Quantity
	TotalRequestsCPUCores              float64
	TotalLimitsCPU                     resource.Quantity
	TotalLimitsCPUCores                float64
	TotalAvailableCPU                  resource.Quantity
	TotalAvailableCPUCores             float64
	TotalRequestsMemory                resource.Quantity
	TotalRequestsMemoryGiB             float64
	TotalLimitsMemory                  resource.Quantity
	TotalLimitsMemoryGiB               float64
	TotalAvailableMemory               resource.Quantity
	TotalAvailableMemoryGiB            float64
	TotalRequestsEphemeralStorage      resource.Quantity
	TotalRequestsEphemeralStorageGB    float64
	TotalLimitsEphemeralStorage        resource.Quantity
	TotalLimitsEphemeralStorageGB      float64
	TotalAvailableEphemeralStorage     resource.Quantity
	TotalAvailableEphemeralStorageGB   float64
}

type ClusterSizeData struct {
	// Cluster APIs
	Namespace          int
	Node               int
	PersistentVolume   int
	ServiceAccount     int
	ClusterRole        int
	ClusterRoleBinding int
	Role               int
	RoleBinding        int
	ResourceQuota      int
	NetworkPolicy      int
	// Workloads APIs
	Container         int
	Pod               int
	ReplicaSet        int
	ReplicaController int
	Deployment        int
	Daemonset         int
	StatefulSet       int
	CronJob           int
	Job               int
	// Service APIs
	EndPoints int
	Service   int
	Ingress   int
	// Config And Storage APIs
	Configmap             int
	Secret                int
	PersistentVolumeClaim int
	StorageClass          int
	VolumeAttachment      int
	// Metadata APIs
	Event               int
	LimitRange          int
	PodDisruptionBudget int
	PodSecurityPolicy   int
}

type NodeCapacityData struct {
	TotalPodCount                      int
	TotalNonTermPodCount               int
	Roles                              sets.String
	Ready                              bool
	Schedulable                        bool
	TotalCapacityPods                  resource.Quantity
	TotalCapacityCPU                   resource.Quantity
	TotalCapacityCPUCores              float64
	TotalCapacityMemory                resource.Quantity
	TotalCapacityMemoryGiB             float64
	TotalCapacityEphemeralStorage      resource.Quantity
	TotalCapacityEphemeralStorageGB    float64
	TotalAllocatablePods               resource.Quantity
	TotalAllocatableCPU                resource.Quantity
	TotalAllocatableCPUCores           float64
	TotalAllocatableMemory             resource.Quantity
	TotalAllocatableMemoryGiB          float64
	TotalAllocatableEphemeralStorage   resource.Quantity
	TotalAllocatableEphemeralStorageGB float64
	TotalAvailablePods                 int
	TotalRequestsCPU                   resource.Quantity
	TotalRequestsCPUCores              float64
	TotalLimitsCPU                     resource.Quantity
	TotalLimitsCPUCores                float64
	TotalAvailableCPU                  resource.Quantity
	TotalAvailableCPUCores             float64
	TotalRequestsMemory                resource.Quantity
	TotalRequestsMemoryGiB             float64
	TotalLimitsMemory                  resource.Quantity
	TotalLimitsMemoryGiB               float64
	TotalAvailableMemory               resource.Quantity
	TotalAvailableMemoryGiB            float64
	TotalRequestsEphemeralStorage      resource.Quantity
	TotalRequestsEphemeralStorageGB    float64
	TotalLimitsEphemeralStorage        resource.Quantity
	TotalLimitsEphemeralStorageGB      float64
	TotalAvailableEphemeralStorage     resource.Quantity
	TotalAvailableEphemeralStorageGB   float64
}

type NamespaceCapacityData struct {
	TotalPodCount                   int
	TotalNonTermPodCount            int
	TotalUnassignedNodePodCount     int
	TotalRequestsCPU                resource.Quantity
	TotalRequestsCPUCores           float64
	TotalLimitsCPU                  resource.Quantity
	TotalLimitsCPUCores             float64
	TotalRequestsMemory             resource.Quantity
	TotalRequestsMemoryGiB          float64
	TotalLimitsMemory               resource.Quantity
	TotalLimitsMemoryGiB            float64
	TotalRequestsEphemeralStorage   resource.Quantity
	TotalRequestsEphemeralStorageGB float64
	TotalLimitsEphemeralStorage     resource.Quantity
	TotalLimitsEphemeralStorageGB   float64
}

func DisplayClusterData(clusterCapacityData ClusterCapacityData, displayDefault bool, displayHeaders bool, displayEphemeralStorage bool, displayFormat string) {
	switch displayFormat {
	case jsonDisplay:
		jsonClusterData, err := json.MarshalIndent(&clusterCapacityData, "", "  ")
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(jsonClusterData))
	case yamlDisplay:
		yamlClusterData, err := yaml.Marshal(clusterCapacityData)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Print(string(yamlClusterData))
	default:
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 5, 1, ' ', 0)
		if displayHeaders {
			if displayDefault {
				fmt.Fprintf(w, "NODES\t\t\t\tPODS\t\t\t\t\tCPU\t\t\t\t\tMEMORY\t\t\t\t\t")
				if displayEphemeralStorage {
					fmt.Fprintf(w, "EPHEMERAL STORAGE")
				}
				fmt.Fprintln(w, "")
			} else {
				fmt.Fprintf(w, "NODES\t\t\t\tPODS\t\t\t\t\tCPU (cores)\t\t\t\t\tMEMORY (GiB)\t\t\t\t\t")
				if displayEphemeralStorage {
					fmt.Fprintf(w, "EPHEMERAL STORAGE (GB)")
				}
				fmt.Fprintln(w, "")
			}
			fmt.Fprintf(w, "Total\tReady\tUnready\tUnsch\tCapacity\tAllocatable\tTotal\tNon-Term\tAvail\tCapacity\tAllocatable\tRequests\tLimits\tAvail\tCapacity\tAllocatable\tRequests\tLimits\tAvail\t")
			if displayEphemeralStorage {
				fmt.Fprintf(w, "Capacity\tAllocatable\tRequests\tLimits\tAvail")
			}
			fmt.Fprintln(w, "")
		}
		fmt.Fprintf(w, "%d\t%d\t%d\t%d\t", clusterCapacityData.TotalNodeCount, clusterCapacityData.TotalReadyNodeCount, clusterCapacityData.TotalUnreadyNodeCount, clusterCapacityData.TotalUnschedulableNodeCount)
		fmt.Fprintf(w, "%s\t%s\t", &clusterCapacityData.TotalCapacityPods, &clusterCapacityData.TotalAllocatablePods)
		fmt.Fprintf(w, "%d\t%d\t", clusterCapacityData.TotalPodCount, clusterCapacityData.TotalNonTermPodCount)
		fmt.Fprintf(w, "%d\t", clusterCapacityData.TotalAvailablePods)
		if displayDefault {
			fmt.Fprintf(w, "%s\t%s\t", &clusterCapacityData.TotalCapacityCPU, &clusterCapacityData.TotalAllocatableCPU)
			fmt.Fprintf(w, "%s\t%s\t", &clusterCapacityData.TotalRequestsCPU, &clusterCapacityData.TotalLimitsCPU)
			fmt.Fprintf(w, "%s\t", &clusterCapacityData.TotalAvailableCPU)
			fmt.Fprintf(w, "%s\t%s\t", &clusterCapacityData.TotalCapacityMemory, &clusterCapacityData.TotalAllocatableMemory)
			fmt.Fprintf(w, "%s\t%s\t", &clusterCapacityData.TotalRequestsMemory, &clusterCapacityData.TotalLimitsMemory)
			fmt.Fprintf(w, "%s\t", &clusterCapacityData.TotalAvailableMemory)
			if displayEphemeralStorage {
				fmt.Fprintf(w, "%s\t%s\t", &clusterCapacityData.TotalCapacityEphemeralStorage, &clusterCapacityData.TotalAllocatableEphemeralStorage)
				fmt.Fprintf(w, "%s\t%s\t", &clusterCapacityData.TotalRequestsEphemeralStorage, &clusterCapacityData.TotalLimitsEphemeralStorage)
				fmt.Fprintf(w, "%s\t", &clusterCapacityData.TotalAvailableEphemeralStorage)
			}
			fmt.Fprintln(w, "")
		} else {
			fmt.Fprintf(w, "%.1f\t%.1f\t", clusterCapacityData.TotalCapacityCPUCores, clusterCapacityData.TotalAllocatableCPUCores)
			fmt.Fprintf(w, "%.1f\t%.1f\t", clusterCapacityData.TotalRequestsCPUCores, clusterCapacityData.TotalLimitsCPUCores)
			fmt.Fprintf(w, "%.1f\t", clusterCapacityData.TotalAvailableCPUCores)
			fmt.Fprintf(w, "%.1f\t%.1f\t", clusterCapacityData.TotalCapacityMemoryGiB, clusterCapacityData.TotalAllocatableMemoryGiB)
			fmt.Fprintf(w, "%.1f\t%.1f\t", clusterCapacityData.TotalRequestsMemoryGiB, clusterCapacityData.TotalLimitsMemoryGiB)
			fmt.Fprintf(w, "%.1f\t", clusterCapacityData.TotalAvailableMemoryGiB)
			if displayEphemeralStorage {
				fmt.Fprintf(w, "%.1f\t%.1f\t", clusterCapacityData.TotalCapacityEphemeralStorageGB, clusterCapacityData.TotalAllocatableEphemeralStorageGB)
				fmt.Fprintf(w, "%.1f\t%.1f\t", clusterCapacityData.TotalRequestsEphemeralStorageGB, clusterCapacityData.TotalLimitsEphemeralStorageGB)
				fmt.Fprintf(w, "%.1f\t", clusterCapacityData.TotalAvailableEphemeralStorageGB)
			}
			fmt.Fprintln(w, "")
		}
		w.Flush()
	}
}

func DisplayClusterSizeData(clusterSizeData ClusterSizeData, displayHeaders bool, displayFormat string) {
	switch displayFormat {
	case jsonDisplay:
		jsonClusterData, err := json.MarshalIndent(&clusterSizeData, "", "  ")
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(jsonClusterData))
	case yamlDisplay:
		yamlClusterData, err := yaml.Marshal(clusterSizeData)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Print(string(yamlClusterData))
	default:
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 5, 1, ' ', 0)
		if displayHeaders {
			fmt.Fprintln(w, "CLUSTER APIs")
			fmt.Fprintln(w, "Namespaces\tNodes\tPersistentVolumes\tServiceAccounts\tClusterRoles\tClusterRoleBindings\tRoles\tRoleBindings\tResourceQuotas\tNetworkPolicies")
		}
		fmt.Fprintf(w, "%d\t%d\t%d\t%d\t", clusterSizeData.Namespace, clusterSizeData.Node, clusterSizeData.PersistentVolume, clusterSizeData.ServiceAccount)
		fmt.Fprintf(w, "%d\t%d\t%d\t%d\t", clusterSizeData.ClusterRole, clusterSizeData.ClusterRoleBinding, clusterSizeData.Role, clusterSizeData.RoleBinding)
		fmt.Fprintf(w, "%d\t%d\n", clusterSizeData.ResourceQuota, clusterSizeData.NetworkPolicy)
		if displayHeaders {
			fmt.Fprintln(w, "WORKLOAD APIs")
			fmt.Fprintln(w, "Containers\tPods\tReplicaSets\tReplicationControllers\tDeployments\tDaemonSets\tStatefulSets\tCronJobs\tJobs")
		}
		fmt.Fprintf(w, "%d\t%d\t%d\t%d\t", clusterSizeData.Container, clusterSizeData.Pod, clusterSizeData.ReplicaSet, clusterSizeData.ReplicaController)
		fmt.Fprintf(w, "%d\t%d\t%d\t%d\t", clusterSizeData.Deployment, clusterSizeData.Daemonset, clusterSizeData.StatefulSet, clusterSizeData.CronJob)
		fmt.Fprintf(w, "%d\n", clusterSizeData.Job)
		if displayHeaders {
			fmt.Fprintln(w, "SERVICE APIs")
			fmt.Fprintln(w, "Endpoints\tIngresses\tServices")
		}
		fmt.Fprintf(w, "%d\t%d\t%d\n", clusterSizeData.EndPoints, clusterSizeData.Ingress, clusterSizeData.Service)
		if displayHeaders {
			fmt.Fprintln(w, "CONFIG And STORAGE APIs")
			fmt.Fprintln(w, "ConfigMaps\tSecrets\tPersistentVolumeClaims\tStorageClasses\tVolumes\tVolumeAttachments")
		}
		fmt.Fprintf(w, "%d\t%d\t%d\t%d\t", clusterSizeData.Configmap, clusterSizeData.Secret, clusterSizeData.PersistentVolumeClaim, clusterSizeData.StorageClass)
		fmt.Fprintf(w, "%d\t\n", clusterSizeData.VolumeAttachment)
		if displayHeaders {
			fmt.Fprintln(w, "METADATA APIs")
			fmt.Fprintln(w, "Events\tLimitRanges\tPodDisruptionBudgets\tPodSecurityPolicies")
		}
		fmt.Fprintf(w, "%d\t%d\t%d\t%d\t\n", clusterSizeData.Event, clusterSizeData.LimitRange, clusterSizeData.PodDisruptionBudget, clusterSizeData.PodSecurityPolicy)

		w.Flush()
	}
}

func DisplayNodeRoleData(nodeRoleCapacityData map[string]*ClusterCapacityData, sortedRoleNames []string, displayDefault bool, displayHeaders bool, displayEphemeralStorage bool, displayFormat string) {
	switch displayFormat {
	case jsonDisplay:
		jsonNodeRoleData, err := json.MarshalIndent(&nodeRoleCapacityData, "", "  ")
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(jsonNodeRoleData))
	case yamlDisplay:
		yamlNodeRoleData, err := yaml.Marshal(nodeRoleCapacityData)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Print(string(yamlNodeRoleData))
	default:
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 5, 1, ' ', 0)
		if displayHeaders {
			if displayDefault {
				fmt.Fprintf(w, "ROLE\tNODES\t\t\t\tPODS\t\t\t\t\tCPU\t\t\t\t\tMEMORY\t\t\t\t\t")
				if displayEphemeralStorage {
					fmt.Fprintf(w, "EPHEMERAL STORAGE")
				}
				fmt.Fprintln(w, "")
			} else {
				fmt.Fprintf(w, "ROLE\tNODES\t\t\t\tPODS\t\t\t\t\tCPU (cores)\t\t\t\t\tMEMORY (GiB)\t\t\t\t\t")
				if displayEphemeralStorage {
					fmt.Fprintf(w, "EPHEMERAL STORAGE (GB)")
				}
				fmt.Fprintln(w, "")
			}
			fmt.Fprintf(w, "\tTotal\tReady\tUnready\tUnsch\tCapacity\tAllocatable\tTotal\tNon-Term\tAvail\tCapacity\tAllocatable\tRequests\tLimits\tAvail\tCapacity\tAllocatable\tRequests\tLimits\tAvail\t")
			if displayEphemeralStorage {
				fmt.Fprintf(w, "Capacity\tAllocatable\tRequests\tLimits\tAvail")
			}
			fmt.Fprintln(w, "")
		}
		for _, k := range sortedRoleNames {
			fmt.Fprintf(w, "%s\t", k)
			fmt.Fprintf(w, "%d\t%d\t%d\t%d\t", nodeRoleCapacityData[k].TotalNodeCount, nodeRoleCapacityData[k].TotalReadyNodeCount, nodeRoleCapacityData[k].TotalUnreadyNodeCount, nodeRoleCapacityData[k].TotalUnschedulableNodeCount)
			fmt.Fprintf(w, "%s\t%s\t", &nodeRoleCapacityData[k].TotalCapacityPods, &nodeRoleCapacityData[k].TotalAllocatablePods)
			fmt.Fprintf(w, "%d\t%d\t", nodeRoleCapacityData[k].TotalPodCount, nodeRoleCapacityData[k].TotalNonTermPodCount)
			fmt.Fprintf(w, "%d\t", nodeRoleCapacityData[k].TotalAvailablePods)
			if displayDefault {
				fmt.Fprintf(w, "%s\t%s\t", &nodeRoleCapacityData[k].TotalCapacityCPU, &nodeRoleCapacityData[k].TotalAllocatableCPU)
				fmt.Fprintf(w, "%s\t%s\t", &nodeRoleCapacityData[k].TotalRequestsCPU, &nodeRoleCapacityData[k].TotalLimitsCPU)
				fmt.Fprintf(w, "%s\t", &nodeRoleCapacityData[k].TotalAvailableCPU)
				fmt.Fprintf(w, "%s\t%s\t", &nodeRoleCapacityData[k].TotalCapacityMemory, &nodeRoleCapacityData[k].TotalAllocatableMemory)
				fmt.Fprintf(w, "%s\t%s\t", &nodeRoleCapacityData[k].TotalRequestsMemory, &nodeRoleCapacityData[k].TotalLimitsMemory)
				fmt.Fprintf(w, "%s\t", &nodeRoleCapacityData[k].TotalAvailableMemory)
				if displayEphemeralStorage {
					fmt.Fprintf(w, "%s\t%s\t", &nodeRoleCapacityData[k].TotalCapacityEphemeralStorage, &nodeRoleCapacityData[k].TotalAllocatableEphemeralStorage)
					fmt.Fprintf(w, "%s\t%s\t", &nodeRoleCapacityData[k].TotalRequestsEphemeralStorage, &nodeRoleCapacityData[k].TotalLimitsEphemeralStorage)
					fmt.Fprintf(w, "%s\t", &nodeRoleCapacityData[k].TotalAvailableEphemeralStorage)
				}
				fmt.Fprintln(w, "")
			} else {
				fmt.Fprintf(w, "%.1f\t%.1f\t", nodeRoleCapacityData[k].TotalCapacityCPUCores, nodeRoleCapacityData[k].TotalAllocatableCPUCores)
				fmt.Fprintf(w, "%.1f\t%.1f\t", nodeRoleCapacityData[k].TotalRequestsCPUCores, nodeRoleCapacityData[k].TotalLimitsCPUCores)
				fmt.Fprintf(w, "%.1f\t", nodeRoleCapacityData[k].TotalAvailableCPUCores)
				fmt.Fprintf(w, "%.1f\t%.1f\t", nodeRoleCapacityData[k].TotalCapacityMemoryGiB, nodeRoleCapacityData[k].TotalAllocatableMemoryGiB)
				fmt.Fprintf(w, "%.1f\t%.1f\t", nodeRoleCapacityData[k].TotalRequestsMemoryGiB, nodeRoleCapacityData[k].TotalLimitsMemoryGiB)
				fmt.Fprintf(w, "%.1f\t", nodeRoleCapacityData[k].TotalAvailableMemoryGiB)
				if displayEphemeralStorage {
					fmt.Fprintf(w, "%.1f\t%.1f\t", nodeRoleCapacityData[k].TotalCapacityEphemeralStorageGB, nodeRoleCapacityData[k].TotalAllocatableEphemeralStorageGB)
					fmt.Fprintf(w, "%.1f\t%.1f\t", nodeRoleCapacityData[k].TotalRequestsEphemeralStorageGB, nodeRoleCapacityData[k].TotalLimitsEphemeralStorageGB)
					fmt.Fprintf(w, "%.1f\t", nodeRoleCapacityData[k].TotalAvailableEphemeralStorageGB)
				}
				fmt.Fprintln(w, "")
			}
		}
		w.Flush()
	}
}

func DisplayNodeData(nodesCapacityData map[string]*NodeCapacityData, sortedNodeNames []string, displayDefault bool, displayHeaders bool, displayEphemeralStorage bool, displayFormat string, sortByRole bool, nodesByRole map[string][]string) {
	switch displayFormat {
	case jsonDisplay:
		jsonNodeData, err := json.MarshalIndent(&nodesCapacityData, "", "  ")
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(jsonNodeData))
	case yamlDisplay:
		yamlNodeData, err := yaml.Marshal(nodesCapacityData)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Print(string(yamlNodeData))
	default:
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 5, 1, ' ', 0)
		if displayHeaders {
			if displayDefault {
				fmt.Fprintf(w, "NAME\tSTATUS\tROLES\tPODS\t\t\t\t\tCPU\t\t\t\t\tMEMORY\t\t\t\t\t")
				if displayEphemeralStorage {
					fmt.Fprintf(w, "EPHEMERAL STORAGE")
				}
				fmt.Fprintln(w, "")
			} else {
				fmt.Fprintf(w, "NAME\tSTATUS\tROLES\tPODS\t\t\t\t\tCPU (cores)\t\t\t\t\tMEMORY (GiB)\t\t\t\t\t")
				if displayEphemeralStorage {
					fmt.Fprintf(w, "EPHEMERAL STORAGE (GB)")
				}
				fmt.Fprintln(w, "")
			}
			fmt.Fprintf(w, "\t\t\tCapacity\tAllocatable\tTotal\tNon-Term\tAvail\tCapacity\tAllocatable\tRequests\tLimits\tAvail\tCapacity\tAllocatable\tRequests\tLimits\tAvail\t")
			if displayEphemeralStorage {
				fmt.Fprintf(w, "Capacity\tAllocatable\tRequests\tLimits\tAvail")
			}
			fmt.Fprintln(w, "")
		}

		if sortByRole {
			// Sort by role
			roles := make([]string, 0, len(nodesByRole))
			for role := range nodesByRole {
				roles = append(roles, role)
			}
			sort.Strings(roles)

			for _, role := range roles {
				for _, node := range nodesByRole[role] {
					printNodeData(w, node, nodesCapacityData[node], displayDefault, displayEphemeralStorage)
				}
			}
		} else {
			// Sort by Node Name
			for _, k := range sortedNodeNames {
				printNodeData(w, k, nodesCapacityData[k], displayDefault, displayEphemeralStorage)
			}
		}

		w.Flush()
	}
}

func printNodeData(w *tabwriter.Writer, nodeName string, nodeData *NodeCapacityData, displayDefault bool, displayEphemeralStorage bool) {
	fmt.Fprintf(w, "%s\t", nodeName)
	if nodeName != "*unassigned*" && nodeName != "*total*" {
		if nodeData.Ready {
			fmt.Fprint(w, "Ready")
		} else {
			fmt.Fprint(w, "NotReady")
		}
		if !nodeData.Schedulable {
			fmt.Fprintf(w, ",Unschedulable")
		}
	}
	fmt.Fprintf(w, "\t")
	fmt.Fprintf(w, "%s\t", strings.Join(nodeData.Roles.List(), ","))
	fmt.Fprintf(w, "%s\t%s\t", &nodeData.TotalCapacityPods, &nodeData.TotalCapacityPods)
	fmt.Fprintf(w, "%d\t%d\t", nodeData.TotalPodCount, nodeData.TotalNonTermPodCount)
	fmt.Fprintf(w, "%d\t", nodeData.TotalAvailablePods)
	if displayDefault {
		fmt.Fprintf(w, "%s\t%s\t", &nodeData.TotalCapacityCPU, &nodeData.TotalAllocatableCPU)
		fmt.Fprintf(w, "%s\t%s\t", &nodeData.TotalRequestsCPU, &nodeData.TotalLimitsCPU)
		fmt.Fprintf(w, "%s\t", &nodeData.TotalAvailableCPU)
		fmt.Fprintf(w, "%s\t%s\t", &nodeData.TotalCapacityMemory, &nodeData.TotalAllocatableMemory)
		fmt.Fprintf(w, "%s\t%s\t", &nodeData.TotalRequestsMemory, &nodeData.TotalLimitsMemory)
		fmt.Fprintf(w, "%s\t", &nodeData.TotalAvailableMemory)
		if displayEphemeralStorage {
			fmt.Fprintf(w, "%s\t%s\t", &nodeData.TotalCapacityEphemeralStorage, &nodeData.TotalAllocatableEphemeralStorage)
			fmt.Fprintf(w, "%s\t%s\t", &nodeData.TotalRequestsEphemeralStorage, &nodeData.TotalLimitsEphemeralStorage)
			fmt.Fprintf(w, "%s\t", &nodeData.TotalAvailableEphemeralStorage)
		}
		fmt.Fprintln(w, "")
	} else {
		fmt.Fprintf(w, "%.1f\t%.1f\t", nodeData.TotalCapacityCPUCores, nodeData.TotalAllocatableCPUCores)
		fmt.Fprintf(w, "%.1f\t%.1f\t", nodeData.TotalRequestsCPUCores, nodeData.TotalLimitsCPUCores)
		fmt.Fprintf(w, "%.1f\t", nodeData.TotalAvailableCPUCores)
		fmt.Fprintf(w, "%.1f\t%.1f\t", nodeData.TotalCapacityMemoryGiB, nodeData.TotalAllocatableMemoryGiB)
		fmt.Fprintf(w, "%.1f\t%.1f\t", nodeData.TotalRequestsMemoryGiB, nodeData.TotalLimitsMemoryGiB)
		fmt.Fprintf(w, "%.1f\t", nodeData.TotalAvailableMemoryGiB)
		if displayEphemeralStorage {
			fmt.Fprintf(w, "%.1f\t%.1f\t", nodeData.TotalCapacityEphemeralStorageGB, nodeData.TotalAllocatableEphemeralStorageGB)
			fmt.Fprintf(w, "%.1f\t%.1f\t", nodeData.TotalRequestsEphemeralStorageGB, nodeData.TotalLimitsEphemeralStorageGB)
			fmt.Fprintf(w, "%.1f\t", nodeData.TotalAvailableEphemeralStorageGB)
		}
		fmt.Fprintln(w, "")
	}
}

func DisplayNamespaceData(namespaceCapacityData map[string]*NamespaceCapacityData, sortedNamespaceNames []string, displayDefault bool, displayHeaders bool, displayEphemeralStorage bool, displayFormat string, displayAllNamespaces bool) {
	switch displayFormat {
	case jsonDisplay:
		jsonNamespaceData, err := json.MarshalIndent(&namespaceCapacityData, "", "  ")
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(jsonNamespaceData))
	case yamlDisplay:
		yamlNamespaceData, err := yaml.Marshal(namespaceCapacityData)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Print(string(yamlNamespaceData))
	default:
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 5, 1, ' ', 0)
		if displayHeaders {
			if displayDefault {
				fmt.Fprintf(w, "NAMESPACE\tPODS\t\t\tCPU\t\tMEMORY\t\t")
				if displayEphemeralStorage {
					fmt.Fprintf(w, "EPHEMERAL STORAGE")
				}
				fmt.Fprintln(w, "")
			} else {
				fmt.Fprintf(w, "NAMESPACE\tPODS\t\t\tCPU (cores)\t\tMEMORY (GiB)\t\t")
				if displayEphemeralStorage {
					fmt.Fprintf(w, "EPHEMERAL STORAGE (GB)")
				}
				fmt.Fprintln(w, "")
			}
			fmt.Fprintf(w, "\tTotal\tNon-Term\tUnassigned\tRequests\tLimits\tRequests\tLimits\t")
			if displayEphemeralStorage {
				fmt.Fprintf(w, "Requests\tLimits")
			}
			fmt.Fprintln(w, "")
		}
		for _, k := range sortedNamespaceNames {
			if (namespaceCapacityData[k].TotalPodCount != 0) || displayAllNamespaces {
				fmt.Fprintf(w, "%s\t", k)
				fmt.Fprintf(w, "%d\t%d\t%d\t", namespaceCapacityData[k].TotalPodCount, namespaceCapacityData[k].TotalNonTermPodCount, namespaceCapacityData[k].TotalUnassignedNodePodCount)
				if displayDefault {
					fmt.Fprintf(w, "%s\t%s\t", &namespaceCapacityData[k].TotalRequestsCPU, &namespaceCapacityData[k].TotalLimitsCPU)
					fmt.Fprintf(w, "%s\t%s\t", &namespaceCapacityData[k].TotalRequestsMemory, &namespaceCapacityData[k].TotalLimitsMemory)
					if displayEphemeralStorage {
						fmt.Fprintf(w, "%s\t%s\t", &namespaceCapacityData[k].TotalRequestsEphemeralStorage, &namespaceCapacityData[k].TotalLimitsEphemeralStorage)
					}
					fmt.Fprintln(w, "")
				} else {
					fmt.Fprintf(w, "%.1f\t%.1f\t", namespaceCapacityData[k].TotalRequestsCPUCores, namespaceCapacityData[k].TotalLimitsCPUCores)
					fmt.Fprintf(w, "%.1f\t%.1f\t", namespaceCapacityData[k].TotalRequestsMemoryGiB, namespaceCapacityData[k].TotalLimitsMemoryGiB)
					if displayEphemeralStorage {
						fmt.Fprintf(w, "%.1f\t%.1f\t", namespaceCapacityData[k].TotalRequestsEphemeralStorageGB, namespaceCapacityData[k].TotalLimitsEphemeralStorageGB)
					}
					fmt.Fprintln(w, "")
				}
			}
		}
		w.Flush()
	}
}

func ValidateOutput(cmd cobra.Command) error {
	displayFormat, err := cmd.Flags().GetString("output")
	if err != nil {
		return fmt.Errorf("unable to get output display format")
	}
	validOutputs := []string{tableDisplay, jsonDisplay, yamlDisplay}
	for _, validOutputFormat := range validOutputs {
		if displayFormat == validOutputFormat {
			return nil
		}
	}
	return fmt.Errorf("Display Format \"%s\" is invalid. Valid values are %v", displayFormat, validOutputs)
}
