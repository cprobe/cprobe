package httpd

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cprobe/cprobe/flags"
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
	flag.StringVar(&flags.HTTPListen, "http.listen", "0.0.0.0:5858", "Address to listen for http connections.")
	flag.StringVar(&flags.HTTPUsername, "http.username", "", "Username for basic http authentication. No authentication is performed if username is empty.")
	flag.StringVar(&flags.HTTPPassword, "http.password", "", "Password for basic http authentication. No authentication is performed if password is empty.")
	flag.StringVar(&flags.HTTPMode, "http.mode", "release", "Gin mode. One of: {debug|release|test}")
	flag.BoolVar(&flags.HTTPPProf, "http.pprof", false, "Enable pprof http handlers. This is insecure and should be disabled in production.")
	flag.DurationVar(&flags.HTTPReadHeaderTimeout, "http.readTimeout", time.Second*5, "Maximum duration for reading request header.")
	flag.DurationVar(&flags.HTTPIdleTimeout, "http.idleTimeout", time.Minute, "Maximum amount of time to wait for the next request when keep-alives are enabled.")
	flag.DurationVar(&flags.HTTPConnTimeout, "http.connTimeout", 2*time.Minute, `Incoming http connections are closed after the configured timeout. This may help to spread the incoming load among a cluster of services behind a load balancer. Please note that the real timeout may be bigger by up to 10% as a protection against the thundering herd problem`)
	flag.BoolVar(&flags.HTTPTLSEnable, "http.tls", false, "Whether to enable TLS for incoming HTTP requests at -http.listen (aka https). -http.tlsCertFile and -http.tlsKeyFile must be set if -http.tls is set")
	flag.StringVar(&flags.HTTPTLSCertFile, "http.tlsCertFile", "", "Path to file with TLS certificate if -http.tls is set. Prefer ECDSA certs instead of RSA certs as RSA certs are slower. The provided certificate file is automatically re-read every second, so it can be dynamically updated")
	flag.StringVar(&flags.HTTPTLSKeyFile, "http.tlsKeyFile", "", "Path to file with TLS key if -http.tls is set. The provided key file is automatically re-read every second, so it can be dynamically updated")
	flag.StringVar(&flags.HTTPTLSCipherSuitesString, "http.tlsCipherSuites", "", "Optional list of TLS cipher suites for incoming requests over HTTPS if -http.tls is set. split by comma. See the list of supported cipher suites at https://pkg.go.dev/crypto/tls#pkg-constants")
	flag.StringVar(&flags.HTTPTLSMinVersion, "http.tlsMinVersion", "", "Optional minimum TLS version to use for incoming requests over HTTPS if -http.tls is set. "+
		"Supported values: TLS10, TLS11, TLS12, TLS13")
	flag.DurationVar(&flags.HTTPMaxGracefulShutdownDuration, "http.maxGracefulShutdownDuration", 7*time.Second, `The maximum duration for a graceful shutdown of the HTTP server. A highly loaded server may require increased value for a graceful shutdown`)

	pair := strings.Split(flags.HTTPListen, ":")
	if len(pair) == 1 {
		log.Fatalf("invalid http.listen address: %s", flags.HTTPListen)
	}

	port := pair[len(pair)-1]
	if p, e := strconv.ParseInt(port, 10, 64); e != nil || p < 1 || p > 65535 {
		log.Fatalf("invalid http.listen port: %s", port)
	} else {
		flags.HTTPPort = int(p)
	}

	if flags.HTTPTLSCipherSuitesString != "" {
		flags.HTTPTLSCipherSuitesArray = strings.Fields(strings.ReplaceAll(flags.HTTPTLSCipherSuitesString, ",", " "))
	}
}

type HTTPRouter struct {
	engine *gin.Engine
}

func Router() *HTTPRouter {
	gin.SetMode(flags.HTTPMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(ginx.BombRecovery())

	if flags.HTTPUsername != "" && flags.HTTPPassword != "" {
		r.Use(gin.BasicAuth(gin.Accounts{
			flags.HTTPUsername: flags.HTTPPassword,
		}))
	}

	if flags.HTTPPProf {
		pprof.Register(r, "/debug/pprof")
	}

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	r.GET("/pid", func(c *gin.Context) {
		c.String(200, fmt.Sprintf("%d", os.Getpid()))
	})

	r.GET("/ppid", func(c *gin.Context) {
		c.String(200, fmt.Sprintf("%d", os.Getppid()))
	})

	r.GET("/remoteaddr", func(c *gin.Context) {
		c.String(200, c.Request.RemoteAddr)
	})

	r.GET("/version", func(c *gin.Context) {
		c.String(200, buildinfo.Version)
	})

	return &HTTPRouter{engine: r}
}

// Init initializes http server and return close function
func (r *HTTPRouter) Start() func() error {
	server := &http.Server{
		Handler:           r.engine,
		IdleTimeout:       flags.HTTPIdleTimeout,
		ReadHeaderTimeout: flags.HTTPReadHeaderTimeout,
		ErrorLog:          logger.StdErrorLogger(),

		ConnContext: func(ctx context.Context, c net.Conn) context.Context {
			timeoutSec := flags.HTTPConnTimeout.Seconds()
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
		if flags.HTTPTLSEnable {
			tc, err := httptls.GetServerTLSConfig(flags.HTTPTLSCertFile, flags.HTTPTLSKeyFile, flags.HTTPTLSMinVersion, flags.HTTPTLSCipherSuitesArray)
			if err != nil {
				logger.Fatalf("cannot get TLS config for http server: %s", err)
			}
			tlsConfig = tc
		}

		listner, err := net.Listen("tcp", flags.HTTPListen)
		if err != nil {
			logger.Fatalf("cannot listen %q: %s", flags.HTTPListen, err)
		}

		if tlsConfig != nil {
			listner = tls.NewListener(listner, tlsConfig)
		}

		if err := server.Serve(listner); err != nil {
			if err == http.ErrServerClosed {
				// The server gracefully closed.
				return
			}
			logger.Fatalf("cannot serve http at %s: %s", flags.HTTPListen, err)
		}
	}()

	return func() error {
		ctx, cancel := context.WithTimeout(context.Background(), flags.HTTPMaxGracefulShutdownDuration)
		defer cancel()
		return server.Shutdown(ctx)
	}
}
