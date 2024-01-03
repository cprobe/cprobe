// Copyright 2020 The Prometheus Authors
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

package exporter

import (
	"net"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func TestParseStatsSettings(t *testing.T) {
	addr, err := net.ResolveIPAddr("ip4", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		var statsSettings = map[net.Addr]map[string]string{
			addr: {
				"maxconns":              "10",
				"lru_crawler":           "yes",
				"lru_crawler_sleep":     "100",
				"lru_crawler_tocrawl":   "0",
				"lru_maintainer_thread": "no",
				"hot_lru_pct":           "20",
				"warm_lru_pct":          "40",
				"hot_max_factor":        "0.20",
				"warm_max_factor":       "2.00",
				"accepting_conns":       "1",
			},
		}
		ch := make(chan prometheus.Metric, 100)
		e := New("", 100*time.Millisecond, nil)
		if err := e.parseStatsSettings(ch, statsSettings); err != nil {
			t.Errorf("expect return error, error: %v", err)
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Parallel()
		var statsSettings = map[net.Addr]map[string]string{
			addr: {
				"maxconns":              "10",
				"lru_crawler":           "yes",
				"lru_crawler_sleep":     "100",
				"lru_crawler_tocrawl":   "0",
				"lru_maintainer_thread": "fail",
				"hot_lru_pct":           "20",
				"warm_lru_pct":          "40",
				"hot_max_factor":        "0.20",
				"warm_max_factor":       "2.00",
				"accepting_conns":       "fail",
			},
		}
		ch := make(chan prometheus.Metric, 100)
		e := New("", 100*time.Millisecond, nil)
		if err := e.parseStatsSettings(ch, statsSettings); err == nil {
			t.Error("expect return error but not")
		}
	})
}

func TestParseTimeval(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		_, err := parseTimeval(map[string]string{"rusage_system": "3.5"}, "rusage_system")
		if err != nil {
			t.Errorf("expect return error, error: %v", err)
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Parallel()
		_, err := parseTimeval(map[string]string{"rusage_system": "35"}, "rusage_system")
		if err == nil {
			t.Error("expect return error but not")
		}
	})
}
