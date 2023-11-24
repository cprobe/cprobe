package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/cprobe/cprobe/lib/promutils"
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

func Scrape(ctx context.Context, labels *promutils.Labels, cfg *Config) error {
	fmt.Println(">> labels", labels)
	fmt.Println(">> username", cfg.Global.Username)
	fmt.Println(">> password", cfg.Global.Password)
	fmt.Println(">> extra_status_metrics", cfg.Global.ExtraStatusMetrics)
	fmt.Println(">> extra_innodb_metrics", cfg.Global.ExtraInnodbMetrics)

	for i := range cfg.Queries {
		fmt.Println(">>>> mesurement", cfg.Queries[i].Mesurement)
		fmt.Println(">>>> metric_fields", cfg.Queries[i].MetricFields)
		fmt.Println(">>>> label_fields", cfg.Queries[i].LabelFields)
		fmt.Println(">>>> timeout", cfg.Queries[i].Timeout)
		fmt.Println(">>>> request", cfg.Queries[i].Request)
	}

	return nil
}
