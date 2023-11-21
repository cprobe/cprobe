package writer

import (
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/cprobe/cprobe/lib/prompbmarshal"
	"github.com/golang/snappy"
)

type Vector struct {
	Labels map[string]string
	Clock  int64
	Value  float64
}

func WriteVectors(vs []*Vector) {
	if len(vs) == 0 || len(WriterConfig.Writers) == 0 {
		return
	}

	// append global extra leabels
	if WriterConfig.Global != nil {
		for key, value := range WriterConfig.Global.ExtraLabels {
			for _, v := range vs {
				v.Labels[key] = value
			}
		}
	}

	if len(WriterConfig.Writers) == 1 {
		WriterConfig.Writers[0].writeVectors(vs)
		return
	}

	for i := range WriterConfig.Writers {
		newVectors := make([]*Vector, len(vs))
		for j := range vs {
			newVectors[j] = &Vector{
				Labels: make(map[string]string, len(vs[j].Labels)),
				Clock:  vs[j].Clock,
				Value:  vs[j].Value,
			}
			for key, value := range vs[j].Labels {
				newVectors[j].Labels[key] = value
			}
		}
		WriterConfig.Writers[i].writeVectors(newVectors)
	}
}

func (w *Writer) writeVectors(vs []*Vector) {
	// append writer extra labels
	for key, value := range w.ExtraLabels {
		for _, v := range vs {
			v.Labels[key] = value
		}
	}

	// convert vectors to time series
	tss := w.makeTimeSeries(vs)

	// relabel
	if WriterConfig.Global.ParsedRelabelConfigs.Len() > 0 {
		tss = new(relabelCtx).applyRelabeling(tss, WriterConfig.Global.ParsedRelabelConfigs)
	}

	if w.ParsedRelabelConfigs.Len() > 0 {
		tss = new(relabelCtx).applyRelabeling(tss, w.ParsedRelabelConfigs)
	}

	req := prompbmarshal.WriteRequest{
		Timeseries: tss,
	}

	bs, err := req.Marshal()
	if err != nil {
		logger.Warnf("failed to marshal WriteRequest: %s", err)
		return
	}

	httpReq, err := w.NewRequest(snappy.Encode(nil, bs))
	if err != nil {
		logger.Warnf("failed to create http request: %s", err)
		return
	}

	w.RequestQueue.PushFront(httpReq)
}

func (w *Writer) makeTimeSeries(vs []*Vector) []prompbmarshal.TimeSeries {
	tss := make([]prompbmarshal.TimeSeries, len(vs))
	for i := range vs {
		tss[i] = prompbmarshal.TimeSeries{
			Labels:  make([]prompbmarshal.Label, 0, len(vs[i].Labels)),
			Samples: []prompbmarshal.Sample{{Value: vs[i].Value, Timestamp: vs[i].Clock}},
		}
		for key, value := range vs[i].Labels {
			tss[i].Labels = append(tss[i].Labels, prompbmarshal.Label{Name: key, Value: value})
		}
	}
	return tss
}
