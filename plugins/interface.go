package plugins

import (
	"context"
	"fmt"
	"github.com/cprobe/cprobe/plugins/mysql"
	"github.com/cprobe/cprobe/types"
	"github.com/pkg/errors"
)

type Plugin interface {
	// ParseConfig is used to parse config
	ParseConfig(bs []byte) (any, error)
	// Scrape is used to scrape metrics, cfg need to be cost specific cfg
	Scrape(ctx context.Context, target string, cfg any, ss *types.Samples) error
}

var plugin = make(map[string]Plugin)

func InitPlugin() {
	plugin[types.PluginMySQL] = &mysql.Mysql{Name: types.PluginMySQL}
	plugin[types.PluginRedis] = &mysql.Mysql{Name: types.PluginRedis}
}
func GetPlugin(pluginName string) (Plugin, error) {
	p, ok := plugin[pluginName]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Unknown plugin: %s", pluginName))
	}
	return p, nil
}
