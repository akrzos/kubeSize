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
	TotalNodeCount              int
	TotalReadyNodeCount         int
	TotalUnreadyNodeCount       int
	TotalUnschedulableNodeCount int
	TotalPodCount               int
	TotalNonTermPodCount        int
	TotalCapacityPods           resource.Quantity
	TotalCapacityCPU            resource.Quantity
	TotalCapacityCPUCores       float64
	TotalCapacityMemory         resource.Quantity
	TotalCapacityMemoryGiB      float64
	TotalAllocatablePods        resource.Quantity
	TotalAllocatableCPU         resource.Quantity
	TotalAllocatableCPUCores    float64
	TotalAllocatableMemory      resource.Quantity
	TotalAllocatableMemoryGiB   float64
	TotalAvailablePods          int
	TotalRequestsCPU            resource.Quantity
	TotalRequestsCPUCores       float64
	TotalLimitsCPU              resource.Quantity
	TotalLimitsCPUCores         float64
	TotalAvailableCPU           resource.Quantity
	TotalAvailableCPUCores      float64
	TotalRequestsMemory         resource.Quantity
	TotalRequestsMemoryGiB      float64
	TotalLimitsMemory           resource.Quantity
	TotalLimitsMemoryGiB        float64
	TotalAvailableMemory        resource.Quantity
	TotalAvailableMemoryGiB     float64
}

type NodeCapacityData struct {
	TotalPodCount          int
	TotalNonTermPodCount   int
	Roles                  sets.String
	Ready                  bool
	Schedulable            bool
	TotalCapacityPods      resource.Quantity
	TotalCapacityCPU       resource.Quantity
	TotalCapacityMemory    resource.Quantity
	TotalAllocatablePods   resource.Quantity
	TotalAllocatableCPU    resource.Quantity
	TotalAllocatableMemory resource.Quantity
	TotalAvailablePods     int
	TotalRequestsCPU       resource.Quantity
	TotalLimitsCPU         resource.Quantity
	TotalAvailableCPU      resource.Quantity
	TotalRequestsMemory    resource.Quantity
	TotalLimitsMemory      resource.Quantity
	TotalAvailableMemory   resource.Quantity
}

type NamespaceCapacityData struct {
	TotalPodCount               int
	TotalNonTermPodCount        int
	TotalUnassignedNodePodCount int
	TotalRequestsCPU            resource.Quantity
	TotalLimitsCPU              resource.Quantity
	TotalRequestsMemory         resource.Quantity
	TotalLimitsMemory           resource.Quantity
}

func DisplayClusterData(clusterCapacityData ClusterCapacityData, displayDefault bool, displayHeaders bool, displayFormat string) {
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
				fmt.Fprintln(w, "NODES\t\t\t\tPODS\t\t\t\t\tCPU\t\t\t\t\tMEMORY\t\t\t")
			} else {
				fmt.Fprintln(w, "NODES\t\t\t\tPODS\t\t\t\t\tCPU (cores)\t\t\t\t\tMEMORY (GiB)\t\t\t")
			}
			fmt.Fprintln(w, "Total\tReady\tUnready\tUnsch\tCapacity\tAllocatable\tTotal\tNon-Term\tAvail\tCapacity\tAllocatable\tRequests\tLimits\tAvail\tCapacity\tAllocatable\tRequests\tLimits\tAvail")
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
			fmt.Fprintf(w, "%s\t\n", &clusterCapacityData.TotalAvailableMemory)
		} else {
			fmt.Fprintf(w, "%.1f\t%.1f\t", clusterCapacityData.TotalCapacityCPUCores, clusterCapacityData.TotalAllocatableCPUCores)
			fmt.Fprintf(w, "%.1f\t%.1f\t", clusterCapacityData.TotalRequestsCPUCores, clusterCapacityData.TotalLimitsCPUCores)
			fmt.Fprintf(w, "%.1f\t", clusterCapacityData.TotalAvailableCPUCores)
			fmt.Fprintf(w, "%.1f\t%.1f\t", clusterCapacityData.TotalCapacityMemoryGiB, clusterCapacityData.TotalAllocatableMemoryGiB)
			fmt.Fprintf(w, "%.1f\t%.1f\t", clusterCapacityData.TotalRequestsMemoryGiB, clusterCapacityData.TotalLimitsMemoryGiB)
			fmt.Fprintf(w, "%.1f\t\n", clusterCapacityData.TotalAvailableMemoryGiB)
		}
		w.Flush()
	}
}

