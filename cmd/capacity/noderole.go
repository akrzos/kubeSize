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

var nodeRoleCmd = &cobra.Command{
	Use:     "node-role",
	Aliases: []string{"nr"},
	Short:   "Get cluster capacity data grouped by node role",
	Long:    `Get metrics and data related to cluster capacity grouped by node role`,
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

		nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list nodes")
		}

		pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list pods")
		}

		nodeRoleCapacityData := make(map[string]*output.ClusterCapacityData)
		nodeRoles := make(map[string][]string)
		roleNames := make([]string, 0)

		for _, node := range nodes.Items {
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
			for role := range roles {
				if !capacity.StringInSlice(role, roleNames) {
					roleNames = append(roleNames, role)
					nodeRoleCapacityData[role] = new(output.ClusterCapacityData)
				}
				nodeRoleCapacityData[role].TotalNodeCount++
				for _, condition := range node.Status.Conditions {
					if (condition.Type == "Ready") && condition.Status == corev1.ConditionTrue {
						nodeRoleCapacityData[role].TotalReadyNodeCount++
					}
				}
				if node.Spec.Unschedulable {
					nodeRoleCapacityData[role].TotalUnschedulableNodeCount++
				}
				nodeRoleCapacityData[role].TotalCapacityPods.Add(*node.Status.Capacity.Pods())
				nodeRoleCapacityData[role].TotalCapacityCPU.Add(*node.Status.Capacity.Cpu())
				nodeRoleCapacityData[role].TotalCapacityMemory.Add(*node.Status.Capacity.Memory())
				nodeRoleCapacityData[role].TotalCapacityEphemeralStorage.Add(*node.Status.Capacity.StorageEphemeral())
				nodeRoleCapacityData[role].TotalAllocatablePods.Add(*node.Status.Allocatable.Pods())
				nodeRoleCapacityData[role].TotalAllocatableCPU.Add(*node.Status.Allocatable.Cpu())
				nodeRoleCapacityData[role].TotalAllocatableMemory.Add(*node.Status.Allocatable.Memory())
				nodeRoleCapacityData[role].TotalAllocatableEphemeralStorage.Add(*node.Status.Allocatable.StorageEphemeral())
			}
			nodeRoles[node.Name] = roles.List()
		}

		nodeRoleCapacityData["*unassigned*"] = new(output.ClusterCapacityData)
		nodeRoles["*unassigned*"] = []string{"*unassigned*"}

		for _, pod := range pods.Items {
			podNode := pod.Spec.NodeName
			if pod.Spec.NodeName == "" {
				podNode = "*unassigned*"
			}
			for _, role := range nodeRoles[podNode] {
				nodeRoleCapacityData[role].TotalPodCount++
				if (pod.Status.Phase != corev1.PodSucceeded) && (pod.Status.Phase != corev1.PodFailed) {
					nodeRoleCapacityData[role].TotalNonTermPodCount++
					for _, container := range pod.Spec.Containers {
						nodeRoleCapacityData[role].TotalRequestsCPU.Add(*container.Resources.Requests.Cpu())
						nodeRoleCapacityData[role].TotalLimitsCPU.Add(*container.Resources.Limits.Cpu())
						nodeRoleCapacityData[role].TotalRequestsMemory.Add(*container.Resources.Requests.Memory())
						nodeRoleCapacityData[role].TotalLimitsMemory.Add(*container.Resources.Limits.Memory())
						nodeRoleCapacityData[role].TotalRequestsEphemeralStorage.Add(*container.Resources.Requests.StorageEphemeral())
						nodeRoleCapacityData[role].TotalLimitsEphemeralStorage.Add(*container.Resources.Limits.StorageEphemeral())
					}
				}
			}
		}

		for _, role := range roleNames {
			nodeRoleCapacityData[role].TotalUnreadyNodeCount = nodeRoleCapacityData[role].TotalNodeCount - nodeRoleCapacityData[role].TotalReadyNodeCount
			nodeRoleCapacityData[role].TotalAvailablePods = int(nodeRoleCapacityData[role].TotalAllocatablePods.Value()) - nodeRoleCapacityData[role].TotalNonTermPodCount
			nodeRoleCapacityData[role].TotalAvailableCPU = nodeRoleCapacityData[role].TotalAllocatableCPU
			nodeRoleCapacityData[role].TotalAvailableCPU.Sub(nodeRoleCapacityData[role].TotalRequestsCPU)
			nodeRoleCapacityData[role].TotalAvailableMemory = nodeRoleCapacityData[role].TotalAllocatableMemory
			nodeRoleCapacityData[role].TotalAvailableMemory.Sub(nodeRoleCapacityData[role].TotalRequestsMemory)
			nodeRoleCapacityData[role].TotalAvailableEphemeralStorage = nodeRoleCapacityData[role].TotalAllocatableEphemeralStorage
			nodeRoleCapacityData[role].TotalAvailableEphemeralStorage.Sub(nodeRoleCapacityData[role].TotalRequestsEphemeralStorage)
		}

		displayDefault, _ := cmd.Flags().GetBool("default-format")

		displayEphemeralStorage, _ := cmd.Flags().GetBool("ephemeral-storage")

		displayNoHeaders, _ := cmd.Flags().GetBool("no-headers")

		displayFormat, _ := cmd.Flags().GetString("output")

		sort.Strings(roleNames)
		if displayUnassigned, _ := cmd.Flags().GetBool("unassigned"); displayUnassigned {
			roleNames = append(roleNames, "*unassigned*")
		}

		// Populate "Human" readable capacity data values
		for _, role := range roleNames {
			nodeRoleCapacityData[role].TotalCapacityCPUCores = capacity.ReadableCPU(nodeRoleCapacityData[role].TotalCapacityCPU)
			nodeRoleCapacityData[role].TotalCapacityMemoryGiB = capacity.ReadableMem(nodeRoleCapacityData[role].TotalCapacityMemory)
			nodeRoleCapacityData[role].TotalCapacityEphemeralStorageGB = capacity.ReadableStorage(nodeRoleCapacityData[role].TotalCapacityEphemeralStorage)
			nodeRoleCapacityData[role].TotalAllocatableCPUCores = capacity.ReadableCPU(nodeRoleCapacityData[role].TotalAllocatableCPU)
			nodeRoleCapacityData[role].TotalAllocatableMemoryGiB = capacity.ReadableMem(nodeRoleCapacityData[role].TotalAllocatableMemory)
			nodeRoleCapacityData[role].TotalAllocatableEphemeralStorageGB = capacity.ReadableStorage(nodeRoleCapacityData[role].TotalAllocatableEphemeralStorage)
			nodeRoleCapacityData[role].TotalRequestsCPUCores = capacity.ReadableCPU(nodeRoleCapacityData[role].TotalRequestsCPU)
			nodeRoleCapacityData[role].TotalLimitsCPUCores = capacity.ReadableCPU(nodeRoleCapacityData[role].TotalLimitsCPU)
			nodeRoleCapacityData[role].TotalAvailableCPUCores = capacity.ReadableCPU(nodeRoleCapacityData[role].TotalAvailableCPU)
			nodeRoleCapacityData[role].TotalRequestsMemoryGiB = capacity.ReadableMem(nodeRoleCapacityData[role].TotalRequestsMemory)
			nodeRoleCapacityData[role].TotalLimitsMemoryGiB = capacity.ReadableMem(nodeRoleCapacityData[role].TotalLimitsMemory)
			nodeRoleCapacityData[role].TotalAvailableMemoryGiB = capacity.ReadableMem(nodeRoleCapacityData[role].TotalAvailableMemory)
			nodeRoleCapacityData[role].TotalRequestsEphemeralStorageGB = capacity.ReadableStorage(nodeRoleCapacityData[role].TotalRequestsEphemeralStorage)
			nodeRoleCapacityData[role].TotalLimitsEphemeralStorageGB = capacity.ReadableStorage(nodeRoleCapacityData[role].TotalLimitsEphemeralStorage)
			nodeRoleCapacityData[role].TotalAvailableEphemeralStorageGB = capacity.ReadableStorage(nodeRoleCapacityData[role].TotalAvailableEphemeralStorage)
		}

		output.DisplayNodeRoleData(nodeRoleCapacityData, roleNames, displayDefault, !displayNoHeaders, displayEphemeralStorage, displayFormat)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(nodeRoleCmd)
	nodeRoleCmd.Flags().BoolP("ephemeral-storage", "e", false, "Include ephemeral storage capacity data in table output")
	nodeRoleCmd.Flags().BoolP("unassigned", "u", false, "Include unassigned pod row, pods which do not have a node")
}
