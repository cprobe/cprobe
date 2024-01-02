## 原理

对 Tomcat 的监控，实际是请求的 Tomcat 的状态页面，然后解析页面内容，获取到需要的信息。Tomcat 状态页面不但可以返回 HTML 格式，也可以返回 JSON 格式，这里使用 JSON 格式。比如我们要监控某个 Tomcat 的地址是：`127.0.0.1:8080`，实际 cprobe 拉取数据请求的是：`http://127.0.0.1:8080/manager/status/all?JSON=true`。

## 配置

修改 `conf/tomcat-users.xml` 配置，增加 role 和 user，比如：

```xml
<tomcat-users xmlns="http://tomcat.apache.org/xml"
              xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
              xsi:schemaLocation="http://tomcat.apache.org/xml tomcat-users.xsd"
              version="1.0">
  <role rolename="manager-gui"/>
  <user username="tomcat" password="s3cret" roles="manager-gui"/>
</tomcat-users>
```

其次，通常 cprobe 和 tomcat 部署在不同的机器上，需要修改 `webapps/manager/META-INF/context.xml` 配置，把下面的部分注释掉：

```xml
<Valve className="org.apache.catalina.valves.RemoteAddrValve"
         allow="127\.\d+\.\d+\.\d+|::1|0:0:0:0:0:0:0:1" />
```

xml 的注释使用 `<!-- -->`，所以，最终注释之后变成：

```xml
<!--
<Valve className="org.apache.catalina.valves.RemoteAddrValve"
         allow="127\.\d+\.\d+\.\d+|::1|0:0:0:0:0:0:0:1" />
-->
```

## 仪表盘

TODO

## 告警规则

TODO

## 声明

cprobe 是一个缝合怪，类似 grafana-agent，相当于集成了众多 exporter 为一个二进制。本插件并没有其他文档，如果上面的信息不足以帮到你，你可能需要自行阅读源码了。当然，并非所有人都有能力阅读源码，所以欢迎大家提 PR 一起完善这个文档，这才是开源的正确协作模式。



