package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"dario.cat/mergo"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/cprobe/cprobe/types"
)

// clusterSettingsResponse is a representation of a Elasticsearch Cluster Settings
type clusterSettingsResponse struct {
	Defaults   clusterSettingsSection `json:"defaults"`
	Persistent clusterSettingsSection `json:"persistent"`
	Transient  clusterSettingsSection `json:"transient"`
}

// clusterSettingsSection is a representation of a Elasticsearch Cluster Settings
type clusterSettingsSection struct {
	Cluster clusterSettingsCluster `json:"cluster"`
}

// clusterSettingsCluster is a representation of a Elasticsearch clusterSettingsCluster Settings
type clusterSettingsCluster struct {
	Routing clusterSettingsRouting `json:"routing"`
	// This can be either a JSON object (which does not contain the value we are interested in) or a string
	MaxShardsPerNode interface{} `json:"max_shards_per_node"`
}

// clusterSettingsRouting is a representation of a Elasticsearch Cluster shard routing configuration
type clusterSettingsRouting struct {
	Allocation clusterSettingsAllocation `json:"allocation"`
}

// clusterSettingsAllocation is a representation of a Elasticsearch Cluster shard routing allocation settings
type clusterSettingsAllocation struct {
	Enabled string              `json:"enable"`
	Disk    clusterSettingsDisk `json:"disk"`
}

// clusterSettingsDisk is a representation of a Elasticsearch Cluster shard routing disk allocation settings
type clusterSettingsDisk struct {
	ThresholdEnabled string                   `json:"threshold_enabled"`
	Watermark        clusterSettingsWatermark `json:"watermark"`
}

// clusterSettingsWatermark is representation of Elasticsearch Cluster shard routing disk allocation watermark settings
type clusterSettingsWatermark struct {
	FloodStage string `json:"flood_stage"`
	High       string `json:"high"`
	Low        string `json:"low"`
}

func (c *Config) gatherClusterSettings(ctx context.Context, target *url.URL, hc *http.Client, ss *types.Samples) error {
	if !c.GatherClusterSettings {
		return nil
	}

	u := target.ResolveReference(&url.URL{Path: "_cluster/settings"})
	q := u.Query()
	q.Set("include_defaults", "true")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return err
	}

	resp, err := hc.Do(req)
	if err != nil {
		return err
	}

	if resp.Body == nil {
		return fmt.Errorf("empty response body")
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var data clusterSettingsResponse
	err = json.Unmarshal(b, &data)
	if err != nil {
		return err
	}

	// Merge all settings into one struct
	merged := data.Defaults

	err = mergo.Merge(&merged, data.Persistent, mergo.WithOverride)
	if err != nil {
		return err
	}

	err = mergo.Merge(&merged, data.Transient, mergo.WithOverride)
	if err != nil {
		return err
	}

	// Max shards per node
	if maxShardsPerNodeString, ok := merged.Cluster.MaxShardsPerNode.(string); ok {
		maxShardsPerNode, err := strconv.ParseInt(maxShardsPerNodeString, 10, 64)
		if err == nil {
			ss.AddMetric(namespace, map[string]interface{}{
				"clustersettings_stats_max_shards_per_node": float64(maxShardsPerNode),
			})
		}
	}

	// Shard allocation enabled
	shardAllocationMap := map[string]int{
		"all":           0,
		"primaries":     1,
		"new_primaries": 2,
		"none":          3,
	}

	ss.AddMetric(namespace, map[string]interface{}{
		"clustersettings_stats_shard_allocation_enabled": float64(shardAllocationMap[merged.Cluster.Routing.Allocation.Enabled]),
	})

	// Threshold enabled
	thresholdMap := map[string]int{
		"false": 0,
		"true":  1,
	}

	ss.AddMetric(namespace, map[string]interface{}{
		"clustersettings_allocation_threshold_enabled": float64(thresholdMap[merged.Cluster.Routing.Allocation.Disk.ThresholdEnabled]),
	})

	// Watermark bytes or ratio metrics
	if strings.HasSuffix(merged.Cluster.Routing.Allocation.Disk.Watermark.High, "b") {
		flooodStageBytes, err := getValueInBytes(merged.Cluster.Routing.Allocation.Disk.Watermark.FloodStage)
		if err != nil {
			logger.Errorf("failed to parse flood_stage bytes: %v", err)
		} else {
			ss.AddMetric(namespace, map[string]interface{}{
				"clustersettings_allocation_watermark_flood_stage_bytes": flooodStageBytes,
			})
		}

		highBytes, err := getValueInBytes(merged.Cluster.Routing.Allocation.Disk.Watermark.High)
		if err != nil {
			logger.Errorf("failed to parse high bytes: %v", err)
		} else {
			ss.AddMetric(namespace, map[string]interface{}{
				"clustersettings_allocation_watermark_high_bytes": highBytes,
			})
		}

		lowBytes, err := getValueInBytes(merged.Cluster.Routing.Allocation.Disk.Watermark.Low)
		if err != nil {
			logger.Errorf("failed to parse low bytes: %v", err)
		} else {
			ss.AddMetric(namespace, map[string]interface{}{
				"clustersettings_allocation_watermark_low_bytes": lowBytes,
			})
		}

		return nil
	}

	// Watermark ratio metrics
	floodRatio, err := getValueAsRatio(merged.Cluster.Routing.Allocation.Disk.Watermark.FloodStage)
	if err != nil {
		logger.Errorf("failed to parse flood_stage ratio: %v", err)
	} else {
		ss.AddMetric(namespace, map[string]interface{}{
			"clustersettings_allocation_watermark_flood_stage_ratio": floodRatio,
		})
	}

	highRatio, err := getValueAsRatio(merged.Cluster.Routing.Allocation.Disk.Watermark.High)
	if err != nil {
		logger.Errorf("failed to parse high ratio: %v", err)
	} else {
		ss.AddMetric(namespace, map[string]interface{}{
			"clustersettings_allocation_watermark_high_ratio": highRatio,
		})
	}

	lowRatio, err := getValueAsRatio(merged.Cluster.Routing.Allocation.Disk.Watermark.Low)
	if err != nil {
		logger.Errorf("failed to parse low ratio: %v", err)
	} else {
		ss.AddMetric(namespace, map[string]interface{}{
			"clustersettings_allocation_watermark_low_ratio": lowRatio,
		})
	}

	return nil
}

func getValueInBytes(value string) (float64, error) {
	type UnitValue struct {
		unit string
		val  float64
	}

	unitValues := []UnitValue{
		{"pb", 1024 * 1024 * 1024 * 1024 * 1024},
		{"tb", 1024 * 1024 * 1024 * 1024},
		{"gb", 1024 * 1024 * 1024},
		{"mb", 1024 * 1024},
		{"kb", 1024},
		{"b", 1},
	}

	for _, uv := range unitValues {
		if strings.HasSuffix(value, uv.unit) {
			numberStr := strings.TrimSuffix(value, uv.unit)

			number, err := strconv.ParseFloat(numberStr, 64)
			if err != nil {
				return 0, err
			}
			return number * uv.val, nil
		}
	}

	return 0, fmt.Errorf("failed to convert unit %s to bytes", value)
}

func getValueAsRatio(value string) (float64, error) {
	if strings.HasSuffix(value, "%") {
		percentValue, err := strconv.Atoi(strings.TrimSpace(strings.TrimSuffix(value, "%")))
		if err != nil {
			return 0, err
		}

		return float64(percentValue) / 100, nil
	}

	ratio, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}

	return ratio, nil
}
