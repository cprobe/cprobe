## 说明

这个插件用于获取 Prometheus 协议的数据，类似 vmagent 或者 prometheus agent mode 的抓取功能。正常来讲，是不需要这个插件的，用 vmagent 就可以了，但是有的时候，通过 vmagent 获取某些数据的时候会报这样的错误：

```
second HELP line for metric name "xx"
```

这种情况是因为源端的数据不规范，导致 vmagent 无法解析，用 promtool 来校验也会报同样的错，这个时候，可以使用这个插件来获取数据。这个插件的 rule.toml 里配置：

```toml
split_body = true
```

## 声明

cprobe 是一个缝合怪，类似 grafana-agent，相当于集成了众多 exporter 为一个二进制。本插件并没有其他文档，如果上面的信息不足以帮到你，你可能需要自行阅读源码了。当然，并非所有人都有能力阅读源码，所以欢迎大家提 PR 一起完善这个文档，这才是开源的正确协作模式。

