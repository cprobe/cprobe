package writer

import (
	"net/http"
	"strings"
	"time"

	"github.com/cprobe/cprobe/lib/logger"
)

func (w *Writer) StartSender() {
	semaphone := make(chan struct{}, w.Concurrency)

	for {
		rs := w.RequestQueue.PopBackN(1)
		if len(rs) == 0 {
			time.Sleep(time.Millisecond * 300)
			continue
		}

		semaphone <- struct{}{}
		go func(req *http.Request) {
			defer func() {
				<-semaphone
			}()

			w.send(req)
		}(rs[0])
	}
}

func (w *Writer) send(req *http.Request) {
	for i := 0; i < w.RetryTimes; i++ {
		res, err := w.Client.Do(req)
		if err == nil {
			if res.StatusCode/100 != 2 {
				logger.Errorf("unexpected status code %d from %q", res.StatusCode, req.URL)
				return
			}
			return
		}

		if strings.Contains(err.Error(), "onnection") {
			logger.Errorf("error sending request to %q: %s, retry #%d", req.URL, err, i+1)
			time.Sleep(time.Duration(w.RequestTimeoutMillis) * time.Millisecond)
			continue
		}

		logger.Errorf("error sending request to %q: %s", req.URL, err)
	}
}
