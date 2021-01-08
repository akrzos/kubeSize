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

type ClusterCapacityData struct {
	TotalNodeCount              int
	TotalReadyNodeCount         int
	TotalUnreadyNodeCount       int
	TotalUnschedulableNodeCount int
	TotalPodCount               int
	TotalNonTermPodCount        int
	TotalCapacityPods           resource.Quantity
	TotalCapacityCPU            resource.Quantity
	TotalCapacityMemory         resource.Quantity
	TotalAllocatablePods        resource.Quantity
	TotalAllocatableCPU         resource.Quantity
	TotalAllocatableMemory      resource.Quantity
	TotalRequestsCPU            resource.Quantity
	TotalLimitsCPU              resource.Quantity
	TotalRequestsMemory         resource.Quantity
	TotalLimitsMemory           resource.Quantity
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
	TotalRequestsCPU       resource.Quantity
	TotalLimitsCPU         resource.Quantity
	TotalRequestsMemory    resource.Quantity
	TotalLimitsMemory      resource.Quantity
}

type NamespaceCapacityData struct {
	TotalPodCount        int
	TotalNonTermPodCount int
	TotalRequestsCPU     resource.Quantity
	TotalLimitsCPU       resource.Quantity
	TotalRequestsMemory  resource.Quantity
	TotalLimitsMemory    resource.Quantity
}

func DisplayClusterData(clusterCapacityData ClusterCapacityData, displayReadable bool, displayFormat string) {
	if displayFormat == "table" {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 5, 1, ' ', 0)
		if displayReadable == true {
			fmt.Fprintln(w, "NODES\t\t\t\tPODS\t\t\t\tCPU (cores)\t\t\t\tMEMORY (GiB)\t\t")
		} else {
			fmt.Fprintln(w, "NODES\t\t\t\tPODS\t\t\t\tCPU\t\t\t\tMEMORY\t\t")
		}
		fmt.Fprintln(w, "Total\tReady\tUnready\tUnschedulable\tCapacity\tAllocatable\tTotal\tNon-Term\tCapacity\tAllocatable\tRequests\tLimits\tCapacity\tAllocatable\tRequests\tLimits")
		fmt.Fprintf(w, "%d\t%d\t%d\t%d\t", clusterCapacityData.TotalNodeCount, clusterCapacityData.TotalReadyNodeCount, clusterCapacityData.TotalUnreadyNodeCount, clusterCapacityData.TotalUnschedulableNodeCount)
		fmt.Fprintf(w, "%s\t%s\t", &clusterCapacityData.TotalCapacityPods, &clusterCapacityData.TotalAllocatablePods)
		fmt.Fprintf(w, "%d\t%d\t", clusterCapacityData.TotalPodCount, clusterCapacityData.TotalNonTermPodCount)
		if displayReadable == true {
			fmt.Fprintf(w, "%.1f\t%.1f\t", readableCPU(clusterCapacityData.TotalCapacityCPU), readableCPU(clusterCapacityData.TotalAllocatableCPU))
			fmt.Fprintf(w, "%.1f\t%.1f\t", readableCPU(clusterCapacityData.TotalRequestsCPU), readableCPU(clusterCapacityData.TotalLimitsCPU))
			fmt.Fprintf(w, "%.1f\t%.1f\t", readableMem(clusterCapacityData.TotalCapacityMemory), readableMem(clusterCapacityData.TotalAllocatableMemory))
			fmt.Fprintf(w, "%.1f\t%.1f\t\n", readableMem(clusterCapacityData.TotalRequestsMemory), readableMem(clusterCapacityData.TotalLimitsMemory))
		} else {
			fmt.Fprintf(w, "%s\t%s\t", &clusterCapacityData.TotalCapacityCPU, &clusterCapacityData.TotalAllocatableCPU)
			fmt.Fprintf(w, "%s\t%s\t", &clusterCapacityData.TotalRequestsCPU, &clusterCapacityData.TotalLimitsCPU)
			fmt.Fprintf(w, "%s\t%s\t", &clusterCapacityData.TotalCapacityMemory, &clusterCapacityData.TotalAllocatableMemory)
			fmt.Fprintf(w, "%s\t%s\t\n", &clusterCapacityData.TotalRequestsMemory, &clusterCapacityData.TotalLimitsMemory)
		}
		w.Flush()
	} else if displayFormat == "json" {
		jsonClusterData, err := json.MarshalIndent(&clusterCapacityData, "", "  ")
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(jsonClusterData))
	} else if displayFormat == "yaml" {
		yamlClusterData, err := yaml.Marshal(clusterCapacityData)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf(string(yamlClusterData))
	}
}