func DisplayNodeRoleData(nodeRoleCapacityData map[string]*ClusterCapacityData, sortedRoleNames []string, displayDefault bool, displayHeaders bool, displayFormat string) {
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
				fmt.Fprintln(w, "ROLE\tNODES\t\t\t\tPODS\t\t\t\t\tCPU\t\t\t\t\tMEMORY\t\t\t")
			} else {
				fmt.Fprintln(w, "ROLE\tNODES\t\t\t\tPODS\t\t\t\t\tCPU (cores)\t\t\t\t\tMEMORY (GiB)\t\t\t")
			}
			fmt.Fprintln(w, "\tTotal\tReady\tUnready\tUnsch\tCapacity\tAllocatable\tTotal\tNon-Term\tAvail\tCapacity\tAllocatable\tRequests\tLimits\tAvail\tCapacity\tAllocatable\tRequests\tLimits\tAvail")
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
				fmt.Fprintf(w, "%s\t\n", &nodeRoleCapacityData[k].TotalAvailableMemory)
			} else {
				fmt.Fprintf(w, "%.1f\t%.1f\t", ReadableCPU(nodeRoleCapacityData[k].TotalCapacityCPU), ReadableCPU(nodeRoleCapacityData[k].TotalAllocatableCPU))
				fmt.Fprintf(w, "%.1f\t%.1f\t", ReadableCPU(nodeRoleCapacityData[k].TotalRequestsCPU), ReadableCPU(nodeRoleCapacityData[k].TotalLimitsCPU))
				fmt.Fprintf(w, "%.1f\t", ReadableCPU(nodeRoleCapacityData[k].TotalAvailableCPU))
				fmt.Fprintf(w, "%.1f\t%.1f\t", ReadableMem(nodeRoleCapacityData[k].TotalCapacityMemory), ReadableMem(nodeRoleCapacityData[k].TotalAllocatableMemory))
				fmt.Fprintf(w, "%.1f\t%.1f\t", ReadableMem(nodeRoleCapacityData[k].TotalRequestsMemory), ReadableMem(nodeRoleCapacityData[k].TotalLimitsMemory))
				fmt.Fprintf(w, "%.1f\t\n", ReadableMem(nodeRoleCapacityData[k].TotalAvailableMemory))
			}
		}
		w.Flush()
	}
}

