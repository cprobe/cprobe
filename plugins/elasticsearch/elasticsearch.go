package elasticsearch

import (
	"context"

	"github.com/BurntSushi/toml"
	"github.com/cprobe/cprobe/plugins"
	"github.com/cprobe/cprobe/plugins/elasticsearch/collector"
	"github.com/cprobe/cprobe/types"
)

func init() {
	plugins.RegisterPlugin(types.PluginElasticSearch, &ElasticSearch{})
}

type ElasticSearch struct{}

func (*ElasticSearch) ParseConfig(baseDir string, bs []byte) (any, error) {
	var c collector.Config
	err := toml.Unmarshal(bs, &c)
	if err != nil {
		return nil, err
	}
	c.BaseDir = baseDir
	return &c, nil
}

func (*ElasticSearch) Scrape(ctx context.Context, target string, c any, ss *types.Samples) error {
	cfg := c.(*collector.Config)
	return cfg.Scrape(ctx, target, ss)
}