func DisplayNodeRoleData(nodeRoleCapacityData map[string]*ClusterCapacityData, sortedRoleNames []string, displayReadable bool, displayFormat string) {
	if displayFormat == "table" {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 5, 1, ' ', 0)
		if displayReadable == true {
			fmt.Fprintln(w, "ROLE\tNODES\t\t\t\tPODS\t\t\t\tCPU (cores)\t\t\t\tMEMORY (GiB)\t\t")
		} else {
			fmt.Fprintln(w, "ROLE\tNODES\t\t\t\tPODS\t\t\t\tCPU\t\t\t\tMEMORY\t\t")
		}
		fmt.Fprintln(w, "\tTotal\tReady\tUnready\tUnschedulable\tCapacity\tAllocatable\tTotal\tNon-Term\tCapacity\tAllocatable\tRequests\tLimits\tCapacity\tAllocatable\tRequests\tLimits")

		for _, k := range sortedRoleNames {
			fmt.Fprintf(w, "%s\t", k)
			fmt.Fprintf(w, "%d\t%d\t%d\t%d\t", nodeRoleCapacityData[k].TotalNodeCount, nodeRoleCapacityData[k].TotalReadyNodeCount, nodeRoleCapacityData[k].TotalUnreadyNodeCount, nodeRoleCapacityData[k].TotalUnschedulableNodeCount)
			fmt.Fprintf(w, "%s\t%s\t", &nodeRoleCapacityData[k].TotalCapacityPods, &nodeRoleCapacityData[k].TotalAllocatablePods)
			fmt.Fprintf(w, "%d\t%d\t", nodeRoleCapacityData[k].TotalPodCount, nodeRoleCapacityData[k].TotalNonTermPodCount)
			if displayReadable == true {
				fmt.Fprintf(w, "%.1f\t%.1f\t", readableCPU(nodeRoleCapacityData[k].TotalCapacityCPU), readableCPU(nodeRoleCapacityData[k].TotalAllocatableCPU))
				fmt.Fprintf(w, "%.1f\t%.1f\t", readableCPU(nodeRoleCapacityData[k].TotalRequestsCPU), readableCPU(nodeRoleCapacityData[k].TotalLimitsCPU))
				fmt.Fprintf(w, "%.1f\t%.1f\t", readableMem(nodeRoleCapacityData[k].TotalCapacityMemory), readableMem(nodeRoleCapacityData[k].TotalAllocatableMemory))
				fmt.Fprintf(w, "%.1f\t%.1f\t\n", readableMem(nodeRoleCapacityData[k].TotalRequestsMemory), readableMem(nodeRoleCapacityData[k].TotalLimitsMemory))
			} else {
				fmt.Fprintf(w, "%s\t%s\t", &nodeRoleCapacityData[k].TotalCapacityCPU, &nodeRoleCapacityData[k].TotalAllocatableCPU)
				fmt.Fprintf(w, "%s\t%s\t", &nodeRoleCapacityData[k].TotalRequestsCPU, &nodeRoleCapacityData[k].TotalLimitsCPU)
				fmt.Fprintf(w, "%s\t%s\t", &nodeRoleCapacityData[k].TotalCapacityMemory, &nodeRoleCapacityData[k].TotalAllocatableMemory)
				fmt.Fprintf(w, "%s\t%s\t\n", &nodeRoleCapacityData[k].TotalRequestsMemory, &nodeRoleCapacityData[k].TotalLimitsMemory)
			}
		}
		w.Flush()
	} else if displayFormat == "json" {
		jsonNodeRoleData, err := json.MarshalIndent(&nodeRoleCapacityData, "", "  ")
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(jsonNodeRoleData))
	} else if displayFormat == "yaml" {
		yamlNodeRoleData, err := yaml.Marshal(nodeRoleCapacityData)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf(string(yamlNodeRoleData))
	}
}

