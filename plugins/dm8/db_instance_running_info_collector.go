package dm8

import (
	"context"
	"database/sql"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"time"
)

type DBInstanceRunningInfoCollector struct {
	db                  *sql.DB
	startTimeDesc       *prometheus.Desc
	statusDesc          *prometheus.Desc
	modeDesc            *prometheus.Desc
	trxNumDesc          *prometheus.Desc
	deadlockDesc        *prometheus.Desc
	threadNumDesc       *prometheus.Desc
	statusOccursDesc    *prometheus.Desc
	switchingOccursDesc *prometheus.Desc
	dbStartDayDesc      *prometheus.Desc
	config              *Config
}

const (
	DB_INSTANCE_STATUS_MOUNT_2   float64 = 2
	DB_INSTANCE_STATUS_SUSPEND_3 float64 = 3
	AlarmStatus_Normal                   = 1
	AlarmStatus_Unusual                  = 0
	AlarmSwitchOccur                     = "InitiateAnAlarm_SwitchOccur"
	AlarmSwitchStr                       = "switchingOccurStr"
)

func NewDBInstanceRunningInfoCollector(db *sql.DB, config *Config) MetricCollector {
	return &DBInstanceRunningInfoCollector{
		db:     db,
		config: config,
		startTimeDesc: prometheus.NewDesc(
			dmdbms_start_time_info,
			"Database status time",
			[]string{"host_name"}, // 添加标签
			nil,
		),
		statusDesc: prometheus.NewDesc(
			dmdbms_status_info,
			"Database status",
			[]string{"host_name"}, // 添加标签
			nil,
		),
		modeDesc: prometheus.NewDesc(
			dmdbms_mode_info,
			"Database mode",
			[]string{"host_name"}, // 添加标签
			nil,
		),
		trxNumDesc: prometheus.NewDesc(
			dmdbms_trx_info,
			"Number of transactions",
			[]string{"host_name"}, // 添加标签
			nil,
		),
		deadlockDesc: prometheus.NewDesc(
			dmdbms_dead_lock_num_info,
			"Number of deadlocks",
			[]string{"host_name"}, // 添加标签
			nil,
		),
		threadNumDesc: prometheus.NewDesc(
			dmdbms_thread_num_info,
			"Number of threads",
			[]string{"host_name"}, // 添加标签
			nil,
		),
		statusOccursDesc: prometheus.NewDesc( //这个是数据库状态切换的标识  OPEN
			dmdbms_db_status_occurs,
			"status changes status, error is 0 , true is 1",
			[]string{"host_name"}, // 添加标签
			nil,
		),
		switchingOccursDesc: prometheus.NewDesc( //这个是集群切换的标识
			dmdbms_switching_occurs,
			"Database instance switching occurs， error is 0 , true is 1  ",
			[]string{"host_name"}, // 添加标签
			nil,
		),
		dbStartDayDesc: prometheus.NewDesc( //这个是集群切换的标识
			dmdbms_start_day,
			"Database instance start_day ",
			[]string{"host_name"}, // 添加标签
			nil,
		),
	}
}

func (c *DBInstanceRunningInfoCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.startTimeDesc
	ch <- c.statusDesc
	ch <- c.modeDesc
	ch <- c.trxNumDesc
	ch <- c.deadlockDesc
	ch <- c.threadNumDesc
	ch <- c.statusOccursDesc
	ch <- c.switchingOccursDesc
	ch <- c.dbStartDayDesc
}

