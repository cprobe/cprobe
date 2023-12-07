package prometheus

import (
	"context"

	"github.com/BurntSushi/toml"
	"github.com/cprobe/cprobe/plugins"
	"github.com/cprobe/cprobe/types"
)

func init() {
	plugins.RegisterPlugin(types.PluginPrometheus, &Prometheus{})
}

type Prometheus struct{}

func (*Prometheus) ParseConfig(baseDir string, bs []byte) (any, error) {
	var c Config
	err := toml.Unmarshal(bs, &c)
	if err != nil {
		return nil, err
	}
	c.BaseDir = baseDir
	return &c, nil
}

func (*Prometheus) Scrape(ctx context.Context, target string, c any, ss *types.Samples) error {
	cfg := c.(*Config)
	if err := cfg.Scrape(ctx, target, ss); err != nil {
		ss.AddMetric(cfg.Global.Namespace, map[string]interface{}{"up": 0.0})
		return err
	}
	ss.AddMetric(cfg.Global.Namespace, map[string]interface{}{"up": 1.0})
	return nil
}
