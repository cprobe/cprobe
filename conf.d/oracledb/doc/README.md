## 授权

Oracle 数据库监控，需要创建一个监控用户，并授予该用户一些权限，举例如下：

```sql
-- Create the monitoring user "cprobe"
CREATE USER cprobe IDENTIFIED BY <YOUR-PASSWORD>;

-- Grant the "cprobe" user the required permissions
GRANT CONNECT TO cprobe;
GRANT SELECT ON SYS.GV_$RESOURCE_LIMIT to cprobe;
GRANT SELECT ON SYS.V_$SESSION to cprobe;
GRANT SELECT ON SYS.V_$WAITCLASSMETRIC to cprobe;
GRANT SELECT ON SYS.GV_$PROCESS to cprobe;
GRANT SELECT ON SYS.GV_$SYSSTAT to cprobe;
GRANT SELECT ON SYS.V_$DATAFILE to cprobe;
GRANT SELECT ON SYS.V_$ASM_DISKGROUP_STAT to cprobe;
GRANT SELECT ON SYS.V_$SYSTEM_WAIT_CLASS to cprobe;
GRANT SELECT ON SYS.DBA_TABLESPACE_USAGE_METRICS to cprobe;
GRANT SELECT ON SYS.DBA_TABLESPACES to cprobe;
GRANT SELECT ON SYS.GLOBAL_NAME to cprobe;
```

## 配置

具体监控哪些内容，就看你的配置文件中有哪些 `[[queries]]` 配置段了，下面是一个例子：

```toml
[[queries]]
mesurement = "sessions"
label_fields = [ "status", "type" ]
value_fields = [ "value" ]
request = "SELECT status, type, COUNT(*) as value FROM v$session GROUP BY status, type"
```

每个 `[[queries]]` 配置段包含以下几个属性：

- `mesurement`：指标名称前缀
- `value_fields`：SQL 会查到多个字段，这里指定哪些字段作为指标输出，对应的字段的字段名作为指标名称后缀，字段值作为指标值
- `label_fields`：SQL 会查到多个字段，这里指定哪些字段作为标签输出
- `metric_name_field`：SQL 会查到多个字段，这里指定哪个字段作为指标名称
- `timeout`：SQL 执行超时时间
- `request`：SQL 语句

## 仪表盘

Oracle 的监控，每个语句都是自定义的，只提供了一个最简单的仪表盘，仅供参考。[在这里](./dash/grafana_oracledb_01.json)。欢迎对 Oracle 有深入了解的朋友提 PR 完善仪表盘。

## 声明

cprobe 是一个缝合怪，类似 grafana-agent，相当于集成了众多 exporter 为一个二进制。Oracle 的监控集成的是 [oracledb_exporter](https://github.com/iamseth/oracledb_exporter)。更多使用方法请参考原项目的文档。

另外，本插件并没有其他文档，如果上面的信息不足以帮到你，你可能需要自行阅读源码了。当然，并非所有人都有能力阅读源码，所以欢迎大家提 PR 一起完善这个文档，这才是开源的正确协作模式。
