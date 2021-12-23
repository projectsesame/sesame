// Copyright Project Contour Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// NewCoreClient returns a new Kubernetes core API client using
// the supplied kubeconfig path, or the cluster environment
// variables if inCluster is true.
func NewCoreClient(kubeconfig string, inCluster bool) (*kubernetes.Clientset, error) {
	config, err := NewRestConfig(kubeconfig, inCluster)
	if err != nil {
		return nil, err
	}

	coreClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return coreClient, nil
}

// NewRestConfig returns a *rest.Config for the supplied kubeconfig
// path, or the cluster environment variables if inCluster is true.
func NewRestConfig(kubeconfig string, inCluster bool) (*rest.Config, error) {
	if kubeconfig != "" && !inCluster {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}
