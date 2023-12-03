package clienttls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// ClientConfig represents the standard client TLS config.
type ClientConfig struct {
	TLSCA              string `yaml:"tls_ca" toml:"tls_ca"`
	TLSCert            string `yaml:"tls_cert" toml:"tls_cert"`
	TLSKey             string `yaml:"tls_key" toml:"tls_key"`
	TLSKeyPwd          string `yaml:"tls_key_pwd" toml:"tls_key_pwd"`
	InsecureSkipVerify bool   `yaml:"tls_skip_verify" toml:"tls_skip_verify"`
	ServerName         string `yaml:"tls_server_name" toml:"tls_server_name"`
	TLSMinVersion      string `yaml:"tls_min_version" toml:"tls_min_version"`
	TLSMaxVersion      string `yaml:"tls_max_version" toml:"tls_max_version"`
}

// TLSConfig returns a tls.Config, may be nil without error if TLS is not
// configured.
func (c *ClientConfig) TLSConfig() (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: c.InsecureSkipVerify,
		Renegotiation:      tls.RenegotiateNever,
	}

	if c.TLSCA != "" {
		pool, err := makeCertPool([]string{c.TLSCA})
		if err != nil {
			return nil, err
		}
		tlsConfig.RootCAs = pool
	}

	if c.TLSCert != "" && c.TLSKey != "" {
		err := loadCertificate(tlsConfig, c.TLSCert, c.TLSKey)
		if err != nil {
			return nil, err
		}
	}

	if c.ServerName != "" {
		tlsConfig.ServerName = c.ServerName
	}

	if c.TLSMinVersion == "1.0" {
		tlsConfig.MinVersion = tls.VersionTLS10
	} else if c.TLSMinVersion == "1.1" {
		tlsConfig.MinVersion = tls.VersionTLS11
	} else if c.TLSMinVersion == "1.2" {
		tlsConfig.MinVersion = tls.VersionTLS12
	} else if c.TLSMinVersion == "1.3" {
		tlsConfig.MinVersion = tls.VersionTLS13
	}

	if c.TLSMaxVersion == "1.0" {
		tlsConfig.MaxVersion = tls.VersionTLS10
	} else if c.TLSMaxVersion == "1.1" {
		tlsConfig.MaxVersion = tls.VersionTLS11
	} else if c.TLSMaxVersion == "1.2" {
		tlsConfig.MaxVersion = tls.VersionTLS12
	} else if c.TLSMaxVersion == "1.3" {
		tlsConfig.MaxVersion = tls.VersionTLS13
	}

	return tlsConfig, nil
}

func makeCertPool(certFiles []string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	for _, certFile := range certFiles {
		pem, err := os.ReadFile(certFile)
		if err != nil {
			return nil, fmt.Errorf(
				"could not read certificate %q: %v", certFile, err)
		}
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf(
				"could not parse any PEM certificates %q: %v", certFile, err)
		}
	}
	return pool, nil
}

func loadCertificate(config *tls.Config, certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf(
			"could not load keypair %s:%s: %v", certFile, keyFile, err)
	}

	config.Certificates = []tls.Certificate{cert}
	// config.BuildNameToCertificate()
	return nil
}
