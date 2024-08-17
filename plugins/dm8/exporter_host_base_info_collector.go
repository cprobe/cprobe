package dm8

import (
	"fmt"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/prometheus/client_golang/prometheus"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// SystemInfo 结构体，用于存储系统信息
type SystemInfo struct {
	OSVersion    string
	OSName       string
	Hostname     string
	CoreNum      string
	MemSize      string
	Architecture string
}

// 定义收集器结构体
type SystemInfoCollector struct {
	systemInfoDesc *prometheus.Desc
}

// 初始化收集器
func NewSystemInfoCollector() *SystemInfoCollector {
	return &SystemInfoCollector{
		systemInfoDesc: prometheus.NewDesc(
			dmdbms_node_uname_info,
			"System information",
			[]string{"host_name", "osName", "osVersion", "coreNum", "memSize", "architecture"},
			nil,
		),
	}
}

// Describe 方法
func (c *SystemInfoCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.systemInfoDesc
}

// Collect 方法
func (c *SystemInfoCollector) Collect(ch chan<- prometheus.Metric) {
	funcStart := time.Now()
	defer func() {
		duration := time.Since(funcStart)
		logger.Infof("host func exec time: %vms", duration.Milliseconds())
	}()

	systemInfo := getSystemInfo()

	ch <- prometheus.MustNewConstMetric(
		c.systemInfoDesc,
		prometheus.GaugeValue,
		1,
		systemInfo.Hostname,
		systemInfo.OSName,
		systemInfo.OSVersion,
		systemInfo.CoreNum,
		systemInfo.MemSize,
		systemInfo.Architecture,
	)
}

// 获取系统信息
func getSystemInfo() SystemInfo {
	var systemInfo SystemInfo

	// 获取主机名
	hostname, err := os.Hostname()
	if err != nil {
		logger.Errorf("Error getting hostname: %v\n", err)
		hostname = "unknown"
	}
	systemInfo.Hostname = hostname

	// 获取操作系统名称和版本
	osName, osVersion := getOSNameAndVersion()
	systemInfo.OSName = osName
	systemInfo.OSVersion = osVersion

	// 获取CPU核数
	coreNum := getCoreNum()
	systemInfo.CoreNum = coreNum

	// 获取内存大小
	memSize := getMemSize()
	systemInfo.MemSize = memSize

	// 获取系统架构
	architecture := getArchitecture()
	systemInfo.Architecture = architecture

	return systemInfo
}

// 获取操作系统名称和版本
func getOSNameAndVersion() (string, string) {
	osName := runtime.GOOS
	var osVersion string

	if osName == "windows" {
		/*cmd := exec.Command("cmd", "ver")
		output, err := cmd.Output()
		if err != nil {
			logger.Logger.Error("Error getting OS version: %v\n", err)
			return osName, "Windows"
		}
		if !utf8.ValidString(osVersion) {
			// 如果不是有效的UTF-8编码，可以逐个字节处理或者采取其他处理方法
			// 这里简单地去掉非ASCII字符
			var filteredBytes []byte
			for _, b := range output {
				if b < 128 {
					filteredBytes = append(filteredBytes, byte(b))
				}
			}
			osVersion = string(filteredBytes)
		}
		// 移除字符串中的换行符和空格等无关字符*/
		osVersion = strings.TrimSpace(osVersion)
		return osName, osVersion
	} else {
		cmd := exec.Command("uname", "-r")
		output, err := cmd.Output()
		if err != nil {
			logger.Errorf("Error getting OS version: %v\n", err)
			return osName, "unknown"
		}
		osVersion = strings.TrimSpace(string(output))
	}

	return osName, osVersion
}

// 获取CPU核数
func getCoreNum() string {
	return fmt.Sprintf("%d", runtime.NumCPU())
}

// 获取内存大小
func getMemSize() string {
	var memSize string
	if runtime.GOOS == "windows" {
		memSize = "0"
	} else {
		cmd := exec.Command("awk", "/MemTotal/ {print $2}", "/proc/meminfo")
		output, err := cmd.Output()
		if err != nil {
			logger.Errorf("Error getting memory size: %v\n", err)
			memSize = "unknown"
		} else {
			memSize = strings.TrimSpace(string(output))
		}
	}
	return memSize
}

// 获取系统架构
func getArchitecture() string {

	switch runtime.GOOS {
	case "linux":
		//	cmd := exec.Command("uname", "-m")
		output, err := exec.Command("arch").Output()
		if err == nil {
			return strings.TrimSpace(string(output))
		}
		/*	default:
			return "0"*/
	}
	return "0"
}
