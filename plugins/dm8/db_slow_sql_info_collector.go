package dm8

import (
	"context"
	"database/sql"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type SessionInfoCollector struct {
	db              *sql.DB
	slowSQLInfoDesc *prometheus.Desc
	config          *Config
}

// 定义数据结构
type SessionInfo struct {
	ExecTime     sql.NullFloat64
	SlowSQL      sql.NullString
	SessID       sql.NullString
	CurrSch      sql.NullString
	ThrdID       sql.NullString
	LastRecvTime sql.NullTime
	ConnIP       sql.NullString
}

func NewSlowSessionInfoCollector(db *sql.DB, config *Config) MetricCollector {
	return &SessionInfoCollector{
		db:     db,
		config: config,
		slowSQLInfoDesc: prometheus.NewDesc(
			dmdbms_slow_sql_info,
			"Information about slow SQL statements",
			[]string{"host_name", "sess_id", "curr_sch", "thrd_id", "last_recv_time", "conn_ip", "slow_sql"},
			nil,
		),
	}
}

func (c *SessionInfoCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.slowSQLInfoDesc
}

func (c *SessionInfoCollector) Collect(ch chan<- prometheus.Metric) {
	if !c.config.CheckSlowSql {
		logger.Infof("CheckSlowSQL is false, skip collecting slow SQL info")
		return
	}
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

	rows, err := c.db.QueryContext(ctx, QueryDbSlowSqlInfoSqlStr, c.config.SlowSqlTime, c.config.SlowSqlMaxRows)
	if err != nil {
		handleDbQueryError(err)
		return
	}
	defer rows.Close()

	var sessionInfos []SessionInfo
	for rows.Next() {
		var info SessionInfo
		if err := rows.Scan(&info.ExecTime, &info.SlowSQL, &info.SessID, &info.CurrSch, &info.ThrdID, &info.LastRecvTime, &info.ConnIP); err != nil {
			logger.Errorf("Error scanning row", err)
			continue
		}
		sessionInfos = append(sessionInfos, info)
	}

	if err := rows.Err(); err != nil {
		logger.Errorf("Error with rows", err)
	}
	// 发送数据到 Prometheus
	for _, info := range sessionInfos {
		hostName := Hostname
		sessionID := NullStringToString(info.SessID)
		currentSchema := NullStringToString(info.CurrSch)
		threadID := NullStringToString(info.ThrdID)
		lastRecvTime := NullTimeToString(info.LastRecvTime)
		connIP := NullStringToString(info.ConnIP)
		slowSQL := NullStringToString(info.SlowSQL)

		ch <- prometheus.MustNewConstMetric(
			c.slowSQLInfoDesc,
			prometheus.GaugeValue,
			NullFloat64ToFloat64(info.ExecTime),
			hostName, sessionID, currentSchema, threadID, lastRecvTime, connIP, slowSQL,
		)
	}
}
