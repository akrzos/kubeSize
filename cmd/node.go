package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/sets"
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

		fmt.Printf("Node Capacity\n")

		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 5, 1, ' ', 0)
		fmt.Fprintln(w, "NAME\tROLES\tPODS\t\tPODS\t\tCPU\t\tCPU\t\tMEMORY\t\tMEMORY")
		fmt.Fprintln(w, "\t\tCapacity\tAllocatable\tTotal\tNon-Term\tCapacity\tAllocatable\tRequests\tLimits\tCapacity\tAllocatable\tRequests\tLimits")

		for _, v := range nodes.Items {

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
			totalPodCount := len(nodePodsList.Items)

			nonTerminatedFieldSelector, err := fields.ParseSelector("spec.nodeName=" + v.Name + ",status.phase!=" + string(corev1.PodSucceeded) + ",status.phase!=" + string(corev1.PodFailed))
			if err != nil {
				return errors.Wrap(err, "failed to create fieldSelector")
			}
			totalNonTermPodsList, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{FieldSelector: nonTerminatedFieldSelector.String()})
			nonTerminatedPodCount := len(totalNonTermPodsList.Items)

			var totalCPURequests, totalCPULimits, totalMemoryRequests, totalMemoryLimits resource.Quantity

			for _, pod := range totalNonTermPodsList.Items {
				for _, container := range pod.Spec.Containers {
					totalCPURequests.Add(*container.Resources.Requests.Cpu())
					totalCPULimits.Add(*container.Resources.Limits.Cpu())
					totalMemoryRequests.Add(*container.Resources.Requests.Memory())
					totalMemoryLimits.Add(*container.Resources.Limits.Memory())
				}
			}

			fmt.Fprintf(w, "%s\t", v.Name)
			fmt.Fprintf(w, "%s\t", strings.Join(roles.List(), ","))
			fmt.Fprintf(w, "%s\t%s\t", v.Status.Capacity.Pods(), v.Status.Allocatable.Pods())
			fmt.Fprintf(w, "%d\t%d\t", totalPodCount, nonTerminatedPodCount)
			fmt.Fprintf(w, "%s\t%s\t", v.Status.Capacity.Cpu(), v.Status.Allocatable.Cpu())
			fmt.Fprintf(w, "%s\t%s\t", &totalCPURequests, &totalCPULimits)
			fmt.Fprintf(w, "%.1f\t%.1f\t", float64(v.Status.Capacity.Memory().Value())/1024/1024/1024, float64(v.Status.Allocatable.Memory().Value())/1024/1024/1024)
			fmt.Fprintf(w, "%s\t%s\t\n", &totalMemoryRequests, &totalMemoryLimits)

		}
		w.Flush()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(nodeCmd)
}
