package oracledb

import (
	"context"

	"github.com/BurntSushi/toml"
	"github.com/cprobe/cprobe/plugins"
	"github.com/cprobe/cprobe/types"
)

func init() {
	plugins.RegisterPlugin(types.PluginOracleDB, &OracleDB{})
}

type OracleDB struct{}

func (*OracleDB) ParseConfig(baseDir string, bs []byte) (any, error) {
	var c Config
	err := toml.Unmarshal(bs, &c)
	if err != nil {
		return nil, err
	}
	c.BaseDir = baseDir
	return &c, nil
}

func (*OracleDB) Scrape(ctx context.Context, target string, c any, ss *types.Samples) error {
	cfg := c.(*Config)
	return cfg.Scrape(ctx, target, ss)
}
