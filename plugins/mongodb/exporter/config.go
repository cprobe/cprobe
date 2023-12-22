package exporter

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cprobe/cprobe/types"
	"github.com/pkg/errors"
)

type Config struct {
	BaseDir                                    string        `toml:"-"`
	User                                       string        `toml:"user"`
	Password                                   string        `toml:"password"`
	DirectConnect                              bool          `toml:"direct_connect"`
	ConnectTimeout                             time.Duration `toml:"connect_timeout"`
	CollstatsColls                             []string      `toml:"collstats_colls"`
	IndexstatsColls                            []string      `toml:"indexstats_colls"`
	CollectDiagnosticdata                      bool          `toml:"collect_diagnosticdata"`
	CollectReplicasetstatus                    bool          `toml:"collect_replicasetstatus"`
	CollectDBStats                             bool          `toml:"collect_dbstats"`
	CollectDBStatsFreeStorage                  bool          `toml:"collect_dbstatsfreestorage"`
	CollectTopMetrics                          bool          `toml:"collect_topmetrics"`
	CollectCurrentopMetrics                    bool          `toml:"collect_currentopmetrics"`
	CollectIndexStats                          bool          `toml:"collect_indexstats"`
	CollectCollStats                           bool          `toml:"collect_collstats"`
	CollectProfile                             bool          `toml:"collect_profile"`
	CollectProfileSlowqueriesTimeWindowSeconds int           `toml:"collect_profile_slowqueries_time_window_seconds"`
	MetricsOverrideDescendingIndex             bool          `toml:"metrics_override_descending_index"`
	DisableCollstatsIfCollcountMoreThan        int           `toml:"disable_collstats_if_collcount_more_than"`
	DiscoveringMode                            bool          `toml:"discovering_mode"`
	CompatibleMode                             bool          `toml:"compatible_mode"`
}

func (c *Config) Scrape(ctx context.Context, target string, ss *types.Samples) error {
	uri := buildURI(target, c.User, c.Password)

	exporterOpts := &Opts{
		CollStatsNamespaces:   c.CollstatsColls,
		CompatibleMode:        c.CompatibleMode,
		DiscoveringMode:       c.DiscoveringMode,
		IndexStatsCollections: c.IndexstatsColls,
		URI:                   uri,
		GlobalConnPool:        false,
		DirectConnect:         c.DirectConnect,
		ConnectTimeout:        c.ConnectTimeout,
		TimeoutOffset:         0,

		EnableDiagnosticData:     c.CollectDiagnosticdata,
		EnableReplicasetStatus:   c.CollectReplicasetstatus,
		EnableCurrentopMetrics:   c.CollectCurrentopMetrics,
		EnableTopMetrics:         c.CollectTopMetrics,
		EnableDBStats:            c.CollectDBStats,
		EnableDBStatsFreeStorage: c.CollectDBStatsFreeStorage,
		EnableIndexStats:         c.CollectIndexStats,
		EnableCollStats:          c.CollectCollStats,
		EnableProfile:            c.CollectProfile,

		EnableOverrideDescendingIndex: c.MetricsOverrideDescendingIndex,

		CollStatsLimit: c.DisableCollstatsIfCollcountMoreThan,
		ProfileTimeTS:  c.CollectProfileSlowqueriesTimeWindowSeconds,
	}

	e := New(exporterOpts)

	client, err := e.GetClient(ctx)
	if err != nil {
		return errors.WithMessage(err, "cannot get mongodb client")
	}

	if client == nil {
		return errors.New("mongodb client is nil")
	}

	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, nil); err != nil {
		return errors.WithMessage(err, "cannot ping mongodb")
	}

	count, err := NonSystemCollectionsCount(ctx, client, nil, nil)
	if err != nil {
		return errors.Wrap(err, "cannot get collections count")
	}

	e.totalCollectionsCount = count
	ti := newTopologyInfo(ctx, client)
	registry := e.makeRegistry(ctx, client, ti)

	mfs, err := registry.Gather()
	if err != nil {
		return errors.WithMessage(err, "failed to gather metrics from mongodb registry")
	}

	ss.AddMetricFamilies(mfs)

	return nil
}

func buildURI(uri string, user string, password string) string {
	// IF user@pass not contained in uri AND custom user and pass supplied in arguments
	// DO concat a new uri with user and pass arguments value
	if !strings.Contains(uri, "@") && user != "" && password != "" {
		// trim mongodb:// prefix to handle user and pass logic
		uri = strings.TrimPrefix(uri, "mongodb://")
		// add user and pass to the uri
		uri = fmt.Sprintf("%s:%s@%s", user, password, uri)
	}
	if !strings.HasPrefix(uri, "mongodb") {
		uri = "mongodb://" + uri
	}

	return uri
}
