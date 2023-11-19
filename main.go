package main

import (
	"context"
	"flag"
	"os"

	"github.com/cprobe/cprobe/flags"
	"github.com/cprobe/cprobe/httpd"
	"github.com/cprobe/cprobe/lib/buildinfo"
	"github.com/cprobe/cprobe/lib/envflag"
	"github.com/cprobe/cprobe/lib/flagutil"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/cprobe/cprobe/lib/procutil"
	"github.com/cprobe/cprobe/lib/runner"
	"github.com/cprobe/cprobe/probe"
)

func main() {
	// Write flags and help message to stdout, since it is easier to grep or pipe.
	flag.CommandLine.SetOutput(os.Stdout)
	flag.Usage = usage
	envflag.Parse()

	buildinfo.Init()
	logger.Init()
	runner.PrintRuntime()

	ctx, cancel := context.WithCancel(context.Background())

	if err := probe.Start(ctx); err != nil {
		logger.Fatalf("failed to start probe: %v", err)
	}

	// http server
	logger.Infof("http server listening on: %s", flags.HTTPListen)
	closeHTTP := httpd.Router().Config().Start()

	// stop process
	sig := procutil.WaitForSigterm()
	logger.Infof("service received signal %s", sig)
	if err := closeHTTP(); err != nil {
		logger.Fatalf("failed to stop the webservice: %s", err)
	}

	cancel()
}

func usage() {
	const s = `
cprobe is a frankenstein made up of vmagent and exporters.
`
	flagutil.Usage(s)
}
