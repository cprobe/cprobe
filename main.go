package main

import (
	"context"
	"flag"
	"os"

	"github.com/cprobe/cprobe/httpd"
	"github.com/cprobe/cprobe/lib/buildinfo"
	"github.com/cprobe/cprobe/lib/envflag"
	"github.com/cprobe/cprobe/lib/flagutil"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/cprobe/cprobe/lib/procutil"
	"github.com/cprobe/cprobe/lib/runner"
	"github.com/cprobe/cprobe/probe"
	"github.com/cprobe/cprobe/writer"
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

	writer.Init()

	if err := probe.Start(ctx); err != nil {
		logger.Fatalf("cannot start probe: %v", err)
	}

	// http server
	closeHTTP := httpd.Router().Config().Start()

	// stop process
	sig := procutil.WaitForSigterm()
	logger.Infof("service received signal %s", sig)
	if err := closeHTTP(); err != nil {
		logger.Fatalf("cannot stop the webservice: %s", err)
	}

	cancel()
}

func usage() {
	const s = `
cprobe is a frankenstein made up of vmagent and exporters.
`
	flagutil.Usage(s)
}
