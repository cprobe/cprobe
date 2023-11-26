package probe

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/cprobe/cprobe/lib/fileutil"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/pkg/errors"
)

var (
	pluginsFilter = flag.String("plugins", "", "Filter plugins, separated by comma, e.g. -plugins=mysql,redis")
)

func listPlugins(configDirectory string) ([]string, error) {
	pluginDirs, err := fileutil.DirsUnder(configDirectory)
	if err != nil {
		return nil, errors.Wrap(err, "cannot list plugin dirs")
	}

	if *pluginsFilter == "" {
		return pluginDirs, nil
	}

	ret := make([]string, 0, len(pluginDirs))

	filters := strings.Split(strings.ReplaceAll(*pluginsFilter, ":", ","), ",")
	for i := 0; i < len(pluginDirs); i++ {
		for j := 0; j < len(filters); j++ {
			if pluginDirs[i] == filters[j] {
				ret = append(ret, pluginDirs[i])
			}
		}
	}

	return ret, nil
}

// Start starts the probe goroutines.
func Start(ctx context.Context, configDirectory string) error {
	pluginDirs, err := listPlugins(configDirectory)
	if err != nil {
		return err
	}

	if len(pluginDirs) == 0 {
		return fmt.Errorf("no plugin dirs found under %s", configDirectory)
	}

	for i := 0; i < len(pluginDirs); i++ {
		if err := startPlugin(ctx, configDirectory, pluginDirs[i]); err != nil {
			return errors.Wrapf(err, "cannot start plugin %s", pluginDirs[i])
		}
	}

	return nil
}

func startPlugin(ctx context.Context, configDirectory, pluginDir string) error {
	pluginDirPath := filepath.Join(configDirectory, pluginDir)
	entryYamlFilePaths, err := filepath.Glob(filepath.Join(pluginDirPath, "main*.yaml"))
	if err != nil {
		return errors.Wrapf(err, "cannot glob main*.yaml under %s", pluginDirPath)
	}

	if len(entryYamlFilePaths) == 0 {
		return nil
	}

	for i := 0; i < len(entryYamlFilePaths); i++ {
		if err = startEntry(ctx, pluginDir, entryYamlFilePaths[i]); err != nil {
			return errors.Wrapf(err, "cannot start entry %s", entryYamlFilePaths[i])
		}
	}

	return nil
}

func startEntry(ctx context.Context, pluginName, entryYamlFilePath string) error {
	cfg, err := loadConfig(entryYamlFilePath)
	if err != nil {
		return err
	}

	pluginJobs, has := Jobs[pluginName]
	if !has {
		return fmt.Errorf("unsupported plugin %s", pluginName)
	}

	for i := range cfg.ScrapeConfigs {
		if cfg.ScrapeConfigs[i] == nil {
			continue
		}

		jobID := JobID{YamlFile: entryYamlFilePath, JobName: cfg.ScrapeConfigs[i].JobName}
		jobGoroutine := NewJobGoroutine(pluginName, cfg.ScrapeConfigs[i])
		pluginJobs[jobID] = jobGoroutine

		// 启动 goroutine，稍微 sleep 一下，避免所有 goroutine 同时启动
		time.Sleep(time.Millisecond * 10)
		go jobGoroutine.Start(ctx)
	}

	return nil
}

// Reload 读取磁盘配置文件，与内存中的配置文件进行比较，增删 JobGoroutine
func Reload(ctx context.Context, configDirectory string) {
	newJobs, err := readFiles(configDirectory)
	if err != nil {
		logger.Errorf("cannot read files: %s", err)
		return
	}

	// 遍历内存中的老 Jobs，如果磁盘上的新 Jobs 中没有，就删除
	for pluginName, jobs := range Jobs {
		newPluginJobs := newJobs[pluginName]

		for jobID, jobGoroutine := range jobs {
			_, has := newPluginJobs[jobID]
			if !has {
				jobGoroutine.Stop()
				delete(jobs, jobID)
			}
		}
	}

	// 遍历磁盘中的新 Jobs，如果内存中老 Jobs 没有，就新增，有就更新
	for pluginName, jobs := range newJobs {
		oldPluginJobs := Jobs[pluginName]

		for jobID, jobGoroutine := range jobs {
			oldJobGoroutine, has := oldPluginJobs[jobID]
			if !has {
				oldPluginJobs[jobID] = jobGoroutine

				time.Sleep(time.Millisecond * 20)
				go oldPluginJobs[jobID].Start(ctx)

				continue
			}

			oldJobGoroutine.UpdateConfig(jobGoroutine.scrapeConfig)
		}
	}
}

func readFiles(configDirectory string) (map[string]map[JobID]*JobGoroutine, error) {
	pluginDirs, err := listPlugins(configDirectory)
	if err != nil {
		return nil, err
	}

	newJobs := makeJobs()

	for i := 0; i < len(pluginDirs); i++ {
		pluginDir := pluginDirs[i]
		pluginDirPath := filepath.Join(configDirectory, pluginDir)

		entryYamlFilePaths, err := filepath.Glob(filepath.Join(pluginDirPath, "main*.yaml"))
		if err != nil {
			return nil, fmt.Errorf("cannot glob main*.yaml under %s: %s", pluginDirPath, err)
		}

		for i := 0; i < len(entryYamlFilePaths); i++ {
			entryYamlFilePath := entryYamlFilePaths[i]

			cfg, err := loadConfig(entryYamlFilePath)
			if err != nil {
				return nil, fmt.Errorf("cannot load config %s: %s", entryYamlFilePath, err)
			}

			pluginJobs, has := newJobs[pluginDir]
			if !has {
				return nil, fmt.Errorf("unsupported plugin %s", pluginDir)
			}

			for i := range cfg.ScrapeConfigs {
				if cfg.ScrapeConfigs[i] == nil {
					continue
				}

				jobID := JobID{YamlFile: entryYamlFilePath, JobName: cfg.ScrapeConfigs[i].JobName}
				jobGoroutine := NewJobGoroutine(pluginDir, cfg.ScrapeConfigs[i])
				pluginJobs[jobID] = jobGoroutine
			}
		}
	}

	return newJobs, nil
}
