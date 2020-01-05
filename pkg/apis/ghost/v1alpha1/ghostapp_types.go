// Copyright 2019 Fossil Dev
//
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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// GhostDatabaseConnectionSpec defines ghost database connection.
type GhostDatabaseConnectionSpec struct {
	// sqlite filename.
	// +optional
	Filename string `json:"filename,omitempty"`
	// mysql host
	// +optional
	Host string `json:"host,omitempty"`
	// mysql port
	// +optional
	Port intstr.IntOrString `json:"port,omitempty"`
	// mysql database user
	// +optional
	User string `json:"user,omitempty"`
	// mysql database password of user
	// +optional
	Password string `json:"password,omitempty"`
	// mysql database name
	// +optional
	Database string `json:"database,omitempty"`
}

type GhostServerSpec struct {
	Host string             `json:"host"`
	Port intstr.IntOrString `json:"port"`
}

// GhostDatabaseSpec defines ghost database config.
// https://ghost.org/docs/concepts/config/#database
type GhostDatabaseSpec struct {
	// Client is ghost database client.
	// +kubebuilder:validation:Enum=sqlite3;mysql
	Client string `json:"client"`
	// +optional
	Connection GhostDatabaseConnectionSpec `json:"connection"`
}

// GhostConfigSpec defines related ghost configuration based on https://ghost.org/docs/concepts/config
// TODO (prksu): we need support all ghost configuration since we reference this spec as ghost config too.
// TODO (prksu): move ghost config to another file.
type GhostConfigSpec struct {
	URL string `json:"url"`

	Database GhostDatabaseSpec `json:"database"`
	// +optional
	Server GhostServerSpec `json:"server"`
}

// GhostPersistentSpec defines peristent volume
type GhostPersistentSpec struct {
	Enabled bool `json:"enabled"`
	// If defined, will create persistentVolumeClaim with spesific storageClass name.
	// If undefined (the default) or set to null, no storageClassName spec is set, choosing the default provisioner.
	// +nullable
	StorageClass *string `json:"storageClass,omitempty"`
	// size of storage
	Size resource.Quantity `json:"size"`
}

// GhostIngressTLSSpec defines ingress tls
type GhostIngressTLSSpec struct {
	Enabled    bool   `json:"enabled"`
	SecretName string `json:"secretName"`
}

// GhostIngressSpec defines ingress
type GhostIngressSpec struct {
	Enabled bool `json:"enabled"`
	// +optional
	// +listType=set
	Hosts []string `json:"hosts,omitempty"`
	// +optional
	TLS GhostIngressTLSSpec `json:"tls,omitempty"`
	// Additional annotations passed to ".metadata.annotations" in networking.k8s.io/ingress object.
	// This is useful for configuring ingress through annotation field like: ingress-class, static-ip, etc
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

// GhostAppSpec defines the desired state of GhostApp
// +k8s:openapi-gen=true
type GhostAppSpec struct {
	// Ghost deployment repicas
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`
	// Ghost container image, by default using latest ghost image from docker hub registry.
	// NOTE: This operator only support ghost image from docker official image. https://hub.docker.com/_/ghost/
	// +optional
	Image string `json:"image,omitempty"`
	// Ghost configuration. This field will be written as ghost configuration. Saved in configmap and mounted
	// in /etc/ghost/config/config.json and symlinked to /var/lib/ghost/config.production.json
	Config GhostConfigSpec `json:"config"`
	// +optional
	Persistent GhostPersistentSpec `json:"persistent,omitempty"`
	// +optional
	Ingress GhostIngressSpec `json:"ingress,omitempty"`
}

// GhostAppPhaseType represents the current phase of GhostApp instances
// +k8s:openapi-gen=true
type GhostAppPhaseType string

const (
	// GhostAppPhaseCreating indicates that the GhostApp is under provisioning
	// +k8s:openapi-gen=true
	GhostAppPhaseCreating GhostAppPhaseType = "Creating"

	// GhostAppPhaseRunning indicates that the GhostApp is ready and running
	// +k8s:openapi-gen=true
	GhostAppPhaseRunning GhostAppPhaseType = "Running"

	// GhostAppPhaseUpdating indicates that the GhostApp is under updating
	// +k8s:openapi-gen=true
	GhostAppPhaseUpdating GhostAppPhaseType = "Updating"

	// GhostAppPhaseFailure indicates that the GhostApp failed to be provisioned
	// +k8s:openapi-gen=true
	GhostAppPhaseFailure GhostAppPhaseType = "Failure"
)

// GhostAppStatus defines the observed state of GhostApp
// +k8s:openapi-gen=true
type GhostAppStatus struct {
	Replicas int32 `json:"replicas,omitempty"`
	// Represents the latest available observations of a ghostapp current state.
	Phase GhostAppPhaseType `json:"phase,omitempty"`

	Reason string `json:"reason,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GhostApp is the Schema for the ghostapps API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=ghostapps,scope=Namespaced
// +kubebuilder:printcolumn:name="replicas",type="string",JSONPath=".status.replicas"
// +kubebuilder:printcolumn:name="phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="age",type="date",JSONPath=".metadata.creationTimestamp"
type GhostApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GhostAppSpec   `json:"spec,omitempty"`
	Status GhostAppStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GhostAppList contains a list of GhostApp
type GhostAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GhostApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GhostApp{}, &GhostAppList{})
}