func DisplayNodeData(nodesCapacityData map[string]*NodeCapacityData, sortedNodeNames []string, displayReadable bool, displayFormat string) {
	if displayFormat == "table" {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 5, 1, ' ', 0)
		if displayReadable == true {
			fmt.Fprintln(w, "NAME\tSTATUS\tROLES\tPODS\t\t\t\tCPU (cores)\t\t\t\tMEMORY (GiB)\t\t")
		} else {
			fmt.Fprintln(w, "NAME\tSTATUS\tROLES\tPODS\t\t\t\tCPU\t\t\t\tMEMORY\t\t")
		}
		fmt.Fprintln(w, "\t\t\tCapacity\tAllocatable\tTotal\tNon-Term\tCapacity\tAllocatable\tRequests\tLimits\tCapacity\tAllocatable\tRequests\tLimits")

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
			if displayReadable == true {
				fmt.Fprintf(w, "%.1f\t%.1f\t", readableCPU(nodesCapacityData[k].TotalCapacityCPU), readableCPU(nodesCapacityData[k].TotalAllocatableCPU))
				fmt.Fprintf(w, "%.1f\t%.1f\t", readableCPU(nodesCapacityData[k].TotalRequestsCPU), readableCPU(nodesCapacityData[k].TotalLimitsCPU))
				fmt.Fprintf(w, "%.1f\t%.1f\t", readableMem(nodesCapacityData[k].TotalCapacityMemory), readableMem(nodesCapacityData[k].TotalAllocatableMemory))
				fmt.Fprintf(w, "%.1f\t%.1f\t\n", readableMem(nodesCapacityData[k].TotalRequestsMemory), readableMem(nodesCapacityData[k].TotalLimitsMemory))
			} else {
				fmt.Fprintf(w, "%s\t%s\t", &nodesCapacityData[k].TotalCapacityCPU, &nodesCapacityData[k].TotalAllocatableCPU)
				fmt.Fprintf(w, "%s\t%s\t", &nodesCapacityData[k].TotalRequestsCPU, &nodesCapacityData[k].TotalLimitsCPU)
				fmt.Fprintf(w, "%s\t%s\t", &nodesCapacityData[k].TotalCapacityMemory, &nodesCapacityData[k].TotalAllocatableMemory)
				fmt.Fprintf(w, "%s\t%s\t\n", &nodesCapacityData[k].TotalRequestsMemory, &nodesCapacityData[k].TotalLimitsMemory)
			}
		}
		w.Flush()
	} else if displayFormat == "json" {
		jsonNodeData, err := json.MarshalIndent(&nodesCapacityData, "", "  ")
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(jsonNodeData))
	} else if displayFormat == "yaml" {
		yamlNodeData, err := yaml.Marshal(nodesCapacityData)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf(string(yamlNodeData))
	}
}

func DisplayNamespaceData(namespaceCapacityData map[string]*NamespaceCapacityData, sortedNamespaceNames []string, displayReadable bool, displayFormat string, displayAllNamespaces bool) {
	if displayFormat == "table" {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 5, 1, ' ', 0)
		if displayReadable == true {
			fmt.Fprintln(w, "NAMESPACE\tPODS\t\tCPU (cores)\t\tMEMORY (GiB)\t\t")
		} else {
			fmt.Fprintln(w, "NAMESPACE\tPODS\t\tCPU\t\tMEMORY\t\t")
		}
		fmt.Fprintln(w, "\tTotal\tNon-Term\tRequests\tLimits\tRequests\tLimits")
		for _, k := range sortedNamespaceNames {
			if (namespaceCapacityData[k].TotalPodCount != 0) || displayAllNamespaces {
				fmt.Fprintf(w, "%s\t", k)
				fmt.Fprintf(w, "%d\t%d\t", namespaceCapacityData[k].TotalPodCount, namespaceCapacityData[k].TotalNonTermPodCount)
				if displayReadable == true {
					fmt.Fprintf(w, "%.1f\t%.1f\t", readableCPU(namespaceCapacityData[k].TotalRequestsCPU), readableCPU(namespaceCapacityData[k].TotalLimitsCPU))
					fmt.Fprintf(w, "%.1f\t%.1f\t\n", readableMem(namespaceCapacityData[k].TotalRequestsMemory), readableMem(namespaceCapacityData[k].TotalLimitsMemory))
				} else {
					fmt.Fprintf(w, "%s\t%s\t", &namespaceCapacityData[k].TotalRequestsCPU, &namespaceCapacityData[k].TotalLimitsCPU)
					fmt.Fprintf(w, "%s\t%s\t\n", &namespaceCapacityData[k].TotalRequestsMemory, &namespaceCapacityData[k].TotalLimitsMemory)
				}
			}
		}
		w.Flush()
	} else if displayFormat == "json" {
		jsonNamespaceData, err := json.MarshalIndent(&namespaceCapacityData, "", "  ")
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(jsonNamespaceData))
	} else if displayFormat == "yaml" {
		yamlNamespaceData, err := yaml.Marshal(namespaceCapacityData)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf(string(yamlNamespaceData))
	}
}

func ValidateOutput(cmd cobra.Command) error {
	displayFormat, _ := cmd.Flags().GetString("output")
	validOutputs := []string{"table", "json", "yaml"}
	for _, validOutputFormat := range validOutputs {
		if displayFormat == validOutputFormat {
			return nil
		}
	}
	return fmt.Errorf("Display Format \"%s\" is invalid. Valid values are %v", displayFormat, validOutputs)
}

func readableCPU(cpu resource.Quantity) float64 {
	// Convert millicores to cores
	return float64(cpu.MilliValue()) / 1000
}

func readableMem(mem resource.Quantity) float64 {
	// Convert from KiB to GiB
	return float64(mem.Value()) / 1024 / 1024 / 1024
}
