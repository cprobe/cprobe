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

// Package exporter implements the collectors and metrics handlers.
package exporter

import (
	"context"
	"fmt"
	_ "net/http/pprof"
	"sync"
	"time"

	"github.com/cprobe/cprobe/lib/logger"
	"github.com/cprobe/cprobe/plugins/mongodb/exporter/dsn_fix"
	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/mongo-driver/mongo"
)

// Exporter holds Exporter methods and attributes.
type Exporter struct {
	opts *Opts
	lock *sync.Mutex
}

// Opts holds new exporter options.
type Opts struct {
	// Only get stats for the collections matching this list of namespaces.
	// Example: db1.col1,db.col1
	CollStatsNamespaces    []string
	CollStatsLimit         int
	CompatibleMode         bool
	DirectConnect          bool
	ConnectTimeout         time.Duration
	DisableDefaultRegistry bool
	DiscoveringMode        bool
	ProfileTimeTS          int

	EnableDBStats            bool
	EnableDBStatsFreeStorage bool
	EnableDiagnosticData     bool
	EnableReplicasetStatus   bool
	EnableCurrentopMetrics   bool
	EnableTopMetrics         bool
	EnableIndexStats         bool
	EnableCollStats          bool
	EnableProfile            bool

	EnableOverrideDescendingIndex bool

	IndexStatsCollections []string

	URI string
}

var (
	errCannotHandleType   = fmt.Errorf("don't know how to handle data type")
	errUnexpectedDataType = fmt.Errorf("unexpected data type")
)

const (
	defaultCacheSize = 1000
)

// New connects to the database and returns a new Exporter instance.
func New(opts *Opts) *Exporter {
	if opts == nil {
		opts = new(Opts)
	}

	exp := &Exporter{
		opts: opts,
		lock: &sync.Mutex{},
	}

	return exp
}

func (e *Exporter) makeRegistry(ctx context.Context, client *mongo.Client, topologyInfo labelsGetter) *prometheus.Registry {
	registry := prometheus.NewRegistry()

	nodeType, err := getNodeType(ctx, client)
	if err != nil {
		logger.Errorf("Registry - Cannot get node type to check if this is a mongos : %s", err)
	}

	isArbiter, err := isArbiter(ctx, client)
	if err != nil {
		logger.Errorf("Registry - Cannot get arbiterOnly to check if this is arbiter role : %s", err)
	}

	// Enable collectors like collstats and indexstats depending on the number of collections
	// present in the database.
	limitsOk := false

	// arbiter only have isMaster privileges
	if isArbiter {
		e.opts.EnableDBStats = false
		e.opts.EnableDBStatsFreeStorage = false
		e.opts.EnableCollStats = false
		e.opts.EnableTopMetrics = false
		e.opts.EnableReplicasetStatus = false
		e.opts.EnableIndexStats = false
		e.opts.EnableCurrentopMetrics = false
		e.opts.EnableProfile = false
	} else {
		if e.opts.CollStatsLimit <= 0 {
			limitsOk = true
		} else {
			count, err := nonSystemCollectionsCount(ctx, client, nil, nil)
			if err != nil {
				logger.Errorf("Registry - Cannot get collections count : %s", err)
			} else {
				if count <= e.opts.CollStatsLimit {
					limitsOk = true
				}
			}
		}
	}

	// If we manually set the collection names we want or auto discovery is set.
	if (len(e.opts.CollStatsNamespaces) > 0 || e.opts.DiscoveringMode) && e.opts.EnableCollStats && limitsOk {
		cc := newCollectionStatsCollector(ctx, client, topologyInfo, e.opts)
		registry.MustRegister(cc)
	}

	// If we manually set the collection names we want or auto discovery is set.
	if (len(e.opts.IndexStatsCollections) > 0 || e.opts.DiscoveringMode) && e.opts.EnableIndexStats && limitsOk {
		ic := newIndexStatsCollector(ctx, client, topologyInfo, e.opts)
		registry.MustRegister(ic)
	}

	if e.opts.EnableDiagnosticData {
		ddc := newDiagnosticDataCollector(ctx, client,
			e.opts.CompatibleMode, topologyInfo)
		registry.MustRegister(ddc)
	}

	if e.opts.EnableDBStats && limitsOk {
		cc := newDBStatsCollector(ctx, client,
			e.opts.CompatibleMode, topologyInfo, nil, e.opts.EnableDBStatsFreeStorage, e.opts)
		registry.MustRegister(cc)
	}

	if e.opts.EnableCurrentopMetrics && nodeType != typeMongos && limitsOk {
		coc := newCurrentopCollector(ctx, client,
			e.opts.CompatibleMode, topologyInfo)
		registry.MustRegister(coc)
	}

	if e.opts.EnableProfile && nodeType != typeMongos && limitsOk && e.opts.ProfileTimeTS != 0 {
		pc := newProfileCollector(ctx, client,
			e.opts.CompatibleMode, topologyInfo, e.opts.ProfileTimeTS)
		registry.MustRegister(pc)
	}

	if e.opts.EnableTopMetrics && nodeType != typeMongos && limitsOk {
		tc := newTopCollector(ctx, client,
			e.opts.CompatibleMode, topologyInfo)
		registry.MustRegister(tc)
	}

	// replSetGetStatus is not supported through mongos.
	if e.opts.EnableReplicasetStatus && nodeType != typeMongos {
		rsgsc := newReplicationSetStatusCollector(ctx, client,
			e.opts.CompatibleMode, topologyInfo)
		registry.MustRegister(rsgsc)
	}

	return registry
}

func (e *Exporter) GetClient(ctx context.Context) (*mongo.Client, error) {
	client, err := connect(ctx, e.opts)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func connect(ctx context.Context, opts *Opts) (*mongo.Client, error) {
	clientOpts, err := dsn_fix.ClientOptionsForDSN(opts.URI)
	if err != nil {
		return nil, fmt.Errorf("invalid dsn: %w", err)
	}

	clientOpts.SetDirect(opts.DirectConnect)
	clientOpts.SetAppName("cprobe")

	connectTimeout := opts.ConnectTimeout
	if opts.ConnectTimeout == 0 {
		connectTimeout = 10 * time.Second
	}

	clientOpts.SetConnectTimeout(connectTimeout)
	clientOpts.SetServerSelectionTimeout(connectTimeout)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("invalid MongoDB options: %w", err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		// Ping failed. Close background connections. Error is ignored since the ping error is more relevant.
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("cannot connect to MongoDB: %w", err)
	}

	return client, nil
}
