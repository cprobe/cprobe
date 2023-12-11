package json

import (
	"context"

	yaml "gopkg.in/yaml.v3"

	"github.com/cprobe/cprobe/lib/logger"
	"github.com/cprobe/cprobe/plugins"
	"github.com/cprobe/cprobe/plugins/json/config"
	"github.com/cprobe/cprobe/plugins/json/exporter"
	"github.com/cprobe/cprobe/types"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type Json struct {
}

func init() {
	plugins.RegisterPlugin(types.PluginJson, &Json{})
}

func (p *Json) ParseConfig(baseDir string, bs []byte) (any, error) {
	var moduleConfig config.Module
	err := yaml.Unmarshal(bs, &moduleConfig)
	if err != nil {
		return nil, err
	}
	//TODO 是否有必要设置 BaseDir
	moduleConfig.BaseDir = baseDir
	return &moduleConfig, nil
}

func (p *Json) Scrape(ctx context.Context, address string, c any, ss *types.Samples) error {
	module := c.(*config.Module)

	registry := prometheus.NewPedanticRegistry()
	metrics, err := exporter.CreateMetricsList(*module)
	if err != nil {
		logger.Errorf("Failed to create metrics list from config, err: %s", err)
	}

	jsonMetricCollector := exporter.JSONMetricCollector{JSONMetrics: metrics}
	//TODO 判断address 不为空
	fetcher := exporter.NewJSONFetcher(ctx, *module, nil)
	data, err := fetcher.FetchJSON(address)
	if err != nil {
		return errors.WithMessagef(err, "fetch json response failed, address: %s, error: %v", address, err)
	}
	jsonMetricCollector.Data = data
	registry.MustRegister(jsonMetricCollector)

	mfs, err := registry.Gather()
	if err != nil {
		return errors.WithMessagef(err, "gather metrics from registry failed, address: %s, error: %v", address, err)
	}
	ss.AddMetricFamilies(mfs)
	return nil
}
