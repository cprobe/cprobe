package exporter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3" //TODO 版本问题
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/cprobe/cprobe/plugins/json/config"

	"github.com/prometheus/client_golang/prometheus"
	pconfig "github.com/prometheus/common/config"
)

func MakeMetricName(parts ...string) string {
	return strings.Join(parts, "_")
}

func SanitizeValue(s string) (float64, error) {
	var err error
	var value float64
	var resultErr string

	if value, err = strconv.ParseFloat(s, 64); err == nil {
		return value, nil
	}
	resultErr = fmt.Sprintf("%s", err)

	if boolValue, err := strconv.ParseBool(s); err == nil {
		if boolValue {
			return 1.0, nil
		}
		return 0.0, nil
	}
	resultErr = resultErr + "; " + fmt.Sprintf("%s", err)

	if s == "<nil>" {
		return math.NaN(), nil
	}
	return value, fmt.Errorf(resultErr)
}

func SanitizeIntValue(s string) (int64, error) {
	var err error
	var value int64
	var resultErr string

	if value, err = strconv.ParseInt(s, 10, 64); err == nil {
		return value, nil
	}
	resultErr = fmt.Sprintf("%s", err)

	return value, fmt.Errorf(resultErr)
}

func CreateMetricsList(c config.Module) ([]JSONMetric, error) {
	var (
		metrics   []JSONMetric
		valueType prometheus.ValueType
	)
	for _, metric := range c.Metrics {
		switch metric.ValueType {
		case config.ValueTypeGauge:
			valueType = prometheus.GaugeValue
		case config.ValueTypeCounter:
			valueType = prometheus.CounterValue
		default:
			valueType = prometheus.UntypedValue
		}
		switch metric.Type {
		case config.ValueScrape:
			var variableLabels, variableLabelsValues []string
			for k, v := range metric.Labels {
				variableLabels = append(variableLabels, k)
				variableLabelsValues = append(variableLabelsValues, v)
			}
			jsonMetric := JSONMetric{
				Type: config.ValueScrape,
				Desc: prometheus.NewDesc(
					metric.Name,
					metric.Help,
					variableLabels,
					nil,
				),
				KeyJSONPath:            metric.Path,
				LabelsJSONPaths:        variableLabelsValues,
				ValueType:              valueType,
				EpochTimestampJSONPath: metric.EpochTimestamp,
			}
			metrics = append(metrics, jsonMetric)
		case config.ObjectScrape:
			for subName, valuePath := range metric.Values {
				name := MakeMetricName(metric.Name, subName)
				var variableLabels, variableLabelsValues []string
				for k, v := range metric.Labels {
					variableLabels = append(variableLabels, k)
					variableLabelsValues = append(variableLabelsValues, v)
				}
				jsonMetric := JSONMetric{
					Type: config.ObjectScrape,
					Desc: prometheus.NewDesc(
						name,
						metric.Help,
						variableLabels,
						nil,
					),
					KeyJSONPath:            metric.Path,
					ValueJSONPath:          valuePath,
					LabelsJSONPaths:        variableLabelsValues,
					ValueType:              valueType,
					EpochTimestampJSONPath: metric.EpochTimestamp,
				}
				metrics = append(metrics, jsonMetric)
			}
		default:
			return nil, fmt.Errorf("Unknown metric type: '%s', for metric: '%s'", metric.Type, metric.Name)
		}
	}
	return metrics, nil
}

type JSONFetcher struct {
	module config.Module
	ctx    context.Context
	method string
	body   io.Reader
}

func NewJSONFetcher(ctx context.Context, m config.Module, tplValues url.Values) *JSONFetcher {
	method, body := renderBody(m.Body, tplValues)
	return &JSONFetcher{
		module: m,
		ctx:    ctx,
		method: method,
		body:   body,
	}
}

func (f *JSONFetcher) FetchJSON(endpoint string) ([]byte, error) {
	httpClientConfig := f.module.HTTPClientConfig
	client, err := pconfig.NewClientFromConfig(httpClientConfig, "fetch_json", pconfig.WithKeepAlivesDisabled(), pconfig.WithHTTP2Disabled())
	if err != nil {
		logger.Errorf("Error generating HTTP client, err: %s", err)
		return nil, err
	}

	var req *http.Request
	req, err = http.NewRequest(f.method, endpoint, f.body)
	req = req.WithContext(f.ctx)
	if err != nil {
		logger.Errorf("Failed to create request, err: %s", err)
		return nil, err
	}

	for key, value := range f.module.Headers {
		req.Header.Add(key, value)
	}
	if req.Header.Get("Accept") == "" {
		req.Header.Add("Accept", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		if _, err := io.Copy(io.Discard, resp.Body); err != nil {
			logger.Errorf("Failed to discard body, err: %s", err)
		}
		resp.Body.Close()
	}()

	if len(f.module.ValidStatusCodes) != 0 {
		success := false
		for _, code := range f.module.ValidStatusCodes {
			if resp.StatusCode == code {
				success = true
				break
			}
		}
		if !success {
			return nil, errors.New(resp.Status)
		}
	} else if resp.StatusCode/100 != 2 {
		return nil, errors.New(resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Use the configured template to render the body if enabled
// Do not treat template errors as fatal, on such errors just log them
// and continue with static body content
func renderBody(body config.Body, tplValues url.Values) (method string, br io.Reader) {
	method = "POST"
	if body.Content == "" {
		return "GET", nil
	}
	br = strings.NewReader(body.Content)
	if body.Templatize {
		tpl, err := template.New("base").Funcs(sprig.TxtFuncMap()).Parse(body.Content)
		if err != nil {
			logger.Errorf("Failed to create a new template from body content, err: %s", err)
			return
		}
		tpl = tpl.Option("missingkey=zero")
		var b strings.Builder
		if err := tpl.Execute(&b, tplValues); err != nil {
			logger.Errorf("Failed to render template with values, err: %s", err)
			return
		}
		br = strings.NewReader(b.String())
	}
	return
}
