package whois

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/araddon/dateparse"
	"github.com/domainr/whois"

	"github.com/cprobe/cprobe/plugins"
	"github.com/cprobe/cprobe/types"
)

var (
	expiryRegex = regexp.MustCompile(`(?i)(\[有効期限]|Registry Expiry Date|paid-till|Expiration Date|Expiration Time|Expiry.*|expires.*|expire-date)[?:|\s][ \t](.*)`)
)

type Whois struct {
}

func init() {
	plugins.RegisterPlugin(types.PluginWhois, &Whois{})
}
func (wh *Whois) ParseConfig(baseDir string, bs []byte) (any, error) {

	return nil, nil
}

func (wh *Whois) Scrape(ctx context.Context, target string, cfg any, ss *types.Samples) error {
	req, err := whois.NewRequest(target)
	if err != nil {
		return err
	}

	res, err := whois.DefaultClient.Fetch(req)
	if err != nil {
		return err
	}

	date, err := parse(target, res.Body)
	if err != nil {
		return err
	}

	ss.AddMetric("whois", map[string]interface{}{
		"domain_expiration": date})
	return nil
}

func parse(host string, res []byte) (float64, error) {
	results := expiryRegex.FindStringSubmatch(string(res))
	if len(results) < 1 {
		return -2, fmt.Errorf("parse domain: %s err", host)
	}

	if parsedTime, err := dateparse.ParseAny(strings.TrimSpace(results[2])); err == nil {
		return float64(parsedTime.Unix()), nil
	}

	return -1, fmt.Errorf("Unable to parse date: %s, for %s\n", strings.TrimSpace(results[2]), host)
}
