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
		var totalPodCount, totalNonTermPodCount = 0, 0

		for _, v := range nodes.Items {
			totalCapacityPods.Add(*v.Status.Capacity.Pods())
			totalCapacityCPU.Add(*v.Status.Capacity.Cpu())
			totalCapacityMemory.Add(*v.Status.Capacity.Memory())
			totalAllocatablePods.Add(*v.Status.Allocatable.Pods())
			totalAllocatableCPU.Add(*v.Status.Allocatable.Cpu())
			totalAllocatableMemory.Add(*v.Status.Allocatable.Memory())
		}

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

		fmt.Printf("Total Cluster Capacity\n")

		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 5, 1, ' ', 0)
		fmt.Fprintln(w, "PODS\t\tPODS\t\tCPU\t\tCPU\t\tMEMORY\t\tMEMORY")
		fmt.Fprintln(w, "Capacity\tAllocatable\tTotal\tNon-Term\tCapacity\tAllocatable\tRequests\tLimits\tCapacity\tAllocatable\tRequests\tLimits")
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
