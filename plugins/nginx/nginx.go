package nginx

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/cprobe/cprobe/lib/buildinfo"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/cprobe/cprobe/plugins"
	"github.com/cprobe/cprobe/plugins/nginx/client"
	"github.com/cprobe/cprobe/plugins/nginx/collector"
	"github.com/cprobe/cprobe/types"
	plusclient "github.com/nginxinc/nginx-plus-go-client/client"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type Config struct {
	NginxPlus     bool          `toml:"nginx_plus"`
	SSLCACert     string        `toml:"ssl_ca_cert"`
	SSLClientCert string        `toml:"ssl_client_cert"`
	SSLClientKey  string        `toml:"ssl_client_key"`
	SSLVerify     bool          `toml:"ssl_verify"`
	Timeout       time.Duration `toml:"timeout"`
}

type Nginx struct {
}

func init() {
	plugins.RegisterPlugin(types.PluginNginx, &Nginx{})
}

func (n *Nginx) ParseConfig(baseDir string, bs []byte) (any, error) {
	var c Config
	err := toml.Unmarshal(bs, &c)
	if err != nil {
		return nil, err
	}

	if c.Timeout == 0 {
		c.Timeout = time.Millisecond * 500
	}
	return &c, nil
}

func (n *Nginx) Scrape(ctx context.Context, target string, cfg any, ss *types.Samples) error {
	conf := cfg.(*Config)
	if !strings.HasPrefix(target, "http") {
		target = "http://" + target
	}

	// #nosec G402
	sslConfig := &tls.Config{InsecureSkipVerify: !conf.SSLVerify}
	if conf.SSLCACert != "" {
		caCert, err := os.ReadFile(conf.SSLCACert)
		if err != nil {
			logger.Errorf("Loading CA cert failed: %s", err)
			return err
		}
		sslCaCertPool := x509.NewCertPool()
		ok := sslCaCertPool.AppendCertsFromPEM(caCert)
		if !ok {
			logger.Errorf("Loading CA cert failed: %s", err)
			return err
		}
		sslConfig.RootCAs = sslCaCertPool
	}

	if conf.SSLClientCert != "" && conf.SSLClientKey != "" {
		clientCert, err := tls.LoadX509KeyPair(conf.SSLClientCert, conf.SSLClientKey)
		if err != nil {
			logger.Errorf("Loading client certificate failed: %s", err)
			return err
		}
		sslConfig.Certificates = []tls.Certificate{clientCert}
	}
	transport := &http.Transport{
		TLSClientConfig: sslConfig,
	}
	collect, err := registerCollector(transport, target, nil, conf)
	if err != nil {
		return err
	}

	ch := make(chan prometheus.Metric)
	go func() {
		collect.Collect(ch)
		close(ch)
	}()

	for m := range ch {
		if err := ss.AddPromMetric(m); err != nil {
			logger.Warnf("failed to transform prometheus metric: %s", err)
		}
	}
	return nil
}

func registerCollector(transport *http.Transport,
	addr string, labels map[string]string, conf *Config) (prometheus.Collector, error) {
	if strings.HasPrefix(addr, "unix:") {
		socketPath, requestPath, err := parseUnixSocketAddress(addr)
		if err != nil {
			logger.Warnf("Parsing unix domain socket scrape address failed url:%s", addr, err)
			return nil, err
		}

		transport.DialContext = func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		}
		addr = "http://unix" + requestPath
	}

	userAgent := fmt.Sprintf("Cprobe-NGINX-Exporter/v%v", buildinfo.Version)

	httpClient := &http.Client{
		Timeout: conf.Timeout,
		Transport: &userAgentRoundTripper{
			agent: userAgent,
			rt:    transport,
		},
	}

	if conf.NginxPlus {
		plusClient, err := plusclient.NewNginxClient(addr, plusclient.WithHTTPClient(httpClient))
		if err != nil {
			logger.Warnf("Could not create Nginx Plus Client", err)
			return nil, err
		}
		variableLabelNames := collector.NewVariableLabelNames(nil, nil, nil, nil, nil, nil, nil, nil)
		plusCollector := collector.NewNginxPlusCollector(plusClient, "nginxplus", variableLabelNames, labels)
		return plusCollector, nil
	} else {
		ossClient := client.NewNginxClient(httpClient, addr)
		nginxCollector := collector.NewNginxCollector(ossClient, "nginx", labels)
		return nginxCollector, nil
	}
}

func parseUnixSocketAddress(address string) (string, string, error) {
	addressParts := strings.Split(address, ":")
	addressPartsLength := len(addressParts)

	if addressPartsLength > 3 || addressPartsLength < 1 {
		return "", "", errors.New("address for unix domain socket has wrong format")
	}

	unixSocketPath := addressParts[1]
	requestPath := ""
	if addressPartsLength == 3 {
		requestPath = addressParts[2]
	}
	return unixSocketPath, requestPath, nil
}

type userAgentRoundTripper struct {
	rt    http.RoundTripper
	agent string
}

func (rt *userAgentRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req = cloneRequest(req)
	req.Header.Set("User-Agent", rt.agent)
	return rt.rt.RoundTrip(req)
}

func cloneRequest(req *http.Request) *http.Request {
	r := new(http.Request)
	*r = *req // shallow clone

	// deep copy headers
	r.Header = make(http.Header, len(req.Header))
	for key, values := range req.Header {
		newValues := make([]string, len(values))
		copy(newValues, values)
		r.Header[key] = newValues
	}
	return r
}
