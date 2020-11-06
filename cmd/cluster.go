package cmd

import (
	"fmt"

	// "github.com/akrzos/k8sCube/pkg/logger"
	// "github.com/akrzos/k8sCube/pkg/plugin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	// "k8s.io/cli-runtime/pkg/genericclioptions"
)

// var (
// 	KubernetesConfigFlags *genericclioptions.ConfigFlags
// )

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
		fmt.Printf("TODO: Cluster Capacity\n")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(clusterCmd)
}
