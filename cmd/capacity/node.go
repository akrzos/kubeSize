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
	Long:    `Get individual node size and capacity metrics grouped by node role`,
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
			nodesCapacityData[node.Name].TotalAllocatablePods.Add(*node.Status.Allocatable.Pods())
			nodesCapacityData[node.Name].TotalAllocatableCPU.Add(*node.Status.Allocatable.Cpu())
			nodesCapacityData[node.Name].TotalAllocatableMemory.Add(*node.Status.Allocatable.Memory())
			for _, role := range roles.List() {
				nodesByRole[role] = append(nodesByRole[role], node.Name)
			}
		}
		nodesCapacityData["unassigned"] = new(output.NodeCapacityData)

		for _, pod := range pods.Items {
			podNode := pod.Spec.NodeName
			if pod.Spec.NodeName == "" {
				podNode = "unassigned"
			}
			nodesCapacityData[podNode].TotalPodCount++

			if (pod.Status.Phase != corev1.PodSucceeded) && (pod.Status.Phase != corev1.PodFailed) {
				nodesCapacityData[podNode].TotalNonTermPodCount++
				for _, container := range pod.Spec.Containers {
					nodesCapacityData[podNode].TotalRequestsCPU.Add(*container.Resources.Requests.Cpu())
					nodesCapacityData[podNode].TotalLimitsCPU.Add(*container.Resources.Limits.Cpu())
					nodesCapacityData[podNode].TotalRequestsMemory.Add(*container.Resources.Requests.Memory())
					nodesCapacityData[podNode].TotalLimitsMemory.Add(*container.Resources.Limits.Memory())
				}
			}
		}

		for _, node := range nodeNames {
			nodesCapacityData[node].TotalAvailablePods = int(nodesCapacityData[node].TotalAllocatablePods.Value()) - nodesCapacityData[node].TotalNonTermPodCount
			nodesCapacityData[node].TotalAvailableCPU = nodesCapacityData[node].TotalAllocatableCPU
			nodesCapacityData[node].TotalAvailableCPU.Sub(nodesCapacityData[node].TotalRequestsCPU)
			nodesCapacityData[node].TotalAvailableMemory = nodesCapacityData[node].TotalAllocatableMemory
			nodesCapacityData[node].TotalAvailableMemory.Sub(nodesCapacityData[node].TotalRequestsMemory)
		}

		displayDefault, _ := cmd.Flags().GetBool("default-format")

		displayNoHeaders, _ := cmd.Flags().GetBool("no-headers")

		displayFormat, _ := cmd.Flags().GetString("output")

		sort.Strings(nodeNames)
		if displayUnassigned, _ := cmd.Flags().GetBool("unassigned"); displayUnassigned {
			nodeNames = append(nodeNames, "unassigned")
			nodesByRole["~"] = append(nodesByRole["~"], "unassigned")
		}

		sortByRole, _ := cmd.Flags().GetBool("sort-by-role")

		output.DisplayNodeData(nodesCapacityData, nodeNames, displayDefault, !displayNoHeaders, displayFormat, sortByRole, nodesByRole)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(nodeCmd)
	nodeCmd.Flags().BoolP("unassigned", "u", false, "Include unassigned pod row, pods which do not have a node")
	nodeCmd.Flags().BoolP("sort-by-role", "r", false, "Sort output by node-role")
}
