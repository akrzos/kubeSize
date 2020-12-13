package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

		for _, v := range nodes.Items {
			totalCapacityPods.Add(*v.Status.Capacity.Pods())
			totalCapacityCPU.Add(*v.Status.Capacity.Cpu())
			totalCapacityMemory.Add(*v.Status.Capacity.Memory())
			totalAllocatablePods.Add(*v.Status.Allocatable.Pods())
			totalAllocatableCPU.Add(*v.Status.Allocatable.Cpu())
			totalAllocatableMemory.Add(*v.Status.Allocatable.Memory())
		}

		var capacityMemoryGiB = float64(totalCapacityMemory.Value()) / 1024 / 1024 / 1024
		var allocatableMemoryGiB = float64(totalAllocatableMemory.Value()) / 1024 / 1024 / 1024

		fmt.Println("Total Cluster Capacity and Allocatable")
		fmt.Printf("             PODS        CPU (cores)       Memory (GiB)\n")
		fmt.Printf("Capacity     ")
		fmt.Printf("%-12s", &totalCapacityPods)
		fmt.Printf("%-18s", &totalCapacityCPU)
		fmt.Printf("%-16.1f\n", capacityMemoryGiB)
		fmt.Printf("Allocatable  ")
		fmt.Printf("%-12s", &totalAllocatablePods)
		fmt.Printf("%-18s", &totalAllocatableCPU)
		fmt.Printf("%-16.1f\n", allocatableMemoryGiB)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(clusterCmd)
}
