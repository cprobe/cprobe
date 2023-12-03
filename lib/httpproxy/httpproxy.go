package httpproxy

import (
	"fmt"
	"net/http"
	"net/url"
)

type proxyFunc func(req *http.Request) (*url.URL, error)

func GetProxyFunc(proxyURL string) (proxyFunc, error) {
	if len(proxyURL) > 0 {
		address, err := url.Parse(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("error parsing proxy url %q: %w", proxyURL, err)
		}
		return http.ProxyURL(address), nil
	}
	return http.ProxyFromEnvironment, nil
}
