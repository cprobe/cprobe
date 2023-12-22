package mongodb

import (
	"context"

	"github.com/BurntSushi/toml"
	"github.com/cprobe/cprobe/plugins"
	"github.com/cprobe/cprobe/plugins/mongodb/exporter"
	"github.com/cprobe/cprobe/types"
)

func init() {
	plugins.RegisterPlugin(types.PluginMongoDB, &MongoDB{})
}

type MongoDB struct{}

func (*MongoDB) ParseConfig(baseDir string, bs []byte) (any, error) {
	var c exporter.Config
	err := toml.Unmarshal(bs, &c)
	if err != nil {
		return nil, err
	}
	c.BaseDir = baseDir
	return &c, nil
}

func (*MongoDB) Scrape(ctx context.Context, target string, c any, ss *types.Samples) error {
	cfg := c.(*exporter.Config)
	err := cfg.Scrape(ctx, target, ss)

	// 其实框架层面是提供了 mongodb_cprobe_up 指标的，但是很多仪表盘都是依赖 mongodb_up 指标
	// 所以这里也重复上报了 mongodb_up 指标，冗余一下，用户改造成本就小了
	if err != nil {
		ss.AddMetric("mongodb", map[string]interface{}{"up": 0.0})
	} else {
		ss.AddMetric("mongodb", map[string]interface{}{"up": 0.0})
	}

	return err
}
