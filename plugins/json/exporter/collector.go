package exporter

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/cprobe/cprobe/lib/logger"
	"github.com/cprobe/cprobe/plugins/json/config"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/util/jsonpath"
)

type JSONMetricCollector struct {
	JSONMetrics []JSONMetric
	Data        []byte
}

type JSONMetric struct {
	Desc                   *prometheus.Desc
	Type                   config.ScrapeType
	KeyJSONPath            string
	ValueJSONPath          string
	LabelsJSONPaths        []string
	ValueType              prometheus.ValueType
	EpochTimestampJSONPath string
}

func (mc JSONMetricCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range mc.JSONMetrics {
		ch <- m.Desc
	}
}

func (mc JSONMetricCollector) Collect(ch chan<- prometheus.Metric) {
	for _, m := range mc.JSONMetrics {
		switch m.Type {
		case config.ValueScrape:
			value, err := extractValue(mc.Data, m.KeyJSONPath, false)
			if err != nil {
				logger.Errorf("Failed to extract value for metric, path: %s, err: %s", m.KeyJSONPath, err)
				continue
			}

			if floatValue, err := SanitizeValue(value); err == nil {
				metric := prometheus.MustNewConstMetric(
					m.Desc,
					m.ValueType,
					floatValue,
					extractLabels(mc.Data, m.LabelsJSONPaths)...,
				)
				ch <- timestampMetric(m, mc.Data, metric)
			} else {
				logger.Errorf("Failed to convert extracted value to float64, path: %s, value: %s, err: %s, metric: %s", m.KeyJSONPath, value, err, m.Desc)
				continue
			}

		case config.ObjectScrape:
			values, err := extractValue(mc.Data, m.KeyJSONPath, true)
			if err != nil {
				logger.Errorf("Failed to extract json objects for metric, path: %s, err: %s, metric: %s", m.KeyJSONPath, err, m.Desc)
				continue
			}

			var jsonData []interface{}
			if err := json.Unmarshal([]byte(values), &jsonData); err == nil {
				for _, data := range jsonData {
					jdata, err := json.Marshal(data)
					if err != nil {
						logger.Errorf("Failed to marshal data to json, path: %s, err: %s, metric: %s, data: %s", m.ValueJSONPath, err, m.Desc, data)
						continue
					}
					value, err := extractValue(jdata, m.ValueJSONPath, false)
					if err != nil {
						logger.Errorf("Failed to extract value for metric, path: %s, err: %s, metric: %s", m.ValueJSONPath, err, m.Desc)
						continue
					}

					if floatValue, err := SanitizeValue(value); err == nil {
						metric := prometheus.MustNewConstMetric(
							m.Desc,
							m.ValueType,
							floatValue,
							extractLabels(jdata, m.LabelsJSONPaths)...,
						)
						ch <- timestampMetric(m, jdata, metric)
					} else {
						logger.Errorf("Failed to convert extracted value to float64, path: %s, value: %s, err: %s, metric: %s", m.ValueJSONPath, value, err, m.Desc)
						continue
					}
				}
			} else {
				logger.Errorf("Failed to convert extracted objects to json, path: %s, err: %s, metric: %s", m.KeyJSONPath, err, m.Desc)
				continue
			}
		default:
			logger.Errorf("Unknown scrape config type, type: %s, metric: %s", m.Type, m.Desc)
			continue
		}
	}
}

// Returns the last matching value at the given json path
func extractValue(data []byte, path string, enableJSONOutput bool) (string, error) {
	var jsonData interface{}
	buf := new(bytes.Buffer)

	j := jsonpath.New("jp")
	if enableJSONOutput {
		j.EnableJSONOutput(true)
	}

	if err := json.Unmarshal(data, &jsonData); err != nil {
		logger.Errorf("Failed to unmarshal data to json, err: %s, data: %s", err, data)
		return "", err
	}

	if err := j.Parse(path); err != nil {
		logger.Errorf("Failed to parse jsonpath, err: %s, path: %s, data: %s", err, path, data)
		return "", err
	}

	if err := j.Execute(buf, jsonData); err != nil {
		logger.Errorf("Failed to execute jsonpath, err: %s, path: %s, data: %s", err, path, data)
		return "", err
	}

	// Since we are finally going to extract only float64, unquote if necessary
	if res, err := jsonpath.UnquoteExtend(buf.String()); err == nil {
		return res, nil
	}

	return buf.String(), nil
}

// Returns the list of labels created from the list of provided json paths
func extractLabels(data []byte, paths []string) []string {
	labels := make([]string, len(paths))
	for i, path := range paths {
		if result, err := extractValue(data, path, false); err == nil {
			labels[i] = result
		} else {
			logger.Errorf("Failed to extract label value, err: %s, path: %s, data: %s", err, path, data)
		}
	}
	return labels
}

func timestampMetric(m JSONMetric, data []byte, pm prometheus.Metric) prometheus.Metric {
	if m.EpochTimestampJSONPath == "" {
		return pm
	}
	ts, err := extractValue(data, m.EpochTimestampJSONPath, false)
	if err != nil {
		logger.Errorf("Failed to extract timestamp for metric, path: %s, err: %s, metric: %s", m.KeyJSONPath, err, m.Desc)
		return pm
	}
	epochTime, err := SanitizeIntValue(ts)
	if err != nil {
		logger.Errorf("Failed to parse timestamp for metric, path: %s, err: %s, metric: %s", m.KeyJSONPath, err, m.Desc)
		return pm
	}
	timestamp := time.UnixMilli(epochTime)
	return prometheus.NewMetricWithTimestamp(timestamp, pm)
}
