package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

// CPUStats json structure
type CPUStats struct {
	M1  float64 `json:"1"`
	M5  float64 `json:"5"`
	M15 float64 `json:"15"`
}

// System json structure
type System struct {
	CPU struct {
		Cores int64 `json:"cores"`
	} `json:"cpu"`
	Load struct {
		CPUStats
		Norm CPUStats `json:"norm"`
	} `json:"load"`
}
type systemCollector struct {
	beatInfo *BeatInfo
	stats    *Stats
	metrics  exportedMetrics
}

// NewSystemCollector constructor
func NewSystemCollector(beatInfo *BeatInfo, stats *Stats) prometheus.Collector {
	return &systemCollector{
		beatInfo: beatInfo,
		stats:    stats,
		metrics: exportedMetrics{
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "system_cpu", "cores_total"),
					"cpu cores",
					nil, nil,
				),
				eval:    func(stats *Stats) float64 { return float64(stats.System.CPU.Cores) },
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "system", "load"),
					"system load",
					nil, prometheus.Labels{"period": "1"},
				),
				eval:    func(stats *Stats) float64 { return stats.System.Load.M1 },
				valType: prometheus.GaugeValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "system", "load"),
					"system load",
					nil, prometheus.Labels{"period": "5"},
				),
				eval:    func(stats *Stats) float64 { return stats.System.Load.M5 },
				valType: prometheus.GaugeValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "system", "load"),
					"system load",
					nil, prometheus.Labels{"period": "15"},
				),
				eval:    func(stats *Stats) float64 { return stats.System.Load.M15 },
				valType: prometheus.GaugeValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "system_load", "norm"),
					"system load",
					nil, prometheus.Labels{"period": "1"},
				),
				eval:    func(stats *Stats) float64 { return stats.System.Load.Norm.M1 },
				valType: prometheus.GaugeValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "system_load", "norm"),
					"system load",
					nil, prometheus.Labels{"period": "5"},
				),
				eval:    func(stats *Stats) float64 { return stats.System.Load.Norm.M5 },
				valType: prometheus.GaugeValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "system_load", "norm"),
					"system load",
					nil, prometheus.Labels{"period": "15"},
				),
				eval:    func(stats *Stats) float64 { return stats.System.Load.Norm.M15 },
				valType: prometheus.GaugeValue,
			},
		},
	}
}

// Describe returns all descriptions of the collector.
func (c *systemCollector) Describe(ch chan<- *prometheus.Desc) {

	for _, metric := range c.metrics {
		ch <- metric.desc
	}

}

// Collect returns the current state of all metrics of the collector.
func (c *systemCollector) Collect(ch chan<- prometheus.Metric) {

	for _, i := range c.metrics {
		ch <- prometheus.MustNewConstMetric(i.desc, i.valType, i.eval(c.stats))
	}

}
