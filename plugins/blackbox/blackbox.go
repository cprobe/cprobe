package blackbox

import (
	"context"
	"fmt"
	"time"

	yaml "gopkg.in/yaml.v3"

	"github.com/cprobe/cprobe/plugins"
	"github.com/cprobe/cprobe/plugins/blackbox/prober"
	"github.com/cprobe/cprobe/types"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type Blackbox struct {
}

func init() {
	plugins.RegisterPlugin(types.PluginBlackbox, &Blackbox{})
}

func (p *Blackbox) ParseConfig(baseDir string, bs []byte) (any, error) {
	var moduleConfig prober.Module
	err := yaml.Unmarshal(bs, &moduleConfig)
	if err != nil {
		return nil, err
	}
	moduleConfig.BaseDir = baseDir
	return &moduleConfig, nil
}

func (p *Blackbox) Scrape(ctx context.Context, address string, c any, ss *types.Samples) error {
	err := p.scrape(ctx, address, c, ss)
	// 冗余一份 probe_success 的指标，方便仪表盘展示，避免用户再去修改仪表盘了
	if err != nil {
		ss.AddMetric("probe", map[string]interface{}{"success": 0})
	} else {
		ss.AddMetric("probe", map[string]interface{}{"success": 1})
	}
	return err
}

func (p *Blackbox) scrape(ctx context.Context, address string, c any, ss *types.Samples) error {
	module := c.(*prober.Module)

	if module.Timeout == 0 {
		module.Timeout = time.Second * 10
	}

	ctx, cancel := context.WithTimeout(ctx, module.Timeout)
	defer cancel()

	prober, ok := prober.Probers[module.Prober]
	if !ok {
		return fmt.Errorf("unknown prober %q, address: %s", module.Prober, address)
	}

	registry := prometheus.NewRegistry()

	success := prober(ctx, address, *module, registry)

	mfs, err := registry.Gather()
	if err != nil {
		return errors.WithMessagef(err, "gather metrics from registry failed, address: %s, error: %v", address, err)
	}

	ss.AddMetricFamilies(mfs)

	if !success {
		return errors.Errorf("prober failed, address: %s", address)
	}

	return nil
}
