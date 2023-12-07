package probe

import (
	"github.com/cprobe/cprobe/types"

	_ "github.com/cprobe/cprobe/plugins/blackbox"
	_ "github.com/cprobe/cprobe/plugins/kafka"
	_ "github.com/cprobe/cprobe/plugins/mysql"
	_ "github.com/cprobe/cprobe/plugins/prometheus"
	_ "github.com/cprobe/cprobe/plugins/redis"
)

func makeJobs() map[string]map[JobID]*JobGoroutine {
	return map[string]map[JobID]*JobGoroutine{
		types.PluginMySQL:         make(map[JobID]*JobGoroutine),
		types.PluginRedis:         make(map[JobID]*JobGoroutine),
		types.PluginMongoDB:       make(map[JobID]*JobGoroutine),
		types.PluginPostgreSQL:    make(map[JobID]*JobGoroutine),
		types.PluginElasticSearch: make(map[JobID]*JobGoroutine),
		types.PluginKafka:         make(map[JobID]*JobGoroutine),
		types.PluginBlackbox:      make(map[JobID]*JobGoroutine),
		types.PluginPrometheus:    make(map[JobID]*JobGoroutine),
	}
}
