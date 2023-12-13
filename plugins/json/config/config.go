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

// JsonExporter 这个插件使用 yaml 配置文件，而不是 toml 配置文件
// 一般来讲，一个 scrape job 只会关联一个 解释json module
type Module struct {
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
