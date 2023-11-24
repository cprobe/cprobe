package probe

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cprobe/cprobe/lib/logger"
	"github.com/cprobe/cprobe/lib/promutils"
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

	targets := j.getTargets()
	for _, target := range targets {
		fmt.Println("target:", target)
	}
}

func (j *JobGoroutine) Stop() {
	close(j.quitChan)
}

func (j *JobGoroutine) getTargets() (targets []*promutils.Labels) {
	for _, c := range j.scrapeConfig.HTTPSDConfigs {
		arr, err := c.GetLabels(j.scrapeConfig.ConfigRef.BaseDir)
		if err != nil {
			logger.Errorf("job(%s) http_sd_configs(%s) get targets error: %s", j.scrapeConfig.JobName, c.URL, err)
			continue
		}
		targets = append(targets, arr...)
	}

	// TODO: 下面的代码是 copilot 自动生成的，尚未验证过，对于 cprobe 而言，核心就是 static、file_sd、http_sd 基本就够用了

	for _, c := range j.scrapeConfig.DNSSDConfigs {
		arr, err := c.GetLabels(j.scrapeConfig.ConfigRef.BaseDir)
		if err != nil {
			logger.Errorf("job(%s) dns_sd_configs(%s) get targets error: %s", j.scrapeConfig.JobName, c.Names, err)
			continue
		}
		targets = append(targets, arr...)
	}

	for _, c := range j.scrapeConfig.AzureSDConfigs {
		arr, err := c.GetLabels(j.scrapeConfig.ConfigRef.BaseDir)
		if err != nil {
			logger.Errorf("job(%s) azure_sd_configs(%s) get targets error: %s", j.scrapeConfig.JobName, c.SubscriptionID, err)
			continue
		}
		targets = append(targets, arr...)
	}

	for _, c := range j.scrapeConfig.DockerSDConfigs {
		arr, err := c.GetLabels(j.scrapeConfig.ConfigRef.BaseDir)
		if err != nil {
			logger.Errorf("job(%s) docker_sd_configs(%s) get targets error: %s", j.scrapeConfig.JobName, c.Host, err)
			continue
		}
		targets = append(targets, arr...)
	}

	for _, c := range j.scrapeConfig.DockerSwarmSDConfigs {
		arr, err := c.GetLabels(j.scrapeConfig.ConfigRef.BaseDir)
		if err != nil {
			logger.Errorf("job(%s) dockerswarm_sd_configs(%s) get targets error: %s", j.scrapeConfig.JobName, c.Host, err)
			continue
		}
		targets = append(targets, arr...)
	}

	for _, c := range j.scrapeConfig.EC2SDConfigs {
		arr, err := c.GetLabels(j.scrapeConfig.ConfigRef.BaseDir)
		if err != nil {
			logger.Errorf("job(%s) ec2_sd_configs(%s) get targets error: %s", j.scrapeConfig.JobName, c.Region, err)
			continue
		}
		targets = append(targets, arr...)
	}

	for _, c := range j.scrapeConfig.EurekaSDConfigs {
		arr, err := c.GetLabels(j.scrapeConfig.ConfigRef.BaseDir)
		if err != nil {
			logger.Errorf("job(%s) eureka_sd_configs(%s) get targets error: %s", j.scrapeConfig.JobName, c.Server, err)
			continue
		}
		targets = append(targets, arr...)
	}

	for _, c := range j.scrapeConfig.GCESDConfigs {
		arr, err := c.GetLabels(j.scrapeConfig.ConfigRef.BaseDir)
		if err != nil {
			logger.Errorf("job(%s) gce_sd_configs(%s) get targets error: %s", j.scrapeConfig.JobName, c.Project, err)
			continue
		}
		targets = append(targets, arr...)
	}

	for _, c := range j.scrapeConfig.DigitaloceanSDConfigs {
		arr, err := c.GetLabels(j.scrapeConfig.ConfigRef.BaseDir)
		if err != nil {
			logger.Errorf("job(%s) digitalocean_sd_configs(%s:%d) get targets error: %s", j.scrapeConfig.JobName, c.Server, c.Port, err)
			continue
		}
		targets = append(targets, arr...)
	}

	for _, c := range j.scrapeConfig.OpenStackSDConfigs {
		arr, err := c.GetLabels(j.scrapeConfig.ConfigRef.BaseDir)
		if err != nil {
			logger.Errorf("job(%s) openstack_sd_configs(%s) get targets error: %s", j.scrapeConfig.JobName, c.IdentityEndpoint, err)
			continue
		}
		targets = append(targets, arr...)
	}

	for _, c := range j.scrapeConfig.YandexCloudSDConfigs {
		arr, err := c.GetLabels(j.scrapeConfig.ConfigRef.BaseDir)
		if err != nil {
			logger.Errorf("job(%s) yandexcloud_sd_configs(%s) get targets error: %s", j.scrapeConfig.JobName, c.APIEndpoint, err)
			continue
		}
		targets = append(targets, arr...)
	}

	return
}
