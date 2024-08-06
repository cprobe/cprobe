package dm8

import (
	"context"
	"database/sql"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

// 定义数据结构
type IniParameterInfo struct {
	ParaName  sql.NullString
	ParaValue sql.NullFloat64
}

// 定义收集器结构体
type IniParameterCollector struct {
	db                *sql.DB
	parameterInfoDesc *prometheus.Desc
	config            *Config
}

func NewIniParameterCollector(db *sql.DB, config *Config) MetricCollector {
	return &IniParameterCollector{
		db:     db,
		config: config,
		parameterInfoDesc: prometheus.NewDesc(
			dmdbms_parameter_info,
			"Information about DM database parameters",
			[]string{"host_name", "param_name"},
			nil,
		),
	}

}

func (c *IniParameterCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.parameterInfoDesc
}

func (c *IniParameterCollector) Collect(ch chan<- prometheus.Metric) {
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

	rows, err := c.db.QueryContext(ctx, QueryParameterInfoSql)
	if err != nil {
		handleDbQueryError(err)
		return
	}
	defer rows.Close()

	var iniParameterInfos []IniParameterInfo
	for rows.Next() {
		var info IniParameterInfo
		if err := rows.Scan(&info.ParaName, &info.ParaValue); err != nil {
			logger.Errorf("Error scanning row", err)
			continue
		}
		iniParameterInfos = append(iniParameterInfos, info)
	}
	if err := rows.Err(); err != nil {
		logger.Errorf("Error with rows", err)
		return
	}

	// 发送数据到 Prometheus
	hostname := Hostname
	for _, info := range iniParameterInfos {
		paramName := NullStringToString(info.ParaName)
		ch <- prometheus.MustNewConstMetric(
			c.parameterInfoDesc,
			prometheus.GaugeValue,
			NullFloat64ToFloat64(info.ParaValue),
			hostname, paramName,
		)
	}
}
