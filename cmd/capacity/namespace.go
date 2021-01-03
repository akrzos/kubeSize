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
	"sort"

	"github.com/akrzos/kubeSize/internal/kube"
	"github.com/akrzos/kubeSize/internal/output"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

var namespaceCmd = &cobra.Command{
	Use:   "namespace",
	Short: "Get namespace size",
	Long:  `Get namespace size and capacity metrics`,
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

		namespaceCapacityData := make(map[string]*output.NamespaceCapacityData)
		sortedNamespaceNames := make([]string, 0, len(namespaces.Items))

		for _, v := range namespaces.Items {
			sortedNamespaceNames = append(sortedNamespaceNames, v.Name)

			namespaceFieldSelector, err := fields.ParseSelector("metadata.namespace=" + v.Name)
			if err != nil {
				return errors.Wrap(err, "failed to create fieldSelector")
			}
			namespacePodsList, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{FieldSelector: namespaceFieldSelector.String()})

			nonTerminatedFieldSelector, err := fields.ParseSelector("metadata.namespace=" + v.Name + ",status.phase!=" + string(corev1.PodSucceeded) + ",status.phase!=" + string(corev1.PodFailed))
			if err != nil {
				return errors.Wrap(err, "failed to create fieldSelector")
			}
			totalNonTermPodsList, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{FieldSelector: nonTerminatedFieldSelector.String()})

			newNamespaceData := new(output.NamespaceCapacityData)
			newNamespaceData.TotalPodCount = len(namespacePodsList.Items)
			newNamespaceData.TotalNonTermPodCount = len(totalNonTermPodsList.Items)

			for _, pod := range totalNonTermPodsList.Items {
				for _, container := range pod.Spec.Containers {
					newNamespaceData.TotalRequestsCPU.Add(*container.Resources.Requests.Cpu())
					newNamespaceData.TotalLimitsCPU.Add(*container.Resources.Limits.Cpu())
					newNamespaceData.TotalRequestsMemory.Add(*container.Resources.Requests.Memory())
					newNamespaceData.TotalLimitsMemory.Add(*container.Resources.Limits.Memory())
				}
			}
			namespaceCapacityData[v.Name] = newNamespaceData
		}

		sort.Strings(sortedNamespaceNames)

		displayReadable, _ := cmd.Flags().GetBool("readable")

		output.DisplayNamespaceData(namespaceCapacityData, sortedNamespaceNames, displayReadable)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(namespaceCmd)
}
