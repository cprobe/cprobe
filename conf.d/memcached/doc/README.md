## 告警规则

```yaml
groups:
    - name: Memcached
      rules:
        - alert: '[Memcached] Instance Down'
          expr: |
            memcached_up == 0
          for: 5m
          labels:
            severity: critical
          annotations:
            description: Instance is not reachable
        - alert: '[Memcached] Low UpTime'
          expr: |
            memcached_uptime_seconds < 3600
          for: 5m
          labels:
            severity: critical
          annotations:
            description: Uptime of less than 1 hour in a Memcached instance
        - alert: '[Memcached] Connection Throttled'
          expr: |
            rate(memcached_connections_yielded_total[5m]) > 5
          for: 10m
          labels:
            severity: critical
          annotations:
            description: Connection throttled because max number of requests per event process reached
        - alert: '[Memcached] Connections Close To The Limit 85%'
          expr: |
            memcached_current_connections/ memcached_max_connections > 0.85
          for: 5m
          labels:
            severity: warning
          annotations:
            description: The mumber of connections are close to the limit
        - alert: '[Memcached] Connections Limit Reached'
          expr: |
            rate(memcached_connections_listener_disabled_total[5m]) > 0
          for: 5m
          labels:
            severity: critical
          annotations:
            description: Reached the number of maximum connections and caused a connection error
```

## 仪表盘

- [https://grafana.com/grafana/dashboards/37-prometheus-memcached/](https://grafana.com/grafana/dashboards/37-prometheus-memcached/)
- [https://grafana.com/grafana/dashboards/11527-memcached-overview/](https://grafana.com/grafana/dashboards/11527-memcached-overview/)

## 声明

cprobe 是一个缝合怪，类似 grafana-agent，相当于集成了众多 exporter 为一个二进制。 Memcached 的监控集成的是 [memcached_exporter](https://github.com/prometheus/memcached_exporter)。更多使用方法请参考原项目的文档。

另外，本插件并没有其他文档，如果上面的信息不足以帮到你，你可能需要自行阅读源码了。当然，并非所有人都有能力阅读源码，所以欢迎大家提 PR 一起完善这个文档，这才是开源的正确协作模式。
