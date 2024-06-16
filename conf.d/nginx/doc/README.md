This exporter is referring by [https://github.com/nginxinc/nginx-prometheus-exporter](https://github.com/nginxinc/nginx-prometheus-exporter).

## Prerequisites

We assume that you have already installed Prometheus and NGINX or NGINX Plus. Additionally, you need to:

- Expose the built-in metrics in NGINX/NGINX Plus:
    - For NGINX, expose the [stub_status
      page](https://nginx.org/en/docs/http/ngx_http_stub_status_module.html#stub_status) at `/stub_status` on port `8080`.
    - For NGINX Plus, expose the [API](https://nginx.org/en/docs/http/ngx_http_api_module.html#api) at `/api` on port
      `8080`.

# Dashboard

- [Grafana](./dash/nginx.json)
