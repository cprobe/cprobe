package probe

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	"github.com/cprobe/cprobe/lib/fileutil"
	"github.com/pkg/errors"
)

var (
	probeDir = flag.String("probe.dir", "probe.d", "Filepath to probe.d .")
)

func checkFlag() error {
	if *probeDir == "" {
		return fmt.Errorf("-probe.dir is empty")
	}

	if !fileutil.IsExist(*probeDir) {
		return fmt.Errorf("-probe.dir %s does not exist", *probeDir)
	}

	if !fileutil.IsDir(*probeDir) {
		return fmt.Errorf("-probe.dir %s is not a directory", *probeDir)
	}

	return nil
}

// Start starts the probe goroutines.
func Start(ctx context.Context) error {
	if err := checkFlag(); err != nil {
		return err
	}

	pluginDirs, err := fileutil.DirsUnder(*probeDir)
	if err != nil {
		return errors.Wrap(err, "cannot list plugin dirs")
	}

	if len(pluginDirs) == 0 {
		return fmt.Errorf("no plugin dirs found under %s", *probeDir)
	}

	for i := 0; i < len(pluginDirs); i++ {
		if err := startPlugin(ctx, pluginDirs[i]); err != nil {
			return errors.Wrapf(err, "cannot start plugin %s", pluginDirs[i])
		}
	}

	return nil
}

func startPlugin(ctx context.Context, pluginDir string) error {
	pluginDirPath := filepath.Join(*probeDir, pluginDir)
	entryYamlFilePaths, err := filepath.Glob(filepath.Join(pluginDirPath, "main*.yaml"))
	if err != nil {
		return errors.Wrapf(err, "cannot glob main*.yaml under %s", pluginDirPath)
	}

	if len(entryYamlFilePaths) == 0 {
		return nil
	}

	for i := 0; i < len(entryYamlFilePaths); i++ {
		if err = startEntry(ctx, entryYamlFilePaths[i]); err != nil {
			return errors.Wrapf(err, "cannot start entry %s", entryYamlFilePaths[i])
		}
	}

	return nil
}

// main*.yaml 文件可能会发生变化，引用的其他文件也可能会变化，HTTP SD 的话远端的目标也可能会变化。
func startEntry(ctx context.Context, entryYamlFilePath string) error {
	cfg, err := loadConfig(entryYamlFilePath)
	if err != nil {
		return err
	}

	fmt.Println(">>>", cfg.BaseDir)

	return nil
}
