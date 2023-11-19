package main

import (
	"flag"
	"os"

	"github.com/cprobe/cprobe/lib/buildinfo"
	"github.com/cprobe/cprobe/lib/envflag"
	"github.com/cprobe/cprobe/lib/flagutil"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/cprobe/cprobe/lib/runner"
)

func main() {
	// Write flags and help message to stdout, since it is easier to grep or pipe.
	flag.CommandLine.SetOutput(os.Stdout)
	flag.Usage = usage
	envflag.Parse()

	buildinfo.Init()
	logger.Init()
	runner.PrintRuntime()

	logger.Infof("starting cprobe %s...", buildinfo.Version)
}

func usage() {
	const s = `
cprobe is a frankenstein made up of vmagent and exporters.
`
	flagutil.Usage(s)
}
