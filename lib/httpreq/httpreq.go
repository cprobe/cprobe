package httpreq

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cprobe/cprobe/lib/httpproxy"
	"github.com/cprobe/cprobe/lib/netutil"
)

type RequestOptions struct {
	ConnectTimeoutMillis int64 `toml:"connect_timeout_millis"`
	RequestTimeoutMillis int64 `toml:"request_timeout_millis"`
	MaxIdleConnsPerHost  int   `toml:"max_idle_conns_per_host"`

	BasicAuthUser   string   `toml:"basic_auth_user"`
	BasicAuthPass   string   `toml:"basic_auth_pass"`
	BearerToken     string   `toml:"bearer_token"`
	BearerTokeFile  string   `toml:"bearer_token_file"`
	Headers         []string `toml:"headers"`
	ProxyURL        string   `toml:"proxy_url"`
	FollowRedirects bool     `toml:"follow_redirects"`
	Interface       string   `toml:"interface"`
}

func (o *RequestOptions) NewClient(tlsConfig *tls.Config, disableKeepAlives bool) (*http.Client, error) {
	connTimeout := time.Duration(o.ConnectTimeoutMillis) * time.Millisecond
	if connTimeout <= 0 {
		connTimeout = 500 * time.Millisecond
	}

	reqTimeout := time.Duration(o.RequestTimeoutMillis) * time.Millisecond
	if reqTimeout <= 0 {
		reqTimeout = 5000 * time.Millisecond
	}

	dialer := &net.Dialer{
		Timeout: connTimeout,
	}

	var err error
	if o.Interface != "" {
		dialer.LocalAddr, err = netutil.LocalAddressByInterfaceName(o.Interface)
		if err != nil {
			return nil, err
		}
	}

	proxy, err := httpproxy.GetProxyFunc(o.ProxyURL)
	if err != nil {
		return nil, err
	}

	trans := &http.Transport{
		Proxy:               proxy,
		DialContext:         dialer.DialContext,
		MaxIdleConnsPerHost: o.MaxIdleConnsPerHost,
		DisableKeepAlives:   disableKeepAlives,
	}

	if tlsConfig != nil {
		trans.TLSClientConfig = tlsConfig
	}

	cli := &http.Client{
		Transport: trans,
		Timeout:   reqTimeout,
	}

	if !o.FollowRedirects {
		cli.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return cli, nil
}

func (o *RequestOptions) FillHeaders(req *http.Request, baseDirs ...string) error {
	for _, h := range o.Headers {
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

	if o.BasicAuthUser != "" && o.BasicAuthPass != "" {
		req.SetBasicAuth(o.BasicAuthUser, o.BasicAuthPass)
	}

	if o.BearerToken == "" && o.BearerTokeFile != "" {
		tokenFilePath := o.BearerTokeFile
		if !filepath.IsAbs(tokenFilePath) {
			if len(baseDirs) > 0 {
				tokenFilePath = filepath.Join(baseDirs[0], tokenFilePath)
			} else {
				return fmt.Errorf("bearer token file path is not absolute, path: %s", tokenFilePath)
			}
		}
		bs, err := os.ReadFile(tokenFilePath)
		if err != nil {
			return fmt.Errorf("read bearer token file failed, path: %s", tokenFilePath)
		}
		o.BearerToken = strings.TrimSpace(string(bs))
	}

	if o.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+o.BearerToken)
	}

	return nil
}
