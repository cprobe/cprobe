package writer

import (
	"strings"

	"github.com/cprobe/cprobe/lib/logger"
	"github.com/cprobe/cprobe/lib/prompbmarshal"
	"github.com/golang/snappy"
)

func WriteTimeSeries(tss []prompbmarshal.TimeSeries) {
	if len(tss) == 0 {
		return
	}

	if *writerDisable {
		for i := range tss {
			point := tss[i]
			var sb strings.Builder
			for j := range point.Labels {
				sb.WriteString(point.Labels[j].Name)
				sb.WriteString("=")
				sb.WriteString(point.Labels[j].Value)
				sb.WriteString(" ")
			}
			logger.Infof(">> %s %d %f ", sb.String(), point.Samples[0].Timestamp, point.Samples[0].Value)
		}
		return
	}

	if len(WriterConfig.Writers) == 0 {
		return
	}

	// append global extra leabels
	if WriterConfig.Global != nil && WriterConfig.Global.ExtraLabels != nil && len(WriterConfig.Global.ExtraLabels.Labels) > 0 {
		new(relabelCtx).appendExtraLabels(tss, WriterConfig.Global.ExtraLabels.Labels)
	}

	if len(WriterConfig.Writers) == 1 {
		WriterConfig.Writers[0].writeTimeSeries(tss)
		return
	}

	for i := range WriterConfig.Writers {
		if i == len(WriterConfig.Writers)-1 {
			// last one
			WriterConfig.Writers[i].writeTimeSeries(tss)
		} else {
			newVectors := make([]prompbmarshal.TimeSeries, len(tss))
			for j := range tss {
				newVectors[j] = prompbmarshal.TimeSeries{
					Labels:  make([]prompbmarshal.Label, 0, len(tss[j].Labels)),
					Samples: tss[j].Samples,
				}
				newVectors[j].Labels = append(newVectors[j].Labels, tss[j].Labels...)
			}
			WriterConfig.Writers[i].writeTimeSeries(newVectors)
		}
	}
}

func (w *Writer) writeTimeSeries(tss []prompbmarshal.TimeSeries) {
	// append writer extra labels
	if w.ExtraLabels != nil && len(w.ExtraLabels.Labels) > 0 {
		new(relabelCtx).appendExtraLabels(tss, w.ExtraLabels.Labels)
	}

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
		logger.Warnf("cannot marshal WriteRequest: %s", err)
		return
	}

	httpReq, err := w.NewRequest(snappy.Encode(nil, bs))
	if err != nil {
		logger.Warnf("cannot create http request: %s", err)
		return
	}

	w.RequestQueue.PushFront(httpReq)
}
