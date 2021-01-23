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

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	KubernetesConfigFlags *genericclioptions.ConfigFlags
)

var rootCmd = &cobra.Command{
	Use:           "capacity",
	Short:         "Get cluster size and capacity",
	Long:          `Exposes size and capacity data for Kubernetes clusters`,
	SilenceErrors: true,
	SilenceUsage:  true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	KubernetesConfigFlags = genericclioptions.NewConfigFlags(false)
	KubernetesConfigFlags.AddFlags(rootCmd.PersistentFlags())
	rootCmd.PersistentFlags().BoolP("default-format", "d", false, "Use default format of displaying resource quantities")
	rootCmd.PersistentFlags().BoolP("no-headers", "", false, "No headers in table output format")
	rootCmd.PersistentFlags().StringP("output", "o", "table", "Output format. One of: table|json|yaml")
}
