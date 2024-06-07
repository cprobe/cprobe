package zookeeper

import (
	"context"
	"crypto/tls"
	"github.com/BurntSushi/toml"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/cprobe/cprobe/plugins"
	"github.com/cprobe/cprobe/plugins/zookeeper/exporter"
	"github.com/cprobe/cprobe/types"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"time"
)

type Config struct {
	Timeout       time.Duration `toml:"timeout"`
	ResetOnScrape bool          `toml:"resetOnScrape"`
	EnableTLS     bool          `toml:"enableTLS"`
	TlsCert       string        `toml:"tlsCert"`
	TlsKey        string        `toml:"tlsKey"`
}

type Zookeeper struct {
	clientCert *tls.Certificate
}

func init() {
	plugins.RegisterPlugin(types.PluginZookeeper, &Zookeeper{})
}

func (f *Zookeeper) ParseConfig(baseDir string, bs []byte) (any, error) {
	var c Config
	//fmt.Printf("zk: %s\n", string(bs))
	if err := toml.Unmarshal(bs, &c); err != nil {
		return nil, err
	}

	if c.Timeout == 0 {
		c.Timeout = time.Second * 10
	}
	if c.EnableTLS {
		if c.TlsKey == "" || c.TlsCert == "" {
			log.Fatal("tlsKey and tlsCert flags are required when enableTLS is true")
		}
		clientCert, err := tls.LoadX509KeyPair(c.TlsCert, c.TlsKey)
		if err != nil {
			log.Fatalf("fatal: can't load keypair %s, %s: %v", c.TlsCert, c.TlsKey, err)
		}
		f.clientCert = &clientCert
	}

	return &c, nil
}

func (f *Zookeeper) Scrape(ctx context.Context, target string, cfg any, ss *types.Samples) error {

	conf := cfg.(*Config)

	exp := exporter.NewZookeeperExporter(exporter.Options{
		Timeout:       conf.Timeout,
		Host:          target,
		ClientCert:    f.clientCert,
		EnableTLS:     conf.EnableTLS,
		ResetOnScrape: conf.ResetOnScrape,
	})

	ch, errCh := make(chan prometheus.Metric), make(chan error, 1)

	go func() {
		errCh <- exp.Collect(ch)
		close(ch)
		close(errCh)
	}()

	for m := range ch {
		if err := ss.AddPromMetric(m); err != nil {
			logger.Warnf("failed to transform prometheus metric: %s", err)
		}
	}

	return <-errCh
}
