package prober

import "github.com/prometheus/common/config"

var (
	// DefaultModule set default configuration for the Module
	DefaultModule = Module{
		HTTP: DefaultHTTPProbe,
		TCP:  DefaultTCPProbe,
		ICMP: DefaultICMPProbe,
		DNS:  DefaultDNSProbe,
	}

	// DefaultHTTPProbe set default value for HTTPProbe
	DefaultHTTPProbe = HTTPProbe{
		IPProtocolFallback: true,
		HTTPClientConfig:   config.DefaultHTTPClientConfig,
	}

	// DefaultGRPCProbe set default value for HTTPProbe
	DefaultGRPCProbe = GRPCProbe{
		Service:            "",
		IPProtocolFallback: true,
	}

	// DefaultTCPProbe set default value for TCPProbe
	DefaultTCPProbe = TCPProbe{
		IPProtocolFallback: true,
	}

	// DefaultICMPProbe set default value for ICMPProbe
	DefaultICMPTTL   = 64
	DefaultICMPProbe = ICMPProbe{
		IPProtocolFallback: true,
		TTL:                DefaultICMPTTL,
	}

	// DefaultDNSProbe set default value for DNSProbe
	DefaultDNSProbe = DNSProbe{
		IPProtocolFallback: true,
		Recursion:          true,
	}
)
