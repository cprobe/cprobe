// mongodb_exporter
// Copyright (C) 2017 Percona LLC
//
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

package exporter

import (
	"context"
	"strings"

	"github.com/cprobe/cprobe/lib/logger"
	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type collstatsCollector struct {
	ctx  context.Context
	base *baseCollector

	compatibleMode  bool
	discoveringMode bool
	topologyInfo    labelsGetter

	collections []string
	opts        *Opts
}

// newCollectionStatsCollector creates a collector for statistics about collections.
func newCollectionStatsCollector(ctx context.Context, client *mongo.Client, topology labelsGetter, opts *Opts) *collstatsCollector {
	return &collstatsCollector{
		ctx:  ctx,
		base: newBaseCollector(client),

		compatibleMode:  opts.CompatibleMode,
		discoveringMode: opts.DiscoveringMode,
		topologyInfo:    topology,

		collections: opts.CollStatsNamespaces,
		opts:        opts,
	}
}

func (d *collstatsCollector) Describe(ch chan<- *prometheus.Desc) {
	d.base.Describe(d.ctx, ch, d.collect)
}

func (d *collstatsCollector) Collect(ch chan<- prometheus.Metric) {
	d.base.Collect(ch)
}

func (d *collstatsCollector) collect(ch chan<- prometheus.Metric) {
	defer measureCollectTime(ch, "mongodb", "collstats")()

	collections := d.collections

	client := d.base.client

	if d.discoveringMode {
		namespaces, err := listAllCollections(d.ctx, client, d.collections, systemDBs)
		if err != nil {
			logger.Errorf("cannot auto discover databases and collections: %s, target uri: %v", err.Error(), d.opts.URI)
			return
		}

		collections = fromMapToSlice(namespaces)
	}

	for _, dbCollection := range collections {
		parts := strings.Split(dbCollection, ".")
		if len(parts) < 2 { //nolint:gomnd
			continue
		}

		database := parts[0]
		collection := strings.Join(parts[1:], ".") // support collections having a .

		aggregation := bson.D{
			{
				Key: "$collStats", Value: bson.M{
					// TODO: PMM-9568 : Add support to handle histogram metrics
					"latencyStats": bson.M{"histograms": false},
					"storageStats": bson.M{"scale": 1},
				},
			},
		}
		project := bson.D{
			{
				Key: "$project", Value: bson.M{
					"storageStats.wiredTiger":   0,
					"storageStats.indexDetails": 0,
				},
			},
		}

		cursor, err := client.Database(database).Collection(collection).Aggregate(d.ctx, mongo.Pipeline{aggregation, project})
		if err != nil {
			logger.Errorf("cannot get $collstats cursor for collection %s.%s: %s", database, collection, err)

			continue
		}

		var stats []bson.M
		if err = cursor.All(d.ctx, &stats); err != nil {
			logger.Errorf("cannot get $collstats for collection %s.%s: %s", database, collection, err)
			continue
		}

		prefix := "collstats"
		labels := d.topologyInfo.baseLabels()
		labels["database"] = database
		labels["collection"] = collection

		for _, metrics := range stats {
			for _, metric := range makeMetrics(prefix, metrics, labels, d.compatibleMode) {
				ch <- metric
			}
		}
	}
}

func fromMapToSlice(databases map[string][]string) []string {
	var collections []string
	for db, cols := range databases {
		for _, value := range cols {
			collections = append(collections, db+"."+value)
		}
	}

	return collections
}

var _ prometheus.Collector = (*collstatsCollector)(nil)
