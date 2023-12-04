package types

import (
	"fmt"
	"math"

	"github.com/cprobe/cprobe/lib/listx"
	"github.com/cprobe/cprobe/types/metric"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type Samples struct {
	slist *listx.SafeList[metric.Metric]
}

func NewSamples() *Samples {
	return &Samples{
		slist: listx.NewSafeList[metric.Metric](),
	}
}

func (s *Samples) AddPromMetric(m prometheus.Metric) error {
	desc := m.Desc()
	if desc.Err() != nil {
		return desc.Err()
	}

	pb := &dto.Metric{}
	err := m.Write(pb)
	if err != nil {
		return err
	}

	tags := make(map[string]string)

	for _, kv := range desc.ConstLabels() {
		tags[kv.GetName()] = kv.GetValue()
	}

	for _, v := range pb.Label {
		tags[v.GetName()] = v.GetValue()
	}

	if pb.Gauge != nil {
		s.AddMetric(desc.Name(), map[string]interface{}{
			"": pb.Gauge.GetValue(),
		}, tags)
	} else if pb.Counter != nil {
		s.AddMetric(desc.Name(), map[string]interface{}{
			"": pb.Counter.GetValue(),
		}, tags)
	} else if pb.Summary != nil {
		s.handleSummary(pb, desc.Name(), tags)
	} else if pb.Histogram != nil {
		s.handleHistogram(pb, desc.Name(), tags)
	} else {
		s.AddMetric(desc.Name(), map[string]interface{}{
			"": pb.Untyped.GetValue(),
		}, tags)
	}

	return nil
}

func (s *Samples) handleSummary(pb *dto.Metric, metricName string, tags map[string]string) {
	count := pb.GetSummary().GetSampleCount()
	sum := pb.GetSummary().GetSampleSum()

	s.AddMetric(metricName, map[string]interface{}{
		"count": count,
		"sum":   sum,
	}, tags)

	for _, q := range pb.GetSummary().Quantile {
		s.AddMetric(metricName, map[string]interface{}{
			"quantile": q.GetValue(),
		}, tags, map[string]string{
			"quantile": fmt.Sprint(q.GetQuantile()),
		})
	}
}

func (s *Samples) handleHistogram(pb *dto.Metric, metricName string, tags map[string]string) {
	count := pb.GetHistogram().GetSampleCount()
	sum := pb.GetHistogram().GetSampleSum()

	s.AddMetric(metricName, map[string]interface{}{
		"count": count,
		"sum":   sum,
	}, tags)

	s.AddMetric(metricName, map[string]interface{}{
		"bucket": count,
	}, tags, map[string]string{
		"le": "+Inf",
	})

	for _, b := range pb.GetHistogram().Bucket {
		le := fmt.Sprint(b.GetUpperBound())
		value := float64(b.GetCumulativeCount())
		s.AddMetric(metricName, map[string]interface{}{
			"bucket": value,
		}, tags, map[string]string{
			"le": le,
		})
	}
}

func getNameAndValue(m *dto.Metric, metricName string) map[string]interface{} {
	fields := make(map[string]interface{})
	if m.Gauge != nil {
		if !math.IsNaN(m.GetGauge().GetValue()) {
			fields[metricName] = m.GetGauge().GetValue()
		}
	} else if m.Counter != nil {
		if !math.IsNaN(m.GetCounter().GetValue()) {
			fields[metricName] = m.GetCounter().GetValue()
		}
	} else if m.Untyped != nil {
		if !math.IsNaN(m.GetUntyped().GetValue()) {
			fields[metricName] = m.GetUntyped().GetValue()
		}
	}
	return fields
}

func (s *Samples) AddMetricFamilies(mfs []*dto.MetricFamily) {
	for i := range mfs {
		mf := mfs[i]
		metricName := mf.GetName()

		for _, m := range mf.GetMetric() {

			tags := make(map[string]string)
			for _, lb := range m.GetLabel() {
				tags[lb.GetName()] = lb.GetValue()
			}

			if mf.GetType() == dto.MetricType_SUMMARY {
				s.handleSummary(m, metricName, tags)
			} else if mf.GetType() == dto.MetricType_HISTOGRAM {
				s.handleHistogram(m, metricName, tags)
			} else {
				fields := getNameAndValue(m, metricName)
				s.AddMetric("", fields, tags)
			}
		}
	}
}

func (s *Samples) AddMetric(mesurement string, fields map[string]interface{}, tagss ...map[string]string) {
	tags := make(map[string]string)
	for i := range tagss {
		for k, v := range tagss[i] {
			tags[k] = v
		}
	}

	m := metric.New(mesurement, tags, fields, 0)
	s.slist.PushFront(m)
}

func (s *Samples) PushFront(m metric.Metric) {
	s.slist.PushFront(m)
}

func (s *Samples) PushFrontN(ms []metric.Metric) {
	s.slist.PushFrontN(ms)
}

func (s *Samples) PopBackN(n int) []metric.Metric {
	return s.slist.PopBackN(n)
}

func (s *Samples) PopBackAll() []metric.Metric {
	return s.slist.PopBackAll()
}

func (s *Samples) Len() int {
	return s.slist.Len()
}
