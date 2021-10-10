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
package capacity

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/akrzos/kubeSize/internal/capacity"
	"github.com/akrzos/kubeSize/internal/kube"
	"github.com/akrzos/kubeSize/internal/output"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

var nodeCmd = &cobra.Command{
	Use:     "node",
	Aliases: []string{"no"},
	Short:   "Get individual node capacity",
	Long:    `Get metrics and data related to node capacity`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if err := output.ValidateOutput(*cmd); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		clientset, err := kube.CreateClientSet(KubernetesConfigFlags)
		if err != nil {
			return errors.Wrap(err, "failed to create clientset")
		}

		nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list nodes")
		}

		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list pods")
		}

		nodesCapacityData := make(map[string]*output.NodeCapacityData)
		nodeNames := make([]string, 0, len(nodes.Items))
		nodesByRole := make(map[string][]string)

		for _, node := range nodes.Items {
			nodeNames = append(nodeNames, node.Name)
			nodesCapacityData[node.Name] = new(output.NodeCapacityData)

			roles := sets.NewString()
			for labelKey, labelValue := range node.Labels {
				switch {
				case strings.HasPrefix(labelKey, "node-role.kubernetes.io/"):
					if role := strings.TrimPrefix(labelKey, "node-role.kubernetes.io/"); len(role) > 0 {
						roles.Insert(role)
					}
				case labelKey == "kubernetes.io/role" && labelValue != "":
					roles.Insert(labelValue)
				}
			}
			if len(roles) == 0 {
				roles.Insert("<none>")
			}

			nodesCapacityData[node.Name].Ready = false
			for _, condition := range node.Status.Conditions {
				if (condition.Type == "Ready") && condition.Status == corev1.ConditionTrue {
					nodesCapacityData[node.Name].Ready = true
					break
				}
			}

			nodesCapacityData[node.Name].Schedulable = !node.Spec.Unschedulable
			nodesCapacityData[node.Name].Roles = roles
			nodesCapacityData[node.Name].TotalCapacityPods.Add(*node.Status.Capacity.Pods())
			nodesCapacityData[node.Name].TotalCapacityCPU.Add(*node.Status.Capacity.Cpu())
			nodesCapacityData[node.Name].TotalCapacityMemory.Add(*node.Status.Capacity.Memory())
			nodesCapacityData[node.Name].TotalCapacityEphemeralStorage.Add(*node.Status.Capacity.StorageEphemeral())
			nodesCapacityData[node.Name].TotalAllocatablePods.Add(*node.Status.Allocatable.Pods())
			nodesCapacityData[node.Name].TotalAllocatableCPU.Add(*node.Status.Allocatable.Cpu())
			nodesCapacityData[node.Name].TotalAllocatableMemory.Add(*node.Status.Allocatable.Memory())
			nodesCapacityData[node.Name].TotalAllocatableEphemeralStorage.Add(*node.Status.Allocatable.StorageEphemeral())
			rolesIndex := strings.Join(roles.List(), ",")
			nodesByRole[rolesIndex] = append(nodesByRole[rolesIndex], node.Name)
		}
		nodesCapacityData["*unassigned*"] = new(output.NodeCapacityData)
		nodesCapacityData["*total*"] = new(output.NodeCapacityData)

		for _, pod := range pods.Items {
			podNode := pod.Spec.NodeName
			if pod.Spec.NodeName == "" {
				podNode = "*unassigned*"
			}
			nodesCapacityData[podNode].TotalPodCount++

			if (pod.Status.Phase != corev1.PodSucceeded) && (pod.Status.Phase != corev1.PodFailed) {
				nodesCapacityData[podNode].TotalNonTermPodCount++
				for _, container := range pod.Spec.Containers {
					nodesCapacityData[podNode].TotalRequestsCPU.Add(*container.Resources.Requests.Cpu())
					nodesCapacityData[podNode].TotalLimitsCPU.Add(*container.Resources.Limits.Cpu())
					nodesCapacityData[podNode].TotalRequestsMemory.Add(*container.Resources.Requests.Memory())
					nodesCapacityData[podNode].TotalLimitsMemory.Add(*container.Resources.Limits.Memory())
					nodesCapacityData[podNode].TotalRequestsEphemeralStorage.Add(*container.Resources.Requests.StorageEphemeral())
					nodesCapacityData[podNode].TotalLimitsEphemeralStorage.Add(*container.Resources.Limits.StorageEphemeral())
				}
			}
		}

		for _, node := range nodeNames {
			nodesCapacityData[node].TotalAvailablePods = int(nodesCapacityData[node].TotalAllocatablePods.Value()) - nodesCapacityData[node].TotalNonTermPodCount
			nodesCapacityData[node].TotalAvailableCPU = nodesCapacityData[node].TotalAllocatableCPU
			nodesCapacityData[node].TotalAvailableCPU.Sub(nodesCapacityData[node].TotalRequestsCPU)
			nodesCapacityData[node].TotalAvailableMemory = nodesCapacityData[node].TotalAllocatableMemory
			nodesCapacityData[node].TotalAvailableMemory.Sub(nodesCapacityData[node].TotalRequestsMemory)
			nodesCapacityData[node].TotalAvailableEphemeralStorage = nodesCapacityData[node].TotalAllocatableEphemeralStorage
			nodesCapacityData[node].TotalAvailableEphemeralStorage.Sub(nodesCapacityData[node].TotalRequestsEphemeralStorage)
		}

		displayDefault, _ := cmd.Flags().GetBool("default-format")

		displayEphemeralStorage, _ := cmd.Flags().GetBool("ephemeral-storage")

		displayNoHeaders, _ := cmd.Flags().GetBool("no-headers")

		displayFormat, _ := cmd.Flags().GetString("output")

		sort.Strings(nodeNames)
		if displayUnassigned, _ := cmd.Flags().GetBool("unassigned"); displayUnassigned {
			nodeNames = append(nodeNames, "*unassigned*")
			nodesByRole["~"] = append(nodesByRole["~"], "*unassigned*")
		}

		// Populate "Human" readable capacity data values and the *total* "node"
		for _, node := range nodeNames {
			nodesCapacityData[node].TotalCapacityCPUCores = capacity.ReadableCPU(nodesCapacityData[node].TotalCapacityCPU)
			nodesCapacityData[node].TotalCapacityMemoryGiB = capacity.ReadableMem(nodesCapacityData[node].TotalCapacityMemory)
			nodesCapacityData[node].TotalCapacityEphemeralStorageGB = capacity.ReadableStorage(nodesCapacityData[node].TotalCapacityEphemeralStorage)
			nodesCapacityData[node].TotalAllocatableCPUCores = capacity.ReadableCPU(nodesCapacityData[node].TotalAllocatableCPU)
			nodesCapacityData[node].TotalAllocatableMemoryGiB = capacity.ReadableMem(nodesCapacityData[node].TotalAllocatableMemory)
			nodesCapacityData[node].TotalAllocatableEphemeralStorageGB = capacity.ReadableStorage(nodesCapacityData[node].TotalAllocatableEphemeralStorage)
			nodesCapacityData[node].TotalRequestsCPUCores = capacity.ReadableCPU(nodesCapacityData[node].TotalRequestsCPU)
			nodesCapacityData[node].TotalLimitsCPUCores = capacity.ReadableCPU(nodesCapacityData[node].TotalLimitsCPU)
			nodesCapacityData[node].TotalAvailableCPUCores = capacity.ReadableCPU(nodesCapacityData[node].TotalAvailableCPU)
			nodesCapacityData[node].TotalRequestsMemoryGiB = capacity.ReadableMem(nodesCapacityData[node].TotalRequestsMemory)
			nodesCapacityData[node].TotalLimitsMemoryGiB = capacity.ReadableMem(nodesCapacityData[node].TotalLimitsMemory)
			nodesCapacityData[node].TotalAvailableMemoryGiB = capacity.ReadableMem(nodesCapacityData[node].TotalAvailableMemory)
			nodesCapacityData[node].TotalRequestsEphemeralStorageGB = capacity.ReadableStorage(nodesCapacityData[node].TotalRequestsEphemeralStorage)
			nodesCapacityData[node].TotalLimitsEphemeralStorageGB = capacity.ReadableStorage(nodesCapacityData[node].TotalLimitsEphemeralStorage)
			nodesCapacityData[node].TotalAvailableEphemeralStorageGB = capacity.ReadableStorage(nodesCapacityData[node].TotalAvailableEphemeralStorage)
			nodesCapacityData["*total*"].TotalPodCount += nodesCapacityData[node].TotalPodCount
			nodesCapacityData["*total*"].TotalNonTermPodCount += nodesCapacityData[node].TotalNonTermPodCount
			nodesCapacityData["*total*"].TotalCapacityPods.Add(nodesCapacityData[node].TotalCapacityPods)
			nodesCapacityData["*total*"].TotalCapacityCPU.Add(nodesCapacityData[node].TotalCapacityCPU)
			nodesCapacityData["*total*"].TotalCapacityCPUCores += nodesCapacityData[node].TotalCapacityCPUCores
			nodesCapacityData["*total*"].TotalCapacityMemory.Add(nodesCapacityData[node].TotalCapacityMemory)
			nodesCapacityData["*total*"].TotalCapacityMemoryGiB += nodesCapacityData[node].TotalCapacityMemoryGiB
			nodesCapacityData["*total*"].TotalCapacityEphemeralStorage.Add(nodesCapacityData[node].TotalCapacityEphemeralStorage)
			nodesCapacityData["*total*"].TotalCapacityEphemeralStorageGB += nodesCapacityData[node].TotalCapacityEphemeralStorageGB
			nodesCapacityData["*total*"].TotalAllocatablePods.Add(nodesCapacityData[node].TotalAllocatablePods)
			nodesCapacityData["*total*"].TotalAllocatableCPU.Add(nodesCapacityData[node].TotalAllocatableCPU)
			nodesCapacityData["*total*"].TotalAllocatableCPUCores += nodesCapacityData[node].TotalAllocatableCPUCores
			nodesCapacityData["*total*"].TotalAllocatableMemory.Add(nodesCapacityData[node].TotalAllocatableMemory)
			nodesCapacityData["*total*"].TotalAllocatableMemoryGiB += nodesCapacityData[node].TotalAllocatableMemoryGiB
			nodesCapacityData["*total*"].TotalAllocatableEphemeralStorage.Add(nodesCapacityData[node].TotalAllocatableEphemeralStorage)
			nodesCapacityData["*total*"].TotalAllocatableEphemeralStorageGB += nodesCapacityData[node].TotalAllocatableEphemeralStorageGB
			nodesCapacityData["*total*"].TotalAvailablePods += nodesCapacityData[node].TotalAvailablePods
			nodesCapacityData["*total*"].TotalRequestsCPU.Add(nodesCapacityData[node].TotalRequestsCPU)
			nodesCapacityData["*total*"].TotalRequestsCPUCores += nodesCapacityData[node].TotalRequestsCPUCores
			nodesCapacityData["*total*"].TotalLimitsCPU.Add(nodesCapacityData[node].TotalLimitsCPU)
			nodesCapacityData["*total*"].TotalLimitsCPUCores += nodesCapacityData[node].TotalLimitsCPUCores
			nodesCapacityData["*total*"].TotalAvailableCPU.Add(nodesCapacityData[node].TotalAvailableCPU)
			nodesCapacityData["*total*"].TotalAvailableCPUCores += nodesCapacityData[node].TotalAvailableCPUCores
			nodesCapacityData["*total*"].TotalRequestsMemory.Add(nodesCapacityData[node].TotalRequestsMemory)
			nodesCapacityData["*total*"].TotalRequestsMemoryGiB += nodesCapacityData[node].TotalRequestsMemoryGiB
			nodesCapacityData["*total*"].TotalLimitsMemory.Add(nodesCapacityData[node].TotalLimitsMemory)
			nodesCapacityData["*total*"].TotalLimitsMemoryGiB += nodesCapacityData[node].TotalLimitsMemoryGiB
			nodesCapacityData["*total*"].TotalAvailableMemory.Add(nodesCapacityData[node].TotalAvailableMemory)
			nodesCapacityData["*total*"].TotalAvailableMemoryGiB += nodesCapacityData[node].TotalAvailableMemoryGiB
			nodesCapacityData["*total*"].TotalRequestsEphemeralStorage.Add(nodesCapacityData[node].TotalRequestsEphemeralStorage)
			nodesCapacityData["*total*"].TotalRequestsEphemeralStorageGB += nodesCapacityData[node].TotalRequestsEphemeralStorageGB
			nodesCapacityData["*total*"].TotalLimitsEphemeralStorage.Add(nodesCapacityData[node].TotalLimitsEphemeralStorage)
			nodesCapacityData["*total*"].TotalLimitsEphemeralStorageGB += nodesCapacityData[node].TotalLimitsEphemeralStorageGB
			nodesCapacityData["*total*"].TotalAvailableEphemeralStorage.Add(nodesCapacityData[node].TotalAvailableEphemeralStorage)
			nodesCapacityData["*total*"].TotalAvailableEphemeralStorageGB += nodesCapacityData[node].TotalAvailableEphemeralStorageGB
		}

		sortByRole, _ := cmd.Flags().GetBool("sort-by-role")

		displayTotal, _ := cmd.Flags().GetBool("display-total")

		if displayTotal {
			nodeNames = append(nodeNames, "*total*")
			nodesByRole["~"] = append(nodesByRole["~"], "*total*")
		}

		output.DisplayNodeData(nodesCapacityData, nodeNames, displayDefault, !displayNoHeaders, displayEphemeralStorage, displayFormat, sortByRole, nodesByRole)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(nodeCmd)
	nodeCmd.Flags().BoolP("ephemeral-storage", "e", false, "Include ephemeral storage capacity data in table output")
	nodeCmd.Flags().BoolP("sort-by-role", "r", false, "Sort output by node-role")
	nodeCmd.Flags().BoolP("display-total", "t", false, "Display sum of all node capacity data in table output")
	nodeCmd.Flags().BoolP("unassigned", "u", false, "Include unassigned pod row, pods which do not have a node")
}
