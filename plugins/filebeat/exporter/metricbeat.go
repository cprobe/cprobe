package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

// MetricbeatEvent json structure
type MetricbeatEvent struct {
	Failures float64 `json:"failures"`
	Success  float64 `json:"success"`
}

// Metricbeat json structure
type Metricbeat struct {
	System struct {
		CPU            MetricbeatEvent `json:"cpu"`
		Filesystem     MetricbeatEvent `json:"filesystem"`
		Fsstat         MetricbeatEvent `json:"fsstat"`
		Load           MetricbeatEvent `json:"load"`
		Memory         MetricbeatEvent `json:"memory"`
		Network        MetricbeatEvent `json:"network"`
		Process        MetricbeatEvent `json:"process"`
		ProcessSummary MetricbeatEvent `json:"process_summary"`
		Uptime         MetricbeatEvent `json:"uptime"`
	} `json:"system"`
}

type metricbeatCollector struct {
	beatInfo *BeatInfo
	stats    *Stats
	metrics  exportedMetrics
}

// NewMetricbeatCollector constructor
func NewMetricbeatCollector(beatInfo *BeatInfo, stats *Stats) prometheus.Collector {
	return &metricbeatCollector{
		beatInfo: beatInfo,
		stats:    stats,
		metrics: exportedMetrics{
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "metricbeat_system", "cpu"),
					"system.cpu",
					nil, prometheus.Labels{"event": "success"},
				),
				eval:    func(stats *Stats) float64 { return stats.Metricbeat.System.CPU.Success },
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "metricbeat_system", "cpu"),
					"system.cpu",
					nil, prometheus.Labels{"event": "failures"},
				),
				eval:    func(stats *Stats) float64 { return stats.Metricbeat.System.CPU.Failures },
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "metricbeat_system", "filesystem"),
					"system.filesystem",
					nil, prometheus.Labels{"event": "success"},
				),
				eval:    func(stats *Stats) float64 { return stats.Metricbeat.System.Filesystem.Success },
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "metricbeat_system", "filesystem"),
					"system.filesystem",
					nil, prometheus.Labels{"event": "failures"},
				),
				eval:    func(stats *Stats) float64 { return stats.Metricbeat.System.Filesystem.Failures },
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "metricbeat_system", "fsstat"),
					"system.fsstat",
					nil, prometheus.Labels{"event": "success"},
				),
				eval:    func(stats *Stats) float64 { return stats.Metricbeat.System.Fsstat.Success },
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "metricbeat_system", "fsstat"),
					"system.fsstat",
					nil, prometheus.Labels{"event": "failures"},
				),
				eval:    func(stats *Stats) float64 { return stats.Metricbeat.System.Fsstat.Failures },
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "metricbeat_system", "load"),
					"system.load",
					nil, prometheus.Labels{"event": "success"},
				),
				eval:    func(stats *Stats) float64 { return stats.Metricbeat.System.Load.Success },
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "metricbeat_system", "load"),
					"system.load",
					nil, prometheus.Labels{"event": "failures"},
				),
				eval:    func(stats *Stats) float64 { return stats.Metricbeat.System.Load.Failures },
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "metricbeat_system", "memory"),
					"system.memory",
					nil, prometheus.Labels{"event": "success"},
				),
				eval:    func(stats *Stats) float64 { return stats.Metricbeat.System.Memory.Success },
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "metricbeat_system", "memory"),
					"system.memory",
					nil, prometheus.Labels{"event": "failures"},
				),
				eval:    func(stats *Stats) float64 { return stats.Metricbeat.System.Memory.Failures },
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "metricbeat_system", "network"),
					"system.network",
					nil, prometheus.Labels{"event": "success"},
				),
				eval:    func(stats *Stats) float64 { return stats.Metricbeat.System.Network.Success },
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "metricbeat_system", "network"),
					"system.network",
					nil, prometheus.Labels{"event": "failures"},
				),
				eval:    func(stats *Stats) float64 { return stats.Metricbeat.System.Network.Failures },
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "metricbeat_system", "process"),
					"system.process",
					nil, prometheus.Labels{"event": "success"},
				),
				eval:    func(stats *Stats) float64 { return stats.Metricbeat.System.Process.Success },
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "metricbeat_system", "process"),
					"system.process",
					nil, prometheus.Labels{"event": "failures"},
				),
				eval:    func(stats *Stats) float64 { return stats.Metricbeat.System.Process.Failures },
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "metricbeat_system", "process_summary"),
					"system.process_summary",
					nil, prometheus.Labels{"event": "success"},
				),
				eval:    func(stats *Stats) float64 { return stats.Metricbeat.System.ProcessSummary.Success },
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "metricbeat_system", "process_summary"),
					"system.process_summary",
					nil, prometheus.Labels{"event": "failures"},
				),
				eval:    func(stats *Stats) float64 { return stats.Metricbeat.System.ProcessSummary.Failures },
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "metricbeat_system", "uptime"),
					"system.uptime",
					nil, prometheus.Labels{"event": "success"},
				),
				eval:    func(stats *Stats) float64 { return stats.Metricbeat.System.Uptime.Success },
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "metricbeat_system", "uptime"),
					"system.uptime",
					nil, prometheus.Labels{"event": "failures"},
				),
				eval:    func(stats *Stats) float64 { return stats.Metricbeat.System.Uptime.Failures },
				valType: prometheus.CounterValue,
			},
		},
	}
}

// Describe returns all descriptions of the collector.
func (c *metricbeatCollector) Describe(ch chan<- *prometheus.Desc) {

	for _, metric := range c.metrics {
		ch <- metric.desc
	}

}

// Collect returns the current state of all metrics of the collector.
func (c *metricbeatCollector) Collect(ch chan<- prometheus.Metric) {

	for _, i := range c.metrics {
		ch <- prometheus.MustNewConstMetric(i.desc, i.valType, i.eval(c.stats))
	}

}
