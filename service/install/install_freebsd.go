package install

import (
	"fmt"
	"os"

	"github.com/cprobe/cprobe/flags"
	"github.com/kardianos/service"
)

const (
	// freebsd的服务名中间不能有"-"
	ServiceName = "cprobe"

	SysvScript = `#!/bin/sh
#
# PROVIDE: {{.Name}}
# REQUIRE: networking syslog
# KEYWORD:
# Add the following lines to /etc/rc.conf to enable the {{.Name}}:
#
# {{.Name}}_enable="YES"
#
. /etc/rc.subr
name="{{.Name}}"
rcvar="{{.Name}}_enable"
command="{{.Path}}"
pidfile="/var/run/$name.pid"
start_cmd="/opt/cprobe/cprobe %s --loggerLevel=INFO -loggerTimezone=Asia/Shanghai"
load_rc_config $name
run_rc_command "$1"
`
)

var (
	serviceConfig = &service.Config{
		// 服务显示名称
		Name: ServiceName,
		// 服务名称
		DisplayName: "cprobe",
		// 服务描述
		Description: "Frankenstein made up of vmagent and exporters",
		// Option: service.KeyValue{
		// 	"SysvScript": SysvScript,
		// },
	}
)

func ServiceConfig() (*service.Config, error) {
	if flags.ConfigDirectory == "" {
		return nil, fmt.Errorf("-conf.d is blank")
	}

	absPath, err := filepath.Abs(flags.ConfigDirectory)
	if err != nil {
		return nil, err
	}

	serviceConfig.Option = service.KeyValue{
		"SysvScript": fmt.Sprintf(SysvScript, fmt.Sprintf("-conf.d=%s", absPath))
	}

	return serviceConfig, nil
}
