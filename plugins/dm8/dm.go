package dm8

import (
	"context"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/cprobe/cprobe/plugins"
	"github.com/cprobe/cprobe/types"
	"github.com/pkg/errors"
	"os"
	"time"
)

type Config struct {
	QueryTimeout            time.Duration `toml:"query_timeout"`
	MaxIdleConns            int           `toml:"max_idle_conns"`
	MaxOpenConns            int           `toml:"max_open_conns"`
	ConnMaxLifetime         time.Duration `toml:"conn_max_lifetime"`
	DbUser                  string        `toml:"db_user"`
	DbPwd                   string        `toml:"db_pwd"`
	BigKeyDataCacheTime     time.Duration `toml:"big_key_data_cache_time"`
	AlarmKeyCacheTime       time.Duration `toml:"alarm_key_cache_time"`
	RegisterHostMetrics     bool          `toml:"register_host_metrics"`
	RegisterDatabaseMetrics bool          `toml:"register_database_metrics"`
	RegisterDmhsMetrics     bool          `toml:"register_dmhs_metrics"`
	CheckSlowSql            bool          `toml:"check_slow_sql"`
	SlowSqlTime             int           `toml:"slow_sql_time"`
	SlowSqlMaxRows          int           `toml:"slow_sql_max_rows"`
}
type Dm struct {
}

var Hostname string

func init() {
	plugins.RegisterPlugin(types.PluginDm, &Dm{})
}

func (d *Dm) ParseConfig(baseDir string, bs []byte) (any, error) {
	var c Config
	err := toml.Unmarshal(bs, &c)
	if err != nil {
		return nil, err
	}

	if c.QueryTimeout == 0 {
		c.QueryTimeout = time.Millisecond * 500
	}
	return &c, nil
}

func (d *Dm) Scrape(ctx context.Context, target string, cfg any, ss *types.Samples) error {
	config := cfg.(*Config)
	// DSN (Data Source Name) format: user/password@host:port/service_name
	dsn := buildDSN(config.DbUser, config.DbPwd, target)
	hn, err := os.Hostname()
	if err != nil {
		logger.Errorf("Failed to get Hostname: %s", err)
		return err
	}
	Hostname = hn

	// 初始化数据库连接池
	err = InitDBPool(dsn, config)
	if err != nil {
		logger.Errorf("Failed to initialize database pool: %v", err)
		return err
	}
	defer CloseDBPool()

	registry := RegisterCollectors(config)
	mfs, err := registry.Gather()
	if err != nil {
		return errors.WithMessage(err, "failed to gather metrics from mongodb registry")
	}

	ss.AddMetricFamilies(mfs)

	return nil
}

func buildDSN(user, password, host string) string {
	//dsn := "dm://SYSDBA:SYSDBA@120.53.103.235:5236?autoCommit=true"
	return fmt.Sprintf("dm://%s:%s@%s?autoCommit=true", user, password, host)
}
