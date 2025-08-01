/*
Copyright 2025.

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

package v1

import (
	rApi "github.com/kloudlite/kloudlite/operator/toolkit/reconciler"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Peer struct {
	Name       string   `json:"name,omitempty"`
	IP         *string  `json:"ip,omitempty"`
	PrivateKey *string  `json:"privateKey,omitempty"`
	PublicKey  *string  `json:"publicKey,omitempty"`
	AllowedIPs []string `json:"allowedIPs,omitempty"`
	Endpoint   *string  `json:"endpoint,omitempty"`
}

type ProxyPort struct {
	Peer               string `json:"peer"`
	corev1.ServicePort `json:",inline"`
}

// ServerSpec defines the desired state of Server.
type ServerSpec struct {
	TargetNamespace string `json:"targetNamespace,omitempty"`

	// +kubebuilder:default="10.13.0.1"
	IP *string `json:"ip,omitempty"`

	// +kubebuilder:default="10.13.0.0/24"
	CIDR *string `json:"cidr,omitempty"`

	PrivateKey *string `json:"privateKey,omitempty"`
	PublicKey  *string `json:"publicKey,omitempty"`

	Endpoint string `json:"endpoint,omitempty"`

	// KeepAlive duration in seconds, defaults to 0 (disabled)
	KeepAlive int32 `json:"keepAlive,omitempty"`

	Expose Expose `json:"expose,omitempty"`

	DNS DNS `json:"dns,omitempty"`

	Peers []Peer `json:"peers,omitempty"`

	Proxy []ProxyPort `json:"proxy,omitempty"`
}

type Expose struct {
	// +kubebuilder:default=NodePort
	// +kubebuilder:validation:Enum=ClusterIP;NodePort;LoadBalancer
	ServiceType string `json:"serviceType,omitempty"`

	// +kubebuilder:default=31820
	Port uint16 `json:"port,omitempty"`
}

type DNS struct {
	Localhosts []string `json:"localhosts,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Seen,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.operator\\.kloudlite\\.io\\/checks",name=Checks,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.operator\\.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// Server is the Schema for the servers API.
type Server struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServerSpec  `json:"spec,omitempty"`
	Status rApi.Status `json:"status,omitempty"`
}

func (s *Server) EnsureGVK() {
	if s != nil {
		s.SetGroupVersionKind(GroupVersion.WithKind("Server"))
	}
}

func (s *Server) GetStatus() *rApi.Status {
	return &s.Status
}

func (s *Server) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (s *Server) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

// +kubebuilder:object:root=true

// ServerList contains a list of Server.
type ServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Server `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Server{}, &ServerList{})
}
