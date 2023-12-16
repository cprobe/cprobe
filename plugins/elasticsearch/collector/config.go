package collector

import "time"

type Config struct {
	BaseDir               string        `toml:"-"`
	Username              string        `toml:"username"`
	Password              string        `toml:"password"`
	ApiKey                string        `toml:"apikey"`
	Timeout               time.Duration `toml:"timeout"`
	GatherNode            string        `toml:"gather_node"`
	GatherClusterInfo     bool          `toml:"gather_cluster_info"`
	GatherClusterSettings bool          `toml:"gather_cluster_settings"`
	GatherSnapshots       bool          `toml:"gather_snapshots"`
	GatherIndices         bool          `toml:"gather_indices"`
	GatherIndicesShards   bool          `toml:"gather_indices_shards"`
	GatherIndicesSettings bool          `toml:"gather_indices_settings"`
	GatherIndicesMappings bool          `toml:"gather_indices_mappings"`
	GatherIndicesUseAlias bool          `toml:"gather_indices_use_alias"`
	GatherIlm             bool          `toml:"gather_ilm"`
	GatherSlm             bool          `toml:"gather_slm"`
	GatherDataStream      bool          `toml:"gather_data_stream"`
	TLSCa                 string        `toml:"tls_ca"`
	TLSClientPrivateKey   string        `toml:"tls_client_private_key"`
	TLSClientCert         string        `toml:"tls_client_cert"`
	TLSSkipVerify         bool          `toml:"tls_skip_verify"`
	AwsRegion             string        `toml:"aws_region"`
	AwsRoleArn            string        `toml:"aws_role_arn"`
}
