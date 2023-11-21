package httpd

import "time"

var (
	HTTPListen                      string
	HTTPPort                        int
	HTTPUsername                    string
	HTTPPassword                    string
	HTTPMode                        string
	HTTPPProf                       bool
	HTTPReadHeaderTimeout           time.Duration
	HTTPIdleTimeout                 time.Duration
	HTTPConnTimeout                 time.Duration
	HTTPTLSEnable                   bool
	HTTPTLSCertFile                 string
	HTTPTLSKeyFile                  string
	HTTPTLSCipherSuitesString       string
	HTTPTLSCipherSuitesArray        []string
	HTTPTLSMinVersion               string
	HTTPMaxGracefulShutdownDuration time.Duration
)
