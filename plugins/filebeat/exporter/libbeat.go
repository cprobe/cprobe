package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

// LibBeat json structure
type LibBeat struct {
	Config struct {
		Module struct {
			Running float64 `json:"running"`
			Starts  float64 `json:"starts"`
			Stops   float64 `json:"stops"`
		} `json:"module"`
		Reloads float64 `json:"reloads"`
	} `json:"config"`
	Output   LibBeatOutput   `json:"output"`
	Pipeline LibBeatPipeline `json:"pipeline"`
}

// LibBeatEvents json structure
type LibBeatEvents struct {
	Acked      float64 `json:"acked"`
	Active     float64 `json:"active"`
	Batches    float64 `json:"batches"`
	Dropped    float64 `json:"dropped"`
	Duplicates float64 `json:"duplicates"`
	Failed     float64 `json:"failed"`
	Filtered   float64 `json:"filtered"`
	Published  float64 `json:"published"`
	Retry      float64 `json:"retry"`
}

// LibBeatOutputBytesErrors json structure
type LibBeatOutputBytesErrors struct {
	Bytes  float64 `json:"bytes"`
	Errors float64 `json:"errors"`
}

// LibBeatOutput json structure
type LibBeatOutput struct {
	Events LibBeatEvents            `json:"events"`
	Read   LibBeatOutputBytesErrors `json:"read"`
	Write  LibBeatOutputBytesErrors `json:"write"`
	Type   string                   `json:"type"`
}

// LibBeatPipeline json structure
type LibBeatPipeline struct {
	Clients float64       `json:"clients"`
	Events  LibBeatEvents `json:"events"`
	Queue   struct {
		Acked float64 `json:"acked"`
	} `json:"queue"`
}

type libbeatCollector struct {
	beatInfo *BeatInfo
	stats    *Stats
	metrics  exportedMetrics
}

var libbeatOutputType *prometheus.Desc

