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

package v1alpha1

import (
	sesame_api_v1 "github.com/projectsesame/sesame/apis/projectsesame/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SesameDeploymentSpec defines the parameters of how a Sesame
// instance should be configured.
type SesameDeploymentSpec struct {
	// Replicas is the desired number of Sesame replicas. If unset,
	// defaults to 2.
	//
	// +kubebuilder:default=2
	// +kubebuilder:validation:Minimum=0
	Replicas int32 `json:"replicas,omitempty"`

	// Config is the config that the instances of Sesame are to utilize.
	Config SesameConfigurationSpec `json:"config"`
}

// SesameDeploymentStatus defines the observed state of a SesameDeployment resource.
type SesameDeploymentStatus struct {
	// Conditions contains the current status of the Sesame resource.
	//
	// Sesame will update a single condition, `Valid`, that is in normal-true polarity.
	//
	// Sesame will not modify any other Conditions set in this block,
	// in case some other controller wants to add a Condition.
	//
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []sesame_api_v1.DetailedCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=sesamedeploy

// SesameDeployment is the schema for a Sesame Deployment.
type SesameDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SesameDeploymentSpec   `json:"spec,omitempty"`
	Status SesameDeploymentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SesameDeploymentList contains a list of Sesame Deployment resources.
type SesameDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SesameDeployment `json:"items"`
}
