# 📊 Service Metrics Documentation (Prometheus + Grafana)

This document describes the Prometheus metrics defined in the project and how to visualize them in Grafana.
It helps developers and operators quickly understand system performance, usage patterns, and error trends.

---

## 🧩 1. Overview of Metrics

The `metrics` package registers the following Prometheus metrics to track API, HTTP, App, and MCP (internal service) request **counts** and **latency**.

| Metric Name                      | Type      | Labels                    | Description                      |
| -------------------------------- | --------- | ------------------------- | -------------------------------- |
| `api_request_total`              | Counter   | `model`                   | Total API requests per model     |
| `api_request_duration_seconds`   | Histogram | `model`                   | API request latency distribution |
| `app_request_total`              | Counter   | `app`                     | Total requests per App           |
| `http_request_total`             | Counter   | `path`                    | Total HTTP requests per path     |
| `http_response_total`            | Counter   | `path`, `code`            | HTTP response count by status    |
| `http_response_duration_seconds` | Histogram | `path`, `code`            | HTTP response latency            |
| `mcp_request_total`              | Counter   | `mcp_service`, `mcp_func` | Total internal MCP calls         |
| `mcp_request_duration_seconds`   | Histogram | `mcp_service`, `mcp_func` | MCP call latency distribution    |

---

## 🧠 2. Metric Details

### 1️⃣ `api_request_total`

* **Type:** CounterVec
* **Labels:** `model`
* **Description:** Counts API calls per model.
* **Use Cases:**

    * Monitor usage frequency per model
    * Identify high-load models
    * Analyze user behavior
* **Example Query:**

  ```promql
  sum(api_request_total) by (model)
  ```

---

### 2️⃣ `api_request_duration_seconds`

* **Type:** HistogramVec
* **Labels:** `model`
* **Description:** Measures API request latency.
* **Use Cases:**

    * Analyze model performance
    * Track average latency and percentiles (P95, P99)
* **Example Queries:**

  ```promql
  rate(api_request_duration_seconds_sum[5m]) / rate(api_request_duration_seconds_count[5m])
  ```

  ```promql
  histogram_quantile(0.95, sum(rate(api_request_duration_seconds_bucket[5m])) by (le, model))
  ```

---

### 3️⃣ `app_request_total`

* **Type:** CounterVec
* **Labels:** `app`
* **Description:** Counts total requests from each application.
* **Use Cases:**

    * Analyze client-side traffic (iOS / Android / Web)
    * Identify most used platforms
* **Example Query:**

  ```promql
  sum(app_request_total) by (app)
  ```

---

### 4️⃣ `http_request_total`

* **Type:** CounterVec
* **Labels:** `path`
* **Description:** Tracks total HTTP requests per endpoint.
* **Use Cases:**

    * Identify frequently accessed endpoints
    * Detect abnormal traffic increases
* **Example Query:**

  ```promql
  sum(increase(http_request_total[5m])) by (path)
  ```

---

### 5️⃣ `http_response_total`

* **Type:** CounterVec
* **Labels:** `path`, `code`
* **Description:** Records response status codes per path.
* **Use Cases:**

    * Compute error rate
    * Identify endpoints returning frequent 5xx/4xx errors
* **Example Queries:**

  ```promql
  sum(http_response_total) by (code)
  ```

  **Error Rate:**

  ```promql
  sum(increase(http_response_total{code=~"5.."}[5m])) / sum(increase(http_response_total[5m]))
  ```

---

### 6️⃣ `http_response_duration_seconds`

* **Type:** HistogramVec
* **Labels:** `path`, `code`
* **Description:** Records HTTP response time distribution.
* **Use Cases:**

    * Identify performance bottlenecks
    * Monitor response time trends
* **Example Queries:**

  ```promql
  rate(http_response_duration_seconds_sum[5m]) / rate(http_response_duration_seconds_count[5m])
  ```

  ```promql
  histogram_quantile(0.95, sum(rate(http_response_duration_seconds_bucket[5m])) by (le, path))
  ```

---

### 7️⃣ `mcp_request_total`

* **Type:** CounterVec
* **Labels:** `mcp_service`, `mcp_func`
* **Description:** Tracks total internal MCP service calls.
* **Use Cases:**

    * Monitor internal RPC call frequency
    * Detect abnormal or heavy inter-service traffic
* **Example Query:**

  ```promql
  sum(increase(mcp_request_total[5m])) by (mcp_service, mcp_func)
  ```

---

### 8️⃣ `mcp_request_duration_seconds`

* **Type:** HistogramVec
* **Labels:** `mcp_service`, `mcp_func`
* **Description:** Measures MCP call latency distribution.
* **Use Cases:**

    * Monitor internal service performance
    * Identify latency spikes in specific calls
* **Example Queries:**

  ```promql
  rate(mcp_request_duration_seconds_sum[5m]) / rate(mcp_request_duration_seconds_count[5m])
  ```

  ```promql
  histogram_quantile(0.95, sum(rate(mcp_request_duration_seconds_bucket[5m])) by (le, mcp_service, mcp_func))
  ```

---

## ⚙️ 3. Grafana Integration Guide

### 1️⃣ Export Dashboard File

A preconfigured Grafana dashboard is available:

```
./conf/grafana/metrics_dashboard.json
```

It includes visualizations for:

* API call count and latency
* HTTP traffic and error rate
* App-level traffic
* MCP internal performance metrics      
![image](https://github.com/user-attachments/assets/536daa93-34ee-4a95-b13f-1ba66a0f87a4)
![image](https://github.com/user-attachments/assets/96b000a2-550b-4fe8-910b-1bd825ff667c)


---

### 2️⃣ Import Steps

1. Log in to Grafana
2. Navigate to **Dashboards → New → Import**
3. Upload the file `metrics_dashboard.json`
4. Choose your **Prometheus data source**
5. Click **Import**
6. Once loaded, you can view panels such as:

| Panel Title                    | Data Source                      |
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

### 3️⃣ Optional Dashboard Variables

You can define Grafana variables for dynamic filtering:

| Variable Name | Source     | Query Example                                  |
| ------------- | ---------- | ---------------------------------------------- |
| `model`       | Prometheus | `label_values(api_request_total, model)`       |
| `path`        | Prometheus | `label_values(http_request_total, path)`       |
| `app`         | Prometheus | `label_values(app_request_total, app)`         |
| `mcp_service` | Prometheus | `label_values(mcp_request_total, mcp_service)` |

---

## 🚀 4. Recommended Combined Metrics

| Metric Name     | PromQL Expression                                                                               | Meaning                      |
| --------------- | ----------------------------------------------------------------------------------------------- | ---------------------------- |
| System QPS      | `sum(rate(http_request_total[1m]))`                                                             | Requests per second          |
| Average Latency | `rate(http_response_duration_seconds_sum[5m]) / rate(http_response_duration_seconds_count[5m])` | Average response time        |
| Error Rate      | `sum(rate(http_response_total{code=~"5.."}[5m])) / sum(rate(http_response_total[5m]))`          | 5xx error ratio              |
| MCP Avg Latency | `rate(mcp_request_duration_seconds_sum[5m]) / rate(mcp_request_duration_seconds_count[5m])`     | Average internal RPC latency |

