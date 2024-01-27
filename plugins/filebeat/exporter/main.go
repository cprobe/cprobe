package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"

	"github.com/cprobe/cprobe/lib/logger"
	"github.com/prometheus/client_golang/prometheus"
)

type mainCollector struct {
	Collectors map[string]prometheus.Collector
	Stats      *Stats
	client     *http.Client
	beatURL    *url.URL
	name       string
	beatInfo   *BeatInfo
	targetDesc *prometheus.Desc
	targetUp   *prometheus.Desc
	metrics    exportedMetrics
	systemBeat bool
}

// HackfixRegex regex to replace JSON part
var HackfixRegex = regexp.MustCompile("\"time\":(\\d+)") // replaces time:123 to time.ms:123, only filebeat has different naming of time metric

// NewMainCollector constructor
func NewMainCollector(client *http.Client, url *url.URL, name string, beatInfo *BeatInfo, systemBeat bool) prometheus.Collector {
	instance := fmt.Sprintf("%s:%s", url.Hostname(), url.Port())
	beat := &mainCollector{
		Collectors: make(map[string]prometheus.Collector),
		Stats:      &Stats{},
		client:     client,
		beatURL:    url,
		name:       name,
		targetDesc: prometheus.NewDesc(
			prometheus.BuildFQName(name, "target", "info"),
			"target information",
			nil,
			prometheus.Labels{"version": beatInfo.Version, "beat": beatInfo.Beat, "uri": instance}),
		targetUp: prometheus.NewDesc(
			prometheus.BuildFQName("", beatInfo.Beat, "up"),
			"Target up",
			nil,
			nil),

		beatInfo:   beatInfo,
		metrics:    exportedMetrics{},
		systemBeat: systemBeat,
	}

	beat.Collectors["system"] = NewSystemCollector(beatInfo, beat.Stats)
	beat.Collectors["beat"] = NewBeatCollector(beatInfo, beat.Stats)
	beat.Collectors["libbeat"] = NewLibBeatCollector(beatInfo, beat.Stats)
	beat.Collectors["registrar"] = NewRegistrarCollector(beatInfo, beat.Stats)
	beat.Collectors["filebeat"] = NewFilebeatCollector(beatInfo, beat.Stats)
	beat.Collectors["metricbeat"] = NewMetricbeatCollector(beatInfo, beat.Stats)
	beat.Collectors["auditd"] = NewAuditdCollector(beatInfo, beat.Stats)

	return beat
}

// Describe returns all descriptions of the collector.
func (b *mainCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- b.targetDesc
	ch <- b.targetUp

	for _, metric := range b.metrics {
		ch <- metric.desc
	}

	// standard collectors for all types of beats
	if b.systemBeat {
		b.Collectors["system"].Describe(ch)
	}
	b.Collectors["beat"].Describe(ch)
	b.Collectors["libbeat"].Describe(ch)
	b.Collectors["auditd"].Describe(ch)

	// Customized collectors per beat type
	switch b.beatInfo.Beat {
	case "filebeat":
		b.Collectors["filebeat"].Describe(ch)
		b.Collectors["registrar"].Describe(ch)
	case "metricbeat":
		b.Collectors["metricbeat"].Describe(ch)
	}

}

// Collect returns the current state of all metrics of the collector.
func (b *mainCollector) Collect(ch chan<- prometheus.Metric) {

	err := b.fetchStatsEndpoint()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(b.targetUp, prometheus.GaugeValue, float64(0)) // set target down
		logger.Errorf("Failed getting /stats endpoint of target: " + err.Error())
		return
	}

	ch <- prometheus.MustNewConstMetric(b.targetDesc, prometheus.GaugeValue, float64(1))
	ch <- prometheus.MustNewConstMetric(b.targetUp, prometheus.GaugeValue, float64(1)) // target up

	for _, i := range b.metrics {
		ch <- prometheus.MustNewConstMetric(i.desc, i.valType, i.eval(b.Stats))
	}

	// standard collectors for all types of beats
	if b.systemBeat {
		b.Collectors["system"].Collect(ch)
	}
	b.Collectors["beat"].Collect(ch)
	b.Collectors["libbeat"].Collect(ch)
	b.Collectors["auditd"].Collect(ch)

	// Customized collectors per beat type
	switch b.beatInfo.Beat {
	case "filebeat":
		b.Collectors["filebeat"].Collect(ch)
		b.Collectors["registrar"].Collect(ch)
	case "metricbeat":
		b.Collectors["metricbeat"].Collect(ch)
	}

}

func (b *mainCollector) fetchStatsEndpoint() error {

	response, err := b.client.Get(b.beatURL.String() + "/stats")
	if err != nil {
		logger.Errorf("Could not fetch stats endpoint of target: %v", b.beatURL.String())
		return err
	}

	defer response.Body.Close()

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		logger.Errorf("Can't read body of response")
		return err
	}

	bodyBytes = HackfixRegex.ReplaceAll(bodyBytes, []byte("\"time\":{\"ms\":$1}"))

	err = json.Unmarshal(bodyBytes, &b.Stats)
	if err != nil {
		logger.Errorf("Could not parse JSON response for target")
		return err
	}

	return nil
}
