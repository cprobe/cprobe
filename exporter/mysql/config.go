package mysql

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/cprobe/cprobe/lib/logger"
	"github.com/go-sql-driver/mysql"
)

func (g Global) FormDSN(target string) (string, error) {
	if target == "" {
		logger.Fatalf("BUG: mysql target is blank")
	}

	config := mysql.NewConfig()
	config.User = g.User
	config.Passwd = g.Password
	config.Net = "tcp"
	if prefix := "unix://"; strings.HasPrefix(target, prefix) {
		config.Net = "unix"
		config.Addr = target[len(prefix):]
	} else {
		if _, _, err := net.SplitHostPort(target); err != nil {
			return "", fmt.Errorf("failed to parse target: %s", err)
		}
		config.Addr = target
	}

	if g.TlsInsecureSkipVerify {
		config.TLSConfig = "skip-verify"
	} else {
		config.TLSConfig = g.Tls
		if g.SslCa != "" {
			if err := g.CustomizeTLS(); err != nil {
				err = fmt.Errorf("failed to register a custom TLS configuration for mysql dsn: %w", err)
				return "", err
			}
			config.TLSConfig = "custom"
		}
	}

	return config.FormatDSN(), nil
}

func (m Global) CustomizeTLS() error {
	var tlsCfg tls.Config
	caBundle := x509.NewCertPool()
	pemCA, err := os.ReadFile(m.SslCa)
	if err != nil {
		return err
	}
	if ok := caBundle.AppendCertsFromPEM(pemCA); ok {
		tlsCfg.RootCAs = caBundle
	} else {
		return fmt.Errorf("failed parse pem-encoded CA certificates from %s", m.SslCa)
	}
	if m.SslCert != "" && m.SslKey != "" {
		certPairs := make([]tls.Certificate, 0, 1)
		keypair, err := tls.LoadX509KeyPair(m.SslCert, m.SslKey)
		if err != nil {
			return fmt.Errorf("failed to parse pem-encoded SSL cert %s or SSL key %s: %w",
				m.SslCert, m.SslKey, err)
		}
		certPairs = append(certPairs, keypair)
		tlsCfg.Certificates = certPairs
	}
	tlsCfg.InsecureSkipVerify = m.TlsInsecureSkipVerify
	mysql.RegisterTLSConfig("custom", &tlsCfg)
	return nil
}
