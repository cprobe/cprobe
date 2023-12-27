// Copyright 2022 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"context"
	"sync"

	"github.com/cprobe/cprobe/lib/logger"
	"github.com/cprobe/cprobe/plugins/postgres/dsn"
	"github.com/prometheus/client_golang/prometheus"
)

type ProbeCollector struct {
	collectors map[string]Collector
	instance   *instance
}

func NewProbeCollector(dsn dsn.DSN, enabledCollectors []string) (*ProbeCollector, error) {
	collectors := make(map[string]Collector)

	for _, key := range enabledCollectors {
		collector, err := factories[key](
			collectorConfig{
				excludeDatabases: []string{},
			})
		if err != nil {
			return nil, err
		}
		collectors[key] = collector
	}

	instance, err := newInstance(dsn.GetConnectionString())
	if err != nil {
		return nil, err
	}

	return &ProbeCollector{
		collectors: collectors,
		instance:   instance,
	}, nil
}

func (pc *ProbeCollector) Describe(ch chan<- *prometheus.Desc) {
}

func (pc *ProbeCollector) Collect(ch chan<- prometheus.Metric) {
	// Set up the database connection for the collector.
	err := pc.instance.setup()
	if err != nil {
		logger.Errorf("Error opening connection to database(%s): %v", pc.instance.dsn, err)
		return
	}
	defer pc.instance.Close()

	wg := sync.WaitGroup{}
	wg.Add(len(pc.collectors))
	for name, c := range pc.collectors {
		go func(name string, c Collector) {
			execute(context.TODO(), name, c, pc.instance, ch)
			wg.Done()
		}(name, c)
	}
	wg.Wait()
}

func (pc *ProbeCollector) Close() error {
	return pc.instance.Close()
}
