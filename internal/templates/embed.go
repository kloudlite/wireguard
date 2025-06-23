package templates

import (
	"embed"
	"path/filepath"

	"github.com/kloudlite/kloudlite/operator/toolkit/templates"
)

//go:embed *
var templatesDir embed.FS

type templateFile string

const (
	ServerDeploymentSpec templateFile = "./server-deployment-spec.yml.tpl"
	ServerServiceSpec    templateFile = "./server-service-spec.yml.tpl"

	WgServerConf templateFile = "./wg-server.conf.tpl"
	WgPeerConf   templateFile = "./wg-peer.conf.tpl"
)

func Read(t templateFile) ([]byte, error) {
	return templatesDir.ReadFile(filepath.Join(string(t)))
}

var ParseBytes = templates.ParseBytes
