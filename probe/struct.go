package probe

import (
	"github.com/cprobe/cprobe/discovery/azure"
	"github.com/cprobe/cprobe/discovery/consul"
	"github.com/cprobe/cprobe/discovery/digitalocean"
	"github.com/cprobe/cprobe/discovery/dns"
	"github.com/cprobe/cprobe/discovery/docker"
	"github.com/cprobe/cprobe/discovery/dockerswarm"
	"github.com/cprobe/cprobe/discovery/ec2"
	"github.com/cprobe/cprobe/discovery/eureka"
	"github.com/cprobe/cprobe/discovery/gce"
	"github.com/cprobe/cprobe/discovery/http"
	"github.com/cprobe/cprobe/discovery/kubernetes"
	"github.com/cprobe/cprobe/discovery/kuma"
	"github.com/cprobe/cprobe/discovery/nomad"
	"github.com/cprobe/cprobe/discovery/openstack"
	"github.com/cprobe/cprobe/discovery/yandexcloud"
	"github.com/cprobe/cprobe/lib/promauth"
	"github.com/cprobe/cprobe/lib/promrelabel"
	"github.com/cprobe/cprobe/lib/promutils"
	"github.com/cprobe/cprobe/lib/proxy"
)

type Config struct {
	Global            GlobalConfig    `yaml:"global,omitempty"`
	ScrapeConfigs     []*ScrapeConfig `yaml:"scrape_configs,omitempty"`
	ScrapeConfigFiles []string        `yaml:"scrape_config_files,omitempty"`
}

// GlobalConfig represents essential parts for `global` section of Prometheus config.
//
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/
type GlobalConfig struct {
	ScrapeInterval *promutils.Duration `yaml:"scrape_interval,omitempty"`
	ScrapeTimeout  *promutils.Duration `yaml:"scrape_timeout,omitempty"`
	ExternalLabels *promutils.Labels   `yaml:"external_labels,omitempty"`
}

// ScrapeConfig represents essential parts for `scrape_config` section of Prometheus config.
//
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config
type ScrapeConfig struct {
	JobName        string              `yaml:"job_name"`
	ScrapeInterval *promutils.Duration `yaml:"scrape_interval,omitempty"`
	ScrapeTimeout  *promutils.Duration `yaml:"scrape_timeout,omitempty"`
	MetricsPath    string              `yaml:"metrics_path,omitempty"`
	HonorLabels    bool                `yaml:"honor_labels,omitempty"`

	// HonorTimestamps is set to false by default contrary to Prometheus, which sets it to true by default,
	// because of the issue with gaps on graphs when scraping cadvisor or similar targets, which export invalid timestamps.
	// See https://github.com/VictoriaMetrics/VictoriaMetrics/issues/4697#issuecomment-1654614799 for details.
	HonorTimestamps bool `yaml:"honor_timestamps,omitempty"`

	Scheme               string                      `yaml:"scheme,omitempty"`
	Params               map[string][]string         `yaml:"params,omitempty"`
	HTTPClientConfig     promauth.HTTPClientConfig   `yaml:",inline"`
	ProxyURL             *proxy.URL                  `yaml:"proxy_url,omitempty"`
	RelabelConfigs       []promrelabel.RelabelConfig `yaml:"relabel_configs,omitempty"`
	MetricRelabelConfigs []promrelabel.RelabelConfig `yaml:"metric_relabel_configs,omitempty"`
	SampleLimit          int                         `yaml:"sample_limit,omitempty"`

	AzureSDConfigs        []azure.SDConfig        `yaml:"azure_sd_configs,omitempty"`
	ConsulSDConfigs       []consul.SDConfig       `yaml:"consul_sd_configs,omitempty"`
	DigitaloceanSDConfigs []digitalocean.SDConfig `yaml:"digitalocean_sd_configs,omitempty"`
	DNSSDConfigs          []dns.SDConfig          `yaml:"dns_sd_configs,omitempty"`
	DockerSDConfigs       []docker.SDConfig       `yaml:"docker_sd_configs,omitempty"`
	DockerSwarmSDConfigs  []dockerswarm.SDConfig  `yaml:"dockerswarm_sd_configs,omitempty"`
	EC2SDConfigs          []ec2.SDConfig          `yaml:"ec2_sd_configs,omitempty"`
	EurekaSDConfigs       []eureka.SDConfig       `yaml:"eureka_sd_configs,omitempty"`
	FileSDConfigs         []FileSDConfig          `yaml:"file_sd_configs,omitempty"`
	GCESDConfigs          []gce.SDConfig          `yaml:"gce_sd_configs,omitempty"`
	HTTPSDConfigs         []http.SDConfig         `yaml:"http_sd_configs,omitempty"`
	KubernetesSDConfigs   []kubernetes.SDConfig   `yaml:"kubernetes_sd_configs,omitempty"`
	KumaSDConfigs         []kuma.SDConfig         `yaml:"kuma_sd_configs,omitempty"`
	NomadSDConfigs        []nomad.SDConfig        `yaml:"nomad_sd_configs,omitempty"`
	OpenStackSDConfigs    []openstack.SDConfig    `yaml:"openstack_sd_configs,omitempty"`
	StaticConfigs         []StaticConfig          `yaml:"static_configs,omitempty"`
	YandexCloudSDConfigs  []yandexcloud.SDConfig  `yaml:"yandexcloud_sd_configs,omitempty"`

	// These options are supported only by lib/promscrape.
	DisableCompression  bool                       `yaml:"disable_compression,omitempty"`
	DisableKeepAlive    bool                       `yaml:"disable_keepalive,omitempty"`
	StreamParse         bool                       `yaml:"stream_parse,omitempty"`
	ScrapeAlignInterval *promutils.Duration        `yaml:"scrape_align_interval,omitempty"`
	ScrapeOffset        *promutils.Duration        `yaml:"scrape_offset,omitempty"`
	SeriesLimit         int                        `yaml:"series_limit,omitempty"`
	NoStaleMarkers      *bool                      `yaml:"no_stale_markers,omitempty"`
	ProxyClientConfig   promauth.ProxyClientConfig `yaml:",inline"`

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
}
