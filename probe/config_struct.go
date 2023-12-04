package probe

import (
	"fmt"
	"time"

	"github.com/cprobe/cprobe/discovery/azure"
	"github.com/cprobe/cprobe/discovery/digitalocean"
	"github.com/cprobe/cprobe/discovery/dns"
	"github.com/cprobe/cprobe/discovery/docker"
	"github.com/cprobe/cprobe/discovery/dockerswarm"
	"github.com/cprobe/cprobe/discovery/ec2"
	"github.com/cprobe/cprobe/discovery/eureka"
	"github.com/cprobe/cprobe/discovery/gce"
	"github.com/cprobe/cprobe/discovery/http"
	"github.com/cprobe/cprobe/discovery/openstack"
	"github.com/cprobe/cprobe/discovery/yandexcloud"
	"github.com/cprobe/cprobe/lib/envtemplate"
	"github.com/cprobe/cprobe/lib/promrelabel"
	"github.com/cprobe/cprobe/lib/promutils"
	"gopkg.in/yaml.v2"
)

const (
	defaultScrapeInterval    = time.Minute
	defaultScrapeTimeout     = 10 * time.Second
	defaultScrapeConcurrency = 50
)

type Config struct {
	Global            GlobalConfig    `yaml:"global,omitempty"`
	ScrapeConfigs     []*ScrapeConfig `yaml:"scrape_configs,omitempty"`
	ScrapeConfigFiles []string        `yaml:"scrape_config_files,omitempty"`

	// This is set to the directory from where the config has been loaded.
	BaseDir string
}

// GlobalConfig represents essential parts for `global` section of Prometheus config.
//
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/
type GlobalConfig struct {
	ScrapeConcurrency int                 `yaml:"scrape_concurrency,omitempty"` // 不能一次性启动太多 target 的抓取，比如 icmp 的抓取，一次性启动太多，会导致 icmp 的抓取超时
	ScrapeInterval    *promutils.Duration `yaml:"scrape_interval,omitempty"`
	// ScrapeTimeout     *promutils.Duration `yaml:"scrape_timeout,omitempty"`
	ExternalLabels *promutils.Labels `yaml:"external_labels,omitempty"`

	MetricRelabelConfigs       []promrelabel.RelabelConfig `yaml:"metric_relabel_configs,omitempty"`
	ParsedMetricRelabelConfigs *promrelabel.ParsedConfigs  `yaml:"-"`
}

