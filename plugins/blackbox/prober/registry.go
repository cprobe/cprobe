package prober

var (
	Probers = map[string]ProbeFn{
		"http": ProbeHTTP,
		"tcp":  ProbeTCP,
		"icmp": ProbeICMP,
		"dns":  ProbeDNS,
		"grpc": ProbeGRPC,
	}
)