func (c *DBInstanceRunningInfoCollector) Collect(ch chan<- prometheus.Metric) {
	funcStart := time.Now()
	// 时间间隔的计算发生在 defer 语句执行时，确保能够获取到正确的函数执行时间。
	defer func() {
		duration := time.Since(funcStart)
		logger.Infof("func exec time：%vms", duration.Milliseconds())
	}()

	if err := checkDBConnection(c.db); err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.QueryTimeout)
	defer cancel()

	rows, err := c.db.QueryContext(ctx, QueryDBInstanceRunningInfoSqlStr)
	if err != nil {
		handleDbQueryError(err)
		return
	}
	defer rows.Close()

	var status, mode, trxNum, deadlockNum, threadNum float64
	var startTimeUnix, dbStartDay int64
	if rows.Next() {
		var startTimeStr, statusStr, modeStr, trxNumStr, deadlockNumStr, threadNumStr string
		if err := rows.Scan(&startTimeStr, &statusStr, &modeStr, &trxNumStr, &deadlockNumStr, &threadNumStr, &dbStartDay); err != nil {
			logger.Errorf("Error scanning row", err)
			return
		}
		status, _ = strconv.ParseFloat(statusStr, 64)
		mode, _ = strconv.ParseFloat(modeStr, 64)
		trxNum, _ = strconv.ParseFloat(trxNumStr, 64)
		deadlockNum, _ = strconv.ParseFloat(deadlockNumStr, 64)
		threadNum, _ = strconv.ParseFloat(threadNumStr, 64)

		// 解析时间戳字符串为 time.Time 类型
		startTime, err := time.Parse("2006-01-02 15:04:05", startTimeStr)
		if err != nil {
			logger.Errorf("Error parsing start time", err)
			// 如果转换失败则赋予默认时间值
			var defaultTime = time.Date(2006, time.January, 1, 0, 0, 0, 0, time.UTC)
			startTime = defaultTime
		}
		// 获取秒级 Unix 时间戳
		startTimeUnix = startTime.Unix()

	}

	var statusOccurs = 0
	//对值进行二次封装处理

	//判断实例状态是否正常，异常时值为0 正常时为1
	//eg： 此处是为了兼容java版本的报错
	if status == DB_INSTANCE_STATUS_MOUNT_2 || status == DB_INSTANCE_STATUS_SUSPEND_3 {
		statusOccurs = 0
	} else {
		statusOccurs = 1
	}

	data := map[string]float64{
		"startTime":    float64(startTimeUnix),
		"status":       status,
		"mode":         mode,
		"trxNum":       trxNum,
		"deadlockNum":  deadlockNum,
		"threadNum":    threadNum,
		"statusOccurs": float64(statusOccurs),
		"dbStartDay":   float64(dbStartDay),
	}
	// 处理数据库模式切换的逻辑（主备集群）
	c.handleDatabaseModeSwitch(ch, mode)

	//注册指标
	c.collectMetrics(ch, data)
	//	logger.Logger.Debugf("Collector DBInstanceRunningInfoCollector success,status: %v", status)

}

func (c *DBInstanceRunningInfoCollector) collectMetrics(ch chan<- prometheus.Metric, data map[string]float64) {
	ch <- prometheus.MustNewConstMetric(c.startTimeDesc, prometheus.GaugeValue, data["startTime"], Hostname)
	ch <- prometheus.MustNewConstMetric(c.statusDesc, prometheus.GaugeValue, data["status"], Hostname)
	ch <- prometheus.MustNewConstMetric(c.modeDesc, prometheus.GaugeValue, data["mode"], Hostname)
	ch <- prometheus.MustNewConstMetric(c.trxNumDesc, prometheus.GaugeValue, data["trxNum"], Hostname)
	ch <- prometheus.MustNewConstMetric(c.deadlockDesc, prometheus.GaugeValue, data["deadlockNum"], Hostname)
	ch <- prometheus.MustNewConstMetric(c.threadNumDesc, prometheus.GaugeValue, data["threadNum"], Hostname)
	ch <- prometheus.MustNewConstMetric(c.statusOccursDesc, prometheus.GaugeValue, data["status"], Hostname)
	ch <- prometheus.MustNewConstMetric(c.dbStartDayDesc, prometheus.GaugeValue, data["dbStartDay"], Hostname)
}

