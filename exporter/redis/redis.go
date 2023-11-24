package redis

import (
	"context"

	"github.com/BurntSushi/toml"
	"github.com/cprobe/cprobe/types"
)

type Config struct {
}

func ParseConfig(bs []byte) (*Config, error) {
	var c Config
	err := toml.Unmarshal(bs, &c)
	return &c, err
}

func Scrape(ctx context.Context, address string, cfg *Config, ss *types.Samples) error {
	return nil
}
