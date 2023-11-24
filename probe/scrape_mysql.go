package probe

import (
	"context"

	"github.com/cprobe/cprobe/lib/promutils"
)

func ScrapeMySQL(ctx context.Context, labels *promutils.Labels, auth *ScrapeAuth, tomlBytes []byte) error {
	return nil
}
