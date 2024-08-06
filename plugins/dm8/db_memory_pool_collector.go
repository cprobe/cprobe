package dm8

import (
	"context"
	"database/sql"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type DbMemoryPoolInfoCollector struct {
	db            *sql.DB
	totalPoolDesc *prometheus.Desc
	currPoolDesc  *prometheus.Desc
	config        *Config
}

type MemoryPoolInfo struct {
	ZoneType sql.NullString
	CurrVal  sql.NullFloat64
	ResVal   sql.NullFloat64
	TotalVal sql.NullFloat64
}

func NewDbMemoryPoolInfoCollector(db *sql.DB, config *Config) MetricCollector {
	return &DbMemoryPoolInfoCollector{
		db:     db,
		config: config,
		totalPoolDesc: prometheus.NewDesc(
			dmdbms_memory_total_pool_info,
			"mem total pool info information",
			[]string{"host_name", "pool_type"}, // 添加标签
			nil,
		),
		currPoolDesc: prometheus.NewDesc(
			dmdbms_memory_curr_pool_info,
			"mem curr pool info information",
			[]string{"host_name", "pool_type"}, // 添加标签
			nil,
		),
	}
}

func (c *DbMemoryPoolInfoCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.totalPoolDesc
	ch <- c.currPoolDesc
}

func (c *DbMemoryPoolInfoCollector) Collect(ch chan<- prometheus.Metric) {
	funcStart := time.Now()
	// 时间间隔的计算发生在 defer 语句执行时，确保能够获取到正确的函数执行时间。
	defer func() {
		duration := time.Since(funcStart)
		logger.Infof("func exec time：%vms", duration.Milliseconds())
	}()

	//保存全局结果对象
	var memoryPoolInfos []MemoryPoolInfo

	if err := c.db.Ping(); err != nil {
		logger.Errorf("Database connection is not available: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.QueryTimeout)
	defer cancel()

	rows, err := c.db.QueryContext(ctx, QueryMemoryPoolInfoSqlStr)
	if err != nil {
		handleDbQueryError(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var info MemoryPoolInfo
		if err := rows.Scan(&info.ZoneType, &info.CurrVal, &info.ResVal, &info.TotalVal); err != nil {
			logger.Errorf("Error scanning row", err)
			continue
		}
		memoryPoolInfos = append(memoryPoolInfos, info)
	}
	if err := rows.Err(); err != nil {
		logger.Errorf("Error with rows", err)
	}
	// 发送数据到 Prometheus
	for _, info := range memoryPoolInfos {
		ch <- prometheus.MustNewConstMetric(c.totalPoolDesc, prometheus.GaugeValue, NullFloat64ToFloat64(info.TotalVal), Hostname, NullStringToString(info.ZoneType))
		ch <- prometheus.MustNewConstMetric(c.currPoolDesc, prometheus.GaugeValue, NullFloat64ToFloat64(info.CurrVal), Hostname, NullStringToString(info.ZoneType))
	}

	//	logger.Logger.Infof("MemoryPoolInfo exec finish")

}
