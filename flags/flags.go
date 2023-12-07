package flags

import (
	"fmt"

	"github.com/cprobe/cprobe/lib/fileutil"
)

var (
	ConfigDirectory string
)

func Check() error {
	if ConfigDirectory == "" {
		return fmt.Errorf("-conf.d is empty")
	}

	if !fileutil.IsExist(ConfigDirectory) {
		return fmt.Errorf("-conf.d %s does not exist", ConfigDirectory)
	}

	if !fileutil.IsDir(ConfigDirectory) {
		return fmt.Errorf("-conf.d %s is not a directory", ConfigDirectory)
	}

	return nil
}
