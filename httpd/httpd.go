package httpd

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/cprobe/cprobe/flags"
	"github.com/cprobe/cprobe/lib/flagutil"
	"github.com/cprobe/cprobe/lib/fs"
	"github.com/cprobe/cprobe/probe"
	"github.com/cprobe/cprobe/writer"
	"gopkg.in/yaml.v2"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cprobe/cprobe/lib/buildinfo"
	"github.com/cprobe/cprobe/lib/fasttime"
	"github.com/cprobe/cprobe/lib/ginx"
	"github.com/cprobe/cprobe/lib/httptls"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/valyala/fastrand"
)

var connDeadlineTimeKey = interface{}("connDeadlineSecs")

func init() {
	flag.StringVar(&HTTPListen, "http.listen", "0.0.0.0:5858", "Address to listen for http connections.")
	flag.StringVar(&HTTPUsername, "http.username", "", "Username for basic http authentication. No authentication is performed if username is empty.")
	flag.StringVar(&HTTPPassword, "http.password", "", "Password for basic http authentication. No authentication is performed if password is empty.")
	flag.StringVar(&HTTPMode, "http.mode", "release", "Gin mode. One of: {debug|release|test}")
	flag.BoolVar(&HTTPPProf, "http.pprof", false, "Enable pprof http handlers. This is insecure and should be disabled in production.")
	flag.DurationVar(&HTTPReadHeaderTimeout, "http.readTimeout", time.Second*5, "Maximum duration for reading request header.")
	flag.DurationVar(&HTTPIdleTimeout, "http.idleTimeout", time.Minute, "Maximum amount of time to wait for the next request when keep-alives are enabled.")
	flag.DurationVar(&HTTPConnTimeout, "http.connTimeout", 2*time.Minute, `Incoming http connections are closed after the configured timeout. This may help to spread the incoming load among a cluster of services behind a load balancer. Please note that the real timeout may be bigger by up to 10% as a protection against the thundering herd problem`)
	flag.BoolVar(&HTTPTLSEnable, "http.tls", false, "Whether to enable TLS for incoming HTTP requests at -http.listen (aka https). -http.tlsCertFile and -http.tlsKeyFile must be set if -http.tls is set")
	flag.StringVar(&HTTPTLSCertFile, "http.tlsCertFile", "", "Path to file with TLS certificate if -http.tls is set. Prefer ECDSA certs instead of RSA certs as RSA certs are slower. The provided certificate file is automatically re-read every second, so it can be dynamically updated")
	flag.StringVar(&HTTPTLSKeyFile, "http.tlsKeyFile", "", "Path to file with TLS key if -http.tls is set. The provided key file is automatically re-read every second, so it can be dynamically updated")
	flag.StringVar(&HTTPTLSCipherSuitesString, "http.tlsCipherSuites", "", "Optional list of TLS cipher suites for incoming requests over HTTPS if -http.tls is set. split by comma. See the list of supported cipher suites at https://pkg.go.dev/crypto/tls#pkg-constants")
	flag.StringVar(&HTTPTLSMinVersion, "http.tlsMinVersion", "", "Optional minimum TLS version to use for incoming requests over HTTPS if -http.tls is set. "+
		"Supported values: TLS10, TLS11, TLS12, TLS13")
	flag.DurationVar(&HTTPMaxGracefulShutdownDuration, "http.maxGracefulShutdownDuration", 7*time.Second, `The maximum duration for a graceful shutdown of the HTTP server. A highly loaded server may require increased value for a graceful shutdown`)

	pair := strings.Split(HTTPListen, ":")
	if len(pair) == 1 {
		log.Fatalf("invalid http.listen address: %s", HTTPListen)
	}

	port := pair[len(pair)-1]
	if p, e := strconv.ParseInt(port, 10, 64); e != nil || p < 1 || p > 65535 {
		log.Fatalf("invalid http.listen port: %s", port)
	} else {
		HTTPPort = int(p)
	}

	if HTTPTLSCipherSuitesString != "" {
		HTTPTLSCipherSuitesArray = strings.Fields(strings.ReplaceAll(HTTPTLSCipherSuitesString, ",", " "))
	}
}

type HTTPRouter struct {
	engine *gin.Engine
}

