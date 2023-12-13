## 改造

由于不同的 exporter 打印日志的方式各异，配置文件的格式各异，命令行的参数各异，有些 exporter 是一对一的设计，即一个 exporter 采集一个实例，或者即便支持一对多，不同目标实例也只能使用完全相同的 exporter 配置，最终还是决定把 json_exporter 的代码直接拷贝过来，然后进行改造。改造的点主要有：

- 统一化日志打印库，和 cprobe 主程序使用同一个日志库，方便日志的统一化
- 把命令行参数、环境变量参数、配置文件参数统一化
- 把原有配置文件做了切分管理，rule.d 下就是一个模块配置文件，不同的 job 需引用不同的 rule 文件

## 使用

```console
## SETUP CPROBE

## TEST with 'default' module
$ python3 -m http.server 8000 &
Serving HTTP on :: port 8000 (http://[::]:8000/) ...

```


## 仪表盘

- 需根据采集指标进行自定义配置

## 告警规则

- 需根据采集指标进行自定义配置

## 声明

cprobe 是一个缝合怪，类似 grafana-agent，相当于集成了众多 exporter 为一个二进制。自定义JSON 的监控集成的是 [json_exporter](https://github.com/prometheus-community/json_exporter)。更多使用方法请参考原项目的文档。

另外，本插件并没有其他文档，如果上面的信息不足以帮到你，你可能需要自行阅读源码了。当然，并非所有人都有能力阅读源码，所以欢迎大家提 PR 一起完善这个文档，这才是开源的正确协作模式。

