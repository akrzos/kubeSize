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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/sets"
)

var nodeCmd = &cobra.Command{
	Use:           "node",
	Short:         "Get individual node capacity",
	Long:          `Get individual node size and capacity metrics grouped by node role`,
	SilenceErrors: true,
	SilenceUsage:  true,
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

		nodesCapacityData := make(map[string]*output.NodeCapacityData)
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

			newNodeData := new(output.NodeCapacityData)
			newNodeData.TotalPodCount = len(nodePodsList.Items)
			newNodeData.TotalNonTermPodCount = len(totalNonTermPodsList.Items)
			newNodeData.Roles = roles
			newNodeData.TotalCapacityPods.Add(*v.Status.Capacity.Pods())
			newNodeData.TotalCapacityCPU.Add(*v.Status.Capacity.Cpu())
			newNodeData.TotalCapacityMemory.Add(*v.Status.Capacity.Memory())
			newNodeData.TotalAllocatablePods.Add(*v.Status.Allocatable.Pods())
			newNodeData.TotalAllocatableCPU.Add(*v.Status.Allocatable.Cpu())
			newNodeData.TotalAllocatableMemory.Add(*v.Status.Capacity.Memory())

			for _, pod := range totalNonTermPodsList.Items {
				for _, container := range pod.Spec.Containers {
					newNodeData.TotalRequestsCPU.Add(*container.Resources.Requests.Cpu())
					newNodeData.TotalLimitsCPU.Add(*container.Resources.Limits.Cpu())
					newNodeData.TotalRequestsMemory.Add(*container.Resources.Requests.Memory())
					newNodeData.TotalLimitsMemory.Add(*container.Resources.Limits.Memory())
				}
			}
			nodesCapacityData[v.Name] = newNodeData

		}

		displayReadable, _ := cmd.Flags().GetBool("readable")

		displayFormat, _ := cmd.Flags().GetString("output")

		sort.Strings(sortedNodeNames)

		output.DisplayNodeData(nodesCapacityData, sortedNodeNames, displayReadable, displayFormat)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(nodeCmd)
}
