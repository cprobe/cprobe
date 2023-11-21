package writer

import (
	"bytes"
	"net/http"
	"strings"

	"github.com/cprobe/cprobe/lib/logger"
)

func (w *Writer) NewRequest(body []byte) (*http.Request, error) {
	reqBody := bytes.NewBuffer(body)
	req, err := http.NewRequest(http.MethodPost, w.URL, reqBody)
	if err != nil {
		logger.Panicf("BUG: unexpected error from http.NewRequest(%q): %s", w.URL, err)
	}

	if w.BasicAuthUser != "" && w.BasicAuthPass != "" {
		req.SetBasicAuth(w.BasicAuthUser, w.BasicAuthPass)
	}

	for _, header := range w.Headers {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) != 2 {
			logger.Panicf("BUG: invalid header %q", header)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		req.Header.Add(key, value)
		if key == "Host" {
			req.Host = value
		}
	}

	req.Header.Set("User-Agent", "cprobe")
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("Content-Encoding", "snappy")
	req.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")

	return req, nil
}
