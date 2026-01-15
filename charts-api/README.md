
# 📊 Charts API

> A high-performance, serverless-ready Go API that generates beautiful SVG charts on-the-fly.

The **Charts API** allows you to generate dynamic charts (Line, Bar, Area, Pie, Scatter) by sending a single HTTP GET request. The configuration is passed as a Base64-encoded JSON string, making it easy to embed charts in emails, Markdown files, or websites without client-side JavaScript libraries.

## ⚡ Features

* **Pure SVG Output:** High-quality, scalable vector graphics.
* **Zero Dependencies:** No client-side JS required; renders entirely on the server.
* **Base64 Configuration:** Full chart config embedded in the URL.
* **Caching:** Built-in HTTP caching headers (`max-age=3600`).
* **5 Chart Types:** Line, Area, Bar, Pie, and Scatter.
* **Self-Documenting:** Interactive documentation included at the root endpoint.

## 🚀 Quick Start

### Prerequisites
* [Go 1.16+](https://golang.org/dl/)

### Installation

```bash
# Clone the repository
git clone [https://github.com/yourusername/charts-api.git](https://github.com/yourusername/charts-api.git)
cd charts-api

# Install dependencies
go mod tidy

# Run the server
go run main.go

```

The server will start on port `8080`.

## 📖 Usage

### 1. The Endpoint

**`GET /chart?data={BASE64_JSON}`**

To generate a chart, you must:

1. Create a JSON configuration object.
2. Base64 URL-encode the JSON string.
3. Pass it to the `data` query parameter.

### 2. Example (Command Line)

Here is how you can test it using `curl` and `base64`:

```bash
# JSON: {"type":"pie","data":{"data":[{"name":"A","value":10},{"name":"B","value":20}]}}
# Base64: eyJ0eXBlIjoicGllIiwiZGF0YSI6eyJkYXRhIjpbeyJuYW1lIjoiQSIsInZhbHVlIjoxMH0seyJuYW1lIjoiQiIsInZhbHVlIjoyMH1dfX0=

curl "http://localhost:8080/chart?data=eyJ0eXBlIjoicGllIiwiZGF0YSI6eyJkYXRhIjpbeyJuYW1lIjoiQSIsInZhbHVlIjoxMH0seyJuYW1lIjoiQiIsInZhbHVlIjoyMH1dfX0=" > chart.svg

```

---

## 🔧 Chart Configurations

All configurations share these base properties:

| Property | Type | Default | Description |
| --- | --- | --- | --- |
| `type` | string | **Required** | `line`, `area`, `bar`, `pie`, `scatter` |
| `title` | string | "" | Title displayed at the top |
| `width` | int | 800 | Width in pixels |
| `height` | int | 600 | Height in pixels |
| `data` | object | **Required** | Specific data for the chart type |

### 📈 Line Chart

```json
{
  "type": "line",
  "title": "Revenue 2024",
  "data": {
    "xAxis": ["Jan", "Feb", "Mar"],
    "series": [
      {
        "name": "Product A",
        "data": [10, 15, 12],
        "color": "#667eea"
      }
    ]
  }
}

```

### 📊 Bar Chart

```json
{
  "type": "bar",
  "title": "User Growth",
  "data": {
    "xAxis": ["Q1", "Q2", "Q3", "Q4"],
    "series": [
      {
        "name": "Users",
        "data": [500, 1200, 1400, 2000]
      }
    ]
  }
}

```

### 🥧 Pie Chart

```json
{
  "type": "pie",
  "data": {
    "data": [
      { "name": "Direct", "value": 40 },
      { "name": "Social", "value": 35 },
      { "name": "Referral", "value": 25 }
    ]
  }
}

```

### ⚫ Scatter Chart

```json
{
  "type": "scatter",
  "data": {
    "series": [
      {
        "name": "Experiment 1",
        "data": [[10, 20], [15, 25], [20, 18]]
      }
    ]
  }
}

```

### 📉 Area Chart

```json
{
  "type": "area",
  "data": {
    "xAxis": ["Day 1", "Day 2", "Day 3"],
    "series": [
      {
        "name": "Temperature",
        "data": [22, 24, 21],
        "color": "#ff0000"
      }
    ]
  }
}

```

## 📄 License

This project is licensed under the **MIT License**.
