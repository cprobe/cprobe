package install

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cprobe/cprobe/flags"
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
	ov, err := os.Executable()
	if err == nil {
		if len(filepath.Dir(ov)) != 0 {
			serviceConfig.WorkingDirectory = filepath.Dir(ov)
		}
	} else {
		return nil, fmt.Errorf("failed to get executable path: %s", err)
	}

	if flags.ConfigDirectory == "" {
		return nil, fmt.Errorf("-conf.d is blank")
	}

	absPath, err := filepath.Abs(flags.ConfigDirectory)
	if err != nil {
		return nil, err
	}

	serviceConfig.Arguments = append(serviceConfig.Arguments, "-conf.d", absPath)
	return serviceConfig, nil
}
