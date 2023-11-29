package kafka

import (
	"context"
	"github.com/BurntSushi/toml"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/cprobe/cprobe/plugins"
	"github.com/cprobe/cprobe/plugins/kafka/exporter"
	"github.com/cprobe/cprobe/types"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"strings"
)

type Global struct {
	KafkaServer  string `toml:"kafka_server" description:"Addresses of Kafka Server"`
	KafkaVersion string `toml:"kafka_version" description:"Kafka broker version"`
	KafkaLabels  string `toml:"kafka_labels" description:"Kafka cluster name"`
	Namespace    string `toml:"namespace"`

	SaslEnabled            *bool  `toml:"sasl_enabled" description:"Connect using SASL/PLAIN"`
	SASLHandshake          *bool  `toml:"sasl_handshake" description:"Only set this to false if using a non-Kafka SASL proxy"`
	SaslUsername           string `toml:"sasl_username" description:"SASL user name"`
	SaslPassword           string `toml:"sasl_password" description:"SASL user password"`
	SaslMechanism          string `toml:"sasl_mechanism" description:"SASL mechanism can be plain, scram-sha512, scram-sha256"`
	SaslServiceName        string `toml:"sasl_service_name" description:"Service name when using Kerberos Auth"`
	SaslKerberosConfigPath string `toml:"sasl_kerberos_config_path" description:"Kerberos config path"`
	SaslRealm              string `toml:"sasl_realm" description:"Kerberos realm"`
	SaslKeytabPath         string `toml:"sasl_keytab_path" description:"Kerberos keytab file path"`
	SaslKerberosAuthType   string `toml:"sasl_kerberos_auth_type" description:"Kerberos auth type. Either 'keytabAuth' or 'userAuth'"`

	TLSEnabled               *bool  `toml:"tls_enabled" description:"Connect to Kafka using TLS"`
	TLSServerName            string `toml:"tls_server_name" description:"Used to verify the hostname on the returned certificates unless tls.insecure-skip-tls-verify is given. The kafka server's name should be given"`
	TLSCAFile                string `toml:"tls_ca_file" description:"The optional certificate authority file for Kafka TLS client authentication"`
	TLSCertFile              string `toml:"tls_cert_file" description:"The optional certificate file for Kafka client authentication"`
	TLSKeyFile               string `toml:"tls_key_file" description:"The optional key file for Kafka client authentication"`
	TLSInsecureSkipTLSVerify *bool  `toml:"tls_insecure_skip_tls_verify" description:"If true, the server's certificate will not be checked for validity"`

	TopicFilter  string `toml:"topic_filter" description:"Regex that determines which topics to collect"`
	TopicExclude string `toml:"topic_exclude" description:"Regex that determines which topics to exclude"`
	GroupFilter  string `toml:"group_filter" description:"Regex that determines which consumer groups to collect"`
	GroupExclude string `toml:"group_exclude" description:"Regex that determines which consumer groups to exclude"`

	UseConsumeLagZookeeper *bool  `toml:"use_consume_lag_zookeeper" description:"if you need to use a group from zookeeper"`
	ZookeeperServer        string `toml:"zookeeper_server" description:"Address (hosts) of zookeeper server"`

	RefreshMetadata  string `toml:"refresh_metadata" description:"Metadata refresh interval"`
	OffsetShowAll    *bool  `toml:"offset_show_all" description:"Whether show the offset/lag for all consumer group, otherwise, only show connected consumer groups"`
	ConcurrentEnable *bool  `toml:"concurrent_enable" description:"If true, all scrapes will trigger kafka operations otherwise, they will share results. WARN: This should be disabled on large clusters"`
	TopicWorks       int    `toml:"topic_works" description:"Number of topic workers"`
	Verbosity        int    `toml:"verbosity" description:"Verbosity log level"`
	LogSarama        *bool  `toml:"log_sarama"`
}

type Config struct {
	Global Global `toml:"global"`
}

type Kafka struct {
	// 这个数据结构中未来如果有变量，千万要小心并发使用变量的问题
}

func init() {
	plugins.RegisterPlugin(types.PluginKafka, &Kafka{})
}

