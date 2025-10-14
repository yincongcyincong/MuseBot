# 📊 服务监控指标说明文档 (Prometheus + Grafana)

本文档介绍项目中定义的 Prometheus 指标及其在 Grafana 中的使用方式，帮助开发者与运维人员快速了解系统运行状况、性能瓶颈与错误趋势。

---

## 🧩 一、指标概览

项目通过 `metrics` 包注册以下监控指标，用于跟踪 API、HTTP、App 和 MCP（内部MCP调用）请求的 **次数** 与 **耗时**。

| 指标名称                             | 类型        | 标签                        | 说明              |
| -------------------------------- | --------- | ------------------------- | --------------- |
| `api_request_total`              | Counter   | `model`                   | 各模型的 API 调用总次数  |
| `api_request_duration_seconds`   | Histogram | `model`                   | 各模型 API 请求耗时分布  |
| `app_request_total`              | Counter   | `app`                     | 各 App 的请求总次数    |
| `http_request_total`             | Counter   | `path`                    | 各 HTTP 路径的请求总次数 |
| `http_response_total`            | Counter   | `path`, `code`            | 各路径返回状态码统计      |
| `http_response_duration_seconds` | Histogram | `path`, `code`            | HTTP 响应耗时分布     |
| `mcp_request_total`              | Counter   | `mcp_service`, `mcp_func` | 各MCP调用的请求总次数    |
| `mcp_request_duration_seconds`   | Histogram | `mcp_service`, `mcp_func` | 各MCP调用的耗时分布     |

---

## 🧠 二、指标详细说明

### 1️⃣ `api_request_total`

* **类型**：CounterVec
* **标签**：`model`
* **说明**：统计每个模型的 API 请求总数。
* **典型用途**：

    * 监控不同模型的使用频率；
    * 判断高负载模型；
    * 分析用户偏好。
* **示例查询**：

  ```promql
  sum(api_request_total) by (model)
  ```

---

### 2️⃣ `api_request_duration_seconds`

* **类型**：HistogramVec
* **标签**：`model`
* **说明**：记录 API 请求的响应时间。
* **典型用途**：

    * 分析模型响应性能；
    * 监控平均耗时和分位数（P95、P99）。
* **示例查询**：

  ```promql
  rate(api_request_duration_seconds_sum[5m]) / rate(api_request_duration_seconds_count[5m])
  ```

  ```promql
  histogram_quantile(0.95, sum(rate(api_request_duration_seconds_bucket[5m])) by (le, model))
  ```

---

### 3️⃣ `app_request_total`

* **类型**：CounterVec
* **标签**：`app`
* **说明**：统计每个 App 的请求数量。
* **典型用途**：

    * 分析不同客户端的访问量；
    * 判断主流使用端（如 iOS / Android / Web）。
* **示例查询**：

  ```promql
  sum(app_request_total) by (app)
  ```

---

### 4️⃣ `http_request_total`

* **类型**：CounterVec
* **标签**：`path`
* **说明**：记录 HTTP 请求的总数。
* **典型用途**：

    * 统计各接口访问频率；
    * 识别热点接口或异常调用增长。
* **示例查询**：

  ```promql
  sum(increase(http_request_total[5m])) by (path)
  ```

---

### 5️⃣ `http_response_total`

* **类型**：CounterVec
* **标签**：`path`, `code`
* **说明**：统计每个接口的响应状态码。
* **典型用途**：

    * 计算错误率；
    * 检测异常接口（5xx、4xx）。
* **示例查询**：

  ```promql
  sum(http_response_total) by (code)
  ```

  **错误率计算：**

  ```promql
  sum(increase(http_response_total{code=~"5.."}[5m])) / sum(increase(http_response_total[5m]))
  ```

---

### 6️⃣ `http_response_duration_seconds`

* **类型**：HistogramVec
* **标签**：`path`, `code`
* **说明**：记录 HTTP 请求响应的耗时分布。
* **典型用途**：

    * 分析系统性能瓶颈；
    * 监控接口延迟。
