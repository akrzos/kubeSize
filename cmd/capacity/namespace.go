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
	Long:    `Get metrics related to the size of a namespace`,
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
		podListOptions := metav1.ListOptions{}

		if nsFlag != "" {
			nsFieldSelector, err := fields.ParseSelector("metadata.name=" + nsFlag)
			if err != nil {
				return errors.Wrap(err, "failed to create fieldSelector")
			}
			podNamespaceFieldSelector, err := fields.ParseSelector("metadata.namespace=" + nsFlag)
			if err != nil {
				return errors.Wrap(err, "failed to create fieldSelector")
			}
			nsListOptions = metav1.ListOptions{FieldSelector: nsFieldSelector.String()}
			podListOptions = metav1.ListOptions{FieldSelector: podNamespaceFieldSelector.String()}
		}

		namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), nsListOptions)
		if err != nil {
			return errors.Wrap(err, "failed to list namespaces")
		}

		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), podListOptions)
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
					namespaceCapacityData[pod.Namespace].TotalRequestsEphemeralStorage.Add(*container.Resources.Requests.StorageEphemeral())
					namespaceCapacityData[pod.Namespace].TotalLimitsEphemeralStorage.Add(*container.Resources.Limits.StorageEphemeral())
				}
			}
		}

		namespaceCapacityData["*total*"] = new(output.NamespaceCapacityData)

		// Populate "Human" readable capacity data values and the *total* "namespace"
		for _, namespace := range namespaceNames {
			namespaceCapacityData[namespace].TotalRequestsCPUCores = capacity.ReadableCPU(namespaceCapacityData[namespace].TotalRequestsCPU)
			namespaceCapacityData[namespace].TotalLimitsCPUCores = capacity.ReadableCPU(namespaceCapacityData[namespace].TotalLimitsCPU)
			namespaceCapacityData[namespace].TotalRequestsMemoryGiB = capacity.ReadableMem(namespaceCapacityData[namespace].TotalRequestsMemory)
			namespaceCapacityData[namespace].TotalLimitsMemoryGiB = capacity.ReadableMem(namespaceCapacityData[namespace].TotalLimitsMemory)
			namespaceCapacityData[namespace].TotalRequestsEphemeralStorageGB = capacity.ReadableStorage(namespaceCapacityData[namespace].TotalRequestsEphemeralStorage)
			namespaceCapacityData[namespace].TotalLimitsEphemeralStorageGB = capacity.ReadableStorage(namespaceCapacityData[namespace].TotalLimitsEphemeralStorage)
			namespaceCapacityData["*total*"].TotalPodCount += namespaceCapacityData[namespace].TotalPodCount
			namespaceCapacityData["*total*"].TotalNonTermPodCount += namespaceCapacityData[namespace].TotalNonTermPodCount
			namespaceCapacityData["*total*"].TotalUnassignedNodePodCount += namespaceCapacityData[namespace].TotalUnassignedNodePodCount
			namespaceCapacityData["*total*"].TotalRequestsCPU.Add(namespaceCapacityData[namespace].TotalRequestsCPU)
			namespaceCapacityData["*total*"].TotalRequestsCPUCores += namespaceCapacityData[namespace].TotalRequestsCPUCores
			namespaceCapacityData["*total*"].TotalLimitsCPU.Add(namespaceCapacityData[namespace].TotalLimitsCPU)
			namespaceCapacityData["*total*"].TotalLimitsCPUCores += namespaceCapacityData[namespace].TotalLimitsCPUCores
			namespaceCapacityData["*total*"].TotalRequestsMemory.Add(namespaceCapacityData[namespace].TotalRequestsMemory)
			namespaceCapacityData["*total*"].TotalRequestsMemoryGiB += namespaceCapacityData[namespace].TotalRequestsMemoryGiB
			namespaceCapacityData["*total*"].TotalLimitsMemory.Add(namespaceCapacityData[namespace].TotalLimitsMemory)
			namespaceCapacityData["*total*"].TotalLimitsMemoryGiB += namespaceCapacityData[namespace].TotalLimitsMemoryGiB
			namespaceCapacityData["*total*"].TotalRequestsEphemeralStorage.Add(namespaceCapacityData[namespace].TotalRequestsEphemeralStorage)
			namespaceCapacityData["*total*"].TotalRequestsEphemeralStorageGB += namespaceCapacityData[namespace].TotalRequestsEphemeralStorageGB
			namespaceCapacityData["*total*"].TotalLimitsEphemeralStorage.Add(namespaceCapacityData[namespace].TotalLimitsEphemeralStorage)
			namespaceCapacityData["*total*"].TotalLimitsEphemeralStorageGB += namespaceCapacityData[namespace].TotalLimitsEphemeralStorageGB
		}

		sort.Strings(namespaceNames)

		displayDefault, _ := cmd.Flags().GetBool("default-format")

		displayEphemeralStorage, _ := cmd.Flags().GetBool("ephemeral-storage")

		displayNoHeaders, _ := cmd.Flags().GetBool("no-headers")

		displayFormat, _ := cmd.Flags().GetString("output")

		displayAllNamespaces, _ := cmd.Flags().GetBool("all-namespaces")

		displayTotal, _ := cmd.Flags().GetBool("display-total")

		if displayTotal {
			namespaceNames = append(namespaceNames, "*total*")
		}

		output.DisplayNamespaceData(namespaceCapacityData, namespaceNames, displayDefault, !displayNoHeaders, displayEphemeralStorage, displayFormat, displayAllNamespaces)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(namespaceCmd)
	namespaceCmd.Flags().BoolP("all-namespaces", "A", false, "Include 0 pod namespaces in table output")
	namespaceCmd.Flags().BoolP("ephemeral-storage", "e", false, "Include ephemeral storage capacity data in table output")
	namespaceCmd.Flags().BoolP("display-total", "t", false, "Display sum of all namespace capacity data in table output")
}
