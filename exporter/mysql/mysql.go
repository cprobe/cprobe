package mysql

import (
	"context"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/cprobe/cprobe/types"
)

type Global struct {
	Username           string `toml:"username"`
	Password           string `toml:"password"`
	ExtraStatusMetrics bool   `toml:"extra_status_metrics"`
	ExtraInnodbMetrics bool   `toml:"extra_innodb_metrics"`
}

type Query struct {
	Mesurement   string        `toml:"mesurement"`
	MetricFields []string      `toml:"metric_fields"`
	LabelFields  []string      `toml:"label_fields"`
	Timeout      time.Duration `toml:"timeout"`
	Request      string        `toml:"request"`
}

type Config struct {
	Global  *Global  `toml:"global"`
	Queries []*Query `toml:"queries"`
}

func ParseConfig(bs []byte) (*Config, error) {
	var c Config
	err := toml.Unmarshal(bs, &c)
	return &c, err
}

func Scrape(ctx context.Context, address string, cfg *Config, ss *types.Samples) error {
	ss.AddMetric("mysql_connections", map[string]interface{}{"connected": 16})
	return nil
}
