package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cprobe/cprobe/httpd"
	"github.com/cprobe/cprobe/lib/buildinfo"
	"github.com/cprobe/cprobe/lib/envflag"
	"github.com/cprobe/cprobe/lib/fileutil"
	"github.com/cprobe/cprobe/lib/flagutil"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/cprobe/cprobe/lib/runner"
	"github.com/cprobe/cprobe/probe"
	"github.com/cprobe/cprobe/writer"
)

var (
	configDirectory = flag.String("conf.d", "conf.d", "Filepath to conf.d")
)

func init() {
	if *configDirectory == "" {
		fmt.Println("-conf.d is empty")
		os.Exit(1)
	}

	if !fileutil.IsExist(*configDirectory) {
		fmt.Printf("-conf.d %s does not exist\n", *configDirectory)
		os.Exit(1)
	}

	if !fileutil.IsDir(*configDirectory) {
		fmt.Printf("-conf.d %s is not a directory\n", *configDirectory)
		os.Exit(1)
	}
}

func main() {
	// Write flags and help message to stdout, since it is easier to grep or pipe.
	flag.CommandLine.SetOutput(os.Stdout)
	flag.Usage = usage
	envflag.Parse()

	buildinfo.Init()
	logger.Init()
	runner.PrintRuntime()

	ctx, cancel := context.WithCancel(context.Background())

	writer.Init(*configDirectory)

	if err := probe.Start(ctx, *configDirectory); err != nil {
		logger.Fatalf("cannot start probe: %v", err)
	}

	// http server
	closeHTTP := httpd.Router().Config().Start()

	sc := make(chan os.Signal, 1)
	// syscall.SIGUSR2 == 0xc , not available on windows
	signal.Notify(sc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGPIPE)

EXIT:
	for {
		sig := <-sc
		logger.Infof("received signal: %v", sig.String())
		switch sig {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			break EXIT
		case syscall.SIGHUP:
			probe.Reload(ctx, *configDirectory)
		case syscall.SIGPIPE:
			// https://pkg.go.dev/os/signal#hdr-SIGPIPE
			// do nothing
		}
	}

	// stop process
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
