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
	"text/tabwriter"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
)

var clusterCmd = &cobra.Command{
	Use:           "cluster",
	Short:         "Get Cluster Capacity",
	Long:          `.`,
	SilenceErrors: true,
	SilenceUsage:  true,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlags(cmd.Flags())
	},
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

		var totalCapacityPods, totalCapacityCPU, totalCapacityMemory resource.Quantity
		var totalAllocatablePods, totalAllocatableCPU, totalAllocatableMemory resource.Quantity
		var totalRequestsCPU, totalLimitsCPU, totalRequestsMemory, totalLimitsMemory resource.Quantity
		var totalNodeCount, totalReadyNodeCount, totalUnreadyNodeCount, totalUnschedulableNodeCount = 0, 0, 0, 0
		var totalPodCount, totalNonTermPodCount = 0, 0

		for _, v := range nodes.Items {
			totalNodeCount++
			for _, condition := range v.Status.Conditions {
				if (condition.Type == "Ready") && condition.Status == corev1.ConditionTrue {
					totalReadyNodeCount++
				}
			}
			if v.Spec.Unschedulable {
				totalUnschedulableNodeCount++
			}
			totalCapacityPods.Add(*v.Status.Capacity.Pods())
			totalCapacityCPU.Add(*v.Status.Capacity.Cpu())
			totalCapacityMemory.Add(*v.Status.Capacity.Memory())
			totalAllocatablePods.Add(*v.Status.Allocatable.Pods())
			totalAllocatableCPU.Add(*v.Status.Allocatable.Cpu())
			totalAllocatableMemory.Add(*v.Status.Allocatable.Memory())
		}
		totalUnreadyNodeCount = totalNodeCount - totalReadyNodeCount

		totalPodsList, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
		totalPodCount = len(totalPodsList.Items)

		fieldSelector, err := fields.ParseSelector("status.phase!=" + string(corev1.PodSucceeded) + ",status.phase!=" + string(corev1.PodFailed))
		if err != nil {
			return errors.Wrap(err, "failed to create fieldSelector")
		}
		totalNonTermPodsList, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{FieldSelector: fieldSelector.String()})
		totalNonTermPodCount = len(totalNonTermPodsList.Items)

		for _, pod := range totalNonTermPodsList.Items {
			for _, container := range pod.Spec.Containers {
				totalRequestsCPU.Add(*container.Resources.Requests.Cpu())
				totalLimitsCPU.Add(*container.Resources.Limits.Cpu())
				totalRequestsMemory.Add(*container.Resources.Requests.Memory())
				totalLimitsMemory.Add(*container.Resources.Limits.Memory())
			}
		}

		var capacityMemoryGiB = float64(totalCapacityMemory.Value()) / 1024 / 1024 / 1024
		var allocatableMemoryGiB = float64(totalAllocatableMemory.Value()) / 1024 / 1024 / 1024

		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 5, 1, ' ', 0)
		fmt.Fprintln(w, "NODES\t\t\t\tPODS\t\t\t\tCPU\t\t\t\tMEMORY\t\t")
		fmt.Fprintln(w, "Total\tReady\tUnready\tUnschedulable\tCapacity\tAllocatable\tTotal\tNon-Term\tCapacity\tAllocatable\tRequests\tLimits\tCapacity\tAllocatable\tRequests\tLimits")
		fmt.Fprintf(w, "%d\t%d\t%d\t%d\t", totalNodeCount, totalReadyNodeCount, totalUnreadyNodeCount, totalUnschedulableNodeCount)
		fmt.Fprintf(w, "%s\t%s\t", &totalCapacityPods, &totalAllocatablePods)
		fmt.Fprintf(w, "%d\t%d\t", totalPodCount, totalNonTermPodCount)
		fmt.Fprintf(w, "%s\t%s\t", &totalCapacityCPU, &totalAllocatableCPU)
		fmt.Fprintf(w, "%s\t%s\t", &totalRequestsCPU, &totalLimitsCPU)
		// fmt.Fprintf(w, "%s : %s\t", &totalCapacityMemory, &totalAllocatableMemory)
		fmt.Fprintf(w, "%.1f\t%.1f\t", capacityMemoryGiB, allocatableMemoryGiB)
		fmt.Fprintf(w, "%s\t%s\t\n", &totalRequestsMemory, &totalLimitsMemory)
		w.Flush()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(clusterCmd)
}
