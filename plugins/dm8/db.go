package dm8

import (
	"github.com/cprobe/cprobe/lib/logger"

	"database/sql"
	"fmt"
	_ "gitee.com/chunanyong/dm"
)

var (
	DBPool *sql.DB
	dsn    string
)

// InitDBPool initializes the database connection pool
func InitDBPool(dsnStr string, config *Config) error {
	var err error
	dsn = dsnStr
	//DBPool, err = sql.Open("godror", dsn)
	//"dm://SYSDBA:SYSDBA@localhost:5236?autoCommit=true"
	DBPool, err = sql.Open("dm", dsnStr)
	if err != nil {
		logger.Errorf("failed to open database: %v", err)
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Set the maximum number of open connections
	DBPool.SetMaxOpenConns(config.MaxOpenConns)
	// Set the maximum number of idle connections
	DBPool.SetMaxIdleConns(config.MaxIdleConns)
	// Set the maximum lifetime of each connection
	DBPool.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Test the database connection
	err = DBPool.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	return nil
}

// CloseDBPool closes the database connection pool
func CloseDBPool() {
	if DBPool != nil {
		DBPool.Close()
	}
}
