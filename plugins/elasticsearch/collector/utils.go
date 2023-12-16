package collector

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func getURL(ctx context.Context, hc *http.Client, u string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s: %v", u, err)
	}

	if resp.Body == nil {
		return nil, fmt.Errorf("empty response body")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP Request failed with code %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return b, nil
}