* **示例查询**：

  ```promql
  rate(http_response_duration_seconds_sum[5m]) / rate(http_response_duration_seconds_count[5m])
  ```

  ```promql
  histogram_quantile(0.95, sum(rate(http_response_duration_seconds_bucket[5m])) by (le, path))
  ```

---

### 7️⃣ `mcp_request_total`

* **类型**：CounterVec
* **标签**：`mcp_service`, `mcp_func`
* **说明**：记录（MCP）间调用的请求次数。
* **典型用途**：

    * 监控内部 RPC 调用频率；
    * 识别高频或异常服务调用。
* **示例查询**：

  ```promql
  sum(increase(mcp_request_total[5m])) by (mcp_service, mcp_func)
  ```

---

### 8️⃣ `mcp_request_duration_seconds`

* **类型**：HistogramVec
* **标签**：`mcp_service`, `mcp_func`
* **说明**：统计各MCP调用的响应耗时。
* **典型用途**：

    * 检查内部调用延迟；
    * 监控跨服务性能。
* **示例查询**：

  ```promql
  rate(mcp_request_duration_seconds_sum[5m]) / rate(mcp_request_duration_seconds_count[5m])
  ```

  ```promql
  histogram_quantile(0.95, sum(rate(mcp_request_duration_seconds_bucket[5m])) by (le, mcp_service, mcp_func))
  ```

---

## ⚙️ 三、Grafana 导入与使用指南

### 1️⃣ 导出 Dashboard 文件

系统已提供一个 Grafana Dashboard 文件：

```
./conf/grafana/metrics_dashboard.json
```

包含所有上述指标的图表展示，包括：

* API 调用次数与耗时；
* HTTP 请求与错误率；
* App 请求统计；
* MCP 内部调用性能。

---

### 2️⃣ Grafana 导入步骤

1. 登录 Grafana 控制台
2. 点击左侧菜单：**Dashboards → New → Import**
3. 在 “Import via file” 处上传 `metrics_dashboard.json` 文件
4. 在 “Prometheus data source” 中选择你的 Prometheus 数据源
5. 点击 **Import**
6. 等待面板加载完成，即可查看以下图表：

| 图表标题                           | 数据来源                             |
| ------------------------------ | -------------------------------- |
| Total API Requests (by model)  | `api_request_total`              |
| API Request Duration (seconds) | `api_request_duration_seconds`   |
| Total App Requests             | `app_request_total`              |
| HTTP Requests per Path         | `http_request_total`             |
| HTTP Responses by Status Code  | `http_response_total`            |
| HTTP Response Duration         | `http_response_duration_seconds` |
| MCP Request Count              | `mcp_request_total`              |
| MCP Request Duration           | `mcp_request_duration_seconds`   |

---

### 3️⃣ 可选的 Dashboard 模板变量

可在 Grafana 面板中添加以下模板变量以支持快速切换：

| 变量名           | 数据源        | 查询语句                                           |
| ------------- | ---------- | ---------------------------------------------- |
| `model`       | Prometheus | `label_values(api_request_total, model)`       |
| `path`        | Prometheus | `label_values(http_request_total, path)`       |
| `app`         | Prometheus | `label_values(app_request_total, app)`         |
| `mcp_service` | Prometheus | `label_values(mcp_request_total, mcp_service)` |

---

## 🚀 四、推荐的综合监控指标

| 指标名称     | PromQL                                                                                          | 含义          |
| -------- | ----------------------------------------------------------------------------------------------- |-------------|
| 系统 QPS   | `sum(rate(http_request_total[1m]))`                                                             | 每秒请求速率      |
| 平均响应时间   | `rate(http_response_duration_seconds_sum[5m]) / rate(http_response_duration_seconds_count[5m])` | 系统平均耗时      |
| 错误率      | `sum(rate(http_response_total{code=~"5.."}[5m])) / sum(rate(http_response_total[5m]))`          | 5xx 错误占比    |
| MCP 平均耗时 | `rate(mcp_request_duration_seconds_sum[5m]) / rate(mcp_request_duration_seconds_count[5m])`     | mcp服务调用平均耗时 |

