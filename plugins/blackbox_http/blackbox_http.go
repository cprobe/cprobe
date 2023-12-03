package blackboxhttp

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/cprobe/cprobe/lib/clienttls"
	"github.com/cprobe/cprobe/lib/httpproxy"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/cprobe/cprobe/plugins"
	"github.com/cprobe/cprobe/types"
	"github.com/pkg/errors"
)

type Global struct {
	Namespace     string        `toml:"namespace"`
	Timeout       time.Duration `toml:"timeout"`
	Method        string        `toml:"method"`
	Payload       string        `toml:"payload"`
	BasicAuthUser string        `toml:"basic_auth_user"`
	BasicAuthPass string        `toml:"basic_auth_pass"`
	Headers       []string      `toml:"headers"`

	ProxyURL string `toml:"proxy_url"`
	clienttls.ClientConfig

	ExpectedStatusCodes        []int          `toml:"expected_status_codes"`
	ExpectedBodyRegexp         string         `toml:"expected_body_regexp"`
	ExpectedBodyRegexpCompiled *regexp.Regexp `toml:"-"`
}

type Config struct {
	BaseDir string `toml:"-"`
	Global  Global `toml:"global"`
}

type BlackboxHTTP struct {
	// 这个数据结构中未来如果有变量，千万要小心并发使用变量的问题
}

func init() {
	plugins.RegisterPlugin(types.PluginBlackboxHTTP, &BlackboxHTTP{})
}

func (p *BlackboxHTTP) ParseConfig(baseDir string, bs []byte) (any, error) {
	var c Config
	err := toml.Unmarshal(bs, &c)
	if err != nil {
		return nil, err
	}

	c.BaseDir = baseDir

	if len(c.Global.ExpectedStatusCodes) == 0 {
		c.Global.ExpectedStatusCodes = []int{200}
	}

	if len(c.Global.ExpectedBodyRegexp) > 0 {
		re, err := regexp.Compile(c.Global.ExpectedBodyRegexp)
		if err != nil {
			return nil, errors.WithMessagef(err, "invalid expected_body_regexp %q", c.Global.ExpectedBodyRegexp)
		}
		c.Global.ExpectedBodyRegexpCompiled = re
	}

	if c.Global.Namespace == "" {
		c.Global.Namespace = "blackbox_http"
	} else if c.Global.Namespace == "-" {
		c.Global.Namespace = ""
	}

	if c.Global.Timeout == 0 {
		c.Global.Timeout = 3 * time.Second
	}

	if c.Global.Method == "" {
		c.Global.Method = "GET"
	}

	return &c, nil
}

const probeErrorMetric = "probe_result_code"

func (p *BlackboxHTTP) Scrape(ctx context.Context, address string, c any, ss *types.Samples) error {
	cfg := c.(*Config)

	if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
		address = "http://" + address
	}

	method := cfg.Global.Method
	if method == "" {
		if cfg.Global.Payload == "" {
			method = "GET"
		} else {
			method = "POST"
		}
	}

	var body io.Reader
	if cfg.Global.Payload != "" {
		body = strings.NewReader(cfg.Global.Payload)
	}

	req, err := http.NewRequest(method, address, body)
	if err != nil {
		ss.AddMetric(cfg.Global.Namespace, map[string]interface{}{
			probeErrorMetric: 1.0, // new request failed
		}, map[string]string{
			"method": method,
		})
		return errors.WithMessagef(err, "error creating request for %q", address)
	}

	// set headers
	for _, h := range cfg.Global.Headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			ss.AddMetric(cfg.Global.Namespace, map[string]interface{}{
				probeErrorMetric: 2.0, // invalid header
			}, map[string]string{
				"method": method,
			})
			return errors.Errorf("invalid header %q", h)
		}

		headerKey := strings.TrimSpace(parts[0])
		headerValue := strings.TrimSpace(parts[1])

		req.Header.Set(headerKey, headerValue)

		if headerKey == "Host" {
			req.Host = headerValue
		}
	}

	// set basic auth
	if cfg.Global.BasicAuthUser != "" && cfg.Global.BasicAuthPass != "" {
		req.SetBasicAuth(cfg.Global.BasicAuthUser, cfg.Global.BasicAuthPass)
	}

	cli := &http.Client{
		Timeout: cfg.Global.Timeout,
	}

	if !strings.HasPrefix(address, "https://") && cfg.Global.ProxyURL == "" {
		// 不单独创建 transport 了，直接使用默认的 client
		return p.Do(cfg, cli, req, ss)
	}

	// create transport, 待会记得要关掉
	trans := &http.Transport{
		DisableKeepAlives: true,
	}

	if cfg.Global.ProxyURL != "" {
		proxy, err := httpproxy.GetProxyFunc(cfg.Global.ProxyURL)
		if err != nil {
			logger.Fatalf("error creating proxy func for %q: %v", cfg.Global.ProxyURL, err)
		}
		trans.Proxy = proxy
	}

	if strings.HasPrefix(address, "https://") {
		tlsConfig, err := cfg.Global.ClientConfig.TLSConfig()
		if err != nil {
			logger.Fatalf("error creating tls config: %v", err)
		}
		trans.TLSClientConfig = tlsConfig
	}

	defer trans.CloseIdleConnections()
	cli.Transport = trans
	return p.Do(cfg, cli, req, ss)
}

