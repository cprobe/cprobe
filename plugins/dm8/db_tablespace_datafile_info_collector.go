package dm8

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type TableSpaceDateFileInfoCollector struct {
	db        *sql.DB
	totalDesc *prometheus.Desc
	freeDesc  *prometheus.Desc
	config    *Config
}

type TableSpaceDateFileInfo struct {
	Path       string
	TotalSize  float64
	FreeSize   float64
	AutoExtend string
	NextSize   string
	MaxSize    string
}

func NewTableSpaceDateFileInfoCollector(db *sql.DB, config *Config) MetricCollector {
	return &TableSpaceDateFileInfoCollector{
		db:     db,
		config: config,
		totalDesc: prometheus.NewDesc(
			dmdbms_tablespace_file_total_info,
			"Tablespace file information",
			[]string{"host_name", "tablespace_name", "auto_extend", "next_size", "max_size"}, // 添加标签
			nil,
		),
		freeDesc: prometheus.NewDesc(
			dmdbms_tablespace_file_free_info,
			"Tablespace file information",
			[]string{"host_name", "tablespace_name", "auto_extend", "next_size", "max_size"}, // 添加标签
			nil,
		),
	}
}

func (c *TableSpaceDateFileInfoCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.totalDesc
	ch <- c.freeDesc
}

func (c *TableSpaceDateFileInfoCollector) Collect(ch chan<- prometheus.Metric) {
	funcStart := time.Now()
	// 时间间隔的计算发生在 defer 语句执行时，确保能够获取到正确的函数执行时间。
	defer func() {
		duration := time.Since(funcStart)
		logger.Infof("func exec time：%vms", duration.Milliseconds())
	}()

	//保存全局结果对象，可以用来做缓存以及序列化
	var tablespaceInfos []TableSpaceDateFileInfo

	// 从缓存中获取数据
	if cachedJSON, found := GetFromCache(dmdbms_tablespace_file_total_info); found {
		// 将缓存中的 JSON 字符串转换为 TablespaceInfo 切片
		if err := json.Unmarshal([]byte(cachedJSON), &tablespaceInfos); err != nil {
			// 处理反序列化错误
			logger.Errorf("Error unmarshaling cached data", err)
			// 反序列化失败，忽略缓存中的数据，继续查询数据库
			cachedJSON = "" // 清空缓存数据，确保后续不使用过期或损坏的数据
		} else {
			logger.Infof("Use cache TablespaceDateFile data")
			// 使用缓存的数据
			for _, info := range tablespaceInfos {
				ch <- prometheus.MustNewConstMetric(c.totalDesc, prometheus.GaugeValue, info.TotalSize, Hostname, info.Path, info.AutoExtend, info.NextSize, info.MaxSize)
				ch <- prometheus.MustNewConstMetric(c.freeDesc, prometheus.GaugeValue, info.FreeSize, Hostname, info.Path, info.AutoExtend, info.NextSize, info.MaxSize)
			}
			return
		}
	}

	if err := c.db.Ping(); err != nil {
		logger.Errorf("Database connection is not available: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.QueryTimeout)
	defer cancel()

	rows, err := c.db.QueryContext(ctx, QueryTablespaceFileSqlStr)
	if err != nil {
		handleDbQueryError(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var info TableSpaceDateFileInfo
		if err := rows.Scan(&info.Path, &info.TotalSize, &info.FreeSize, &info.AutoExtend, &info.NextSize, &info.MaxSize); err != nil {
			logger.Errorf("Error scanning row", err)
			continue
		}
		tablespaceInfos = append(tablespaceInfos, info)
	}
	if err := rows.Err(); err != nil {
		logger.Errorf("Error with rows", err)
	}
	// 发送数据到 Prometheus
	for _, info := range tablespaceInfos {
		ch <- prometheus.MustNewConstMetric(c.totalDesc, prometheus.GaugeValue, info.TotalSize, Hostname, info.Path, info.AutoExtend, info.NextSize, info.MaxSize)
		ch <- prometheus.MustNewConstMetric(c.freeDesc, prometheus.GaugeValue, info.FreeSize, Hostname, info.Path, info.AutoExtend, info.NextSize, info.MaxSize)
	}

	// 将 TablespaceInfo 切片序列化为 JSON 字符串
	valueJSON, err := json.Marshal(tablespaceInfos)
	if err != nil {
		// 处理序列化错误
		logger.Errorf("TablespaceInfo ", err)
		return
	}
	// 将查询结果存入缓存
	SetCache(dmdbms_tablespace_file_total_info, string(valueJSON), c.config.AlarmKeyCacheTime) // 设置缓存有效时间为5分钟
	logger.Infof("TablespaceFileInfoCollector exec finish")

}