// ScrapeConfig represents essential parts for `scrape_config` section of Prometheus config.
//
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config
type ScrapeConfig struct {
	ConfigRef *Config `yaml:"-"`

	JobName           string              `yaml:"job_name"`
	ScrapeConcurrency int                 `yaml:"scrape_concurrency,omitempty"`
	ScrapeInterval    *promutils.Duration `yaml:"scrape_interval,omitempty"`
	// ScrapeTimeout     *promutils.Duration `yaml:"scrape_timeout,omitempty"`

	// 抓取数据的逻辑大变，已经不止是 HTTP /metrics 数据的抓取，可能是抓取的 SNMP、也可能抓的 MySQL
	ScrapeRuleFiles []string `yaml:"scrape_rule_files,omitempty"`

	// move to rules.d
	// MetricsPath    string              `yaml:"metrics_path,omitempty"`
	// HonorLabels    bool                `yaml:"honor_labels,omitempty"`

	// HonorTimestamps is set to false by default contrary to Prometheus, which sets it to true by default,
	// because of the issue with gaps on graphs when scraping cadvisor or similar targets, which export invalid timestamps.
	// See https://github.com/VictoriaMetrics/VictoriaMetrics/issues/4697#issuecomment-1654614799 for details.
	// move to rules.d
	// HonorTimestamps bool `yaml:"honor_timestamps,omitempty"`

	// move to rules.d
	// Scheme               string                      `yaml:"scheme,omitempty"`
	// Params               map[string][]string         `yaml:"params,omitempty"`
	// ProxyURL         *proxy.URL                `yaml:"proxy_url,omitempty"`

	RelabelConfigs       []promrelabel.RelabelConfig `yaml:"relabel_configs,omitempty"`
	MetricRelabelConfigs []promrelabel.RelabelConfig `yaml:"metric_relabel_configs,omitempty"`

	ParsedRelabelConfigs       *promrelabel.ParsedConfigs `yaml:"-"`
	ParsedMetricRelabelConfigs *promrelabel.ParsedConfigs `yaml:"-"`

	// SampleLimit          int                         `yaml:"sample_limit,omitempty"`

	AzureSDConfigs        []azure.SDConfig        `yaml:"azure_sd_configs,omitempty"`
	DigitaloceanSDConfigs []digitalocean.SDConfig `yaml:"digitalocean_sd_configs,omitempty"`
	DNSSDConfigs          []dns.SDConfig          `yaml:"dns_sd_configs,omitempty"`
	DockerSDConfigs       []docker.SDConfig       `yaml:"docker_sd_configs,omitempty"`
	DockerSwarmSDConfigs  []dockerswarm.SDConfig  `yaml:"dockerswarm_sd_configs,omitempty"`
	EC2SDConfigs          []ec2.SDConfig          `yaml:"ec2_sd_configs,omitempty"`
	EurekaSDConfigs       []eureka.SDConfig       `yaml:"eureka_sd_configs,omitempty"`
	FileSDConfigs         []FileSDConfig          `yaml:"file_sd_configs,omitempty"`
	GCESDConfigs          []gce.SDConfig          `yaml:"gce_sd_configs,omitempty"`
	HTTPSDConfigs         []http.SDConfig         `yaml:"http_sd_configs,omitempty"`
	OpenStackSDConfigs    []openstack.SDConfig    `yaml:"openstack_sd_configs,omitempty"`
	StaticConfigs         []StaticConfig          `yaml:"static_configs,omitempty"`
	YandexCloudSDConfigs  []yandexcloud.SDConfig  `yaml:"yandexcloud_sd_configs,omitempty"`

	// move to rules.d
	// These options are supported only by lib/promscrape.
	// DisableCompression  bool                       `yaml:"disable_compression,omitempty"`
	// DisableKeepAlive    bool                       `yaml:"disable_keepalive,omitempty"`
	// StreamParse         bool                       `yaml:"stream_parse,omitempty"`
	// ScrapeAlignInterval *promutils.Duration        `yaml:"scrape_align_interval,omitempty"`
	// ScrapeOffset        *promutils.Duration        `yaml:"scrape_offset,omitempty"`
	// SeriesLimit         int                        `yaml:"series_limit,omitempty"`
	// NoStaleMarkers      *bool                      `yaml:"no_stale_markers,omitempty"`
	// ProxyClientConfig   promauth.ProxyClientConfig `yaml:",inline"`

	// This is set in loadConfig
	// swc *scrapeWorkConfig

}

// FileSDConfig represents file-based service discovery config.
//
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#file_sd_config
type FileSDConfig struct {
	Files []string `yaml:"files"`
	// `refresh_interval` is ignored. See `-promscrape.fileSDCheckInterval`
}

// StaticConfig represents essential parts for `static_config` section of Prometheus config.
//
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#static_config
type StaticConfig struct {
	Targets []string          `yaml:"targets"`
	Labels  *promutils.Labels `yaml:"labels,omitempty"`
	Paths   []string          `yaml:"paths,omitempty"` // TODO: 是否要通过这个字段和 CMDB 打通，纠结
}

func (cfg *Config) unmarshal(data []byte, isStrict bool) error {
	var err error
	data, err = envtemplate.ReplaceBytes(data)
	if err != nil {
		return fmt.Errorf("cannot expand environment variables: %w", err)
	}
	if isStrict {
		if err = yaml.UnmarshalStrict(data, cfg); err != nil {
			err = fmt.Errorf("%w; pass -promscrape.config.strictParse=false command-line flag for ignoring unknown fields in yaml config", err)
		}
	} else {
		err = yaml.Unmarshal(data, cfg)
	}
	return err
}
