package redis

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/cprobe/cprobe/plugins"
	"github.com/cprobe/cprobe/plugins/redis/exporter"
	"github.com/cprobe/cprobe/types"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type Global struct {
	User              string        `toml:"user"`
	Password          string        `toml:"password"`
	Namespace         string        `toml:"namespace"`
	ConnectionTimeout time.Duration `toml:"connection_timeout"`
	PingOnConnect     bool          `toml:"ping_on_connect"`
	ConfigCommand     string        `toml:"config_command"`
	SetClientName     *bool         `toml:"set_client_name"`

	IsTile38  bool `toml:"is_tile38"`
	IsCluster bool `toml:"is_cluster"`

	SkipTLSVerification bool   `toml:"skip_tls_verification"`
	ClientCertFile      string `toml:"client_cert_file"`
	ClientKeyFile       string `toml:"client_key_file"`
	CaCertFile          string `toml:"ca_cert_file"`

	IncludeSystemMetrics *bool `toml:"include_system_metrics"`
	IncludeConfigMetrics *bool `toml:"include_config_metrics"`
	RedactConfigMetrics  *bool `toml:"redact_config_metrics"`

	CheckKeys                 []string `toml:"check_keys"`
	CheckSingleKeys           []string `toml:"check_single_keys"`
	CheckKeysBatchSize        int64    `toml:"check_keys_batch_size"`
	CheckKeyGroups            []string `toml:"check_key_groups"`
	MaxDistinctKeyGroups      int64    `toml:"max_distinct_key_groups"`
	CheckStreams              []string `toml:"check_streams"`
	CheckSingleStreams        []string `toml:"check_single_streams"`
	CountKeys                 []string `toml:"count_keys"`
	LuaScriptFiles            []string `toml:"lua_script_files"`
	DisableExportingKeyValues bool     `toml:"disable_exporting_key_values"`

	ExportClientList         bool `toml:"export_client_list"`
	ExportClientsIncludePort bool `toml:"export_clients_include_port"`
}

type Config struct {
	BaseDir string `toml:"-"`
	Global  Global `toml:"global"`
}

type Redis struct {
	// 这个数据结构中未来如果有变量，千万要小心并发使用变量的问题
}

func init() {
	plugins.RegisterPlugin(types.PluginRedis, &Redis{})
}

func (*Redis) ParseConfig(baseDir string, bs []byte) (any, error) {
	var c Config
	err := toml.Unmarshal(bs, &c)
	if err != nil {
		return nil, err
	}

	c.BaseDir = baseDir

	if c.Global.Namespace == "" {
		c.Global.Namespace = "redis"
	} else if c.Global.Namespace == "-" {
		c.Global.Namespace = ""
	}

	if c.Global.ConnectionTimeout == 0 {
		c.Global.ConnectionTimeout = time.Second
	}

	if c.Global.ConfigCommand == "" {
		c.Global.ConfigCommand = "CONFIG"
	}

	if c.Global.CheckKeysBatchSize == 0 {
		c.Global.CheckKeysBatchSize = 1000
	}

	if c.Global.MaxDistinctKeyGroups == 0 {
		c.Global.MaxDistinctKeyGroups = 100
	}

	if c.Global.SetClientName == nil {
		b := true
		c.Global.SetClientName = &b
	}

	if c.Global.RedactConfigMetrics == nil {
		b := true
		c.Global.RedactConfigMetrics = &b
	}

	if c.Global.IncludeConfigMetrics == nil {
		b := true
		c.Global.IncludeConfigMetrics = &b
	}

	if c.Global.IncludeSystemMetrics == nil {
		b := true
		c.Global.IncludeSystemMetrics = &b
	}

	return &c, nil
}

func (*Redis) Scrape(ctx context.Context, target string, c any, ss *types.Samples) error {
	// 这个方法中如果要对配置 c 变量做修改，一定要 clone 一份之后再修改，因为并发的多个 target 共享了一个 c 变量
	cfg := c.(*Config)
	if !strings.Contains(target, "://") {
		target = "redis://" + target
	}

	u, err := url.Parse(target)
	if err != nil {
		return errors.WithMessagef(err, "failed to parse target %s", target)
	}

	u.User = nil
	target = u.String()

	ls := make(map[string][]byte)
	for _, script := range cfg.Global.LuaScriptFiles {
		if ls[script], err = os.ReadFile(script); err != nil {
			return fmt.Errorf("failed to read redis lua script file %s: %s", script, err)
		}
	}

	conf := cfg.Global
	opts := exporter.Options{
		User:                      conf.User,
		Password:                  conf.Password,
		Namespace:                 conf.Namespace,
		ConnectionTimeout:         conf.ConnectionTimeout,
		PingOnConnect:             conf.PingOnConnect,
		ConfigCommandName:         conf.ConfigCommand,
		SetClientName:             *conf.SetClientName,
		IsTile38:                  conf.IsTile38,
		IsCluster:                 conf.IsCluster,
		SkipTLSVerification:       conf.SkipTLSVerification,
		ClientCertFile:            conf.ClientCertFile,
		ClientKeyFile:             conf.ClientKeyFile,
		CaCertFile:                conf.CaCertFile,
		InclSystemMetrics:         *conf.IncludeSystemMetrics,
		InclConfigMetrics:         *conf.IncludeConfigMetrics,
		RedactConfigMetrics:       *conf.RedactConfigMetrics,
		CheckKeys:                 conf.CheckKeys,
		CheckSingleKeys:           conf.CheckSingleKeys,
		CheckKeysBatchSize:        conf.CheckKeysBatchSize,
		CheckKeyGroups:            conf.CheckKeyGroups,
		MaxDistinctKeyGroups:      conf.MaxDistinctKeyGroups,
		CheckStreams:              conf.CheckStreams,
		CheckSingleStreams:        conf.CheckSingleStreams,
		CountKeys:                 conf.CountKeys,
		LuaScript:                 ls,
		DisableExportingKeyValues: conf.DisableExportingKeyValues,
		ExportClientList:          conf.ExportClientList,
		ExportClientsInclPort:     conf.ExportClientsIncludePort,
	}

	exp, err := exporter.NewRedisExporter(target, opts)
	if err != nil {
		return errors.Wrap(err, "failed to create redis exporter")
	}

	if (conf.ClientCertFile != "") != (conf.ClientKeyFile != "") {
		return fmt.Errorf("client_cert_file and client_key_file must be specified together")
	}

	ch := make(chan prometheus.Metric)
	go func() {
		exp.Collect(ch)
		close(ch)
	}()

	for m := range ch {
		if err := ss.AddPromMetric(m); err != nil {
			logger.Warnf("failed to transform prometheus metric: %s", err)
		}
	}

	return nil
}