func (*Kafka) ParseConfig(bs []byte) (any, error) {
	var c Config
	err := toml.Unmarshal(bs, &c)
	if err != nil {
		return nil, err
	}

	if c.Global.Namespace == "" {
		c.Global.Namespace = "kafka"
	}

	if c.Global.RefreshMetadata == "" {
		c.Global.RefreshMetadata = "30s"
	}

	if c.Global.OffsetShowAll == nil {
		b := true
		c.Global.OffsetShowAll = &b
	}

	if c.Global.ConcurrentEnable == nil {
		b := false
		c.Global.ConcurrentEnable = &b
	}

	if c.Global.TopicWorks == 0 {
		c.Global.TopicWorks = 100
	}

	if c.Global.Verbosity == 0 {
		c.Global.Verbosity = 0
	}

	if c.Global.SaslEnabled == nil {
		b := false
		c.Global.SaslEnabled = &b
	}

	if c.Global.SASLHandshake == nil {
		b := false
		c.Global.SASLHandshake = &b
	}

	if c.Global.TLSEnabled == nil {
		b := false
		c.Global.TLSEnabled = &b
	}

	if c.Global.TLSInsecureSkipTLSVerify == nil {
		b := false
		c.Global.TLSInsecureSkipTLSVerify = &b
	}

	if c.Global.UseConsumeLagZookeeper == nil {
		b := false
		c.Global.UseConsumeLagZookeeper = &b
	}

	return &c, nil
}

func (*Kafka) Scrape(ctx context.Context, target string, c any, ss *types.Samples) error {
	// 这个方法中如果要对配置 c 变量做修改，一定要 clone 一份之后再修改，因为并发的多个 target 共享了一个 c 变量
	cfg := c.(*Config)
	//if !strings.Contains(target, "://") {
	//	target = "kafka://" + target
	//}

	//u, err := url.Parse(target)
	//if err != nil {
	//	return errors.WithMessagef(err, "failed to parse target %s", target)
	//}

	var targets []string
	targets = append(targets, target)

	conf := cfg.Global
	opts := exporter.KafkaOpts{
		Uri:                      targets,
		UseSASL:                  *conf.SaslEnabled,
		UseSASLHandshake:         *conf.SASLHandshake,
		SaslUsername:             conf.SaslUsername,
		SaslPassword:             conf.SaslPassword,
		SaslMechanism:            conf.SaslMechanism,
		SaslDisablePAFXFast:      false,
		UseTLS:                   *conf.TLSEnabled,
		TlsServerName:            conf.TLSServerName,
		TlsCAFile:                "",
		TlsCertFile:              conf.TLSCertFile,
		TlsKeyFile:               conf.TLSKeyFile,
		TlsInsecureSkipTLSVerify: *conf.TLSInsecureSkipTLSVerify,
		KafkaVersion:             conf.KafkaVersion,
		UseZooKeeperLag:          *conf.UseConsumeLagZookeeper,
		UriZookeeper:             nil,
		Labels:                   conf.KafkaLabels,
		MetadataRefreshInterval:  conf.RefreshMetadata,
		ServiceName:              "",
		KerberosConfigPath:       "",
		Realm:                    conf.SaslRealm,
		KeyTabPath:               "",
		KerberosAuthType:         "",
		OffsetShowAll:            *conf.OffsetShowAll,
		TopicWorkers:             conf.TopicWorks,
		AllowConcurrent:          *conf.ConcurrentEnable,
		AllowAutoTopicCreation:   false,
		VerbosityLogLevel:        conf.Verbosity,
	}

	labels := make(map[string]string)
	if opts.Labels != "" {
		for _, label := range strings.Split(opts.Labels, ",") {
			splitted := strings.Split(label, "=")
			if len(splitted) >= 2 {
				labels[splitted[0]] = splitted[1]
			}
		}
	}

	exp, err := exporter.Setup(conf.TopicFilter, conf.TopicExclude, conf.GroupFilter, conf.GroupExclude, opts, labels)

	////exp, err := exporter.NewRedisExporter(target, opts)
	if err != nil {
		return errors.Wrap(err, "failed to create kafka exporter")
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