func DisplayNodeData(nodesCapacityData map[string]*NodeCapacityData, sortedNodeNames []string, displayDefault bool, displayHeaders bool, displayFormat string) {
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
				fmt.Fprintln(w, "NAME\tSTATUS\tROLES\tPODS\t\t\t\t\tCPU\t\t\t\t\tMEMORY\t\t\t")
			} else {
				fmt.Fprintln(w, "NAME\tSTATUS\tROLES\tPODS\t\t\t\t\tCPU (cores)\t\t\t\t\tMEMORY (GiB)\t\t\t")
			}
			fmt.Fprintln(w, "\t\t\tCapacity\tAllocatable\tTotal\tNon-Term\tAvail\tCapacity\tAllocatable\tRequests\tLimits\tAvail\tCapacity\tAllocatable\tRequests\tLimits\tAvail")
		}
		for _, k := range sortedNodeNames {
			fmt.Fprintf(w, "%s\t", k)
			if k != "unassigned" {
				if nodesCapacityData[k].Ready {
					fmt.Fprint(w, "Ready")
				} else {
					fmt.Fprint(w, "NotReady")
				}
				if !nodesCapacityData[k].Schedulable {
					fmt.Fprintf(w, ",Unschedulable")
				}
			}
			fmt.Fprintf(w, "\t")
			fmt.Fprintf(w, "%s\t", strings.Join(nodesCapacityData[k].Roles.List(), ","))
			fmt.Fprintf(w, "%s\t%s\t", &nodesCapacityData[k].TotalCapacityPods, &nodesCapacityData[k].TotalCapacityPods)
			fmt.Fprintf(w, "%d\t%d\t", nodesCapacityData[k].TotalPodCount, nodesCapacityData[k].TotalNonTermPodCount)
			fmt.Fprintf(w, "%d\t", nodesCapacityData[k].TotalAvailablePods)
			if displayDefault {
				fmt.Fprintf(w, "%s\t%s\t", &nodesCapacityData[k].TotalCapacityCPU, &nodesCapacityData[k].TotalAllocatableCPU)
				fmt.Fprintf(w, "%s\t%s\t", &nodesCapacityData[k].TotalRequestsCPU, &nodesCapacityData[k].TotalLimitsCPU)
				fmt.Fprintf(w, "%s\t", &nodesCapacityData[k].TotalAvailableCPU)
				fmt.Fprintf(w, "%s\t%s\t", &nodesCapacityData[k].TotalCapacityMemory, &nodesCapacityData[k].TotalAllocatableMemory)
				fmt.Fprintf(w, "%s\t%s\t", &nodesCapacityData[k].TotalRequestsMemory, &nodesCapacityData[k].TotalLimitsMemory)
				fmt.Fprintf(w, "%s\t\n", &nodesCapacityData[k].TotalAvailableMemory)
			} else {
				fmt.Fprintf(w, "%.1f\t%.1f\t", ReadableCPU(nodesCapacityData[k].TotalCapacityCPU), ReadableCPU(nodesCapacityData[k].TotalAllocatableCPU))
				fmt.Fprintf(w, "%.1f\t%.1f\t", ReadableCPU(nodesCapacityData[k].TotalRequestsCPU), ReadableCPU(nodesCapacityData[k].TotalLimitsCPU))
				fmt.Fprintf(w, "%.1f\t", ReadableCPU(nodesCapacityData[k].TotalAvailableCPU))
				fmt.Fprintf(w, "%.1f\t%.1f\t", ReadableMem(nodesCapacityData[k].TotalCapacityMemory), ReadableMem(nodesCapacityData[k].TotalAllocatableMemory))
				fmt.Fprintf(w, "%.1f\t%.1f\t", ReadableMem(nodesCapacityData[k].TotalRequestsMemory), ReadableMem(nodesCapacityData[k].TotalLimitsMemory))
				fmt.Fprintf(w, "%.1f\t\n", ReadableMem(nodesCapacityData[k].TotalAvailableMemory))
			}
		}
		w.Flush()
	}
}

func DisplayNamespaceData(namespaceCapacityData map[string]*NamespaceCapacityData, sortedNamespaceNames []string, displayDefault bool, displayHeaders bool, displayFormat string, displayAllNamespaces bool) {
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
				fmt.Fprintln(w, "NAMESPACE\tPODS\t\t\tCPU\t\tMEMORY\t\t")
			} else {
				fmt.Fprintln(w, "NAMESPACE\tPODS\t\t\tCPU (cores)\t\tMEMORY (GiB)\t\t")
			}
			fmt.Fprintln(w, "\tTotal\tNon-Term\tUnassigned\tRequests\tLimits\tRequests\tLimits")
		}
		for _, k := range sortedNamespaceNames {
			if (namespaceCapacityData[k].TotalPodCount != 0) || displayAllNamespaces {
				fmt.Fprintf(w, "%s\t", k)
				fmt.Fprintf(w, "%d\t%d\t%d\t", namespaceCapacityData[k].TotalPodCount, namespaceCapacityData[k].TotalNonTermPodCount, namespaceCapacityData[k].TotalUnassignedNodePodCount)
				if displayDefault {
					fmt.Fprintf(w, "%s\t%s\t", &namespaceCapacityData[k].TotalRequestsCPU, &namespaceCapacityData[k].TotalLimitsCPU)
					fmt.Fprintf(w, "%s\t%s\t\n", &namespaceCapacityData[k].TotalRequestsMemory, &namespaceCapacityData[k].TotalLimitsMemory)
				} else {
					fmt.Fprintf(w, "%.1f\t%.1f\t", ReadableCPU(namespaceCapacityData[k].TotalRequestsCPU), ReadableCPU(namespaceCapacityData[k].TotalLimitsCPU))
					fmt.Fprintf(w, "%.1f\t%.1f\t\n", ReadableMem(namespaceCapacityData[k].TotalRequestsMemory), ReadableMem(namespaceCapacityData[k].TotalLimitsMemory))
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

func ReadableCPU(cpu resource.Quantity) float64 {
	// Convert millicores to cores
	return float64(cpu.MilliValue()) / 1000
}

func ReadableMem(mem resource.Quantity) float64 {
	// Convert from KiB to GiB
	return float64(mem.Value()) / 1024 / 1024 / 1024
}
