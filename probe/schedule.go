package probe

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cprobe/cprobe/lib/envtemplate"
	"github.com/cprobe/cprobe/lib/fs"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/cprobe/cprobe/lib/promutils"
	"gopkg.in/yaml.v2"
)

func makeJobs() map[string]map[JobID]*JobGoroutine {
	return map[string]map[JobID]*JobGoroutine{
		"mysql":         make(map[JobID]*JobGoroutine),
		"redis":         make(map[JobID]*JobGoroutine),
		"elasticsearch": make(map[JobID]*JobGoroutine),
		"postgresql":    make(map[JobID]*JobGoroutine),
		"kafka":         make(map[JobID]*JobGoroutine),
		"mongodb":       make(map[JobID]*JobGoroutine),
	}
}

var (
	Jobs = makeJobs()
)

type JobID struct {
	YamlFile string
	JobName  string
}

type JobGoroutine struct {
	plugin       string
	scrapeConfig *ScrapeConfig
	quitChan     chan struct{}
	sync.RWMutex
}

func NewJobGoroutine(plugin string, scrapeConfig *ScrapeConfig) *JobGoroutine {
	return &JobGoroutine{
		plugin:       plugin,
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
			j.run(ctx)
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
func (j *JobGoroutine) run(ctx context.Context) {
	jobName := j.GetJobName()
	targets := j.getTargets()
	for _, target := range targets {
		j.runTarget(ctx, jobName, target)
	}
}

func (j *JobGoroutine) runTarget(ctx context.Context, job string, target *promutils.Labels) {
	labels := promutils.GetLabels()
	defer promutils.PutLabels(labels)

	mergeLabels(labels, job, target, j.scrapeConfig.ConfigRef.Global.ExternalLabels)
	labels.Labels = j.scrapeConfig.ParsedRelabelConfigs.Apply(labels.Labels, 0)
	labels.RemoveMetaLabels()

	if labels.Len() == 0 {
		return
	}

	if labels.Get("__address__") == "" {
		return
	}

	labelsCopy := labels.Clone()
	labelsCopy.Sort()

	// 获取 scrape_rules，就是几个 toml 文件，拼接在一起，根据 plugin 类型获取认证信息
	// 如果是 HTTP 的 plugin，应该有个统一的认证配置方式，比如 basic auth、token 等
	// 目标实例的地址是 address，这个地址信息无需发到最终的 writer，因为有 instance 字段了

	auth := j.scrapeConfig.ScrapeAuth
	if auth == nil {
		auth = j.scrapeConfig.ConfigRef.Global.ScrapeAuth
	}

	// TODO 直接把 toml decode 成 rule 配置，传给抓取函数
	var tomlBytes []byte

	switch j.plugin {
	case "mysql":
		ScrapeMySQL(ctx, labelsCopy, auth, tomlBytes)
	case "redis":
		ScrapeRedis(ctx, labelsCopy, auth, tomlBytes)
	default:
		logger.Errorf("unknown plugin: %s", j.plugin)
	}
}

func mergeLabels(dst *promutils.Labels, job string, targetLabels, extraLabels *promutils.Labels) {
	dst.Add("job", job)

	if extraLabels != nil {
		dst.AddFrom(extraLabels)
	}

	instanceBlank := dst.Get("instance") == ""

	for _, label := range targetLabels.GetLabels() {
		dst.Add(label.Name, label.Value)
		if label.Name == "__address__" {
			if label.Value != "" && instanceBlank {
				dst.Add("instance", label.Value)
			}
		}
	}

	dst.RemoveDuplicates()
}

func (j *JobGoroutine) Stop() {
	close(j.quitChan)
}

func loadStaticConfigs(path string) ([]StaticConfig, error) {
	data, err := fs.ReadFileOrHTTP(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read `static_configs` from %q: %w", path, err)
	}
	data, err = envtemplate.ReplaceBytes(data)
	if err != nil {
		return nil, fmt.Errorf("cannot expand environment vars in %q: %w", path, err)
	}
	var stcs []StaticConfig
	if err := yaml.UnmarshalStrict(data, &stcs); err != nil {
		return nil, fmt.Errorf("cannot unmarshal `static_configs` from %q: %w", path, err)
	}
	return stcs, nil
}

func (j *JobGoroutine) getTargets() (targets []*promutils.Labels) {
	baseDir := j.scrapeConfig.ConfigRef.BaseDir

	for _, c := range j.scrapeConfig.StaticConfigs {
		for _, t := range c.Targets {
			m := promutils.NewLabels(1 + c.Labels.Len())
			m.AddFrom(c.Labels)
			m.Add("__address__", t)
			m.RemoveDuplicates()
			targets = append(targets, m)
		}
	}

	for _, c := range j.scrapeConfig.FileSDConfigs {
		for _, file := range c.Files {
			pathPattern := fs.GetFilepath(baseDir, file)
			paths := []string{pathPattern}
			if strings.Contains(pathPattern, "*") {
				var err error
				paths, err = filepath.Glob(pathPattern)
				if err != nil {
					// Do not return this error, since other files may contain valid scrape configs.
					logger.Errorf("skipping entry %q in `file_sd_config->files` for job_name=%s because of error: %s", file, j.scrapeConfig.JobName, err)
					continue
				}
			}
			for _, path := range paths {
				stcs, err := loadStaticConfigs(path)
				if err != nil {
					// Do not return this error, since other paths may contain valid scrape configs.
					logger.Errorf("skipping file %s for job_name=%s at `file_sd_configs` because of error: %s", path, j.scrapeConfig.JobName, err)
					continue
				}

				pathShort := path
				if strings.HasPrefix(pathShort, baseDir) {
					pathShort = path[len(baseDir):]
					if len(pathShort) > 0 && pathShort[0] == filepath.Separator {
						pathShort = pathShort[1:]
					}
				}

				for _, stc := range stcs {
					for _, t := range stc.Targets {
						m := promutils.NewLabels(2 + stc.Labels.Len())
						m.AddFrom(stc.Labels)
						m.Add("__address__", t)
						m.Add("__meta_filepath", pathShort)
						m.RemoveDuplicates()
						targets = append(targets, m)
					}
				}
			}
		}
	}

	for _, c := range j.scrapeConfig.HTTPSDConfigs {
		arr, err := c.GetLabels(baseDir)
		if err != nil {
			logger.Errorf("job(%s) http_sd_configs(%s) get targets error: %s", j.scrapeConfig.JobName, c.URL, err)
			continue
		}
		targets = append(targets, arr...)
	}

	// TODO: 下面的代码是 copilot 自动生成的，尚未验证过，对于 cprobe 而言，核心就是 static、file_sd、http_sd 基本就够用了

	for _, c := range j.scrapeConfig.DNSSDConfigs {
		arr, err := c.GetLabels(baseDir)
		if err != nil {
			logger.Errorf("job(%s) dns_sd_configs(%s) get targets error: %s", j.scrapeConfig.JobName, c.Names, err)
			continue
		}
		targets = append(targets, arr...)
	}

	for _, c := range j.scrapeConfig.AzureSDConfigs {
		arr, err := c.GetLabels(baseDir)
		if err != nil {
			logger.Errorf("job(%s) azure_sd_configs(%s) get targets error: %s", j.scrapeConfig.JobName, c.SubscriptionID, err)
			continue
		}
		targets = append(targets, arr...)
	}

	for _, c := range j.scrapeConfig.DockerSDConfigs {
		arr, err := c.GetLabels(baseDir)
		if err != nil {
			logger.Errorf("job(%s) docker_sd_configs(%s) get targets error: %s", j.scrapeConfig.JobName, c.Host, err)
			continue
		}
		targets = append(targets, arr...)
	}

	for _, c := range j.scrapeConfig.DockerSwarmSDConfigs {
		arr, err := c.GetLabels(baseDir)
		if err != nil {
			logger.Errorf("job(%s) dockerswarm_sd_configs(%s) get targets error: %s", j.scrapeConfig.JobName, c.Host, err)
			continue
		}
		targets = append(targets, arr...)
	}

	for _, c := range j.scrapeConfig.EC2SDConfigs {
		arr, err := c.GetLabels(baseDir)
		if err != nil {
			logger.Errorf("job(%s) ec2_sd_configs(%s) get targets error: %s", j.scrapeConfig.JobName, c.Region, err)
			continue
		}
		targets = append(targets, arr...)
	}

	for _, c := range j.scrapeConfig.EurekaSDConfigs {
		arr, err := c.GetLabels(baseDir)
		if err != nil {
			logger.Errorf("job(%s) eureka_sd_configs(%s) get targets error: %s", j.scrapeConfig.JobName, c.Server, err)
			continue
		}
		targets = append(targets, arr...)
	}

	for _, c := range j.scrapeConfig.GCESDConfigs {
		arr, err := c.GetLabels(baseDir)
		if err != nil {
			logger.Errorf("job(%s) gce_sd_configs(%s) get targets error: %s", j.scrapeConfig.JobName, c.Project, err)
			continue
		}
		targets = append(targets, arr...)
	}

	for _, c := range j.scrapeConfig.DigitaloceanSDConfigs {
		arr, err := c.GetLabels(baseDir)
		if err != nil {
			logger.Errorf("job(%s) digitalocean_sd_configs(%s:%d) get targets error: %s", j.scrapeConfig.JobName, c.Server, c.Port, err)
			continue
		}
		targets = append(targets, arr...)
	}

	for _, c := range j.scrapeConfig.OpenStackSDConfigs {
		arr, err := c.GetLabels(baseDir)
		if err != nil {
			logger.Errorf("job(%s) openstack_sd_configs(%s) get targets error: %s", j.scrapeConfig.JobName, c.IdentityEndpoint, err)
			continue
		}
		targets = append(targets, arr...)
	}

	for _, c := range j.scrapeConfig.YandexCloudSDConfigs {
		arr, err := c.GetLabels(baseDir)
		if err != nil {
			logger.Errorf("job(%s) yandexcloud_sd_configs(%s) get targets error: %s", j.scrapeConfig.JobName, c.APIEndpoint, err)
			continue
		}
		targets = append(targets, arr...)
	}

	return
}
