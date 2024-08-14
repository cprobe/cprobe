package dm8

import (
	"context"
	"database/sql"
	"errors"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/prometheus/client_golang/prometheus"
	"strings"
	"sync"
)

var (
	registerMux sync.Mutex
	//timeout     = 5 * time.Second
)

const (
	dmdbms_node_uname_info            string = "dmdbms_node_uname_info"
	dmdbms_tablespace_file_total_info string = "dmdbms_tablespace_file_total_info"
	dmdbms_tablespace_file_free_info  string = "dmdbms_tablespace_file_free_info"
	dmdbms_tablespace_size_total_info string = "dmdbms_tablespace_size_total_info"
	dmdbms_tablespace_size_free_info  string = "dmdbms_tablespace_size_free_info"
	dmdbms_start_time_info            string = "dmdbms_start_time_info"
	dmdbms_status_info                string = "dmdbms_status_info"
	dmdbms_mode_info                  string = "dmdbms_mode_info"
	dmdbms_trx_info                   string = "dmdbms_trx_info"
	dmdbms_dead_lock_num_info         string = "dmdbms_dead_lock_num_info"
	dmdbms_thread_num_info            string = "dmdbms_thread_num_info"
	dmdbms_switching_occurs           string = "dmdbms_switching_occurs"
	dmdbms_db_status_occurs           string = "dmdbms_db_status_occurs"

	dmdbms_memory_curr_pool_info  string = "dmdbms_memory_curr_pool_info"
	dmdbms_memory_total_pool_info string = "dmdbms_memory_total_pool_info"

	dmdbms_session_percentage string = "dmdbms_session_percentage"
	dmdbms_session_type_Info  string = "dmdbms_session_type_info"
	dmdbms_ckpttime_info      string = "dmdbms_ckpttime_info"

	dmdbms_joblog_error_num string = "dmdbms_joblog_error_num"

	dmdbms_slow_sql_info            string = "dmdbms_slow_sql_info"
	dmdbms_monitor_info             string = "dmdbms_monitor_info"
	dmdbms_statement_type_info      string = "dmdbms_statement_type_info"
	dmdbms_parameter_info           string = "dmdbms_parameter_info"
	dmdbms_user_list_info           string = "dmdbms_user_list_info"
	dmdbms_license_date             string = "dmdbms_license_date"
	dmdbms_version                  string = "dmdbms_version"
	dmdbms_arch_status              string = "dmdbms_arch_status"
	dmdbms_start_day                string = "dmdbms_start_day"
	dmdbms_rapply_sys_task_mem_used string = "dmdbms_rapply_sys_task_mem_used"
	dmdbms_rapply_sys_task_num      string = "dmdbms_rapply_sys_task_num"
	dmdbms_instance_log_error_info  string = "dmdbms_instance_log_error_info"

	dmdbms_dmap_process_is_exit      string = "dmdbms_dmap_process_is_exit"
	dmdbms_dmserver_process_is_exit  string = "dmdbms_dmserver_process_is_exit"
	dmdbms_dmwatcher_process_is_exit string = "dmdbms_dmwatcher_process_is_exit"
	dmdbms_dmmonitor_process_is_exit string = "dmdbms_dmmonitor_process_is_exit"
	dmdbms_dmagent_process_is_exit   string = "dmdbms_dmagent_process_is_exit"
)

// MetricCollector 接口
type MetricCollector interface {
	Describe(ch chan<- *prometheus.Desc)
	Collect(ch chan<- prometheus.Metric)
}

func RegisterCollectors(config *Config) *prometheus.Registry {
	registerMux.Lock()
	defer registerMux.Unlock()
	reg := prometheus.NewRegistry()
	logger.Infof("exporter running system is %v", GetOS())

	collectors := make([]prometheus.Collector, 0)
	collectors = append(collectors, NewSystemInfoCollector())

	if config.RegisterHostMetrics && strings.Compare(GetOS(), OS_LINUX) == 0 {
		collectors = append(collectors, NewDmapProcessCollector(DBPool, config))
	}
	if config.RegisterDatabaseMetrics {
		//collectors = append(collectors, NewDBSessionsCollector(dm8.DBPool))
		collectors = append(collectors, NewTableSpaceDateFileInfoCollector(DBPool, config))
		collectors = append(collectors, NewTableSpaceInfoCollector(DBPool, config))
		collectors = append(collectors, NewDBInstanceRunningInfoCollector(DBPool, config))
		collectors = append(collectors, NewDbMemoryPoolInfoCollector(DBPool, config))
		collectors = append(collectors, NewDBSessionsStatusCollector(DBPool, config))
		collectors = append(collectors, NewDbJobRunningInfoCollector(DBPool, config))
		collectors = append(collectors, NewSlowSessionInfoCollector(DBPool, config))
		collectors = append(collectors, NewMonitorInfoCollector(DBPool, config))
		collectors = append(collectors, NewDbSqlExecTypeCollector(DBPool, config))
		collectors = append(collectors, NewIniParameterCollector(DBPool, config))
		collectors = append(collectors, NewDbUserCollector(DBPool, config))
		collectors = append(collectors, NewDbLicenseCollector(DBPool, config))
		collectors = append(collectors, NewDbVersionCollector(DBPool, config))
		collectors = append(collectors, NewDbArchStatusCollector(DBPool, config))
		collectors = append(collectors, NewDbRapplySysCollector(DBPool, config))
		collectors = append(collectors, NewInstanceLogErrorCollector(DBPool, config))
		collectors = append(collectors, NewCkptCollector(DBPool, config))

	}
	if config.RegisterDmhsMetrics {
		// Add all middleware collectors here
		// collectors = append(collectors, NewMiddlewareCollector())
	}

	for _, collector := range collectors {
		reg.MustRegister(collector)
	}
	return reg
}

func checkDBConnection(db *sql.DB) error {
	if err := db.Ping(); err != nil {
		logger.Errorf("Database connection is not available", err)
		return err
	}
	return nil
}

func handleDbQueryError(err error) {
	if errors.Is(err, context.DeadlineExceeded) {
		logger.Errorf("Query timed out %v", err)
	} else {
		logger.Errorf("Error querying database %v", err)
	}
}
