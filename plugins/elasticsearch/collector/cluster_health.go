package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/cprobe/cprobe/types"
	"github.com/pkg/errors"
)

var (
	colors = []string{"green", "yellow", "red"}
)

type clusterHealthResponse struct {
	ClusterName                 string  `json:"cluster_name"`
	Status                      string  `json:"status"`
	TimedOut                    bool    `json:"timed_out"`
	NumberOfNodes               int     `json:"number_of_nodes"`
	NumberOfDataNodes           int     `json:"number_of_data_nodes"`
	ActivePrimaryShards         int     `json:"active_primary_shards"`
	ActiveShards                int     `json:"active_shards"`
	RelocatingShards            int     `json:"relocating_shards"`
	InitializingShards          int     `json:"initializing_shards"`
	UnassignedShards            int     `json:"unassigned_shards"`
	DelayedUnassignedShards     int     `json:"delayed_unassigned_shards"`
	NumberOfPendingTasks        int     `json:"number_of_pending_tasks"`
	NumberOfInFlightFetch       int     `json:"number_of_in_flight_fetch"`
	TaskMaxWaitingInQueueMillis int     `json:"task_max_waiting_in_queue_millis"`
	ActiveShardsPercentAsNumber float64 `json:"active_shards_percent_as_number"`
}

func (c *Config) gatherClusterHealth(ctx context.Context, target *url.URL, httpClient *http.Client, ss *types.Samples, clusterName string) error {
	clusterHealthResp, err := c.fetchAndDecodeClusterHealth(target, httpClient)
	if err != nil {
		return errors.WithMessage(err, "failed to fetch and decode cluster health")
	}

	prefix := namespace + "_cluster_health"
	clusterTag := map[string]string{"cluster": clusterName}
	fields := map[string]interface{}{
		"active_primary_shards":            clusterHealthResp.ActivePrimaryShards,
		"active_shards":                    clusterHealthResp.ActiveShards,
		"delayed_unassigned_shards":        clusterHealthResp.DelayedUnassignedShards,
		"initializing_shards":              clusterHealthResp.InitializingShards,
		"number_of_data_nodes":             clusterHealthResp.NumberOfDataNodes,
		"number_of_in_flight_fetch":        clusterHealthResp.NumberOfInFlightFetch,
		"task_max_waiting_in_queue_millis": clusterHealthResp.TaskMaxWaitingInQueueMillis,
		"number_of_nodes":                  clusterHealthResp.NumberOfNodes,
		"number_of_pending_tasks":          clusterHealthResp.NumberOfPendingTasks,
		"relocating_shards":                clusterHealthResp.RelocatingShards,
		"unassigned_shards":                clusterHealthResp.UnassignedShards,
	}

	ss.AddMetric(prefix, fields, clusterTag)

	for _, color := range colors {
		ss.AddMetric(prefix, map[string]interface{}{
			"status": colorValue(clusterHealthResp, color),
		}, map[string]string{
			"cluster": clusterName,
			"status":  color,
		})
	}

	return nil
}

func (c *Config) fetchAndDecodeClusterHealth(target *url.URL, httpClient *http.Client) (clusterHealthResponse, error) {
	var chr clusterHealthResponse

	u := *target
	u.Path = path.Join(u.Path, "/_cluster/health")
	res, err := httpClient.Get(u.String())
	if err != nil {
		return chr, fmt.Errorf("failed to get cluster health from %s://%s:%s%s: %s",
			u.Scheme, u.Hostname(), u.Port(), u.Path, err)
	}

	if res.Body == nil {
		return chr, fmt.Errorf("empty response body")
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return chr, fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	bts, err := io.ReadAll(res.Body)
	if err != nil {
		return chr, err
	}

	if err := json.Unmarshal(bts, &chr); err != nil {
		return chr, err
	}

	return chr, nil
}

func colorValue(clusterHealth clusterHealthResponse, color string) float64 {
	if clusterHealth.Status == color {
		return 1
	}
	return 0
}
