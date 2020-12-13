package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var nodeCmd = &cobra.Command{
	Use:           "node",
	Short:         "Get Node Capacity",
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

		fmt.Println("Node - Capacity : Allocatable")
		fmt.Printf("Node Name                 PODS          CPU (cores)  Memory (GiB)\n")
		for _, v := range nodes.Items {
			fmt.Printf("%-25s", v.Name)
			fmt.Printf(" %-5s", v.Status.Capacity.Pods())
			fmt.Printf(" : %-5s", v.Status.Allocatable.Pods())
			fmt.Printf(" %-4s", v.Status.Capacity.Cpu())
			fmt.Printf(" : %-4s", v.Status.Allocatable.Cpu())
			fmt.Printf("  %-6.1f", float64(v.Status.Capacity.Memory().Value())/1024/1024/1024)
			fmt.Printf(" : %-6.1f\n", float64(v.Status.Allocatable.Memory().Value())/1024/1024/1024)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(nodeCmd)
}
