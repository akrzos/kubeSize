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
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/sets"
)

var nodeRoleCmd = &cobra.Command{
	Use:     "node-role",
	Aliases: []string{"nr"},
	Short:   "Get cluster capacity grouped by node role",
	Long:    `Get Kubernetes cluster size and capacity metrics grouped by node role`,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlags(cmd.Flags())
		if err := output.ValidateOutput(*cmd); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
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

		nodeRoleCapacityData := make(map[string]*output.ClusterCapacityData)
		sortedRoleNames := make([]string, 0)

		for _, v := range nodes.Items {

			roles := sets.NewString()
			for labelKey, labelValue := range v.Labels {
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

			nodeFieldSelector, err := fields.ParseSelector("spec.nodeName=" + v.Name)
			if err != nil {
				return errors.Wrap(err, "failed to create fieldSelector")
			}
			nodePodsList, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{FieldSelector: nodeFieldSelector.String()})
			totalPodCount := len(nodePodsList.Items)

			nonTerminatedFieldSelector, err := fields.ParseSelector("spec.nodeName=" + v.Name + ",status.phase!=" + string(corev1.PodSucceeded) + ",status.phase!=" + string(corev1.PodFailed))
			if err != nil {
				return errors.Wrap(err, "failed to create fieldSelector")
			}
			totalNonTermPodsList, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{FieldSelector: nonTerminatedFieldSelector.String()})
			nonTerminatedPodCount := len(totalNonTermPodsList.Items)

			var totalRequestsCPU, totalLimitssCPU, totalRequestsMemory, totalLimitsMemory resource.Quantity

			for _, pod := range totalNonTermPodsList.Items {
				for _, container := range pod.Spec.Containers {
					totalRequestsCPU.Add(*container.Resources.Requests.Cpu())
					totalLimitssCPU.Add(*container.Resources.Limits.Cpu())
					totalRequestsMemory.Add(*container.Resources.Requests.Memory())
					totalLimitsMemory.Add(*container.Resources.Limits.Memory())
				}
			}

			for role := range roles {
				if nodeRoleData, ok := nodeRoleCapacityData[role]; ok {
					nodeRoleData.TotalNodeCount++
					for _, condition := range v.Status.Conditions {
						if (condition.Type == "Ready") && condition.Status == corev1.ConditionTrue {
							nodeRoleData.TotalReadyNodeCount++
						}
					}
					nodeRoleData.TotalUnreadyNodeCount = nodeRoleData.TotalNodeCount - nodeRoleData.TotalReadyNodeCount
					if v.Spec.Unschedulable {
						nodeRoleData.TotalUnschedulableNodeCount++
					}
					nodeRoleData.TotalCapacityPods.Add(*v.Status.Capacity.Pods())
					nodeRoleData.TotalCapacityCPU.Add(*v.Status.Capacity.Cpu())
					nodeRoleData.TotalCapacityMemory.Add(*v.Status.Capacity.Memory())
					nodeRoleData.TotalAllocatablePods.Add(*v.Status.Allocatable.Pods())
					nodeRoleData.TotalAllocatableCPU.Add(*v.Status.Allocatable.Cpu())
					nodeRoleData.TotalAllocatableMemory.Add(*v.Status.Allocatable.Memory())
					nodeRoleData.TotalRequestsCPU.Add(totalRequestsCPU)
					nodeRoleData.TotalLimitsCPU.Add(totalLimitssCPU)
					nodeRoleData.TotalRequestsMemory.Add(totalRequestsMemory)
					nodeRoleData.TotalLimitsMemory.Add(totalLimitsMemory)
					nodeRoleData.TotalPodCount += totalPodCount
					nodeRoleData.TotalNonTermPodCount += nonTerminatedPodCount
				} else {
					sortedRoleNames = append(sortedRoleNames, role)
					newNodeRoleCapacityData := new(output.ClusterCapacityData)
					newNodeRoleCapacityData.TotalNodeCount = 1
					for _, condition := range v.Status.Conditions {
						if (condition.Type == "Ready") && condition.Status == corev1.ConditionTrue {
							newNodeRoleCapacityData.TotalReadyNodeCount = 1
							newNodeRoleCapacityData.TotalUnreadyNodeCount = 0
						}
					}
					if v.Spec.Unschedulable {
						newNodeRoleCapacityData.TotalUnschedulableNodeCount = 1
					}
					newNodeRoleCapacityData.TotalCapacityPods.Add(*v.Status.Capacity.Pods())
					newNodeRoleCapacityData.TotalCapacityCPU.Add(*v.Status.Capacity.Cpu())
					newNodeRoleCapacityData.TotalCapacityMemory.Add(*v.Status.Capacity.Memory())
					newNodeRoleCapacityData.TotalAllocatablePods.Add(*v.Status.Allocatable.Pods())
					newNodeRoleCapacityData.TotalAllocatableCPU.Add(*v.Status.Allocatable.Cpu())
					newNodeRoleCapacityData.TotalAllocatableMemory.Add(*v.Status.Allocatable.Memory())
					newNodeRoleCapacityData.TotalRequestsCPU.Add(totalRequestsCPU)
					newNodeRoleCapacityData.TotalLimitsCPU.Add(totalLimitssCPU)
					newNodeRoleCapacityData.TotalRequestsMemory.Add(totalRequestsMemory)
					newNodeRoleCapacityData.TotalLimitsMemory.Add(totalLimitsMemory)
					newNodeRoleCapacityData.TotalPodCount += totalPodCount
					newNodeRoleCapacityData.TotalNonTermPodCount += nonTerminatedPodCount
					nodeRoleCapacityData[role] = newNodeRoleCapacityData
				}
			}

		}

		displayReadable, _ := cmd.Flags().GetBool("readable")

		displayFormat, _ := cmd.Flags().GetString("output")

		sort.Strings(sortedRoleNames)

		output.DisplayNodeRoleData(nodeRoleCapacityData, sortedRoleNames, displayReadable, displayFormat)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(nodeRoleCmd)
}
