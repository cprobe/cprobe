package dm8

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"time"
)

// 定义数据结构
type DbLicenseInfo struct {
	ExpiredDate sql.NullString
}

// 定义收集器结构体
type DbLicenseCollector struct {
	db              *sql.DB
	licenseDateDesc *prometheus.Desc
	config          *Config
}

func NewDbLicenseCollector(db *sql.DB, config *Config) MetricCollector {
	return &DbLicenseCollector{
		db:     db,
		config: config,
		licenseDateDesc: prometheus.NewDesc(
			dmdbms_license_date,
			"Information about DM database license expiration date",
			[]string{"host_name", "date_day_str"},
			nil,
		),
	}
}

func (c *DbLicenseCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.licenseDateDesc
}

func (c *DbLicenseCollector) Collect(ch chan<- prometheus.Metric) {
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

	rows, err := c.db.QueryContext(ctx, QueryDbGrantInfoSql)
	if err != nil {
		handleDbQueryError(err)
		return
	}
	defer rows.Close()

	var licenseInfos []DbLicenseInfo
	for rows.Next() {
		var info DbLicenseInfo
		if err := rows.Scan(&info.ExpiredDate); err != nil {
			logger.Errorf("Error scanning row", err)
			continue
		}
		licenseInfos = append(licenseInfos, info)
	}
	if err := rows.Err(); err != nil {
		logger.Errorf("Error with rows", err)
		return
	}

	hostname := Hostname
	for _, info := range licenseInfos {
		expiredDateStr := NullStringToString(info.ExpiredDate)
		var returnDateStr string
		var licenseStatus string
		if expiredDateStr != "" {
			expiredDate, err := time.Parse("20060102", expiredDateStr)
			if err != nil {
				logger.Errorf("Error parsing date", err)
				continue
			}
			betweenDay := expiredDate.Sub(time.Now()).Hours() / 24
			returnDateStr = fmt.Sprintf("%.0f", betweenDay)
			licenseStatus = returnDateStr
			logger.Infof("Check Database License Date Info Success, betweenDay is %s day", returnDateStr)
		} else {
			licenseStatus = "无限制"
			returnDateStr = "-1"
			logger.Infof("Check Database License Date Info Success, Expired Unlimited")
		}

		ch <- prometheus.MustNewConstMetric(
			c.licenseDateDesc,
			prometheus.GaugeValue,
			parseToFloat64(returnDateStr),
			hostname, licenseStatus,
		)
	}

}

// 辅助函数，将 string 转换为 float64
func parseToFloat64(s string) float64 {
	if s == "" {
		return 0
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return val
}
