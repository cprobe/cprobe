package main

import (
	"context"
	"flag"
	"os"

	"github.com/cprobe/cprobe/flags"
	"github.com/cprobe/cprobe/httpd"
	"github.com/cprobe/cprobe/lib/buildinfo"
	"github.com/cprobe/cprobe/lib/envflag"
	"github.com/cprobe/cprobe/lib/fasttime"
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

	var vs []*writer.Vector
	vs = append(vs, &writer.Vector{
		Labels: map[string]string{
			"__name__": "cprobe_test02",
			"job":      "cprobe01",
		},
		Clock: int64(fasttime.UnixTimestamp()) * 1000,
		Value: 1.1,
	})

	vs = append(vs, &writer.Vector{
		Labels: map[string]string{
			"__name__": "cprobe_test02",
			"job":      "cprobe02",
		},
		Clock: int64(fasttime.UnixTimestamp()) * 1000,
		Value: 1.1,
	})

	writer.WriteVectors(vs)

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
