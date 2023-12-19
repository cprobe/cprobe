package oracledb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cprobe/cprobe/plugins/sqlc"
	"github.com/cprobe/cprobe/types"
	go_ora "github.com/sijms/go-ora/v2"
)

type Global struct {
	Username  string            `toml:"username"`
	Password  string            `toml:"password"`
	Options   map[string]string `toml:"options"`
	Namespace string            `toml:"namespace"`
}

type Config struct {
	BaseDir string             `toml:"base_dir"`
	Global  *Global            `toml:"global"`
	Queries []sqlc.CustomQuery `toml:"queries"`
}

// target: ip:port/service
func (c *Config) Scrape(ctx context.Context, target string, ss *types.Samples) error {
	ip, port, service, err := explode(target)
	if err != nil {
		return fmt.Errorf("invalid target: %s", target)
	}

	// default settings
	if c.Global.Namespace == "" {
		c.Global.Namespace = "oracledb"
	} else if c.Global.Namespace == "-" {
		c.Global.Namespace = ""
	}

	connString := go_ora.BuildUrl(ip, port, service, c.Global.Username, c.Global.Password, c.Global.Options)
	conn, err := sql.Open("oracle", connString)
	if err != nil {
		return fmt.Errorf("cannot opening connection to database: %s, error: %s", target, err)
	}

	if conn == nil {
		return fmt.Errorf("cannot opening connection to database: %s", target)
	}

	defer conn.Close()

	conn.SetMaxOpenConns(16)
	conn.SetMaxIdleConns(1)
	conn.SetConnMaxLifetime(time.Minute)

	if err := conn.PingContext(ctx); err != nil {
		return fmt.Errorf("cannot ping database: %s, error: %s", target, err)
	}

	if c.Global.Namespace != "" {
		for i := 0; i < len(c.Queries); i++ {
			c.Queries[i].Mesurement = c.Global.Namespace + "_" + c.Queries[i].Mesurement
		}
	}

	sqlc.CollectCustomQueries(ctx, conn, ss, c.Queries)
	return nil
}

var ErrInvalidTarget = errors.New("invalid target")

func explode(target string) (ip string, port int, service string, err error) {
	parts := strings.Split(target, "/")
	if len(parts) != 2 {
		return "", 0, "", ErrInvalidTarget
	}

	ipPort := strings.Split(parts[0], ":")
	if len(ipPort) != 2 {
		return "", 0, "", ErrInvalidTarget
	}

	port, err = strconv.Atoi(ipPort[1])
	if err != nil {
		return "", 0, "", ErrInvalidTarget
	}

	ip = strings.TrimSpace(ipPort[0])
	if ip == "" {
		return "", 0, "", ErrInvalidTarget
	}

	service = strings.TrimSpace(parts[1])
	if service == "" {
		return "", 0, "", ErrInvalidTarget
	}

	return
}
