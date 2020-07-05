package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KpackDeploySpec defines the desired state of KpackDeploy
type KpackDeploySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// ImageName refers to the kpack Image in this namespace that will be deployed to target environments
	ImageName string `json:"imageName"`

	// Target specifies the location of the K8S Deployment YAML that will control deployment in the target environments
	Target Target `json:"target"`
}

type Target struct {
	// Git describes the location of a repo for storing deployment files in a GitOps environment
	Git Git `json:"git"`
}

// +kubebuilder:validation:Enum=commit;pullrequest
type WriteMethod string

const GIT_COMMIT WriteMethod = "commit"
const GIT_PULL_REQUEST WriteMethod = "pullrequest"

type Git struct {
	// URL of the Git Repo
	Url string `json:"url"`
	// Directories of the supported environments, as per Kustomize layout
	Paths []string `json:"paths"`
	// Branch to commit to
	Branch string `json:"branch"`
	// Filename of the deployment YAML
	DeploymentFile string `json:"deploymentFile"`
	// Git Access token
	AccessToken string `json:"accessToken"`
	// Method of writing changes (commit, pullrequest)
	WriteMethod WriteMethod `json:"writeMethod"`
}

// KpackDeployStatus defines the observed state of KpackDeploy
type KpackDeployStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// Deployment Images are the container images named in the K8S Deployment descriptor for each configured target environment
	DeploymentImages map[string]string `json:"deploymentImages"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KpackDeploy is the Schema for the kpackdeploys API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=kpackdeploys,scope=Namespaced
// +kubebuilder:storageversion
type KpackDeploy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KpackDeploySpec   `json:"spec,omitempty"`
	Status KpackDeployStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KpackDeployList contains a list of KpackDeploy
type KpackDeployList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KpackDeploy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KpackDeploy{}, &KpackDeployList{})
}
