// Copyright 2021 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

func (c *Config) gatherSnapshots(ctx context.Context, target *url.URL, hc *http.Client, ss *types.Samples) error {
	if !c.GatherSnapshots {
		return nil
	}

	// indices
	snapshotsStatsResp := make(map[string]SnapshotStatsResponse)
	u := target.ResolveReference(&url.URL{Path: "/_snapshot"})

	var srr SnapshotRepositoriesResponse
	resp, err := getURL(ctx, hc, u.String())
	if err != nil {
		return err
	}

	err = json.Unmarshal(resp, &srr)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	for repository := range srr {
		pathPart := path.Join("/_snapshot", repository, "/_all")
		u := target.ResolveReference(&url.URL{Path: pathPart})
		var ssr SnapshotStatsResponse
		resp, err := getURL(ctx, hc, u.String())
		if err != nil {
			continue
		}
		err = json.Unmarshal(resp, &ssr)
		if err != nil {
			return fmt.Errorf("failed to unmarshal JSON: %v", err)
		}
		snapshotsStatsResp[repository] = ssr
	}

	prefix := namespace + "_snapshot_stats"

	// Snapshots stats
	for repositoryName, snapshotStats := range snapshotsStatsResp {
		repoNameTag := map[string]string{"repository": repositoryName}

		ss.AddMetric(prefix, map[string]interface{}{
			"number_of_snapshots": float64(len(snapshotStats.Snapshots)),
		}, repoNameTag)

		oldest := float64(0)
		if len(snapshotStats.Snapshots) > 0 {
			oldest = float64(snapshotStats.Snapshots[0].StartTimeInMillis / 1000)
		}

		ss.AddMetric(prefix, map[string]interface{}{
			"oldest_snapshot_timestamp": oldest,
		}, repoNameTag)

		latest := float64(0)
		for i := len(snapshotStats.Snapshots) - 1; i >= 0; i-- {
			var snap = snapshotStats.Snapshots[i]
			if snap.State == "SUCCESS" || snap.State == "PARTIAL" {
				latest = float64(snap.StartTimeInMillis / 1000)
				break
			}
		}

		ss.AddMetric(prefix, map[string]interface{}{
			"latest_snapshot_timestamp_seconds": latest,
		}, repoNameTag)

		if len(snapshotStats.Snapshots) == 0 {
			continue
		}

		lastSnapshot := snapshotStats.Snapshots[len(snapshotStats.Snapshots)-1]

		stateVersionTag := map[string]string{
			"state":   lastSnapshot.State,
			"version": lastSnapshot.Version,
		}

		fields := map[string]interface{}{
			"snapshot_number_of_indices":    float64(len(lastSnapshot.Indices)),
			"snapshot_start_time_timestamp": float64(lastSnapshot.StartTimeInMillis / 1000),
			"snapshot_end_time_timestamp":   float64(lastSnapshot.EndTimeInMillis / 1000),
			"snapshot_number_of_failures":   float64(len(lastSnapshot.Failures)),
			"snapshot_total_shards":         float64(lastSnapshot.Shards.Total),
			"snapshot_failed_shards":        float64(lastSnapshot.Shards.Failed),
			"snapshot_successful_shards":    float64(lastSnapshot.Shards.Successful),
		}

		ss.AddMetric(prefix, fields, repoNameTag, stateVersionTag)
	}

	return nil
}
