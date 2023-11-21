package runner

import (
	"fmt"
	"os"

	"github.com/cprobe/cprobe/lib/logger"
	"go.uber.org/automaxprocs/maxprocs"
)

func Noop(string, ...interface{}) {}

func Init() {
	maxprocs.Set(maxprocs.Logger(Noop))
}

func PrintRuntime() {
	maxprocs.Set(maxprocs.Logger(Noop))

	hostname, err := os.Hostname()
	if err != nil {
		hostname = fmt.Sprintf("cannot get hostname: %s", err)
	}

	logger.Infof("hostname: %s", hostname)
	logger.Infof("runtime.fd_limits: %s", FdLimits())
	logger.Infof("runtime.vm_limits: %s", VMLimits())
}
