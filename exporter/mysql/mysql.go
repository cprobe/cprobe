package mysql

import (
	"context"

	"github.com/cprobe/cprobe/lib/promutils"
)

type Auth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func Scrape(ctx context.Context, labels *promutils.Labels, auth *Auth, tomlBytes []byte) error {
	return nil
}
