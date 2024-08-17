package dm8

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/prometheus/client_golang/prometheus"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// DBInstanceInfo 结构体，用于存储 SQL 查询结果
type DBInstanceInfo struct {
	DBInstancePath string
	PID            string
}

// 定义收集器结构体
type DmapProcessCollector struct {
	db                   *sql.DB
	dmapProcessDesc      *prometheus.Desc
	dmserverProcessDesc  *prometheus.Desc
	dmwatcherProcessDesc *prometheus.Desc
	dmmonitorProcessDesc *prometheus.Desc
	dmagentProcessDesc   *prometheus.Desc
	localInstallBinPath  string
	lastPID              string
	//mutex                sync.Mutex
	config *Config
}

// 初始化收集器
func NewDmapProcessCollector(db *sql.DB, config *Config) *DmapProcessCollector {
	return &DmapProcessCollector{
		db:     db,
		config: config,
		dmapProcessDesc: prometheus.NewDesc(
			dmdbms_dmap_process_is_exit,
			"Information about DM database dmap process existence",
			[]string{"host_name"},
			nil,
		),
		dmserverProcessDesc: prometheus.NewDesc(
			dmdbms_dmserver_process_is_exit,
			"Information about DM database dmserver process existence",
			[]string{"host_name"},
			nil,
		),
		dmwatcherProcessDesc: prometheus.NewDesc(
			dmdbms_dmwatcher_process_is_exit,
			"Information about DM database dmwatcher process existence",
			[]string{"host_name"},
			nil,
		),
		dmmonitorProcessDesc: prometheus.NewDesc(
			dmdbms_dmmonitor_process_is_exit,
			"Information about DM database dmmonitor process existence",
			[]string{"host_name"},
			nil,
		),
		dmagentProcessDesc: prometheus.NewDesc(
			dmdbms_dmagent_process_is_exit,
			"Information about DM database dmagent process existence",
			[]string{"host_name"},
			nil,
		),
	}
}

// Describe 方法
func (c *DmapProcessCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.dmapProcessDesc
	ch <- c.dmserverProcessDesc
	ch <- c.dmwatcherProcessDesc
	ch <- c.dmmonitorProcessDesc
	ch <- c.dmagentProcessDesc
}

// Collect 方法
func (c *DmapProcessCollector) Collect(ch chan<- prometheus.Metric) {
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

	// 获取数据库实例信息
	dbInstanceInfo, err := c.getDbInstanceInfo()
	if err != nil {
		logger.Errorf("Error getting DB instance info: %v\n", err)
		return
	}

	// 如果 PID 发生变化，则更新 localInstallBinPath
	//c.mutex.Lock()
	if c.lastPID != dbInstanceInfo.PID {
		c.localInstallBinPath, err = getLocalInstallBinPath(dbInstanceInfo.PID)
		if err != nil {
			logger.Errorf("Error getting db install bin path: %v", err)
			c.localInstallBinPath = ""
			return
		}
		c.lastPID = dbInstanceInfo.PID
	}
	//c.mutex.Unlock()

	// 检查各个进程
	hostname, _ := os.Hostname()
	ch <- prometheus.MustNewConstMetric(
		c.dmapProcessDesc,
		prometheus.GaugeValue,
		checkProcess(c.localInstallBinPath, dbInstanceInfo.PID, "dmap"),
		hostname,
	)
	ch <- prometheus.MustNewConstMetric(
		c.dmserverProcessDesc,
		prometheus.GaugeValue,
		checkProcess(c.localInstallBinPath, dbInstanceInfo.PID, "dmserver"),
		hostname,
	)
	ch <- prometheus.MustNewConstMetric(
		c.dmwatcherProcessDesc,
		prometheus.GaugeValue,
		checkProcess(c.localInstallBinPath, dbInstanceInfo.PID, "dmwatcher"),
		hostname,
	)
	ch <- prometheus.MustNewConstMetric(
		c.dmmonitorProcessDesc,
		prometheus.GaugeValue,
		checkProcess(c.localInstallBinPath, dbInstanceInfo.PID, "dmmonitor"),
		hostname,
	)
	ch <- prometheus.MustNewConstMetric(
		c.dmagentProcessDesc,
		prometheus.GaugeValue,
		checkProcess(c.localInstallBinPath, dbInstanceInfo.PID, "dmagent"),
		hostname,
	)

}

// 检查进程
func checkProcess(installBinPath, pid, processName string) float64 {
	if installBinPath == "" {
		return 0
	}

	if installBinPath[len(installBinPath)-1:] == "/" || installBinPath[len(installBinPath)-1:] == "\\" {
		installBinPath = installBinPath[:len(installBinPath)-1]
	}

	var shellStr string
	if processName == "dmap" {
		shellStr = fmt.Sprintf("ps -ef | grep %s/dmap | grep -v grep | wc -l", installBinPath)
	} else if processName == "dmserver" {
		shellStr = fmt.Sprintf("ps -ef | grep %s | grep dm.ini | grep -v grep | wc -l", pid)
	} else {
		shellStr = fmt.Sprintf("ps -ef | grep %s/%s | grep -v grep | wc -l", installBinPath, processName)
	}

	cmd := exec.Command("sh", "-c", shellStr)
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error checking %s process: %v\n", processName, err)
		return 0
	}

	processCount := strings.TrimSpace(string(output))
	count, err := strconv.ParseFloat(processCount, 64)
	if err != nil {
		logger.Errorf("Error parsing %s process count: %v\n", processName, err)
		return 0
	}

	return count
}

// 获取数据库实例信息
func (c *DmapProcessCollector) getDbInstanceInfo() (DBInstanceInfo, error) {
	var info DBInstanceInfo

	ctx, cancel := context.WithTimeout(context.Background(), c.config.QueryTimeout)
	defer cancel()

	query := `
		SELECT /*+DMDB_CHECK_FLAG*/ PARA_VALUE AS DB_INSTANCE_PATH, (SELECT PID from V$PROCESS) PID
		FROM V$DM_INI
		WHERE PARA_NAME = 'CONFIG_PATH'
	`
	row := c.db.QueryRowContext(ctx, query)
	err := row.Scan(&info.DBInstancePath, &info.PID)
	if err != nil {
		return info, err
	}
	logger.Infof("DBInstanceInfo: %v\n", info)
	return info, nil
}

// 获取 localInstallBinPath
func getLocalInstallBinPath(pid string) (string, error) {
	cmd := exec.Command("ls", "-l", fmt.Sprintf("/proc/%s/cwd", pid))
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	logger.Infof("exec %v: %v", fmt.Sprintf("/proc/%s/cwd", pid), string(output))
	procStr := strings.TrimSpace(string(output))
	lineList := strings.Fields(procStr)
	lastElement := lineList[len(lineList)-1]
	logger.Infof("lastElement %v", lastElement)

	if strings.Contains(lastElement, "bin") {
		return lastElement, nil
	}

	shellStr := fmt.Sprintf("ls -l %s/dmserver | wc -l", lastElement)
	output, err = exec.Command("sh", "-c", shellStr).Output()
	logger.Infof("exec %v: %v", shellStr, string(output))
	if err != nil {
		return "", err
	}
	serverCount := strings.TrimSpace(string(output))
	if serverCount == "1" {
		return lastElement, nil
	}
	return "", fmt.Errorf("failed to get localInstallBinPath")
}
