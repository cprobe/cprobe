package redis

import (
	"context"

	"github.com/BurntSushi/toml"
	"github.com/cprobe/cprobe/lib/promutils"
)

type Config struct {
}

func ParseConfig(bs []byte) (*Config, error) {
	var c Config
	err := toml.Unmarshal(bs, &c)
	return &c, err
}

func Scrape(ctx context.Context, labels *promutils.Labels, cfg *Config) error {
	return nil
}
