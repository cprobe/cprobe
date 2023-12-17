package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/cprobe/cprobe/types"
)

// ShardResponse has shard's node and index info
type ShardResponse struct {
	Index string `json:"index"`
	Shard string `json:"shard"`
	State string `json:"state"`
	Node  string `json:"node"`
}

func (c *Config) gatherShardsTotal(ctx context.Context, target *url.URL, hc *http.Client, ss *types.Samples, clusterName string) error {
	sr, err := c.fetchAndDecodeShards(target, hc)
	if err != nil {
		return err
	}

	nodeShards := make(map[string]float64)

	for _, shard := range sr {
		if shard.State == "STARTED" {
			nodeShards[shard.Node]++
		}
	}

	for node, shards := range nodeShards {
		ss.AddMetric(namespace, map[string]interface{}{
			"node_shards_total": shards,
		}, map[string]string{
			"node":    node,
			"cluster": clusterName,
		})
	}

	return nil
}

func (c *Config) fetchAndDecodeShards(target *url.URL, hc *http.Client) ([]ShardResponse, error) {
	u := *target
	u.Path = path.Join(u.Path, "/_cat/shards")
	q := u.Query()
	q.Set("format", "json")
	u.RawQuery = q.Encode()
	sfr, err := c.getAndParseShardsURL(&u, hc)
	if err != nil {
		return sfr, err
	}
	return sfr, err
}

func (c *Config) getAndParseShardsURL(u *url.URL, hc *http.Client) ([]ShardResponse, error) {
	res, err := hc.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get from %s://%s:%s%s: %s",
			u.Scheme, u.Hostname(), u.Port(), u.Path, err)
	}

	if res.Body == nil {
		return nil, fmt.Errorf("empty response body")
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}
	var sfr []ShardResponse
	if err := json.NewDecoder(res.Body).Decode(&sfr); err != nil {
		return nil, err
	}

	return sfr, nil
}
