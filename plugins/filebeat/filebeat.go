package filebeat

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/cprobe/cprobe/plugins"
	collector "github.com/cprobe/cprobe/plugins/filebeat/exporter"
	"github.com/cprobe/cprobe/types"
	"github.com/prometheus/client_golang/prometheus"
)

type Config struct {
	BeatSystem  bool          `toml:"beat_system"`
	BeatTimeout time.Duration `toml:"beat_timeout"`
}

type Filebeat struct {
}

func init() {
	plugins.RegisterPlugin(types.PluginFilebeat, &Filebeat{})
}

func (f *Filebeat) ParseConfig(baseDir string, bs []byte) (any, error) {
	var c Config
	err := toml.Unmarshal(bs, &c)
	if err != nil {
		return nil, err
	}

	if c.BeatTimeout == 0 {
		c.BeatTimeout = time.Second * 10
	}
	return &c, nil
}

func (f *Filebeat) Scrape(ctx context.Context, target string, cfg any, ss *types.Samples) error {
	conf := cfg.(*Config)

	if !strings.HasPrefix(target, "http") {
		target = "http://" + target
	}

	beatURL, err := url.Parse(target)

	if err != nil {
		return err
	}
	httpClient := &http.Client{
		Timeout: conf.BeatTimeout,
	}
	var beatInfo *collector.BeatInfo
	stopCh := make(chan bool)

	t := time.NewTicker(1 * time.Second)

beatdiscovery:
	for {
		select {
		case <-t.C:
			beatInfo, err = loadBeatType(httpClient, *beatURL)
			if err != nil {
				logger.Errorf("Could not load beat type, with error: %v, retrying in 1s", err)
				continue
			}

			break beatdiscovery

		case <-stopCh:
			os.Exit(0) // signal received, stop gracefully
		}
	}
	t.Stop()

	mainCollector := collector.NewMainCollector(httpClient, beatURL, "beat_exporter", beatInfo, conf.BeatSystem)
	err = prometheus.NewRegistry().Register(mainCollector)
	if err != nil {
		return err
	}

	ch := make(chan prometheus.Metric)
	go func() {
		mainCollector.Collect(ch)
		close(ch)
	}()

	for m := range ch {
		if err := ss.AddPromMetric(m); err != nil {
			logger.Warnf("failed to transform prometheus metric: %s", err)
		}
	}

	return nil
}

func loadBeatType(client *http.Client, url url.URL) (*collector.BeatInfo, error) {
	beatInfo := &collector.BeatInfo{}

	response, err := client.Get(url.String())
	if err != nil {
		return beatInfo, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		logger.Errorf("Beat URL: %q status code: %d", url.String(), response.StatusCode)
		return beatInfo, err
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		logger.Errorf("Can't read body of response")
		return beatInfo, err
	}

	err = json.Unmarshal(bodyBytes, &beatInfo)
	if err != nil {
		logger.Errorf("Could not parse JSON response for target")
		return beatInfo, err
	}

	logger.Infof("Target beat:%s configuration loaded successfully!", beatInfo.String())

	return beatInfo, nil
}
