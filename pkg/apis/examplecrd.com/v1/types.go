
package v1

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)
// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

//CronTab Specification
type CronTab struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec CronTabDeploymentSpec  `json:"spec"`
	Status CronTabDeploymentStatus `json:"status"`
}

//CronTabSpec Specification
type CronTabDeploymentSpec struct {
	Replicas int32              `json:"replicas"`
	Template CronTabPodTemplate `json:"template"`
	DeploymentName string `json:"deploymentName"`
}
//Status for CustomDeployment
type CronTabDeploymentStatus struct {
	AvailableReplicas   int32 `json:"availableReplicas"`
	CreatingReplicas    int32 `json:"creatingReplicas"`
	TerminatingReplicas int32 `json:"terminatingReplicas"`
}
//CronTabPodTemplate Specification
type CronTabPodTemplate struct {
	metav1.ObjectMeta	`json:"metadata,omitempty"`
	Spec apiv1.PodSpec	`json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

//CronTabDeploymentList
type CronTabList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items []CronTabList `json:"items"`
}
