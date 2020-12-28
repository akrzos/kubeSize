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
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
)

type NodeCapacityData struct {
	totalPodCount          int
	totalNonTermPodCount   int
	roles                  sets.String
	totalCapacityPods      resource.Quantity
	totalCapacityCPU       resource.Quantity
	totalCapacityMemory    resource.Quantity
	totalAllocatablePods   resource.Quantity
	totalAllocatableCPU    resource.Quantity
	totalAllocatableMemory resource.Quantity
	totalRequestsCPU       resource.Quantity
	totalLimitsCPU         resource.Quantity
	totalRequestsMemory    resource.Quantity
	totalLimitsMemory      resource.Quantity
}

var nodeCmd = &cobra.Command{
	Use:           "node",
	Short:         "Get individual node capacity",
	Long:          `Get individual node size and capacity metrics grouped by node role`,
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

		nodesCapacityData := make(map[string]*NodeCapacityData)
		sortedNodeNames := make([]string, 0, len(nodes.Items))

		for _, v := range nodes.Items {
			sortedNodeNames = append(sortedNodeNames, v.Name)

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

			nonTerminatedFieldSelector, err := fields.ParseSelector("spec.nodeName=" + v.Name + ",status.phase!=" + string(corev1.PodSucceeded) + ",status.phase!=" + string(corev1.PodFailed))
			if err != nil {
				return errors.Wrap(err, "failed to create fieldSelector")
			}
			totalNonTermPodsList, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{FieldSelector: nonTerminatedFieldSelector.String()})

			newNodeData := new(NodeCapacityData)
			newNodeData.totalPodCount = len(nodePodsList.Items)
			newNodeData.totalNonTermPodCount = len(totalNonTermPodsList.Items)
			newNodeData.roles = roles
			newNodeData.totalCapacityPods.Add(*v.Status.Capacity.Pods())
			newNodeData.totalCapacityCPU.Add(*v.Status.Capacity.Cpu())
			newNodeData.totalCapacityMemory.Add(*v.Status.Capacity.Memory())
			newNodeData.totalAllocatablePods.Add(*v.Status.Allocatable.Pods())
			newNodeData.totalAllocatableCPU.Add(*v.Status.Allocatable.Cpu())
			newNodeData.totalAllocatableMemory.Add(*v.Status.Capacity.Memory())

			for _, pod := range totalNonTermPodsList.Items {
				for _, container := range pod.Spec.Containers {
					newNodeData.totalRequestsCPU.Add(*container.Resources.Requests.Cpu())
					newNodeData.totalLimitsCPU.Add(*container.Resources.Limits.Cpu())
					newNodeData.totalRequestsMemory.Add(*container.Resources.Requests.Memory())
					newNodeData.totalLimitsMemory.Add(*container.Resources.Limits.Memory())
				}
			}
			nodesCapacityData[v.Name] = newNodeData

		}

		displayReadable, _ := cmd.Flags().GetBool("readable")

		sort.Strings(sortedNodeNames)

		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 5, 1, ' ', 0)
		if displayReadable == true {
			fmt.Fprintln(w, "NAME\tROLES\tPODS\t\t\t\tCPU (cores)\t\t\t\tMEMORY (GiB)\t\t")
		} else {
			fmt.Fprintln(w, "NAME\tROLES\tPODS\t\t\t\tCPU\t\t\t\tMEMORY\t\t")
		}
		fmt.Fprintln(w, "\t\tCapacity\tAllocatable\tTotal\tNon-Term\tCapacity\tAllocatable\tRequests\tLimits\tCapacity\tAllocatable\tRequests\tLimits")

		for _, k := range sortedNodeNames {
			fmt.Fprintf(w, "%s\t", k)
			fmt.Fprintf(w, "%s\t", strings.Join(nodesCapacityData[k].roles.List(), ","))
			fmt.Fprintf(w, "%s\t%s\t", &nodesCapacityData[k].totalCapacityPods, &nodesCapacityData[k].totalCapacityPods)
			fmt.Fprintf(w, "%d\t%d\t", nodesCapacityData[k].totalPodCount, nodesCapacityData[k].totalNonTermPodCount)
			if displayReadable == true {
				fmt.Fprintf(w, "%.1f\t%.1f\t", float64(nodesCapacityData[k].totalCapacityCPU.MilliValue())/1000, float64(nodesCapacityData[k].totalAllocatableCPU.MilliValue())/1000)
				fmt.Fprintf(w, "%.1f\t%.1f\t", float64(nodesCapacityData[k].totalRequestsCPU.MilliValue())/1000, float64(nodesCapacityData[k].totalLimitsCPU.MilliValue())/1000)
				fmt.Fprintf(w, "%.1f\t%.1f\t", float64(nodesCapacityData[k].totalCapacityMemory.Value())/1024/1024/1024, float64(nodesCapacityData[k].totalAllocatableMemory.Value())/1024/1024/1024)
				fmt.Fprintf(w, "%.1f\t%.1f\t\n", float64(nodesCapacityData[k].totalRequestsMemory.Value())/1024/1024/1024, float64(nodesCapacityData[k].totalLimitsMemory.Value())/1024/1024/1024)
			} else {
				fmt.Fprintf(w, "%s\t%s\t", &nodesCapacityData[k].totalCapacityCPU, &nodesCapacityData[k].totalAllocatableCPU)
				fmt.Fprintf(w, "%s\t%s\t", &nodesCapacityData[k].totalRequestsCPU, &nodesCapacityData[k].totalLimitsCPU)
				fmt.Fprintf(w, "%s\t%s\t", &nodesCapacityData[k].totalCapacityMemory, &nodesCapacityData[k].totalAllocatableMemory)
				fmt.Fprintf(w, "%s\t%s\t\n", &nodesCapacityData[k].totalRequestsMemory, &nodesCapacityData[k].totalLimitsMemory)
			}
		}
		w.Flush()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(nodeCmd)
}