func Router() *HTTPRouter {
	gin.SetMode(HTTPMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(ginx.BombRecovery())
	r.LoadHTMLGlob("httpd/templates/*")
	if HTTPUsername != "" && HTTPPassword != "" {
		r.Use(gin.BasicAuth(gin.Accounts{
			HTTPUsername: HTTPPassword,
		}))
	}

	if HTTPPProf {
		pprof.Register(r, "/debug/pprof")
	}

	r.GET("/", func(c *gin.Context) {
		endpoints := map[string]string{
			"targets": "status for discovered active targets",
			"metrics": "available service metrics",
			"flags":   "command-line flags",
			"config":  "cprobe config contents",
			"reload":  "reload configuration",
		}
		if HTTPPProf {
			endpoints["/debug/pprof"] = "pprof"
		}
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"endpoints": endpoints,
			"plugins":   probe.PluginCfgs,
		})
	})
	r.GET("/flags", func(c *gin.Context) {
		flagutil.WriteFlags(c.Writer)
	})
	r.GET("/config", func(c *gin.Context) {
		config := writer.WriterConfig
		out, _ := yaml.Marshal(config)
		fmt.Fprint(c.Writer, string(out))
	})
	r.GET("/plugins/:name", func(c *gin.Context) {
		name := c.Param("name")
		if cfg, ok := probe.PluginCfgs[name]; ok {
			var rules string
			for _, config := range cfg {
				for _, scrapeConfig := range config.ScrapeConfigs {
					jobName := fmt.Sprintf("%s-", scrapeConfig.JobName)
					for _, ruleFile := range scrapeConfig.ScrapeRuleFiles {
						ruleFilePath := fs.GetFilepath(scrapeConfig.ConfigRef.BaseDir, ruleFile)
						data := probe.CacheGetBytes(ruleFilePath)
						if data == nil {
							data, _ = fs.ReadFileOrHTTP(ruleFilePath)
						}
						ruleValue := fmt.Sprintf("%s\n%s", ruleFile, data)
						if rules == "" {
							rules = fmt.Sprintf("%s%s", jobName, ruleValue)
						} else {
							rules = fmt.Sprintf("%s\n%s%s", rules, jobName, ruleValue)
						}
					}

				}
			}
			out, _ := yaml.Marshal(cfg)
			fmt.Fprint(c.Writer, string(out)+"\n"+rules)
		}
	})
	r.GET("/reload", func(c *gin.Context) {
		probe.Reload(c, flags.ConfigDirectory)
		c.String(http.StatusOK, "OK")
	})

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	r.GET("/pid", func(c *gin.Context) {
		c.String(http.StatusOK, fmt.Sprintf("%d", os.Getpid()))
	})

	r.GET("/ppid", func(c *gin.Context) {
		c.String(http.StatusOK, fmt.Sprintf("%d", os.Getppid()))
	})

	r.GET("/remoteaddr", func(c *gin.Context) {
		c.String(http.StatusOK, c.Request.RemoteAddr)
	})

	r.GET("/version", func(c *gin.Context) {
		c.String(http.StatusOK, buildinfo.Version)
	})

	return &HTTPRouter{engine: r}
}

// Init initializes http server and return close function
func (r *HTTPRouter) Start() func() error {
	server := &http.Server{
		Handler:           r.engine,
		IdleTimeout:       HTTPIdleTimeout,
		ReadHeaderTimeout: HTTPReadHeaderTimeout,
		ErrorLog:          logger.StdErrorLogger(),

		ConnContext: func(ctx context.Context, c net.Conn) context.Context {
			timeoutSec := HTTPConnTimeout.Seconds()
			// Add a jitter for connection timeout in order to prevent Thundering herd problem
			// when all the connections are established at the same time.
			// See https://en.wikipedia.org/wiki/Thundering_herd_problem
			jitterSec := fastrand.Uint32n(uint32(timeoutSec / 10))
			deadline := fasttime.UnixTimestamp() + uint64(timeoutSec) + uint64(jitterSec)
			return context.WithValue(ctx, connDeadlineTimeKey, &deadline)
		},
	}

	go func() {
		var tlsConfig *tls.Config
		if HTTPTLSEnable {
			tc, err := httptls.GetServerTLSConfig(HTTPTLSCertFile, HTTPTLSKeyFile, HTTPTLSMinVersion, HTTPTLSCipherSuitesArray)
			if err != nil {
				logger.Fatalf("cannot get TLS config for http server: %s", err)
			}
			tlsConfig = tc
		}

		listner, err := net.Listen("tcp", HTTPListen)
		if err != nil {
			logger.Fatalf("cannot listen %q: %s", HTTPListen, err)
		}

		if tlsConfig != nil {
			listner = tls.NewListener(listner, tlsConfig)
		}

		logger.Infof("listening http on %s", HTTPListen)
		if err := server.Serve(listner); err != nil {
			if err == http.ErrServerClosed {
				// The server gracefully closed.
				return
			}
			logger.Fatalf("cannot serve http at %s: %s", HTTPListen, err)
		}
	}()

	return func() error {
		ctx, cancel := context.WithTimeout(context.Background(), HTTPMaxGracefulShutdownDuration)
		defer cancel()
		return server.Shutdown(ctx)
	}
}
