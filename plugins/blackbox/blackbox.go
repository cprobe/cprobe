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

// 注意：这里没有使用 baseDir 这个变量，所以，如果 yaml 配置中包含路径，需要使用全路径
// TODO：后面想办法优化一下，支持相对路径
func (p *Blackbox) ParseConfig(baseDir string, bs []byte) (any, error) {
	var moduleConfig prober.Module
	err := yaml.Unmarshal(bs, &moduleConfig)
	if err != nil {
		return nil, err
	}
	return &moduleConfig, nil
}

func (p *Blackbox) Scrape(ctx context.Context, address string, c any, ss *types.Samples) error {
	module := c.(*prober.Module)

	if module.Timeout == 0 {
		module.Timeout = time.Second * 10
	}

	ctx, cancel := context.WithTimeout(ctx, module.Timeout)
	defer cancel()

	probeSuccessGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "probe_success",
		Help: "Displays whether or not the probe was a success",
	})
	probeDurationGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "probe_duration_seconds",
		Help: "Returns how long the probe took to complete in seconds",
	})

	prober, ok := prober.Probers[module.Prober]
	if !ok {
		return fmt.Errorf("unknown prober %q, address: %s", module.Prober, address)
	}

	start := time.Now()
	registry := prometheus.NewRegistry()
	registry.MustRegister(probeSuccessGauge)
	registry.MustRegister(probeDurationGauge)
	success := prober(ctx, address, *module, registry)
	duration := time.Since(start).Seconds()
	probeDurationGauge.Set(duration)

	if success {
		probeSuccessGauge.Set(1)
	} else {
		probeSuccessGauge.Set(0)
	}

	mfs, err := registry.Gather()
	if err != nil {
		return errors.WithMessagef(err, "gather metrics from registry failed, address: %s, error: %v", address, err)
	}

	ss.AddMetricFamilies(mfs)

	return nil
}
