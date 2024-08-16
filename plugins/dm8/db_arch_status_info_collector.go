package dm8

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

const (
	DB_ARCH_NO_ENABLE = -1
	DB_ARCH_VALID     = 1
	DB_ARCH_INVALID   = 2
)

type DbArchStatusCollector struct {
	db             *sql.DB
	archStatusDesc *prometheus.Desc
	config         *Config
}

func NewDbArchStatusCollector(db *sql.DB, config *Config) MetricCollector {
	return &DbArchStatusCollector{
		db:     db,
		config: config,
		archStatusDesc: prometheus.NewDesc(
			dmdbms_arch_status,
			"Information about DM database archive status",
			[]string{"host_name"},
			nil,
		),
	}
}

func (c *DbArchStatusCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.archStatusDesc
}

func (c *DbArchStatusCollector) Collect(ch chan<- prometheus.Metric) {
	funcStart := time.Now()
	// 时间间隔的计算发生在 defer 语句执行时，确保能够获取到正确的函数执行时间。
	defer func() {
		duration := time.Since(funcStart)
		logger.Infof("func exec time：%vms", duration.Milliseconds())
	}()

	if err := c.db.Ping(); err != nil {
		logger.Errorf("Database connection is not available: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.QueryTimeout)
	defer cancel()

	// 获取数据库归档状态信息
	dbArchStatus, err := getDbArchStatus(ctx, c.db)
	if err != nil {
		logger.Errorf("exec getDbArchStatus func error", err)
		setArchMetric(ch, c.archStatusDesc, DB_ARCH_INVALID)
		return
	}
	setArchMetric(ch, c.archStatusDesc, dbArchStatus)
}

// 辅助函数：设置指标
func setArchMetric(ch chan<- prometheus.Metric, desc *prometheus.Desc, value int) {
	hostname := Hostname
	ch <- prometheus.MustNewConstMetric(
		desc,
		prometheus.GaugeValue,
		float64(value),
		hostname,
	)
}

// 获取数据库归档状态信息
func getDbArchStatus(ctx context.Context, db *sql.DB) (int, error) {
	var dbArchStatus string

	// 查询 PARA_VALUE
	query := `select /*+DMDB_CHECK_FLAG*/ PARA_VALUE from v$dm_ini where para_name='ARCH_INI'`
	row := db.QueryRowContext(ctx, query)
	err := row.Scan(&dbArchStatus)
	if err != nil {
		return DB_ARCH_INVALID, fmt.Errorf("query error: %v", err)
	}

	// 处理 PARA_VALUE 为 '1' 的情况
	if dbArchStatus == "1" {
		query = `select /*+DMDB_CHECK_FLAG*/ case arch_status when 'VALID' then 1 when 'INVALID' then 0 end ARCH_STATUS from v$arch_status where arch_type='LOCAL'`
		row = db.QueryRowContext(ctx, query)
		err = row.Scan(&dbArchStatus)
		if err != nil {
			return DB_ARCH_INVALID, fmt.Errorf("query error: %v", err)
		}
		if dbArchStatus == "1" {
			return DB_ARCH_VALID, nil
		} else if dbArchStatus == "0" {
			return DB_ARCH_INVALID, nil
		}
	} else if dbArchStatus == "0" {
		return DB_ARCH_NO_ENABLE, nil
	}

	logger.Infof("Check Database Arch Status Info Success")
	return DB_ARCH_INVALID, nil
}
