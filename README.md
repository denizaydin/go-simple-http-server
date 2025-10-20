# 🌀 Go Simple HTTP Server

A lightweight Go-based HTTP service that provides request and environment details in a JSON response.  
Designed for **Kubernetes**, **Cilium**, and **service-mesh** test environments to trace east–west calls between services.

---

## 🚀 Features

- Returns request metadata as **structured JSON**
- Automatically includes:
  - Node name (`NODE_NAME`)
  - Pod name (`POD_NAME`)
  - Hostname (`hostname`)
  - Source / destination IP
  - Full request URL
  - All **incoming HTTP headers**
- Supports **cascaded service calls** using the `CALL_SERVICE` environment variable
- Includes built-in `/healthz` endpoint for liveness / readiness probes
- IPv4 / IPv6 / dual-stack listening via `IP_MODE`
- Lightweight Alpine-based Docker image (~20 MB)

---

## ⚙️ Environment Variables

| Variable | Description | Default |
|-----------|--------------|----------|
| **PORT** | Port the server listens on | `8080` |
| **NODE_NAME** | Node name (from Kubernetes Downward API) | `""` |
| **POD_NAME** | Pod name (from Kubernetes Downward API) | `""` |
| **CALL_SERVICE** | Optional downstream URL (e.g., `app2.default.svc.cluster.local:8080/api`) — if set, the server will make an HTTP/HTTPS call to this target, append its own hop info, and return the combined JSON chain. | unset |
| **IP_MODE** | IP mode: `ipv4`, `ipv6`, or unset (listen on both) | unset |

---

## 🔗 JSON Response Structure

When called directly:

```json
[
  {
    "node_name": "node-1",
    "pod_name": "app1-abc123",
    "hostname": "app1-host",
    "request_source_ip": "10.244.0.10",
    "request_destination_ip": "10.244.0.20",
    "request_url": "http://app1:8080/test",
    "incoming_headers": {
      "User-Agent": ["curl/8.7.1"],
      "Accept": ["*/*"]
    },
    "ts": "2025-10-20T09:40:10.366Z"
  }
]
