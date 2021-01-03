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
package kube

import (
	"github.com/pkg/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

func CreateClientSet(KubernetesConfigFlags *genericclioptions.ConfigFlags) (*kubernetes.Clientset, error) {
	config, err := KubernetesConfigFlags.ToRESTConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read kubeconfig")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create clientset")
	}

	return clientset, nil
}