func (p *BlackboxHTTP) Do(cfg *Config, cli *http.Client, req *http.Request, ss *types.Samples) error {
	fields := make(map[string]interface{})
	tags := map[string]string{"method": req.Method}

	start := time.Now()
	resp, err := cli.Do(req)
	if err != nil {
		fields[probeErrorMetric] = 3.0 // request failed
		ss.AddMetric(cfg.Global.Namespace, fields, tags)
		return errors.WithMessagef(err, "failed to make request %q error: %v", req.URL.String(), err)
	}

	fields["probe_duration_seconds"] = time.Since(start).Seconds()
	fields["probe_response_code"] = float64(resp.StatusCode)

	if len(cfg.Global.ExpectedStatusCodes) > 0 {
		// check status code
		ok := false
		for _, code := range cfg.Global.ExpectedStatusCodes {
			if resp.StatusCode == code {
				ok = true
				break
			}
		}

		if !ok {
			fields[probeErrorMetric] = 4.0 // status code not match
			ss.AddMetric(cfg.Global.Namespace, fields, tags)
			return errors.Errorf("status code %d not match expected status codes %v", resp.StatusCode, cfg.Global.ExpectedStatusCodes)
		}
	}

	defer resp.Body.Close()

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		fields[probeErrorMetric] = 5.0 // read body failed
		ss.AddMetric(cfg.Global.Namespace, fields, tags)
		return errors.WithMessagef(err, "failed to read response body from %q error: %v", req.URL.String(), err)
	}

	if len(cfg.Global.ExpectedBodyRegexp) > 0 {
		// check body regex
		responseBody := string(bs)

		if !cfg.Global.ExpectedBodyRegexpCompiled.MatchString(responseBody) {
			fields[probeErrorMetric] = 6.0 // body not match
			ss.AddMetric(cfg.Global.Namespace, fields, tags)
			return errors.Errorf("response body %q not match expected body regexp %q", responseBody, cfg.Global.ExpectedBodyRegexp)
		}
	}

	fields[probeErrorMetric] = 0.0 // success

	// check tls cert
	if strings.HasPrefix(req.URL.String(), "https://") && resp.TLS != nil {
		fields["cert_expire_timestamp"] = getEarliestCertExpiry(resp.TLS).Unix()
	}

	ss.AddMetric(cfg.Global.Namespace, fields, tags)

	return nil
}

func getEarliestCertExpiry(state *tls.ConnectionState) time.Time {
	earliest := time.Time{}
	for _, cert := range state.PeerCertificates {
		if (earliest.IsZero() || cert.NotAfter.Before(earliest)) && !cert.NotAfter.IsZero() {
			earliest = cert.NotAfter
		}
	}
	return earliest
}
