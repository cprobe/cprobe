package memcached

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/cprobe/cprobe/lib/clienttls"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/cprobe/cprobe/plugins/memcached/exporter"
	"github.com/cprobe/cprobe/types"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type Config struct {
	BaseDir    string        `toml:"-"`
	Timeout    time.Duration `toml:"timeout"`
	TLSEnabled bool          `toml:"tls_enabled"`

	clienttls.ClientConfig
}

func (c *Config) Scrape(ctx context.Context, target string, ss *types.Samples) error {
	var (
		tlsConfig  *tls.Config
		err        error
		serverName string
	)

	if c.TLSEnabled {
		if c.ClientConfig.ServerName == "" {
			serverName, _, err = net.SplitHostPort(target)
			if err != nil {
				return errors.WithMessagef(err, "split host port failed, target: %s", target)
			}
			c.ClientConfig.ServerName = serverName
		}

		tlsConfig, err = c.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}
	}

	timeout := c.Timeout
	if timeout == 0 {
		timeout = time.Second
	}

	ch := make(chan prometheus.Metric)
	exp := exporter.New(target, timeout, tlsConfig)
	go func() {
		exp.Collect(ch)
		close(ch)
	}()
	for m := range ch {
		if err := ss.AddPromMetric(m); err != nil {
			logger.Warnf("failed to transform memcached metric: %s", err)
		}
	}

	return nil
}
