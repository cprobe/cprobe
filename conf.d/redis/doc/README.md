## 改造

cprobe 对 redis 的监控是把 [redis_exporter](https://github.com/oliver006/redis_exporter) 集成了进来，然后进行了一些改造，改造的点主要有：

- 统一化日志打印库，和 cprobe 主程序使用同一个日志库，方便日志的统一化
- 把命令行参数、环境变量参数、配置文件参数统一化，支持了配置文件切分管理
- 干掉了原来的 HTTP Server，毕竟 metrics 直接通过 remote write 方式发走了，不需要通过 `/metrics` 接口暴露了

## 仪表盘

TODO，希望社区能够贡献一些仪表盘，这个文档里直接放置 Grafana 的仪表盘 ID 或链接即可。

## 告警规则

TODO

## 声明

cprobe 是一个缝合怪，类似 grafana-agent，相当于集成了众多 exporter 为一个二进制。Redis 的监控集成的是 [redis_exporter](https://github.com/oliver006/redis_exporter)。更多使用方法请参考原项目的文档。

另外，本插件并没有其他文档，如果上面的信息不足以帮到你，你可能需要自行阅读源码了。当然，并非所有人都有能力阅读源码，所以欢迎大家提 PR 一起完善这个文档，这才是开源的正确协作模式。

