package dm8

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

// 定义收集器结构体
type DbInstanceLogErrorCollector struct {
	db                   *sql.DB
	instanceLogErrorDesc *prometheus.Desc
	config               *Config
}

// 初始化收集器
func NewInstanceLogErrorCollector(db *sql.DB, config *Config) MetricCollector {
	return &DbInstanceLogErrorCollector{
		db:     db,
		config: config,
		instanceLogErrorDesc: prometheus.NewDesc(
			dmdbms_instance_log_error_info,
			"Information about DM database instance log errors",
			[]string{"host_name"},
			nil,
		),
	}
}

// Describe 方法
func (c *DbInstanceLogErrorCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.instanceLogErrorDesc
}

// Collect 方法
func (c *DbInstanceLogErrorCollector) Collect(ch chan<- prometheus.Metric) {
	funcStart := time.Now()
	defer func() {
		duration := time.Since(funcStart)
		logger.Infof("func exec time: %vms", duration.Milliseconds())
	}()

	if err := c.db.Ping(); err != nil {
		logger.Errorf("Database connection is not available", err)
		setMetric(ch, c.instanceLogErrorDesc, 0)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.QueryTimeout)
	defer cancel()

	// 获取数据库实例日志错误信息
	errorCount, err := getDbInstanceLogErrorInfo(ctx, c.db)
	if err != nil {
		logger.Errorf("exec getDbInstanceLogErrorInfo func error", err)
		//setMetric(ch, c.instanceLogErrorDesc, 0)
		return
	}

	setMetric(ch, c.instanceLogErrorDesc, errorCount)
}

// 辅助函数：设置指标
func setMetric(ch chan<- prometheus.Metric, desc *prometheus.Desc, value float64) {
	hostname := Hostname
	ch <- prometheus.MustNewConstMetric(
		desc,
		prometheus.GaugeValue,
		value,
		hostname,
	)
}

// 获取数据库实例日志错误信息
func getDbInstanceLogErrorInfo(ctx context.Context, db *sql.DB) (float64, error) {
	var errorCount float64

	query := `SELECT /*+DM_EXPORTER*/ count(*) error_info FROM V$instance_log_history WHERE level$ IN ('ERROR', 'FATAL')`
	row := db.QueryRowContext(ctx, query)
	err := row.Scan(&errorCount)
	if err != nil {
		return 0, fmt.Errorf("query error: %v", err)
	}

	return errorCount, nil
}
