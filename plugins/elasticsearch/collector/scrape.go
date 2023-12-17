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
