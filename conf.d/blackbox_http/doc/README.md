## 说明

这个插件用于黑盒方式探测远端的 HTTP 地址，可以是 http 的也可以是 https 的，采集的指标包括：

**blackbox_http_probe_result_code**

探测结果码，0 表示探测成功，非 0 表示探测失败。探测失败的情况比较多，不同的 value 代表不同的错误：

- 1：New Request 失败
- 2：Invalid Headers
- 3：发 HTTP 网络请求失败
- 4：Status code 不匹配
- 5：Body 不匹配

**blackbox_http_probe_duration_seconds**

探测耗时，单位是秒。

**blackbox_http_probe_response_code**

HTTP 远端服务器返回的状态码。如果连接失败了，这个指标不会输出，因为压根获取不到 response。

**blackbox_http_cert_expire_timestamp**

如果是 https 的地址，这个指标表示证书的过期时间，单位是秒。如果是 http 的地址，这个指标不会输出。

## 仪表盘

TODO

## 告警规则

```
blackbox_http_probe_result_code != 0
blackbox_http_cert_expire_timestamp - time() < 86400 * 30
```

## 附

本插件并没有其他文档，如果上面的信息不足以帮到你，你可能需要自行阅读源码了。当然，并非所有人都有能力阅读源码，所以欢迎大家提 PR 一起完善这个文档，这才是开源的正确协作模式。

