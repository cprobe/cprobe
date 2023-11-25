package types

import (
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
	} else {
		s.AddMetric(desc.Name(), map[string]interface{}{
			"": pb.Untyped.GetValue(),
		}, tags)
	}

	return nil
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
