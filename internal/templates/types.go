package templates

import (
	v1 "github.com/kloudlite/wireguard/api/v1"
)

type ParamsWgServerConf struct {
	ServerIP         string
	ServerPrivateKey string
	PodCIDR          string
	WithUDP2Raw      bool
	Peers            []v1.Peer
}

type ParamsWgPeerConf struct {
	Name       string
	IP         string
	PrivateKey string

	DNS           string
	DNSLocalhosts []string
	Peers         []v1.Peer
}

type ParamsServerDeploymentSpec struct {
	PodLabels     map[string]string
	Wg0Conf       string
	KubeDNSSvcIP  string
	DNSLocalhosts []string
}

type ParamsServerServiceSpec struct {
	SelectorLabels map[string]string
	ServiceType    string
	Port           uint16
}
