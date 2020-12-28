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
	"text/tabwriter"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
)

type NamespaceCapacityData struct {
	totalPodCount        int
	totalNonTermPodCount int
	totalRequestsCPU     resource.Quantity
	totalLimitsCPU       resource.Quantity
	totalRequestsMemory  resource.Quantity
	totalLimitsMemory    resource.Quantity
}

var namespaceCmd = &cobra.Command{
	Use:   "namespace",
	Short: "Get namespace size",
	Long:  `Get namespace size and capacity metrics`,
	RunE: func(cmd *cobra.Command, args []string) error {

		config, err := KubernetesConfigFlags.ToRESTConfig()
		if err != nil {
			return errors.Wrap(err, "failed to read kubeconfig")
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return errors.Wrap(err, "failed to create clientset")
		}

		namespaces, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list namespaces")
		}

		namespaceCapacityData := make(map[string]*NamespaceCapacityData)
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

			newNamespaceData := new(NamespaceCapacityData)
			newNamespaceData.totalPodCount = len(namespacePodsList.Items)
			newNamespaceData.totalNonTermPodCount = len(totalNonTermPodsList.Items)

			for _, pod := range totalNonTermPodsList.Items {
				for _, container := range pod.Spec.Containers {
					newNamespaceData.totalRequestsCPU.Add(*container.Resources.Requests.Cpu())
					newNamespaceData.totalLimitsCPU.Add(*container.Resources.Limits.Cpu())
					newNamespaceData.totalRequestsMemory.Add(*container.Resources.Requests.Memory())
					newNamespaceData.totalLimitsMemory.Add(*container.Resources.Limits.Memory())
				}
			}
			namespaceCapacityData[v.Name] = newNamespaceData
		}

		sort.Strings(sortedNamespaceNames)

		displayReadable, _ := cmd.Flags().GetBool("readable")

		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 5, 1, ' ', 0)
		if displayReadable == true {
			fmt.Fprintln(w, "NAMESPACE\tPODS\t\tCPU (cores)\t\tMEMORY (GiB)\t\t")
		} else {
			fmt.Fprintln(w, "NAMESPACE\tPODS\t\tCPU\t\tMEMORY\t\t")
		}
		fmt.Fprintln(w, "\tTotal\tNon-Term\tRequests\tLimits\tRequests\tLimits")
		for _, k := range sortedNamespaceNames {
			fmt.Fprintf(w, "%s\t", k)
			fmt.Fprintf(w, "%d\t%d\t", namespaceCapacityData[k].totalPodCount, namespaceCapacityData[k].totalNonTermPodCount)
			if displayReadable == true {
				fmt.Fprintf(w, "%.1f\t%.1f\t", float64(namespaceCapacityData[k].totalRequestsCPU.MilliValue())/1000, float64(namespaceCapacityData[k].totalLimitsCPU.MilliValue())/1000)
				fmt.Fprintf(w, "%.1f\t%.1f\t\n", float64(namespaceCapacityData[k].totalRequestsMemory.Value())/1024/1024/1024, float64(namespaceCapacityData[k].totalLimitsMemory.Value())/1024/1024/1024)
			} else {
				fmt.Fprintf(w, "%s\t%s\t", &namespaceCapacityData[k].totalRequestsCPU, &namespaceCapacityData[k].totalLimitsCPU)
				fmt.Fprintf(w, "%s\t%s\t\n", &namespaceCapacityData[k].totalRequestsMemory, &namespaceCapacityData[k].totalLimitsMemory)
			}
		}
		w.Flush()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(namespaceCmd)
}
