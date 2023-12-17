package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/blang/semver/v4"
	"github.com/cprobe/cprobe/types"
)

// ClusterInfoResponse is the cluster info retrievable from the / endpoint
type ClusterInfoResponse struct {
	Name        string      `json:"name"`
	ClusterName string      `json:"cluster_name"`
	ClusterUUID string      `json:"cluster_uuid"`
	Version     VersionInfo `json:"version"`
	Tagline     string      `json:"tagline"`
}

// VersionInfo is the version info retrievable from the / endpoint, embedded in ClusterInfoResponse
type VersionInfo struct {
	Number        semver.Version `json:"number"`
	BuildHash     string         `json:"build_hash"`
	BuildDate     string         `json:"build_date"`
	BuildSnapshot bool           `json:"build_snapshot"`
	LuceneVersion semver.Version `json:"lucene_version"`
}

func (c *Config) gatherClusterInfo(ctx context.Context, u *url.URL, hc *http.Client, ss *types.Samples) (string, error) {
	if !c.GatherClusterInfo {
		return "", nil
	}

	resp, err := hc.Get(u.String())
	if err != nil {
		return "", err
	}

	if resp.Body == nil {
		return "", fmt.Errorf("empty response body")
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var info ClusterInfoResponse
	err = json.Unmarshal(b, &info)
	if err != nil {
		return "", err
	}

	fields := map[string]interface{}{
		"version": 1.0,
	}

	tags := map[string]string{
		"cluster":        info.ClusterName,
		"cluster_uuid":   info.ClusterUUID,
		"build_date":     info.Version.BuildDate,
		"build_hash":     info.Version.BuildHash,
		"version":        info.Version.Number.String(),
		"lucene_version": info.Version.LuceneVersion.String(),
	}

	ss.AddMetric(namespace, fields, tags)

	return info.ClusterName, nil
}