/*
*
Case 1 (switchOccurExists)：如果 AlarmSwitchOccur 缓存键存在，表示之前发生过切换，设置 switchingOccursDesc 为 AlarmStatus_Unusual。
Case 2 (modeExists && cachedMode == modeStr)：如果 AlarmSwitchStr 缓存键存在且模式没有变化，设置 switchingOccursDesc 为 AlarmStatus_Normal。
Case 3 (modeExists)：如果 AlarmSwitchStr 缓存键存在但模式发生变化，设置 switchingOccursDesc 为 AlarmStatus_Unusual，并更新缓存。
Default Case：如果 AlarmSwitchStr 缓存键不存在，设置 switchingOccursDesc 为 AlarmStatus_Normal 并更新缓存。
*/
func (c *DBInstanceRunningInfoCollector) handleDatabaseModeSwitch(ch chan<- prometheus.Metric, mode float64) {
	modeStr := strconv.FormatFloat(mode, 'f', -1, 64)

	cachedModeValue, modeExists := GetFromCache(AlarmSwitchStr) //这个key存储的是 mode值
	switchOccurExists := GetKeyExists(AlarmSwitchOccur)         //这个key表示已经发生切换了，保留的时间

	switch {
	case switchOccurExists:
		ch <- prometheus.MustNewConstMetric(c.switchingOccursDesc, prometheus.GaugeValue, AlarmStatus_Unusual, Hostname)
	case modeExists && cachedModeValue == modeStr:
		ch <- prometheus.MustNewConstMetric(c.switchingOccursDesc, prometheus.GaugeValue, AlarmStatus_Normal, Hostname)
	case modeExists:
		ch <- prometheus.MustNewConstMetric(c.switchingOccursDesc, prometheus.GaugeValue, AlarmStatus_Unusual, Hostname)
		DeleteFromCache(AlarmSwitchStr)
		SetCache(AlarmSwitchOccur, strconv.Itoa(AlarmStatus_Unusual), c.config.AlarmKeyCacheTime)
	default:
		SetCache(AlarmSwitchStr, modeStr, c.config.AlarmKeyCacheTime)
		ch <- prometheus.MustNewConstMetric(c.switchingOccursDesc, prometheus.GaugeValue, AlarmStatus_Normal, Hostname)
	}
}

/*
func (c *DBInstanceRunningInfoCollector) handleDatabaseModeSwitch(ch chan<- prometheus.Metric, mode float64) {
	//，'f'表示以小数形式输出，-1表示将所有小数位都输出，64表示mode的类型是float64。
	modeStr := strconv.FormatFloat(mode, 'f', -1, 64)

	if config.GetKeyExists(AlarmSwitchOccur) {
		// 如果key存在表名发生过切换
		ch <- prometheus.MustNewConstMetric(c.switchingOccursDesc, prometheus.GaugeValue, AlarmStatus_Unusual, dm8.Hostname)
	} else {
		// 判断是否发生切换
		if config.GetKeyExists(AlarmSwitchStr) {
			// 判断模式是否发生切换
			if cachedMode, found := config.GetFromCache(AlarmSwitchStr); found && cachedMode == modeStr {
				// 模式未发生变化
				ch <- prometheus.MustNewConstMetric(c.switchingOccursDesc, prometheus.GaugeValue, AlarmStatus_Normal, dm8.Hostname)
			} else {
				// 模式发生变化
				ch <- prometheus.MustNewConstMetric(c.switchingOccursDesc, prometheus.GaugeValue, AlarmStatus_Unusual, dm8.Hostname)
				config.DeleteFromCache(AlarmSwitchStr)
				config.SetCache(AlarmSwitchOccur, strconv.Itoa(AlarmStatus_Unusual), 30*time.Minute)
			}
		} else {
			// 第一次出现，更新缓存
			config.SetCache(AlarmSwitchStr, modeStr, 30*time.Minute)
			ch <- prometheus.MustNewConstMetric(c.switchingOccursDesc, prometheus.GaugeValue, AlarmStatus_Normal, dm8.Hostname)
		}
	}
}
*/
