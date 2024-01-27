package collector

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// CPUTimings json structure
type CPUTimings struct {
	Ticks float64 `json:"ticks"`
	Time  struct {
		MS float64 `json:"ms"`
	} `json:"time"`
	Value float64 `json:"value"`
}

// BeatStats json structure
type BeatStats struct {
	CPU struct {
		System CPUTimings `json:"system"`
		Total  CPUTimings `json:"total"`
		User   CPUTimings `json:"user"`
	} `json:"cpu"`
	BeatUptime struct {
		Uptime struct {
			MS float64 `json:"ms"`
		} `json:"uptime"`

		EmphemeralID string `json:"emphemeral_id"`
	} `json:"info"`

	Memstats struct {
		GCNext      float64 `json:"gc_next"`
		MemoryAlloc float64 `json:"memory_alloc"`
		MemoryTotal float64 `json:"memory_total"`
		RSS         float64 `json:"rss"`
	} `json:"memstats"`

	Runtime struct {
		Goroutines uint64 `json:"goroutines"`
	} `json:"runtime"`
}

type beatCollector struct {
	beatInfo *BeatInfo
	stats    *Stats
	metrics  exportedMetrics
}

// NewBeatCollector constructor
func NewBeatCollector(beatInfo *BeatInfo, stats *Stats) prometheus.Collector {
	return &beatCollector{
		beatInfo: beatInfo,
		stats:    stats,
		metrics: exportedMetrics{
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "cpu_time", "seconds_total"),
					"beat.cpu.time",
					nil, prometheus.Labels{"mode": "system"},
				),
				eval: func(stats *Stats) float64 {
					return (time.Duration(stats.Beat.CPU.System.Time.MS) * time.Millisecond).Seconds()
				},
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "cpu_time", "seconds_total"),
					"beat.cpu.time",
					nil, prometheus.Labels{"mode": "user"},
				),
				eval: func(stats *Stats) float64 {
					return (time.Duration(stats.Beat.CPU.User.Time.MS) * time.Millisecond).Seconds()
				},
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "cpu", "ticks_total"),
					"beat.cpu.ticks",
					nil, prometheus.Labels{"mode": "system"},
				),
				eval:    func(stats *Stats) float64 { return stats.Beat.CPU.System.Ticks },
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "cpu", "ticks_total"),
					"beat.cpu.ticks",
					nil, prometheus.Labels{"mode": "user"},
				),
				eval:    func(stats *Stats) float64 { return stats.Beat.CPU.User.Ticks },
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "uptime", "seconds_total"),
					"beat.info.uptime.ms",
					nil, nil,
				),
				eval: func(stats *Stats) float64 {
					return (time.Duration(stats.Beat.BeatUptime.Uptime.MS) * time.Millisecond).Seconds()
				},
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "memstats", "gc_next_total"),
					"beat.memstats.gc_next",
					nil, nil,
				),
				eval: func(stats *Stats) float64 {
					return stats.Beat.Memstats.GCNext
				},
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "memstats", "memory_alloc"),
					"beat.memstats.memory_alloc",
					nil, nil,
				),
				eval: func(stats *Stats) float64 {
					return stats.Beat.Memstats.MemoryAlloc
				},
				valType: prometheus.GaugeValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "memstats", "memory"),
					"beat.memstats.memory_total",
					nil, nil,
				),
				eval: func(stats *Stats) float64 {
					return stats.Beat.Memstats.MemoryTotal
				},
				valType: prometheus.GaugeValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "memstats", "rss"),
					"beat.memstats.rss",
					nil, nil,
				),
				eval: func(stats *Stats) float64 {
					return stats.Beat.Memstats.RSS
				},
				valType: prometheus.GaugeValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "runtime", "goroutines"),
					"beat.runtime.goroutines",
					nil, nil,
				),
				eval: func(stats *Stats) float64 {
					return float64(stats.Beat.Runtime.Goroutines)
				},
				valType: prometheus.GaugeValue,
			},
		},
	}
}

// Describe returns all descriptions of the collector.
func (c *beatCollector) Describe(ch chan<- *prometheus.Desc) {

	for _, metric := range c.metrics {
		ch <- metric.desc
	}

}

// Collect returns the current state of all metrics of the collector.
func (c *beatCollector) Collect(ch chan<- prometheus.Metric) {

	for _, i := range c.metrics {
		ch <- prometheus.MustNewConstMetric(i.desc, i.valType, i.eval(c.stats))
	}

}
