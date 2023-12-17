package collector

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/cprobe/cprobe/lib/logger"
	"github.com/cprobe/cprobe/plugins/elasticsearch/pkg/roundtripper"
	"github.com/cprobe/cprobe/types"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

func (c *Config) Scrape(ctx context.Context, _target string, ss *types.Samples) error {
	target, err := url.Parse(_target)
	if err != nil {
		return errors.WithMessagef(err, "invalid target: %s", _target)
	}

	if c.Username != "" && c.Password != "" {
		target.User = url.UserPassword(c.Username, c.Password)
	}

	// returns nil if not provided and falls back to simple TCP.
	tlsConfig := createTLSConfig(c.TLSCa, c.TLSClientCert, c.TLSClientPrivateKey, c.TLSSkipVerify)
	var httpTransport http.RoundTripper

	httpTransport = &http.Transport{
		TLSClientConfig: tlsConfig,
		Proxy:           http.ProxyFromEnvironment,
	}

	if c.ApiKey != "" {
		httpTransport = &transportWithAPIKey{
			underlyingTransport: httpTransport,
			apiKey:              c.ApiKey,
		}
	}

	httpClient := &http.Client{
		Timeout:   c.Timeout,
		Transport: httpTransport,
	}

	if c.AwsRegion != "" {
		httpClient.Transport, err = roundtripper.NewAWSSigningTransport(httpTransport, c.AwsRegion, c.AwsRoleArn)
		if err != nil {
			return errors.WithMessage(err, "failed to create AWS transport")
		}
	}

	defer httpClient.CloseIdleConnections()

	var clusterName string

	// 下面的几个 gather 方法是魔改了 exporter，不过指标名称尽量保持一致
	// 这个改动比原来更清晰，原来的 exporter 考虑了把 cluster 信息周期性更新，引入了复杂度，实在没必要
	// cprobe 这个场景只需要每次采集的一开始获取一次 cluster 信息即可
	if clusterName, err = c.gatherClusterInfo(ctx, target, httpClient, ss); err != nil {
		return errors.WithMessage(err, "failed to gather cluster info")
	}

	if err = c.gatherClusterHealth(ctx, target, httpClient, ss, clusterName); err != nil {
		return errors.WithMessage(err, "failed to gather cluster health")
	}

	if err = c.gatherClusterSettings(ctx, target, httpClient, ss); err != nil {
		logger.Errorf("failed to gather cluster settings: %s", err)
	}

	if err = c.gatherSnapshots(ctx, target, httpClient, ss); err != nil {
		logger.Errorf("failed to gather snapshots: %s", err)
	}

	// 换成另一个写法，尽量不改动原本的 exporter 的逻辑，传递 channel 收集数据，主要是这部分指标太多了，改起来费劲
	nodes := NewNodes(httpClient, target, c.GatherNode == "*", c.GatherNode)

	nodesCh := make(chan prometheus.Metric)
	go func() {
		nodes.Collect(nodesCh)
		close(nodesCh)
	}()

	for m := range nodesCh {
		if err := ss.AddPromMetric(m); err != nil {
			logger.Warnf("failed to transform nodes metric: %s", err)
		}
	}

	if c.GatherIndices || c.GatherIndicesShards {
		indices := NewIndices(httpClient, target, c.GatherIndicesShards, c.GatherIndicesUseAlias, clusterName)
		indicesCh := make(chan prometheus.Metric)
		go func() {
			indices.Collect(indicesCh)
			close(indicesCh)
		}()
		for m := range indicesCh {
			if err := ss.AddPromMetric(m); err != nil {
				logger.Warnf("failed to transform indices metric: %s", err)
			}
		}

		if err = c.gatherShardsTotal(ctx, target, httpClient, ss, clusterName); err != nil {
			logger.Errorf("failed to gather shards total: %s", err)
		}
	}

	if c.GatherSlm {
		slm := NewSLM(httpClient, target)
		slmCh := make(chan prometheus.Metric)
		go func() {
			slm.Collect(slmCh)
			close(slmCh)
		}()
		for m := range slmCh {
			if err := ss.AddPromMetric(m); err != nil {
				logger.Warnf("failed to transform slm metric: %s", err)
			}
		}
	}

	if c.GatherDataStream {
		dataStream := NewDataStream(httpClient, target)
		dataStreamCh := make(chan prometheus.Metric)
		go func() {
			dataStream.Collect(dataStreamCh)
			close(dataStreamCh)
		}()
		for m := range dataStreamCh {
			if err := ss.AddPromMetric(m); err != nil {
				logger.Warnf("failed to transform data stream metric: %s", err)
			}
		}
	}

	if c.GatherIndicesSettings {
		indicesSettings := NewIndicesSettings(httpClient, target)
		indicesSettingsCh := make(chan prometheus.Metric)
		go func() {
			indicesSettings.Collect(indicesSettingsCh)
			close(indicesSettingsCh)
		}()
		for m := range indicesSettingsCh {
			if err := ss.AddPromMetric(m); err != nil {
				logger.Warnf("failed to transform indices settings metric: %s", err)
			}
		}
	}

	if c.GatherIndicesMappings {
		indicesMappings := NewIndicesMappings(httpClient, target)
		indicesMappingsCh := make(chan prometheus.Metric)
		go func() {
			indicesMappings.Collect(indicesMappingsCh)
			close(indicesMappingsCh)
		}()
		for m := range indicesMappingsCh {
			if err := ss.AddPromMetric(m); err != nil {
				logger.Warnf("failed to transform indices mappings metric: %s", err)
			}
		}
	}

	if c.GatherIlm {
		ilmStatus := NewIlmStatus(httpClient, target)
		ilmStatusCh := make(chan prometheus.Metric)
		go func() {
			ilmStatus.Collect(ilmStatusCh)
			close(ilmStatusCh)
		}()
		for m := range ilmStatusCh {
			if err := ss.AddPromMetric(m); err != nil {
				logger.Warnf("failed to transform ilm status metric: %s", err)
			}
		}

		ilmIndicies := NewIlmIndicies(httpClient, target)
		ilmIndiciesCh := make(chan prometheus.Metric)
		go func() {
			ilmIndicies.Collect(ilmIndiciesCh)
			close(ilmIndiciesCh)
		}()
		for m := range ilmIndiciesCh {
			if err := ss.AddPromMetric(m); err != nil {
				logger.Warnf("failed to transform ilm indicies metric: %s", err)
			}
		}
	}

	return nil
}

type transportWithAPIKey struct {
	underlyingTransport http.RoundTripper
	apiKey              string
}

func (t *transportWithAPIKey) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", fmt.Sprintf("ApiKey %s", t.apiKey))
	return t.underlyingTransport.RoundTrip(req)
}
