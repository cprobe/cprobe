package exporter

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	// template format: command, host_label
	//commandNotAllowedTmpl     = "warning: %q command isn't allowed at %q, see '4lw.commands.whitelist' ZK config parameter"
	instanceNotServingMessage = "This ZooKeeper instance is not currently serving requests"
	cmdNotExecutedSffx        = "is not executed because it is not in the whitelist."
)

type (
	Exporter struct {
		options     Options
		upIndicator *prometheus.Desc
		metrics     map[string]zookeeperMetric
	}

	Options struct {
		Timeout       time.Duration
		Host          string
		ResetOnScrape bool
		EnableTLS     bool
		ClientCert    *tls.Certificate
	}

	zookeeperMetric struct {
		desc          *prometheus.Desc
		extract       func(string) float64
		extractLabels func(s string) []string
		valType       prometheus.ValueType
	}
)

func NewZookeeperExporter(opts Options) *Exporter {

	c := &Exporter{
		options:     opts,
		upIndicator: prometheus.NewDesc("zk_up", "Exporter successful", nil, nil),
		metrics:     initMetricDesc(),
	}
	if c.options.Timeout == 0 {
		c.options.Timeout = 5 * time.Second
	}
	return c
}

func (c *Exporter) Describe(ch chan<- *prometheus.Desc) {

	logger.Infof("Sending %d metrics descriptions", len(c.metrics))
	for _, i := range c.metrics {
		ch <- i.desc
	}
}

func (c *Exporter) Collect(ch chan<- prometheus.Metric) error {
	//logger.Infof("Fetching metrics from Zookeeper")

	data, ok := c.sendZkCommand("mntr")

	if !ok {
		logger.Errorf("Failed to fetch metrics")
		ch <- prometheus.MustNewConstMetric(c.upIndicator, prometheus.GaugeValue, 0)
		return errors.New("Failed to fetch metrics")
	}

	data = strings.TrimSpace(data)
	// get slice of strings from response, like 'zk_avg_latency 0'
	lines := strings.Split(data, "\n")

	// skip instance if it in a leader only state and doesn't server client requests
	if lines[0] == instanceNotServingMessage {
		ch <- prometheus.MustNewConstMetric(c.upIndicator, prometheus.GaugeValue, 1)
		return errors.New("Can not response requests")
	}

	// 'mntr' command isn't allowed in zk config, log as a warning
	if strings.Contains(lines[0], cmdNotExecutedSffx) {
		ch <- prometheus.MustNewConstMetric(c.upIndicator, prometheus.GaugeValue, 0)
		return errors.New("mntr can not executed")
	}

	status := 1.0
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			logger.Warnf("%s, Unexpected format of returned data, expected tab-separated key/value.", line)
			status = 0
			continue
		}
		label, value := parts[0], parts[1]
		if metric, ok := c.metrics[label]; ok {
			//logger.Infof("Sending metric %s=%s", label, value)
			if metric.extractLabels != nil {
				ch <- prometheus.MustNewConstMetric(metric.desc, metric.valType, metric.extract(value), metric.extractLabels(value)...)
			} else {
				ch <- prometheus.MustNewConstMetric(metric.desc, metric.valType, metric.extract(value))
			}
		}
	}
	ch <- prometheus.MustNewConstMetric(c.upIndicator, prometheus.GaugeValue, status)

	if c.options.ResetOnScrape {
		c.resetStatistics()
	}
	return nil
}

func (c *Exporter) resetStatistics() {
	logger.Infof("Resetting Zookeeper statistics")
	_, ok := c.sendZkCommand("srst")
	if !ok {
		logger.Warnf("Failed to reset statistics")
	}
}

func (c *Exporter) sendZkCommand(fourLetterWord string) (string, bool) {
	//log.Debugf("Connecting to Zookeeper at %s", *zookeeperAddr)

	conn, err := c.zkConnect()
	if err != nil {
		logger.Errorf("Unable to open connection to Zookeeper: %s", err.Error())
		return "", false
	}
	defer conn.Close()

	if err = conn.SetDeadline(time.Now().Add(c.options.Timeout)); err != nil {
		logger.Errorf("Failed to set timeout on Zookeeper connection: %s", err.Error())
		return "", false
	}

	//log.WithFields(log.Fields{"command": fourLetterWord}).Debug("Sending four letter word")
	if _, err = conn.Write([]byte(fourLetterWord)); err != nil {
		logger.Errorf("Error sending command to Zookeeper: %s", err.Error())
		return "", false
	}
	scanner := bufio.NewScanner(conn)

	buffer := bytes.Buffer{}
	for scanner.Scan() {
		buffer.WriteString(scanner.Text() + "\n")
	}
	if err = scanner.Err(); err != nil {
		logger.Errorf("Error parsing response from Zookeeper: %s", err.Error())
		return "", false
	}
	//log.Debug("Successfully retrieved reply")

	return buffer.String(), true
}

func (c *Exporter) zkConnect() (net.Conn, error) {

	if c.options.EnableTLS {
		return tls.Dial("tcp", c.options.Host, &tls.Config{
			Certificates: []tls.Certificate{*c.options.ClientCert},
		})
	}

	return net.Dial("tcp", c.options.Host)
}

func parseFloatOrZero(s string) float64 {

	res, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Warnf("Failed to parse to float64: %s", err)
		return 0.0
	}
	return res
}
