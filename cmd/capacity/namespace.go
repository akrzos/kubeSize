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

	"github.com/akrzos/kubeSize/internal/capacity"
	"github.com/akrzos/kubeSize/internal/kube"
	"github.com/akrzos/kubeSize/internal/output"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

var namespaceCmd = &cobra.Command{
	Use:     "namespace",
	Aliases: []string{"ns"},
	Short:   "Get namespace size",
	Long:    `Get namespace size and capacity metrics`,
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

		nsFlag, _ := cmd.Flags().GetString("namespace")
		nsListOptions := metav1.ListOptions{}

		if nsFlag != "" {
			nsFieldSelector, err := fields.ParseSelector("metadata.name=" + nsFlag)
			if err != nil {
				return errors.Wrap(err, "failed to create fieldSelector")
			}
			nsListOptions = metav1.ListOptions{FieldSelector: nsFieldSelector.String()}
		}

		namespaces, err := clientset.CoreV1().Namespaces().List(nsListOptions)
		if err != nil {
			return errors.Wrap(err, "failed to list namespaces")
		}

		pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list pods")
		}

		namespaceCapacityData := make(map[string]*output.NamespaceCapacityData)
		namespaceNames := make([]string, 0, len(namespaces.Items))

		for _, namespace := range namespaces.Items {
			namespaceNames = append(namespaceNames, namespace.Name)
			namespaceCapacityData[namespace.Name] = new(output.NamespaceCapacityData)
		}

		for _, pod := range pods.Items {
			if !capacity.StringInSlice(pod.Namespace, namespaceNames) {
				namespaceNames = append(namespaceNames, pod.Namespace)
				namespaceCapacityData[pod.Namespace] = new(output.NamespaceCapacityData)
			}
			if pod.Spec.NodeName == "" {
				namespaceCapacityData[pod.Namespace].TotalUnassignedNodePodCount++
			}
			namespaceCapacityData[pod.Namespace].TotalPodCount++
			if (pod.Status.Phase != corev1.PodSucceeded) && (pod.Status.Phase != corev1.PodFailed) {
				namespaceCapacityData[pod.Namespace].TotalNonTermPodCount++
				for _, container := range pod.Spec.Containers {
					namespaceCapacityData[pod.Namespace].TotalRequestsCPU.Add(*container.Resources.Requests.Cpu())
					namespaceCapacityData[pod.Namespace].TotalLimitsCPU.Add(*container.Resources.Limits.Cpu())
					namespaceCapacityData[pod.Namespace].TotalRequestsMemory.Add(*container.Resources.Requests.Memory())
					namespaceCapacityData[pod.Namespace].TotalLimitsMemory.Add(*container.Resources.Limits.Memory())
				}
			}
		}

		// Populate "Human" readable capacity data values
		for _, namespace := range namespaceNames {
			namespaceCapacityData[namespace].TotalRequestsCPUCores = capacity.ReadableCPU(namespaceCapacityData[namespace].TotalRequestsCPU)
			namespaceCapacityData[namespace].TotalLimitsCPUCores = capacity.ReadableCPU(namespaceCapacityData[namespace].TotalLimitsCPU)
			namespaceCapacityData[namespace].TotalRequestsMemoryGiB = capacity.ReadableMem(namespaceCapacityData[namespace].TotalRequestsMemory)
			namespaceCapacityData[namespace].TotalLimitsMemoryGiB = capacity.ReadableMem(namespaceCapacityData[namespace].TotalLimitsMemory)
		}

		sort.Strings(namespaceNames)

		displayDefault, _ := cmd.Flags().GetBool("default-format")

		displayNoHeaders, _ := cmd.Flags().GetBool("no-headers")

		displayFormat, _ := cmd.Flags().GetString("output")

		displayAllNamespaces, _ := cmd.Flags().GetBool("all-namespaces")

		output.DisplayNamespaceData(namespaceCapacityData, namespaceNames, displayDefault, !displayNoHeaders, displayFormat, displayAllNamespaces)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(namespaceCmd)
	namespaceCmd.Flags().BoolP("all-namespaces", "A", false, "Include 0 pod namespaces in table output")
}