// NewLibBeatCollector constructor
func NewLibBeatCollector(beatInfo *BeatInfo, stats *Stats) prometheus.Collector {
	return &libbeatCollector{
		beatInfo: beatInfo,
		stats:    stats,
		metrics: exportedMetrics{
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "libbeat_config", "reloads_total"),
					"libbeat.config.reloads",
					nil, nil,
				),
				eval: func(stats *Stats) float64 {
					return stats.LibBeat.Config.Reloads
				},
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "libbeat", "config"),
					"libbeat.config.module",
					nil, prometheus.Labels{"module": "running"},
				),
				eval: func(stats *Stats) float64 {
					return stats.LibBeat.Config.Module.Running
				},
				valType: prometheus.GaugeValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "libbeat", "config"),
					"libbeat.config.module",
					nil, prometheus.Labels{"module": "starts"},
				),
				eval: func(stats *Stats) float64 {
					return stats.LibBeat.Config.Module.Starts
				},
				valType: prometheus.GaugeValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "libbeat", "config"),
					"libbeat.config.module",
					nil, prometheus.Labels{"module": "stops"},
				),
				eval: func(stats *Stats) float64 {
					return stats.LibBeat.Config.Module.Stops
				},
				valType: prometheus.GaugeValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "libbeat", "output_read_bytes_total"),
					"libbeat.output.read.bytes",
					nil, nil,
				),
				eval: func(stats *Stats) float64 {
					return stats.LibBeat.Output.Read.Bytes
				},
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "libbeat", "output_read_errors_total"),
					"libbeat.output.read.errors",
					nil, nil,
				),
				eval: func(stats *Stats) float64 {
					return stats.LibBeat.Output.Read.Errors
				},
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "libbeat", "output_write_bytes_total"),
					"libbeat.output.write.bytes",
					nil, nil,
				),
				eval: func(stats *Stats) float64 {
					return stats.LibBeat.Output.Write.Bytes
				},
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "libbeat", "output_write_errors_total"),
					"libbeat.output.write.errors",
					nil, nil,
				),
				eval: func(stats *Stats) float64 {
					return stats.LibBeat.Output.Write.Errors
				},
				valType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "libbeat", "output_events"),
					"libbeat.output.events",
					nil, prometheus.Labels{"type": "acked"},
				),
				eval: func(stats *Stats) float64 {
					return stats.LibBeat.Output.Events.Acked
				},
				valType: prometheus.UntypedValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "libbeat", "output_events"),
					"libbeat.output.events",
					nil, prometheus.Labels{"type": "active"},
				),
				eval: func(stats *Stats) float64 {
					return stats.LibBeat.Output.Events.Active
				},
				valType: prometheus.UntypedValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "libbeat", "output_events"),
					"libbeat.output.events",
					nil, prometheus.Labels{"type": "batches"},
				),
				eval: func(stats *Stats) float64 {
					return stats.LibBeat.Output.Events.Batches
				},
				valType: prometheus.UntypedValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "libbeat", "output_events"),
					"libbeat.output.events",
					nil, prometheus.Labels{"type": "dropped"},
				),
				eval: func(stats *Stats) float64 {
					return stats.LibBeat.Output.Events.Dropped
				},
				valType: prometheus.UntypedValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "libbeat", "output_events"),
					"libbeat.output.events",
					nil, prometheus.Labels{"type": "duplicates"},
				),
				eval: func(stats *Stats) float64 {
					return stats.LibBeat.Output.Events.Duplicates
				},
				valType: prometheus.UntypedValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "libbeat", "output_events"),
					"libbeat.output.events",
					nil, prometheus.Labels{"type": "failed"},
				),
				eval: func(stats *Stats) float64 {
					return stats.LibBeat.Output.Events.Failed
				},
				valType: prometheus.UntypedValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "libbeat", "pipeline_clients"),
					"libbeat.pipeline.clients",
					nil, nil,
				),
				eval: func(stats *Stats) float64 {
					return stats.LibBeat.Pipeline.Clients
				},
				valType: prometheus.GaugeValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "libbeat", "pipeline_queue"),
					"libbeat.pipeline.queue",
					nil, prometheus.Labels{"type": "acked"},
				),
				eval: func(stats *Stats) float64 {
					return stats.LibBeat.Pipeline.Queue.Acked
				},
				valType: prometheus.UntypedValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "libbeat", "pipeline_events"),
					"libbeat.pipeline.events",
					nil, prometheus.Labels{"type": "active"},
				),
				eval: func(stats *Stats) float64 {
					return stats.LibBeat.Pipeline.Events.Active
				},
				valType: prometheus.UntypedValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "libbeat", "pipeline_events"),
					"libbeat.pipeline.events",
					nil, prometheus.Labels{"type": "dropped"},
				),
				eval: func(stats *Stats) float64 {
					return stats.LibBeat.Pipeline.Events.Dropped
				},
				valType: prometheus.UntypedValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "libbeat", "pipeline_events"),
					"libbeat.pipeline.events",
					nil, prometheus.Labels{"type": "failed"},
				),
				eval: func(stats *Stats) float64 {
					return stats.LibBeat.Pipeline.Events.Failed
				},
				valType: prometheus.UntypedValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "libbeat", "pipeline_events"),
					"libbeat.pipeline.events",
					nil, prometheus.Labels{"type": "filtered"},
				),
				eval: func(stats *Stats) float64 {
					return stats.LibBeat.Pipeline.Events.Filtered
				},
				valType: prometheus.UntypedValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "libbeat", "pipeline_events"),
					"libbeat.pipeline.events",
					nil, prometheus.Labels{"type": "published"},
				),
				eval: func(stats *Stats) float64 {
					return stats.LibBeat.Pipeline.Events.Published
				},
				valType: prometheus.UntypedValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "libbeat", "pipeline_events"),
					"libbeat.pipeline.events",
					nil, prometheus.Labels{"type": "retry"},
				),
				eval: func(stats *Stats) float64 {
					return stats.LibBeat.Pipeline.Events.Retry
				},
				valType: prometheus.UntypedValue,
			},
		},
	}
}

// Describe returns all descriptions of the collector.
func (c *libbeatCollector) Describe(ch chan<- *prometheus.Desc) {

	for _, metric := range c.metrics {
		ch <- metric.desc
	}

	libbeatOutputType = prometheus.NewDesc(
		prometheus.BuildFQName(c.beatInfo.Beat, "libbeat", "output_total"),
		"libbeat.output.type",
		[]string{"type"}, nil,
	)

	ch <- libbeatOutputType

}

// Collect returns the current state of all metrics of the collector.
func (c *libbeatCollector) Collect(ch chan<- prometheus.Metric) {

	for _, i := range c.metrics {
		ch <- prometheus.MustNewConstMetric(i.desc, i.valType, i.eval(c.stats))
	}

	// output.type with dynamic label
	ch <- prometheus.MustNewConstMetric(libbeatOutputType, prometheus.CounterValue, float64(1), c.stats.LibBeat.Output.Type)

}
