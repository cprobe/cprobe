package writer

import (
	"github.com/cprobe/cprobe/lib/prompbmarshal"
	"github.com/cprobe/cprobe/lib/promrelabel"
)

type relabelCtx struct {
	// pool for labels, which are used during the relabeling.
	labels []prompbmarshal.Label
}

func fixPromCompatibleNaming(labels []prompbmarshal.Label) {
	// Replace unsupported Prometheus chars in label names and metric names with underscores.
	for i := range labels {
		label := &labels[i]
		if label.Name == "__name__" {
			label.Value = promrelabel.SanitizeMetricName(label.Value)
		} else {
			label.Name = promrelabel.SanitizeLabelName(label.Name)
		}
	}
}

func (rctx *relabelCtx) applyRelabeling(tss []prompbmarshal.TimeSeries, pcs *promrelabel.ParsedConfigs) []prompbmarshal.TimeSeries {
	if pcs.Len() == 0 {
		// Nothing to change.
		return tss
	}

	tssDst := tss[:0]
	labels := rctx.labels[:0]
	for i := range tss {
		ts := &tss[i]
		labelsLen := len(labels)
		labels = append(labels, ts.Labels...)
		labels = pcs.Apply(labels, labelsLen)
		// labels = promrelabel.FinalizeLabels(labels[:labelsLen], labels[labelsLen:])
		if len(labels) == labelsLen {
			// Drop the current time series, since relabeling removed all the labels.
			continue
		}
		fixPromCompatibleNaming(labels[labelsLen:])
		tssDst = append(tssDst, prompbmarshal.TimeSeries{
			Labels:  labels[labelsLen:],
			Samples: ts.Samples,
		})
	}
	rctx.labels = labels
	return tssDst
}
