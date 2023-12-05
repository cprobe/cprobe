## 改造

本插件是把 [blackbox_exporter](https://github.com/prometheus/blackbox_exporter) 集成了进来，更多使用方法请参考原项目的文档。改造的点主要有：

- 统一化日志打印库，和 cprobe 主程序使用同一个日志库，方便日志的统一化
- 把配置文件做了切分管理，rule.d 下就是采集规则文件，不同的 job 引用不同的 rule 文件

## 仪表盘

可以复用 blackbox_exporter 的仪表盘，比如 [这个](https://grafana.com/grafana/dashboards/7587-prometheus-blackbox-exporter/)。

## 告警规则

```
# 连通性失败
probe_success == 0

# 证书将在半个月内到期
(probe_ssl_earliest_cert_expiry - time())/86400 < 15
```

## 声明

cprobe 是一个缝合怪，类似 grafana-agent，相当于集成了众多 exporter 为一个二进制。本插件并没有其他文档，如果上面的信息不足以帮到你，你可能需要自行阅读源码了。当然，并非所有人都有能力阅读源码，所以欢迎大家提 PR 一起完善这个文档，这才是开源的正确协作模式。

