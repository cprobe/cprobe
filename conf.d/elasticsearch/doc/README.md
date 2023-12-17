## 说明

ElasticSearch 这个插件 fork 自 [elasticsearch_exporter](https://github.com/prometheus-community/elasticsearch_exporter)，所有的指标名称尽量保持一致。更多使用方法请参考原项目的文档。

## 改造

- 统一化日志打印库，和 cprobe 主程序使用同一个日志库，方便日志的统一化
- 把命令行参数、环境变量参数、配置文件参数统一化
- 原来获取 cluster-name 的代码根据 cprobe 的场景做了重构

## 配置

- 如果集群不大，比如小于10个节点，建议 `gather_node = "*"`，只需要连 master 节点，即可采集所有其他节点的数据
- 如果集群比较大，比如大于10个节点，建议 `gather_node = "_local"`，即每个节点分别去发请求采集数据，此时 main.yaml 中就不能只配置 master 节点的地址，而是需要配置所有节点的地址
- 配置文件中的各类 gather 配置建议维持默认，gather_indices gather_indices_shards 等配置都关闭，如果开启，就意味着要采集每个索引的监控指标，指标量会非常非常大

## 仪表盘

- [Grafana 仪表盘](./dash/grafana_elasticsearch_01.json)

## 告警规则

```
# 集群状态是 red 了，主分片都有问题了
elasticsearch_cluster_health_status{status="red"} == 1

# heap 内存使用率超过 90%
elasticsearch_jvm_memory_used_bytes{area="heap"} / elasticsearch_jvm_memory_max_bytes{area="heap"} * 100 > 90

# unassigned shards 数量大于 0
elasticsearch_cluster_health_unassigned_shards > 0

# 其他规则 TODO
```

## 另外

本插件并没有其他文档，如果上面的信息不足以帮到你，你可能需要自行阅读源码了。当然，并非所有人都有能力阅读源码，所以欢迎大家提 PR 一起完善这个文档，这才是开源的正确协作模式。
