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

	"github.com/akrzos/kubeSize/internal/kube"
	"github.com/akrzos/kubeSize/internal/output"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var sizeCmd = &cobra.Command{
	Use:     "size",
	Aliases: []string{"s"},
	Short:   "Get cluster size data",
	Long:    `Get counts of many Kubernetes objects in a cluster`,
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

		clusterSizeData := new(output.ClusterSizeData)

		// Cluster APIs
		namespaces, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list namespaces")
		}
		nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list nodes")
		}
		persistentVolumes, err := clientset.CoreV1().PersistentVolumes().List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list persistent volumes")
		}
		serviceAccounts, err := clientset.CoreV1().ServiceAccounts("").List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list service accounts")
		}
		clusterRoles, err := clientset.RbacV1().ClusterRoles().List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list cluster roles")
		}
		clusterRoleBindings, err := clientset.RbacV1().ClusterRoleBindings().List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list cluster role bindings")
		}
		roles, err := clientset.RbacV1().Roles("").List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list roles")
		}
		roleBindings, err := clientset.RbacV1().RoleBindings("").List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list role bindings")
		}
		resourceQuotas, err := clientset.CoreV1().ResourceQuotas("").List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list resourcequotas")
		}
		networkPolicy, err := clientset.NetworkingV1().NetworkPolicies("").List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list networkpolicy")
		}

		// Workloads APIs
		pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list pods")
		}
		replicaSets, err := clientset.AppsV1().ReplicaSets("").List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list replicasets")
		}
		replicationControllers, err := clientset.CoreV1().ReplicationControllers("").List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list replication controllers")
		}
		deployments, err := clientset.AppsV1().Deployments("").List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list deployments")
		}
		daemonsets, err := clientset.AppsV1().DaemonSets("").List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list daemonsets")
		}
		statefulSets, err := clientset.AppsV1().StatefulSets("").List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list statefulsets")
		}
		cronJobs, err := clientset.BatchV1beta1().CronJobs("").List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list jobs")
		}
		jobs, err := clientset.BatchV1().Jobs("").List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list jobs")
		}

		// Service APIs
		endPoints, err := clientset.CoreV1().Endpoints("").List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list end points")
		}
		services, err := clientset.CoreV1().Services("").List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list services")
		}
		ingresses, err := clientset.NetworkingV1beta1().Ingresses("").List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list ingresses")
		}

		// Config And Storage APIs
		configmaps, err := clientset.CoreV1().ConfigMaps("").List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list configmaps")
		}
		secrets, err := clientset.CoreV1().Secrets("").List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list secrets")
		}
		persistentVolumeClaims, err := clientset.CoreV1().PersistentVolumeClaims("").List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list persistentvolumesclaims")
		}
		storageClasses, err := clientset.StorageV1().StorageClasses().List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list storageclasses")
		}
		volumeAttachments, err := clientset.StorageV1().VolumeAttachments().List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list storageclasses")
		}

		// Metadata APIs
		events, err := clientset.CoreV1().Events("").List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list events")
		}
		limitRanges, err := clientset.CoreV1().LimitRanges("").List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list limitrange")
		}
		podDisruptionBudget, err := clientset.PolicyV1beta1().PodDisruptionBudgets("").List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list poddisruptionbudget")
		}
		podSecurityPolicy, err := clientset.PolicyV1beta1().PodSecurityPolicies().List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list podsecuritypolicy")
		}

		// Cluster APIs
		clusterSizeData.Namespace = len(namespaces.Items)
		clusterSizeData.Node = len(nodes.Items)
		clusterSizeData.PersistentVolume = len(persistentVolumes.Items)
		clusterSizeData.ServiceAccount = len(serviceAccounts.Items)
		clusterSizeData.ClusterRole = len(clusterRoles.Items)
		clusterSizeData.ClusterRoleBinding = len(clusterRoleBindings.Items)
		clusterSizeData.Role = len(roles.Items)
		clusterSizeData.RoleBinding = len(roleBindings.Items)
		clusterSizeData.ResourceQuota = len(resourceQuotas.Items)
		clusterSizeData.NetworkPolicy = len(networkPolicy.Items)

		// Workloads APIs
		for _, pod := range pods.Items {
			for range pod.Spec.Containers {
				clusterSizeData.Container++
			}
		}
		clusterSizeData.Pod = len(pods.Items)
		clusterSizeData.ReplicaSet = len(replicaSets.Items)
		clusterSizeData.ReplicaController = len(replicationControllers.Items)
		clusterSizeData.Deployment = len(deployments.Items)
		clusterSizeData.Daemonset = len(daemonsets.Items)
		clusterSizeData.StatefulSet = len(statefulSets.Items)
		clusterSizeData.CronJob = len(cronJobs.Items)
		clusterSizeData.Job = len(jobs.Items)

		// Service APIs
		clusterSizeData.EndPoints = len(endPoints.Items)
		clusterSizeData.Service = len(services.Items)
		clusterSizeData.Ingress = len(ingresses.Items)

		// Config And Storage APIs
		clusterSizeData.Configmap = len(configmaps.Items)
		clusterSizeData.Secret = len(secrets.Items)
		clusterSizeData.PersistentVolumeClaim = len(persistentVolumeClaims.Items)
		clusterSizeData.StorageClass = len(storageClasses.Items)
		clusterSizeData.VolumeAttachment = len(volumeAttachments.Items)

		// Metadata APIs
		clusterSizeData.Event = len(events.Items)
		clusterSizeData.LimitRange = len(limitRanges.Items)
		clusterSizeData.PodDisruptionBudget = len(podDisruptionBudget.Items)
		clusterSizeData.PodSecurityPolicy = len(podSecurityPolicy.Items)

		displayNoHeaders, _ := cmd.Flags().GetBool("no-headers")

		displayFormat, _ := cmd.Flags().GetString("output")

		output.DisplayClusterSizeData(*clusterSizeData, !displayNoHeaders, displayFormat)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(sizeCmd)
}
