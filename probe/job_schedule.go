package probe

import (
	"context"
	"sync"
	"time"

	"github.com/cprobe/cprobe/lib/logger"
)

var (
	Jobs = makeJobs()
)

type JobID struct {
	YamlFile string
	JobName  string
}

type JobGoroutine struct {
	scrapeConfig *ScrapeConfig
	quitChan     chan struct{}
	sync.RWMutex
}

func NewJobGoroutine(scrapeConfig *ScrapeConfig) *JobGoroutine {
	return &JobGoroutine{
		quitChan:     make(chan struct{}),
		scrapeConfig: scrapeConfig,
	}
}

func (j *JobGoroutine) UpdateConfig(scrapeConfig *ScrapeConfig) {
	j.Lock()
	defer j.Unlock()
	j.scrapeConfig = scrapeConfig
}

func (j *JobGoroutine) GetInterval() time.Duration {
	j.RLock()
	defer j.RUnlock()
	return j.scrapeConfig.ScrapeInterval.Duration()
}

func (j *JobGoroutine) GetJobName() string {
	j.RLock()
	defer j.RUnlock()
	return j.scrapeConfig.JobName
}

func (j *JobGoroutine) Start(ctx context.Context) {
	timer := time.NewTimer(0)
	defer timer.Stop()

	var start time.Time
	for {
		select {
		case <-timer.C:
			start = time.Now()
			j.run()
			next := j.GetInterval() - time.Since(start)
			if next < 0 {
				next = 0
			}
			timer.Reset(next)
		case <-j.quitChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

// 拿到这个 job 相关的 targets + relabel + scrape_rules，按照并发度创建 goroutine 拉取数据
// 不同的 job 其 targets 获取方式和列表可能是类似的，scrape_rules 也可能是一样的，所以这里可以缓存，缓存时间可以短一点，比如 5s
// targets 可能很多，要做一下并发度控制，并发度可以在 job 粒度自定义，每个 yaml 的 global 部分也可以有一个全局的并发度配置
// 通过 wait group 等待所有的 goroutine 抓取完毕，统一做 metric_relabel_configs，然后发送给 writer
func (j *JobGoroutine) run() {
	logger.Errorf("run job %s", j.GetJobName())
	for _, rule := range j.scrapeConfig.ScrapeRuleFiles {
		logger.Errorf("    - %v", rule)
	}
}

func (j *JobGoroutine) Stop() {
	close(j.quitChan)
}
