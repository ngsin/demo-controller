package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DemoAPI is a specification for a Demo resource
type DemoAPI struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DemoAPISpec   `json:"spec"`
	Status DemoAPIStatus `json:"status"`
}

// DemoAPISpec is the spec for a Demo resource
type DemoAPISpec struct {
	DeploymentName string `json:"deploymentName"`
	Replicas       *int32 `json:"replicas"`
	Image          string `json:"image"`
}

// DemoAPIStatus is the status for a Demo resource
type DemoAPIStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DemoAPIList is a list of Demo resources
type DemoAPIList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []DemoAPI `json:"items"`
}
