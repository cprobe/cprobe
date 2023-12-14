package json

import (
	"context"

	yaml "gopkg.in/yaml.v3"

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

func (p *Json) ParseConfig(_ string, bs []byte) (any, error) {
	var moduleConfig config.Module
	err := yaml.Unmarshal(bs, &moduleConfig)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(moduleConfig.Metrics); i++ {
		if moduleConfig.Metrics[i].Type == "" {
			moduleConfig.Metrics[i].Type = config.ValueScrape
		}
		if moduleConfig.Metrics[i].Help == "" {
			moduleConfig.Metrics[i].Help = moduleConfig.Metrics[i].Name
		}
		if moduleConfig.Metrics[i].ValueType == "" {
			moduleConfig.Metrics[i].ValueType = config.ValueTypeUntyped
		}
	}
	return &moduleConfig, nil
}

func (p *Json) Scrape(ctx context.Context, address string, c any, ss *types.Samples) error {
	module := c.(*config.Module)

	registry := prometheus.NewPedanticRegistry()

	metrics, err := exporter.CreateMetricsList(*module)
	if err != nil {
		return errors.WithMessagef(err, "create metrics list from config failed, error: %v", err)
	}

	jsonMetricCollector := exporter.JSONMetricCollector{JSONMetrics: metrics}
	if address == "" {
		return errors.WithMessagef(err, "address is empty")
	}

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
