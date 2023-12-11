package config

import (
	pconfig "github.com/prometheus/common/config"
)

// Metric contains values that define a metric
type Metric struct {
	Name           string
	Path           string
	Labels         map[string]string
	Type           ScrapeType
	ValueType      ValueType
	EpochTimestamp string
	Help           string
	Values         map[string]string
}

type ScrapeType string

const (
	ValueScrape  ScrapeType = "value" // default
	ObjectScrape ScrapeType = "object"
)

type ValueType string

const (
	ValueTypeGauge   ValueType = "gauge"
	ValueTypeCounter ValueType = "counter"
	ValueTypeUntyped ValueType = "untyped"
)

// Config contains multiple modules.
// type Config struct {
// 	Modules map[string]Module `yaml:"modules"`
// }

// JsonExporter 这个插件使用 yaml 配置文件，而不是 toml 配置文件
// 一般来讲，一个 scrape job 只会关联一个 解释json module
type Module struct {
	BaseDir          string                   `yaml:"-"`
	Headers          map[string]string        `yaml:"headers,omitempty"`
	Metrics          []Metric                 `yaml:"metrics"`
	HTTPClientConfig pconfig.HTTPClientConfig `yaml:"http_client_config,omitempty"`
	Body             Body                     `yaml:"body,omitempty"`
	ValidStatusCodes []int                    `yaml:"valid_status_codes,omitempty"`
}

type Body struct {
	Content    string `yaml:"content"`
	Templatize bool   `yaml:"templatize,omitempty"`
}

// func LoadConfig(configPath string) (Config, error) {
// 	var config Config
// 	data, err := os.ReadFile(configPath)
// 	if err != nil {
// 		return config, err
// 	}

// 	if err := yaml.Unmarshal(data, &config); err != nil {
// 		return config, err
// 	}

// 	// Complete Defaults
// 	for _, module := range config.Modules {
// 		for i := 0; i < len(module.Metrics); i++ {
// 			if module.Metrics[i].Type == "" {
// 				module.Metrics[i].Type = ValueScrape
// 			}
// 			if module.Metrics[i].Help == "" {
// 				module.Metrics[i].Help = module.Metrics[i].Name
// 			}
// 			if module.Metrics[i].ValueType == "" {
// 				module.Metrics[i].ValueType = ValueTypeUntyped
// 			}
// 		}
// 	}

// 	return config, nil
// }
