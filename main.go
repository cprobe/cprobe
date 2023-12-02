package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/cprobe/cprobe/flags"
	"github.com/cprobe/cprobe/httpd"
	"github.com/cprobe/cprobe/lib/buildinfo"
	"github.com/cprobe/cprobe/lib/envflag"
	"github.com/cprobe/cprobe/lib/fileutil"
	"github.com/cprobe/cprobe/lib/flagutil"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/cprobe/cprobe/lib/runner"
	"github.com/cprobe/cprobe/probe"
	service_install "github.com/cprobe/cprobe/service/install"
	service_update "github.com/cprobe/cprobe/service/update"
	"github.com/cprobe/cprobe/writer"
	"github.com/kardianos/service"
	"github.com/pkg/errors"
)

var (
	install    = flag.Bool("install", false, "Install service")
	remove     = flag.Bool("remove", false, "Remove service")
	start      = flag.Bool("start", false, "Start service")
	stop       = flag.Bool("stop", false, "Stop service")
	status     = flag.Bool("status", false, "Show service status")
	update     = flag.Bool("update", false, "Update binary")
	updateFile = flag.String("updateFile", "", "new version tar.gz file or url")
	nohttp     = flag.Bool("no-httpd", false, "Disable http server")
)

func init() {
	flag.StringVar(&flags.ConfigDirectory, "conf.d", "conf.d", "Filepath to conf.d")

	if flags.ConfigDirectory == "" {
		fmt.Println("-conf.d is empty")
		os.Exit(1)
	}

	if !fileutil.IsExist(flags.ConfigDirectory) {
		fmt.Printf("-conf.d %s does not exist\n", flags.ConfigDirectory)
		os.Exit(1)
	}

	if !fileutil.IsDir(flags.ConfigDirectory) {
		fmt.Printf("-conf.d %s is not a directory\n", flags.ConfigDirectory)
		os.Exit(1)
	}
}

func main() {
	// Write flags and help message to stdout, since it is easier to grep or pipe.
	flag.CommandLine.SetOutput(os.Stdout)
	flag.Usage = usage
	envflag.Parse()

	if *install || *remove || *start || *stop || *status || *update {
		err := serviceProcess()
		if err != nil {
			fmt.Println("error:", err)
		}
		return
	}

	buildinfo.Init()
	logger.Init()
	runner.PrintRuntime()

	ctx, cancel := context.WithCancel(context.Background())

	if err := writer.Init(flags.ConfigDirectory); err != nil {
		logger.Fatalf("cannot init writer: %v", err)
	}

	if err := probe.Start(ctx, flags.ConfigDirectory); err != nil {
		logger.Fatalf("cannot start probe: %v", err)
	}

	var closeHTTP func() error
	if !*nohttp {
		// http server
		closeHTTP = httpd.Router().Config().Start()
	}

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
			probe.Reload(ctx, flags.ConfigDirectory)
		case syscall.SIGPIPE:
			// https://pkg.go.dev/os/signal#hdr-SIGPIPE
			// do nothing
		}
	}

	if !*nohttp {
		if err := closeHTTP(); err != nil {
			logger.Fatalf("cannot stop the webservice: %s", err)
		}
	}

	cancel()
}

func usage() {
	const s = `
cprobe is a frankenstein made up of vmagent and exporters.
`
	flagutil.Usage(s)
}

type program struct{}

func (p *program) Start(s service.Service) error {
	return nil
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func serviceProcess() error {
	svcConfig, err := service_install.ServiceConfig()
	if err != nil {
		return err
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		return errors.WithMessage(err, "failed to generate service")
	}

	if *stop {
		sts, err := s.Status()
		if err != nil {
			return errors.WithMessage(err, "failed to get service status")
		}

		if sts == service.StatusStopped {
			fmt.Println("service is already stopped")
			return nil
		}

		if err := s.Stop(); err != nil {
			return errors.WithMessage(err, "failed to stop service")
		} else {
			fmt.Println("service stopped")
			return nil
		}

	}

	if *remove {
		if err := s.Uninstall(); err != nil {
			return errors.WithMessage(err, "failed to remove service")
		} else {
			fmt.Println("service removed")
			return nil
		}
	}

	if *install {
		if err := s.Install(); err != nil {
			return errors.WithMessage(err, "failed to install service")
		} else {
			fmt.Println("service installed")
			return nil
		}
	}

	if *start {
		sts, err := s.Status()
		if err != nil {
			return errors.WithMessage(err, "failed to get service status")
		}

		if sts == service.StatusRunning {
			fmt.Println("service is already running")
			return nil
		}

		if err := s.Start(); err != nil {
			return errors.WithMessage(err, "failed to start service")
		} else {
			fmt.Println("service started")
			return nil
		}
	}

	if *status {
		sts, err := s.Status()
		if err != nil {
			return errors.WithMessage(err, "failed to get service status")
		}

		switch sts {
		case service.StatusRunning:
			fmt.Println("service status: running")
		case service.StatusStopped:
			fmt.Println("service status: stopped")
		default:
			fmt.Println("service status: unknown")
		}

		return nil
	}

	if *update {
		if *updateFile == "" {
			return errors.New("--updateFile is empty")
		}

		if _, err := s.Status(); err != nil {
			if strings.Contains(err.Error(), "not installed") {
				fmt.Println("service is not installed")
				return nil
			} else {
				return errors.WithMessage(err, "failed to get service status")
			}
		}

		if err := service_update.Update(*updateFile); err != nil {
			return errors.WithMessage(err, "failed to update service")
		}

		if err := s.Restart(); err != nil {
			return errors.WithMessage(err, "failed to restart service")
		} else {
			fmt.Println("service updated and restarted")
		}
	}

	return nil
}
