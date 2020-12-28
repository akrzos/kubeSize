/*
Copyright © 2020 Alex Krzos akrzos@redhat.com

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
package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
)

type RoleCapacityData struct {
	totalNodeCount              int
	totalReadyNodeCount         int
	totalUnreadyNodeCount       int
	totalUnschedulableNodeCount int
	totalPodCount               int
	totalNonTermPodCount        int
	totalCapacityPods           resource.Quantity
	totalCapacityCPU            resource.Quantity
	totalCapacityMemory         resource.Quantity
	totalAllocatablePods        resource.Quantity
	totalAllocatableCPU         resource.Quantity
	totalAllocatableMemory      resource.Quantity
	totalRequestsCPU            resource.Quantity
	totalLimitsCPU              resource.Quantity
	totalRequestsMemory         resource.Quantity
	totalLimitsMemory           resource.Quantity
}

var nodeRoleCmd = &cobra.Command{
	Use:   "node-role",
	Short: "Get cluster capacity grouped by node role",
	Long:  `Get Kubernetes cluster size and capacity metrics grouped by node role`,
	RunE: func(cmd *cobra.Command, args []string) error {

		config, err := KubernetesConfigFlags.ToRESTConfig()
		if err != nil {
			return errors.Wrap(err, "failed to read kubeconfig")
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return errors.Wrap(err, "failed to create clientset")
		}

		nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list nodes")
		}

		nodeRoleCapacityData := make(map[string]*RoleCapacityData)
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
					nodeRoleData.totalNodeCount++
					for _, condition := range v.Status.Conditions {
						if (condition.Type == "Ready") && condition.Status == corev1.ConditionTrue {
							nodeRoleData.totalReadyNodeCount++
						}
					}
					nodeRoleData.totalUnreadyNodeCount = nodeRoleData.totalNodeCount - nodeRoleData.totalReadyNodeCount
					if v.Spec.Unschedulable {
						nodeRoleData.totalUnschedulableNodeCount++
					}
					nodeRoleData.totalCapacityPods.Add(*v.Status.Capacity.Pods())
					nodeRoleData.totalCapacityCPU.Add(*v.Status.Capacity.Cpu())
					nodeRoleData.totalCapacityMemory.Add(*v.Status.Capacity.Memory())
					nodeRoleData.totalAllocatablePods.Add(*v.Status.Allocatable.Pods())
					nodeRoleData.totalAllocatableCPU.Add(*v.Status.Allocatable.Cpu())
					nodeRoleData.totalAllocatableMemory.Add(*v.Status.Allocatable.Memory())
					nodeRoleData.totalRequestsCPU.Add(totalRequestsCPU)
					nodeRoleData.totalLimitsCPU.Add(totalLimitssCPU)
					nodeRoleData.totalRequestsMemory.Add(totalRequestsMemory)
					nodeRoleData.totalLimitsMemory.Add(totalLimitsMemory)
					nodeRoleData.totalPodCount += totalPodCount
					nodeRoleData.totalNonTermPodCount += nonTerminatedPodCount
				} else {
					sortedRoleNames = append(sortedRoleNames, role)
					newNodeRoleCapacityData := new(RoleCapacityData)
					newNodeRoleCapacityData.totalNodeCount = 1
					for _, condition := range v.Status.Conditions {
						if (condition.Type == "Ready") && condition.Status == corev1.ConditionTrue {
							newNodeRoleCapacityData.totalReadyNodeCount = 1
							newNodeRoleCapacityData.totalUnreadyNodeCount = 0
						}
					}
					if v.Spec.Unschedulable {
						newNodeRoleCapacityData.totalUnschedulableNodeCount = 1
					}
					newNodeRoleCapacityData.totalCapacityPods.Add(*v.Status.Capacity.Pods())
					newNodeRoleCapacityData.totalCapacityCPU.Add(*v.Status.Capacity.Cpu())
					newNodeRoleCapacityData.totalCapacityMemory.Add(*v.Status.Capacity.Memory())
					newNodeRoleCapacityData.totalAllocatablePods.Add(*v.Status.Allocatable.Pods())
					newNodeRoleCapacityData.totalAllocatableCPU.Add(*v.Status.Allocatable.Cpu())
					newNodeRoleCapacityData.totalAllocatableMemory.Add(*v.Status.Allocatable.Memory())
					newNodeRoleCapacityData.totalRequestsCPU.Add(totalRequestsCPU)
					newNodeRoleCapacityData.totalLimitsCPU.Add(totalLimitssCPU)
					newNodeRoleCapacityData.totalRequestsMemory.Add(totalRequestsMemory)
					newNodeRoleCapacityData.totalLimitsMemory.Add(totalLimitsMemory)
					newNodeRoleCapacityData.totalPodCount += totalPodCount
					newNodeRoleCapacityData.totalNonTermPodCount += nonTerminatedPodCount
					nodeRoleCapacityData[role] = newNodeRoleCapacityData
				}
			}

		}

		displayReadable, _ := cmd.Flags().GetBool("readable")

		sort.Strings(sortedRoleNames)

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
			fmt.Fprintf(w, "%d\t%d\t%d\t%d\t", nodeRoleCapacityData[k].totalNodeCount, nodeRoleCapacityData[k].totalReadyNodeCount, nodeRoleCapacityData[k].totalUnreadyNodeCount, nodeRoleCapacityData[k].totalUnschedulableNodeCount)
			fmt.Fprintf(w, "%s\t%s\t", &nodeRoleCapacityData[k].totalCapacityPods, &nodeRoleCapacityData[k].totalAllocatablePods)
			fmt.Fprintf(w, "%d\t%d\t", nodeRoleCapacityData[k].totalPodCount, nodeRoleCapacityData[k].totalNonTermPodCount)
			if displayReadable == true {
				fmt.Fprintf(w, "%.1f\t%.1f\t", float64(nodeRoleCapacityData[k].totalCapacityCPU.MilliValue())/1000, float64(nodeRoleCapacityData[k].totalAllocatableCPU.MilliValue())/1000)
				fmt.Fprintf(w, "%.1f\t%.1f\t", float64(nodeRoleCapacityData[k].totalRequestsCPU.MilliValue())/1000, float64(nodeRoleCapacityData[k].totalLimitsCPU.MilliValue())/1000)
				fmt.Fprintf(w, "%.1f\t%.1f\t", float64(nodeRoleCapacityData[k].totalCapacityMemory.Value())/1024/1024/1024, float64(nodeRoleCapacityData[k].totalAllocatableMemory.Value())/1024/1024/1024)
				fmt.Fprintf(w, "%.1f\t%.1f\t\n", float64(nodeRoleCapacityData[k].totalRequestsMemory.Value())/1024/1024/1024, float64(nodeRoleCapacityData[k].totalLimitsMemory.Value())/1024/1024/1024)
			} else {
				fmt.Fprintf(w, "%s\t%s\t", &nodeRoleCapacityData[k].totalCapacityCPU, &nodeRoleCapacityData[k].totalAllocatableCPU)
				fmt.Fprintf(w, "%s\t%s\t", &nodeRoleCapacityData[k].totalRequestsCPU, &nodeRoleCapacityData[k].totalLimitsCPU)
				fmt.Fprintf(w, "%s\t%s\t", &nodeRoleCapacityData[k].totalCapacityMemory, &nodeRoleCapacityData[k].totalAllocatableMemory)
				fmt.Fprintf(w, "%s\t%s\t\n", &nodeRoleCapacityData[k].totalRequestsMemory, &nodeRoleCapacityData[k].totalLimitsMemory)
			}
		}

		w.Flush()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(nodeRoleCmd)
}
