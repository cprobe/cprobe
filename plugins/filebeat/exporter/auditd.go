package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

// AuditdInfo json structure
type AuditdStats struct {
	KernelLost         float64 `json:"kernel_lost"`
	ReassemblerSeqGaps float64 `json:"reassembler_seq_gaps"`
	ReceivedMsgs       float64 `json:"received_msgs"`
	UserspaceLost      float64 `json:"userspace_lost"`
}

type auditdCollector struct {
	beatInfo *BeatInfo
	stats    *Stats
	metrics  exportedMetrics
}

// NewAuditdCollector constructor
func NewAuditdCollector(beatInfo *BeatInfo, stats *Stats) prometheus.Collector {
	return &auditdCollector{
		beatInfo: beatInfo,
		stats:    stats,
		metrics: exportedMetrics{
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "auditd", "kernel_lost"),
					"auditd.kernel_lost",
					nil, nil,
				),
				eval: func(stats *Stats) float64 {
					return stats.Auditd.KernelLost
				},
				valType: prometheus.GaugeValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "auditd", "reassembler_seq_gaps"),
					"auditd.reassembler_seq_gaps",
					nil, nil,
				),
				eval: func(stats *Stats) float64 {
					return stats.Auditd.ReassemblerSeqGaps
				},
				valType: prometheus.GaugeValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "auditd", "received_msgs"),
					"auditd.received_msgs",
					nil, nil,
				),
				eval: func(stats *Stats) float64 {
					return stats.Auditd.ReceivedMsgs
				},
				valType: prometheus.GaugeValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "auditd", "userspace_lost"),
					"auditd.userspace_lost",
					nil, nil,
				),
				eval: func(stats *Stats) float64 {
					return stats.Auditd.UserspaceLost
				},
				valType: prometheus.GaugeValue,
			},
		},
	}
}

// Describe returns all descriptions of the collector.
func (c *auditdCollector) Describe(ch chan<- *prometheus.Desc) {

	for _, metric := range c.metrics {
		ch <- metric.desc
	}

}

// Collect returns the current state of all metrics of the collector.
func (c *auditdCollector) Collect(ch chan<- prometheus.Metric) {

	for _, i := range c.metrics {
		ch <- prometheus.MustNewConstMetric(i.desc, i.valType, i.eval(c.stats))
	}

}
