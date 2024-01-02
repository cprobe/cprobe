package tomcat

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/cprobe/cprobe/lib/clienttls"
	"github.com/cprobe/cprobe/lib/httpreq"
	"github.com/cprobe/cprobe/types"
	"github.com/pkg/errors"
)

type Config struct {
	BaseDir string `toml:"-"`
	Suffix  string `toml:"suffix"`

	httpreq.RequestOptions
	clienttls.ClientConfig
}

func (c *Config) Scrape(ctx context.Context, target string, ss *types.Samples) error {
	if c.Suffix == "" {
		c.Suffix = "/manager/status/all?JSON=true"
	}

	target = strings.TrimSuffix(target, "/") + c.Suffix
	if !strings.HasPrefix(target, "http") {
		target = "http://" + target
	}

	var tlsConfig *tls.Config
	var err error
	if strings.HasPrefix(target, "https://") {
		tlsConfig, err = c.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}
	}

	cli, err := c.RequestOptions.NewClient(tlsConfig, true)
	if err != nil {
		return errors.WithMessagef(err, "new client failed, target: %s", target)
	}

	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		return errors.WithMessagef(err, "new request failed, target: %s", target)
	}

	err = c.RequestOptions.FillHeaders(req)
	if err != nil {
		return errors.WithMessagef(err, "fill headers failed, target: %s", target)
	}

	res, err := cli.Do(req.WithContext(ctx))
	if err != nil {
		return errors.WithMessagef(err, "do request failed, target: %s", target)
	}

	if res.Body == nil {
		return errors.Errorf("response body is nil, target: %s", target)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		bs, err := io.ReadAll(res.Body)
		if err != nil {
			return errors.WithMessagef(err, "read response body failed, target: %s", target)
		}
		return errors.Errorf("response status code is not 200, target: %s, code: %d, response body: %s", target, res.StatusCode, string(bs))
	}

	var resStruct ResponseStruct
	if err := json.NewDecoder(res.Body).Decode(&resStruct); err != nil {
		return errors.WithMessagef(err, "decode json failed, target: %s", target)
	}

	fields := map[string]interface{}{
		"jvm_memory_free":  resStruct.Tomcat.TomcatJvm.JvmMemory.Free,
		"jvm_memory_total": resStruct.Tomcat.TomcatJvm.JvmMemory.Total,
		"jvm_memory_max":   resStruct.Tomcat.TomcatJvm.JvmMemory.Max,
	}

	ss.AddMetric(types.PluginTomcat, fields)

	// add tomcat_jvm_memorypool measurements
	for _, mp := range resStruct.Tomcat.TomcatJvm.JvmMemoryPools {
		tcmpTags := map[string]string{
			"name": mp.Name,
			"type": mp.Type,
		}

		tcmpFields := map[string]interface{}{
			"jvm_memorypool_init":      mp.UsageInit,
			"jvm_memorypool_committed": mp.UsageCommitted,
			"jvm_memorypool_max":       mp.UsageMax,
			"jvm_memorypool_used":      mp.UsageUsed,
		}

		ss.AddMetric(types.PluginTomcat, tcmpFields, tcmpTags)
	}

	// add tomcat_connector measurements
	for _, c := range resStruct.Tomcat.TomcatConnectors {
		name, err := strconv.Unquote(c.Name)
		if err != nil {
			name = c.Name
		}

		tccTags := map[string]string{
			"name": name,
		}

		tccFields := map[string]interface{}{
			"connector_max_threads":          c.ThreadInfo.MaxThreads,
			"connector_current_thread_count": c.ThreadInfo.CurrentThreadCount,
			"connector_current_threads_busy": c.ThreadInfo.CurrentThreadsBusy,
			"connector_max_time":             c.RequestInfo.MaxTime,
			"connector_processing_time":      c.RequestInfo.ProcessingTime,
			"connector_request_count":        c.RequestInfo.RequestCount,
			"connector_error_count":          c.RequestInfo.ErrorCount,
			"connector_bytes_received":       c.RequestInfo.BytesReceived,
			"connector_bytes_sent":           c.RequestInfo.BytesSent,
		}

		ss.AddMetric(types.PluginTomcat, tccFields, tccTags)
	}

	return nil
}
