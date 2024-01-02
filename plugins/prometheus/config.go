package prometheus

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cprobe/cprobe/lib/clienttls"
	"github.com/cprobe/cprobe/lib/httpproxy"
	"github.com/cprobe/cprobe/lib/netutil"
	"github.com/cprobe/cprobe/types"
	"github.com/pkg/errors"
)

type Config struct {
	BaseDir              string   `toml:"-"`
	Namespace            string   `toml:"namespace"`
	BasicAuthUser        string   `toml:"basic_auth_user"`
	BasicAuthPass        string   `toml:"basic_auth_pass"`
	Headers              []string `toml:"headers"`
	ConnectTimeoutMillis int64    `toml:"connect_timeout_millis"`
	RequestTimeoutMillis int64    `toml:"request_timeout_millis"`
	MaxIdleConnsPerHost  int      `toml:"max_idle_conns_per_host"`
	ProxyURL             string   `toml:"proxy_url"`
	Interface            string   `toml:"interface"`
	FollowRedirects      bool     `toml:"follow_redirects"`
	Method               string   `toml:"method"`
	Payload              string   `toml:"payload"`
	BearerToken          string   `toml:"bearer_token"`
	BearerTokeFile       string   `toml:"bearer_token_file"`
	SplitBody            bool     `toml:"split_body"`

	clienttls.ClientConfig
}

func (cfg *Config) initDefault() {
	if cfg.ConnectTimeoutMillis <= 0 {
		cfg.ConnectTimeoutMillis = 500
	}

	if cfg.RequestTimeoutMillis <= 0 {
		cfg.RequestTimeoutMillis = 5000
	}

	cfg.MaxIdleConnsPerHost = 0

	if cfg.Method == "" {
		cfg.Method = "GET"
	}

	if cfg.Namespace == "" {
		cfg.Namespace = "prometheus"
	} else if cfg.Namespace == "-" {
		cfg.Namespace = ""
	}
}

func (cfg *Config) newClient(isHTTPs bool) (*http.Client, error) {
	dialer := &net.Dialer{
		Timeout: time.Duration(cfg.ConnectTimeoutMillis) * time.Millisecond,
	}

	var err error
	if cfg.Interface != "" {
		dialer.LocalAddr, err = netutil.LocalAddressByInterfaceName(cfg.Interface)
		if err != nil {
			return nil, err
		}
	}

	proxy, err := httpproxy.GetProxyFunc(cfg.ProxyURL)
	if err != nil {
		return nil, err
	}

	trans := &http.Transport{
		Proxy:               proxy,
		DialContext:         dialer.DialContext,
		MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		DisableKeepAlives:   true,
	}

	if isHTTPs {
		tlsConfig, err := cfg.ClientConfig.TLSConfig()
		if err != nil {
			return nil, err
		}
		trans.TLSClientConfig = tlsConfig
	}

	cli := &http.Client{
		Transport: trans,
		Timeout:   time.Duration(cfg.RequestTimeoutMillis) * time.Millisecond,
	}

	if !cfg.FollowRedirects {
		cli.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return cli, nil
}

func (cfg *Config) fillHeaders(req *http.Request) error {
	for _, h := range cfg.Headers {
		kv := strings.SplitN(h, ":", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		req.Header.Set(key, value)
		if key == "Host" {
			req.Host = value
		}
	}

	if cfg.BasicAuthUser != "" && cfg.BasicAuthPass != "" {
		req.SetBasicAuth(cfg.BasicAuthUser, cfg.BasicAuthPass)
	}

	if cfg.BearerToken == "" && cfg.BearerTokeFile != "" {
		tokenFilePath := cfg.BearerTokeFile
		if !filepath.IsAbs(tokenFilePath) {
			tokenFilePath = filepath.Join(cfg.BaseDir, tokenFilePath)
		}
		bs, err := os.ReadFile(tokenFilePath)
		if err != nil {
			return fmt.Errorf("read bearer token file failed, path: %s", tokenFilePath)
		}
		cfg.BearerToken = strings.TrimSpace(string(bs))
	}

	if cfg.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.BearerToken)
	}

	return nil
}

func (cfg *Config) Scrape(ctx context.Context, target string, ss *types.Samples) error {
	cfg.initDefault()

	cli, err := cfg.newClient(strings.HasPrefix(target, "https"))
	if err != nil {
		return errors.WithMessagef(err, "new client failed, target: %s", target)
	}

	var payload io.Reader
	if cfg.Payload != "" {
		payload = strings.NewReader(cfg.Payload)
	}

	req, err := http.NewRequest(cfg.Method, target, payload)
	if err != nil {
		return errors.WithMessagef(err, "new request failed, target: %s", target)
	}

	if err := cfg.fillHeaders(req); err != nil {
		return errors.WithMessagef(err, "fill headers failed, target: %s", target)
	}

	now := time.Now()
	resp, err := cli.Do(req.WithContext(ctx))
	if err != nil {
		return errors.WithMessagef(err, "request failed, target: %s", target)
	}
	ss.AddMetric(cfg.Namespace, map[string]interface{}{"last_scrape_duration_seconds": time.Since(now).Seconds()})

	if resp.Body == nil {
		return errors.WithMessagef(err, "response body is nil, target: %s", target)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.WithMessagef(err, "read response body failed, target: %s", target)
	}

	if err := ss.AddMetricsBody(body, resp.Header, cfg.SplitBody); err != nil {
		return errors.WithMessagef(err, "parse response failed, target: %s", target)
	}

	return nil
}
