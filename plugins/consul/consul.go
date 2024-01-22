package consul

import (
	"context"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/cprobe/cprobe/plugins"
	"github.com/cprobe/cprobe/plugins/consul/exporter"
	"github.com/cprobe/cprobe/types"
	consul_api "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type Config struct {
	AllowStale        bool          `toml:"allow_stale"`
	CAFile            string        `toml:"ca_file"`
	CertFile          string        `toml:"cert_file"`
	HealthSummary     bool          `toml:"health_summary"`
	KeyFile           string        `toml:"key_file"`
	Insecure          bool          `toml:"insecure"`
	RequireConsistent bool          `toml:"require_consistent"`
	ServerName        string        `toml:"server_name"`
	Timeout           time.Duration `toml:"timeout"`
	RequestLimit      int           `toml:"request_limit"`
	KV                KVConfig      `toml:"kv"`
	Meta              MetaConfig    `toml:"meta"`
	AgentOnly         bool          `toml:"agent_only"`
}
type KVConfig struct {
	Prefix string `toml:"prefix"`
	Filter string `toml:"filter"`
}

type MetaConfig struct {
	Filter string `toml:"filter"`
}

type Consul struct {
}

func init() {
	plugins.RegisterPlugin(types.PluginConsul, &Consul{})
}

func (*Consul) ParseConfig(baseDir string, bs []byte) (any, error) {
	var c Config
	err := toml.Unmarshal(bs, &c)
	if err != nil {
		return nil, err
	}

	if c.Timeout == 0 {
		c.Timeout = time.Millisecond * 500
	}
	return &c, nil
}

func (*Consul) Scrape(ctx context.Context, target string, cfg any, ss *types.Samples) error {
	conf := cfg.(*Config)
	opts := exporter.ConsulOpts{
		URI:          target,
		CAFile:       conf.CAFile,
		CertFile:     conf.CertFile,
		KeyFile:      conf.KeyFile,
		ServerName:   conf.ServerName,
		Timeout:      conf.Timeout,
		Insecure:     conf.Insecure,
		RequestLimit: conf.RequestLimit,
		AgentOnly:    conf.AgentOnly,
	}
	queryOptions := consul_api.QueryOptions{
		AllowStale:        conf.AllowStale,
		RequireConsistent: conf.RequireConsistent,
	}
	exporter, err := exporter.New(opts, queryOptions, conf.KV.Prefix, conf.KV.Filter, conf.Meta.Filter, conf.HealthSummary)
	if err != nil {
		return errors.Wrapf(err, "failed to create consul exporter: %s, error: %v", target, err)
	}

	ch := make(chan prometheus.Metric)
	go func() {
		exporter.Collect(ch)
		close(ch)
	}()

	for m := range ch {
		if err := ss.AddPromMetric(m); err != nil {
			logger.Warnf("failed to transform prometheus metric: %s", err)
		}
	}

	return nil
}
