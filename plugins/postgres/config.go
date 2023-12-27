package postgres

import (
	"context"

	"github.com/cprobe/cprobe/lib/logger"
	"github.com/cprobe/cprobe/plugins/postgres/collector"
	"github.com/cprobe/cprobe/plugins/postgres/dsn"
	"github.com/cprobe/cprobe/types"
	"github.com/prometheus/client_golang/prometheus"
)

type Config struct {
	BaseDir                string            `toml:"-"`
	Username               string            `toml:"username"`
	Password               string            `toml:"password"`
	Options                map[string]string `toml:"options"`
	DisableDefaultMetrics  bool              `toml:"disable_default_metrics"`
	DisableSettingsMetrics bool              `toml:"disable_settings_metrics"`
	EnabledCollectors      []string          `toml:"enabled_collectors"`
}

func (c *Config) ConfigureTarget(target string) (dsn.DSN, error) {
	d, err := dsn.DsnFromString(target)
	if err != nil {
		return dsn.DSN{}, err
	}
	d.SetUserAndOptions(c.Username, c.Password, c.Options)
	return d, nil
}

func (c *Config) Scrape(ctx context.Context, target string, ss *types.Samples) error {
	dsn, err := c.ConfigureTarget(target)
	if err != nil {
		return err
	}

	opts := []ExporterOpt{
		DisableDefaultMetrics(c.DisableDefaultMetrics),
		DisableSettingsMetrics(c.DisableSettingsMetrics),
	}

	dsns := []string{dsn.GetConnectionString()}
	exporter := NewExporter(dsns, opts...)
	defer func() {
		exporter.servers.Close()
	}()

	ch := make(chan prometheus.Metric)
	go func() {
		exporter.Collect(ch)
		close(ch)
	}()

	for m := range ch {
		if err := ss.AddPromMetric(m); err != nil {
			logger.Warnf("failed to transform metric: %s", err)
		}
	}

	pc, err := collector.NewProbeCollector(dsn, c.EnabledCollectors)
	if err != nil {
		return err
	}

	defer pc.Close()

	pcCh := make(chan prometheus.Metric)
	go func() {
		pc.Collect(pcCh)
		close(pcCh)
	}()

	for m := range pcCh {
		if err := ss.AddPromMetric(m); err != nil {
			logger.Warnf("failed to transform metric: %s", err)
		}
	}

	return nil
}
