package writer

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/cprobe/cprobe/lib/cgroup"
	"github.com/cprobe/cprobe/lib/clienttls"
	"github.com/cprobe/cprobe/lib/fileutil"
	"github.com/cprobe/cprobe/lib/httpproxy"
	"github.com/cprobe/cprobe/lib/listx"
	"github.com/cprobe/cprobe/lib/netutil"
	"github.com/cprobe/cprobe/lib/promrelabel"
	"github.com/cprobe/cprobe/lib/promutils"
	"github.com/pkg/errors"
)

var (
	writerDisable = flag.Bool("no-writer", false, "Disable remote writer")

	WriterConfig = &WriterYaml{}
)

type Writer struct {
	URL                  string                      `yaml:"url"`
	RetryTimes           int                         `yaml:"retry_times"`
	RetryIntervalMillis  int64                       `yaml:"retry_interval_millis"`
	BasicAuthUser        string                      `yaml:"basic_auth_user"`
	BasicAuthPass        string                      `yaml:"basic_auth_pass"`
	Headers              []string                    `yaml:"headers"`
	ConnectTimeoutMillis int64                       `yaml:"connect_timeout_millis"`
	RequestTimeoutMillis int64                       `yaml:"request_timeout_millis"`
	MaxIdleConnsPerHost  int                         `yaml:"max_idle_conns_per_host"`
	Concurrency          int                         `yaml:"concurrency"`
	ProxyURL             string                      `yaml:"proxy_url"`
	Interface            string                      `yaml:"interface"`
	FollowRedirects      bool                        `yaml:"follow_redirects"`
	ExtraLabels          *promutils.Labels           `yaml:"extra_labels"`
	RelabelConfigs       []promrelabel.RelabelConfig `yaml:"metric_relabel_configs"`
	ParsedRelabelConfigs *promrelabel.ParsedConfigs  `yaml:"-"`

	clienttls.ClientConfig `yaml:",inline"`
	Client                 *http.Client                   `yaml:"-"`
	RequestQueue           *listx.SafeList[*http.Request] `yaml:"-"`
}

func (w *Writer) Parse() error {
	if w.Concurrency <= 0 {
		w.Concurrency = cgroup.AvailableCPUs() * 2
	}

	if w.ConnectTimeoutMillis <= 0 {
		w.ConnectTimeoutMillis = 500
	}

	if w.RequestTimeoutMillis <= 0 {
		w.RequestTimeoutMillis = 5000
	}

	if w.MaxIdleConnsPerHost <= 0 {
		w.MaxIdleConnsPerHost = 2
	}

	// http client
	dialer := &net.Dialer{
		Timeout: time.Duration(w.ConnectTimeoutMillis) * time.Millisecond,
	}

	var err error
	if w.Interface != "" {
		dialer.LocalAddr, err = netutil.LocalAddressByInterfaceName(w.Interface)
		if err != nil {
			return err
		}
	}

	proxy, err := httpproxy.GetProxyFunc(w.ProxyURL)
	if err != nil {
		return err
	}

	trans := &http.Transport{
		Proxy:               proxy,
		DialContext:         dialer.DialContext,
		MaxIdleConnsPerHost: w.MaxIdleConnsPerHost,
	}

	if strings.HasPrefix(w.URL, "https") {
		tlsConfig, err := w.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}

		trans.TLSClientConfig = tlsConfig
	}

	w.Client = &http.Client{
		Transport: trans,
		Timeout:   time.Duration(w.RequestTimeoutMillis) * time.Millisecond,
	}

	if !w.FollowRedirects {
		w.Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	// relabel configs
	w.ParsedRelabelConfigs, err = promrelabel.ParseRelabelConfigs(w.RelabelConfigs)
	if err != nil {
		return err
	}

	// request queue
	w.RequestQueue = listx.NewSafeList[*http.Request]()

	if w.RetryTimes <= 0 {
		w.RetryTimes = 100
	}

	if w.RetryIntervalMillis <= 0 {
		w.RetryIntervalMillis = 3000
	}

	go w.StartSender()

	return nil
}

type Global struct {
	ExtraLabels          *promutils.Labels           `yaml:"extra_labels"`
	RelabelConfigs       []promrelabel.RelabelConfig `yaml:"metric_relabel_configs"`
	ParsedRelabelConfigs *promrelabel.ParsedConfigs  `yaml:"-"`
}

type WriterYaml struct {
	Global  *Global   `yaml:"global"`
	Writers []*Writer `yaml:"writers"`
}

func (wy *WriterYaml) Parse() (err error) {
	for i := range wy.Writers {
		if err = wy.Writers[i].Parse(); err != nil {
			return err
		}
	}

	wy.Global.ParsedRelabelConfigs, err = promrelabel.ParseRelabelConfigs(wy.Global.RelabelConfigs)
	if err != nil {
		return err
	}

	return nil
}

func Init(configDirectory string) error {
	if *writerDisable {
		return nil
	}

	writerFile := filepath.Join(configDirectory, "writer.yaml")

	if !fileutil.IsExist(writerFile) {
		return fmt.Errorf("writer.file %s does not exist", writerFile)
	}

	if !fileutil.IsFile(writerFile) {
		return fmt.Errorf("writer.file %s is not a file", writerFile)
	}

	err := fileutil.ReadYaml(writerFile, WriterConfig)
	if err != nil {
		return errors.Wrap(err, "cannot read writer config")
	}

	if err = WriterConfig.Parse(); err != nil {
		return errors.Wrap(err, "cannot set writer fields")
	}

	return nil
}
