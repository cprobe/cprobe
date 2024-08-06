package dm8

import (
	"context"
	"database/sql"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

// 定义数据结构
type RapplySysInfo struct {
	TaskMemUsed sql.NullFloat64
	TaskNum     sql.NullFloat64
}

// 定义收集器结构体
type DbRapplySysCollector struct {
	db              *sql.DB
	taskMemUsedDesc *prometheus.Desc
	taskNumDesc     *prometheus.Desc
	config          *Config
}

func NewDbRapplySysCollector(db *sql.DB, config *Config) MetricCollector {
	return &DbRapplySysCollector{
		db:     db,
		config: config,
		taskMemUsedDesc: prometheus.NewDesc(
			dmdbms_rapply_sys_task_mem_used,
			"Information about DM database apply system task memory used",
			[]string{"host_name"},
			nil,
		),
		taskNumDesc: prometheus.NewDesc(
			dmdbms_rapply_sys_task_num,
			"Information about DM database apply system task number",
			[]string{"host_name"},
			nil,
		),
	}
}

func (c *DbRapplySysCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.taskMemUsedDesc
	ch <- c.taskNumDesc
}

func (c *DbRapplySysCollector) Collect(ch chan<- prometheus.Metric) {
	funcStart := time.Now()
	defer func() {
		duration := time.Since(funcStart)
		logger.Infof("func exec time: %vms", duration.Milliseconds())
	}()

	if err := c.db.Ping(); err != nil {
		logger.Errorf("Database connection is not available", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.QueryTimeout)
	defer cancel()

	// 执行查询
	rows, err := c.db.QueryContext(ctx, QueryStandbyInfoSql)
	if err != nil {
		handleDbQueryError(err)
		return
	}
	defer rows.Close()

	var rapplySysInfos []RapplySysInfo
	for rows.Next() {
		var info RapplySysInfo
		if err := rows.Scan(&info.TaskMemUsed, &info.TaskNum); err != nil {
			logger.Errorf("Error scanning row", err)
			continue
		}
		rapplySysInfos = append(rapplySysInfos, info)
	}
	if err := rows.Err(); err != nil {
		logger.Errorf("Error with rows", err)
		return
	}

	hostname := Hostname
	for _, info := range rapplySysInfos {
		ch <- prometheus.MustNewConstMetric(
			c.taskMemUsedDesc,
			prometheus.GaugeValue,
			NullFloat64ToFloat64(info.TaskMemUsed),
			hostname,
		)
		ch <- prometheus.MustNewConstMetric(
			c.taskNumDesc,
			prometheus.GaugeValue,
			NullFloat64ToFloat64(info.TaskNum),
			hostname,
		)
	}
}
