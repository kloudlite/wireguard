package templates

import (
	v1 "github.com/kloudlite/wireguard/api/v1"
	corev1 "k8s.io/api/core/v1"
)

type ParamsWgServerConf struct {
	ServerIP         string
	ServerPrivateKey string
	PodCIDR          string
	WithUDP2Raw      bool
	Peers            []v1.Peer
	KeepAlive        int32
}

type ParamsWgPeerConf struct {
	Name       string
	IP         string
	PrivateKey string

	DNS           string
	DNSLocalhosts []string
	Peers         []v1.Peer

	ServerPeer v1.Peer
}

type ParamsServerDeploymentSpec struct {
	PodLabels map[string]string
	Wg0Conf   string

	WgDNSTemplateParams

	WgProxyTemplateParams
}

type WgServiceSpecParams struct {
	SelectorLabels map[string]string
	ServiceType    string
	Port           uint16

	Proxy []v1.ProxyPort
}

type WgDNSTemplateParams struct {
	KubeDNSSvcIP  string
	DNSLocalhosts []string
}

type PortMapping struct {
	Protocol   corev1.Protocol
	Port       int32
	TargetHost string
	TargetPort int32
}

type WgProxyTemplateParams struct {
	PortMappings []PortMapping
}
