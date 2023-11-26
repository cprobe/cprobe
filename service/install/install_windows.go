package install

import (
	"github.com/kardianos/service"
)

const (
	ServiceName = "cprobe"
)

var (
	serviceConfig = &service.Config{
		// 服务显示名称
		Name: ServiceName,
		// 服务名称
		DisplayName: "cprobe",
		// 服务描述
		Description: "Frankenstein made up of vmagent and exporters",
	}
)

func ServiceConfig() (*service.Config, error) {
	return serviceConfig, nil
}
