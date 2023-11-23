package probe

import (
	"flag"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cprobe/cprobe/lib/envtemplate"
	"github.com/cprobe/cprobe/lib/fs"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/cprobe/cprobe/lib/promrelabel"
	"github.com/cprobe/cprobe/lib/promutils"
	"gopkg.in/yaml.v2"
)

var (
	strictParse = flag.Bool("scrape.config.strictParse", true, "Whether to deny unsupported fields in main*.yaml. Set to false in order to silently skip unsupported fields")
)

// loadConfig loads Prometheus config from the given path.
func loadConfig(path string) (*Config, error) {
	data, err := fs.ReadFileOrHTTP(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read Prometheus config from %q: %w", path, err)
	}
	var c Config
	if err := c.parseData(data, path); err != nil {
		return nil, fmt.Errorf("cannot parse Prometheus config from %q: %w", path, err)
	}
	return &c, nil
}

func (cfg *Config) parseData(data []byte, path string) error {
	if err := cfg.unmarshal(data, *strictParse); err != nil {
		cfg.ScrapeConfigs = nil
		return fmt.Errorf("cannot unmarshal data: %w", err)
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		cfg.ScrapeConfigs = nil
		return fmt.Errorf("cannot obtain abs path for %q: %w", path, err)
	}
	cfg.BaseDir = filepath.Dir(absPath)

	// handle GlobalConfig
	cfg.Global.ParsedMetricRelabelConfigs, err = promrelabel.ParseRelabelConfigs(cfg.Global.MetricRelabelConfigs)
	if err != nil {
		return fmt.Errorf("cannot parse global metric_relabel_configs: %w", err)
	}

	// Load cfg.ScrapeConfigFiles into c.ScrapeConfigs
	scs := mustLoadScrapeConfigFiles(cfg.BaseDir, cfg.ScrapeConfigFiles)
	cfg.ScrapeConfigFiles = nil
	cfg.ScrapeConfigs = append(cfg.ScrapeConfigs, scs...)

	// Check that all the scrape configs have unique JobName
	m := make(map[string]struct{}, len(cfg.ScrapeConfigs))
	for _, sc := range cfg.ScrapeConfigs {
		jobName := sc.JobName
		if _, ok := m[jobName]; ok {
			cfg.ScrapeConfigs = nil
			return fmt.Errorf("duplicate `job_name` in `scrape_configs` loaded from %q: %q", path, jobName)
		}
		m[jobName] = struct{}{}
	}

	for i := range cfg.ScrapeConfigs {
		sc := cfg.ScrapeConfigs[i]
		if sc.JobName == "" {
			logger.Errorf("skipping `scrape_config` without `job_name` at %q", path)
			continue
		}

		if len(sc.ScrapeRuleFiles) == 0 {
			logger.Errorf("skipping `scrape_config` without `scrape_rule_files` at %q", path)
			continue
		}

		sc.ParsedRelabelConfigs, err = promrelabel.ParseRelabelConfigs(sc.RelabelConfigs)
		if err != nil {
			logger.Errorf("skipping `scrape_config` for job_name=%s because of parse relabel_configs error: %s", sc.JobName, err)
			continue
		}

		sc.ParsedMetricRelabelConfigs, err = promrelabel.ParseRelabelConfigs(sc.MetricRelabelConfigs)
		if err != nil {
			logger.Errorf("skipping `scrape_config` for job_name=%s because of parse metric_relabel_configs error: %s", sc.JobName, err)
			continue
		}

		scrapeInterval := sc.ScrapeInterval.Duration()
		if scrapeInterval <= 0 {
			scrapeInterval = cfg.Global.ScrapeInterval.Duration()
			if scrapeInterval <= 0 {
				scrapeInterval = defaultScrapeInterval
			}
		}
		scrapeTimeout := sc.ScrapeTimeout.Duration()
		if scrapeTimeout <= 0 {
			scrapeTimeout = cfg.Global.ScrapeTimeout.Duration()
			if scrapeTimeout <= 0 {
				scrapeTimeout = defaultScrapeTimeout
			}
		}
		if scrapeTimeout > scrapeInterval {
			// Limit the `scrape_timeout` with `scrape_interval` like Prometheus does.
			// This guarantees that the scraper can miss only a single scrape if the target sometimes responds slowly.
			// See https://github.com/VictoriaMetrics/VictoriaMetrics/issues/1281#issuecomment-840538907
			scrapeTimeout = scrapeInterval
		}

		sc.ScrapeInterval = promutils.NewDuration(scrapeInterval)
		sc.ScrapeTimeout = promutils.NewDuration(scrapeTimeout)
		sc.GlobalExternalLabels = cfg.Global.ExternalLabels
		sc.GlobalParsedMetricRelabelConfigs = cfg.Global.ParsedMetricRelabelConfigs
		sc.BaseDir = cfg.BaseDir
	}

	return nil
}

func mustLoadScrapeConfigFiles(baseDir string, scrapeConfigFiles []string) []*ScrapeConfig {
	var scrapeConfigs []*ScrapeConfig
	for _, filePath := range scrapeConfigFiles {
		filePath := fs.GetFilepath(baseDir, filePath)
		paths := []string{filePath}
		if strings.Contains(filePath, "*") {
			ps, err := filepath.Glob(filePath)
			if err != nil {
				logger.Errorf("skipping pattern %q at `scrape_config_files` because of error: %s", filePath, err)
				continue
			}
			sort.Strings(ps)
			paths = ps
		}
		for _, path := range paths {
			data, err := fs.ReadFileOrHTTP(path)
			if err != nil {
				logger.Errorf("skipping %q at `scrape_config_files` because of error: %s", path, err)
				continue
			}
			data, err = envtemplate.ReplaceBytes(data)
			if err != nil {
				logger.Errorf("skipping %q at `scrape_config_files` because of failure to expand environment vars: %s", path, err)
				continue
			}
			var scs []*ScrapeConfig
			if err = yaml.UnmarshalStrict(data, &scs); err != nil {
				logger.Errorf("skipping %q at `scrape_config_files` because of failure to parse it: %s", path, err)
				continue
			}
			scrapeConfigs = append(scrapeConfigs, scs...)
		}
	}
	return scrapeConfigs
}
