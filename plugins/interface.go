package plugins

import (
	"context"

	"github.com/cprobe/cprobe/types"
)

type Plugin interface {
	// ParseConfig is used to parse config
	ParseConfig(baseDir string, bs []byte) (any, error)
	// Scrape is used to scrape metrics, cfg need to be cast specific cfg
	Scrape(ctx context.Context, target string, cfg any, ss *types.Samples) error
}

var registry = make(map[string]Plugin)

func GetPlugin(pluginName string) (Plugin, bool) {
	p, ok := registry[pluginName]
	return p, ok
}

func RegisterPlugin(pluginName string, p Plugin) {
	registry[pluginName] = p
}
