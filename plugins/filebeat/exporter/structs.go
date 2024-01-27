package collector

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

// BeatInfo beat info json structure
type BeatInfo struct {
	Beat     string `json:"beat"`
	Hostname string `json:"hostname"`
	Name     string `json:"name"`
	UUID     string `json:"uuid"`
	Version  string `json:"version"`
}

func (b *BeatInfo) String() string {
	return fmt.Sprintf("beat: %s, hostname: %s, name: %s, uuid: %s, version: %s", b.Beat, b.Hostname, b.Name, b.UUID, b.Version)
}

// Stats stats endpoint json structure
type Stats struct {
	System     System      `json:"system"`
	Beat       BeatStats   `json:"beat"`
	LibBeat    LibBeat     `json:"libbeat"`
	Registrar  Registrar   `json:"registrar"`
	Filebeat   Filebeat    `json:"filebeat"`
	Metricbeat Metricbeat  `json:"metricbeat"`
	Auditd     AuditdStats `json:"auditd"`
}

type exportedMetrics []struct {
	desc    *prometheus.Desc
	eval    func(stats *Stats) float64
	valType prometheus.ValueType
}
