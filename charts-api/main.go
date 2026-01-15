package main

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

// ChartConfig represents the configuration for any chart type
type ChartConfig struct {
	Type   string          `json:"type"`
	Title  string          `json:"title"`
	Width  int             `json:"width"`
	Height int             `json:"height"`
	Data   json.RawMessage `json:"data"`
}

// LineChartData represents line chart specific data
type LineChartData struct {
	XAxis  []string     `json:"xAxis"`
	Series []SeriesData `json:"series"`
}

// AreaChartData represents area chart data
type AreaChartData struct {
	XAxis   []string     `json:"xAxis"`
	Series  []SeriesData `json:"series"`
	Stacked bool         `json:"stacked,omitempty"`
}

// BarChartData represents bar chart data
type BarChartData struct {
	XAxis      []string     `json:"xAxis"`
	Series     []SeriesData `json:"series"`
	Stacked    bool         `json:"stacked,omitempty"`
	Horizontal bool         `json:"horizontal,omitempty"`
}

// PieChartData represents pie chart data
type PieChartData struct {
	Data []PieItem `json:"data"`
}

// ScatterChartData represents scatter chart data
type ScatterChartData struct {
	Series []ScatterSeries `json:"series"`
}

// SeriesData represents a data series
type SeriesData struct {
	Name  string        `json:"name"`
	Data  []interface{} `json:"data"`
	Color string        `json:"color,omitempty"`
}

// ScatterSeries represents scatter plot series
type ScatterSeries struct {
	Name string      `json:"name"`
	Data [][]float64 `json:"data"`
}

// PieItem represents a pie chart item
type PieItem struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

func main() {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)

	// Routes
	r.Get("/", documentationHandler)
	r.Get("/chart", chartHandler)
	r.Get("/health", healthHandler)

	log.Println("Charts API Server starting on :8002")
	log.Fatal(http.ListenAndServe(":8002", r))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func documentationHandler(w http.ResponseWriter, r *http.Request) {
	const apiDocsHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Charts API - Documentation</title>
	<style>
		:root {
			--bg-color: #f8f9fa;
			--text-color: #212529;
			--accent-color: #667eea;
			--accent-secondary: #764ba2;
			--code-bg: #e9ecef;
			--card-bg: #ffffff;
			--border-color: #dee2e6;
			--shadow: 0 4px 6px rgba(0,0,0,0.1);
			--success: #28a745;
			--warning: #ffc107;
			--info: #17a2b8;
		}
		@media (prefers-color-scheme: dark) {
			:root {
				--bg-color: #0d1117;
				--text-color: #e6edf3;
				--accent-color: #8b9fe8;
				--accent-secondary: #9d7bc7;
				--code-bg: #161b22;
				--card-bg: #161b22;
				--border-color: #30363d;
				--shadow: 0 4px 6px rgba(0,0,0,0.4);
				--success: #3fb950;
				--warning: #d29922;
				--info: #58a6ff;
			}
		}
		* { margin: 0; padding: 0; box-sizing: border-box; }
		body { 
			font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; 
			line-height: 1.6; 
			color: var(--text-color); 
			background: var(--bg-color); 
		}
		.header {
			background: linear-gradient(135deg, var(--accent-color) 0%, var(--accent-secondary) 100%);
			color: white;
			padding: 3rem 2rem;
			text-align: center;
		}
		.header h1 { font-size: 2.5rem; margin-bottom: 0.5rem; text-shadow: 2px 2px 4px rgba(0,0,0,0.2); }
		.header p { font-size: 1.1rem; opacity: 0.95; }
		.container { max-width: 1200px; margin: 0 auto; padding: 2rem; }
		.section { 
			background: var(--card-bg); 
			padding: 2rem; 
			border-radius: 8px; 
			box-shadow: var(--shadow); 
			margin-bottom: 2rem;
			border: 1px solid var(--border-color);
		}
		h2 { 
			color: var(--accent-color); 
			margin-bottom: 1rem; 
			border-bottom: 2px solid var(--accent-color); 
			padding-bottom: 0.5rem;
			font-size: 1.8rem;
		}
		h3 { 
			color: var(--accent-color); 
			margin: 1.5rem 0 1rem;
			font-size: 1.3rem;
		}
		.endpoint { 
			background: var(--code-bg); 
			padding: 1rem; 
			border-radius: 5px; 
			font-family: 'Courier New', monospace; 
			border: 1px solid var(--border-color); 
			margin: 1rem 0;
			font-size: 0.95rem;
		}
		.method { 
			color: #fff; 
			background: var(--success); 
			padding: 4px 10px; 
			border-radius: 4px; 
			margin-right: 10px; 
			font-weight: bold;
			font-size: 0.85rem;
		}
		code { 
			background: var(--code-bg); 
			padding: 3px 6px; 
			border-radius: 4px; 
			font-family: 'Courier New', monospace;
			font-size: 0.9em;
		}
		pre {
			background: var(--code-bg);
			padding: 1rem;
			border-radius: 5px;
			overflow-x: auto;
			border: 1px solid var(--border-color);
			margin: 1rem 0;
			font-family: 'Courier New', monospace;
			font-size: 0.9rem;
			line-height: 1.5;
		}
		.chart-grid {
			display: grid;
			grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
			gap: 1.5rem;
			margin: 2rem 0;
		}
		.chart-card {
			background: var(--card-bg);
			border: 1px solid var(--border-color);
			border-radius: 8px;
			padding: 1rem;
			transition: transform 0.2s;
		}
		.chart-card:hover { transform: translateY(-4px); box-shadow: 0 8px 12px rgba(0,0,0,0.2); }
		.chart-card h4 { margin-bottom: 0.5rem; color: var(--accent-color); }
		.chart-card img {
			width: 100%;
			height: auto;
			border-radius: 4px;
			margin-top: 10px;
			background: white;
			border: 1px solid var(--border-color);
		}
		.badge {
			display: inline-block;
			padding: 4px 10px;
			border-radius: 4px;
			font-size: 0.85rem;
			font-weight: 600;
			margin-right: 0.5rem;
		}
		.badge-type { background: var(--info); color: white; }
		.feature-list {
			list-style: none;
			padding-left: 0;
		}
		.feature-list li {
			padding: 0.5rem 0;
			padding-left: 1.5rem;
			position: relative;
		}
		.feature-list li:before {
			content: "✓";
			position: absolute;
			left: 0;
			color: var(--success);
			font-weight: bold;
		}
		.param-table {
			width: 100%;
			border-collapse: collapse;
			margin: 1rem 0;
		}
		.param-table th,
		.param-table td {
			padding: 0.75rem;
			text-align: left;
			border-bottom: 1px solid var(--border-color);
		}
		.param-table th {
			background: var(--code-bg);
			font-weight: 600;
		}
		.param-table tr:hover {
			background: var(--code-bg);
		}
		.footer { 
			text-align: center; 
			padding: 2rem; 
			color: #6c757d; 
			border-top: 1px solid var(--border-color);
		}
		.copy-btn {
			background: var(--accent-color);
			color: white;
			border: none;
			padding: 0.5rem 1rem;
			border-radius: 4px;
			cursor: pointer;
			font-size: 0.9rem;
			margin-top: 0.5rem;
			transition: opacity 0.2s;
		}
		.copy-btn:hover {
			opacity: 0.8;
		}
	</style>
</head>
<body>
	<div class="header">
		<h1>📊 Charts API</h1>
		<p>Generate beautiful SVG charts with a simple HTTP GET request</p>
	</div>

	<div class="container">
		
		<div class="section">
			<h2>🚀 Quick Start</h2>
			<div class="endpoint"><span class="method">GET</span> /chart?data={base64_json}</div>
			<p style="margin-top: 1rem">Create a JSON config, Base64 URL-encode it, and pass it as the <code>data</code> query parameter.</p>
			
			<h3>Example Request</h3>
			<pre>{
  "type": "line",
  "title": "Sales Data",
  "width": 800,
  "height": 600,
  "data": {
    "xAxis": ["Jan", "Feb", "Mar", "Apr", "May"],
    "series": [
      {
        "name": "Revenue",
        "data": [1200, 1900, 1500, 2100, 2400],
        "color": "#667eea"
      }
    ]
  }
}</pre>
			<button class="copy-btn" onclick="copyExample()">Copy Example</button>
		</div>

		<div class="section">
			<h2>📋 API Endpoints</h2>
			
			<h3>Health Check</h3>
			<div class="endpoint"><span class="method">GET</span> /health</div>
			<p>Returns server health status.</p>
			
			<h3>Generate Chart</h3>
			<div class="endpoint"><span class="method">GET</span> /chart?data={base64_json}</div>
			<p>Generates an SVG chart based on the provided configuration.</p>
			
			<table class="param-table">
				<thead>
					<tr>
						<th>Parameter</th>
						<th>Type</th>
						<th>Required</th>
						<th>Description</th>
					</tr>
				</thead>
				<tbody>
					<tr>
						<td><code>data</code></td>
						<td>string</td>
						<td>Yes</td>
						<td>Base64 URL-encoded JSON configuration</td>
					</tr>
				</tbody>
			</table>
		</div>

		<div class="section">
			<h2>📊 Chart Types</h2>
			<p>Below are live SVG examples rendered by this API instance.</p>
			
			<div class="chart-grid">
				<div class="chart-card">
					<h4>📈 Line Chart</h4>
					<span class="badge badge-type">type: "line"</span>
					<img id="demo-line" alt="Line Chart" />
				</div>

				<div class="chart-card">
					<h4>📉 Area Chart</h4>
					<span class="badge badge-type">type: "area"</span>
					<img id="demo-area" alt="Area Chart" />
				</div>

				<div class="chart-card">
					<h4>📊 Bar Chart</h4>
					<span class="badge badge-type">type: "bar"</span>
					<img id="demo-bar" alt="Bar Chart" />
				</div>

				<div class="chart-card">
					<h4>🥧 Pie Chart</h4>
					<span class="badge badge-type">type: "pie"</span>
					<img id="demo-pie" alt="Pie Chart" />
				</div>

				<div class="chart-card">
					<h4>⚫ Scatter Chart</h4>
					<span class="badge badge-type">type: "scatter"</span>
					<img id="demo-scatter" alt="Scatter Chart" />
				</div>
			</div>
		</div>

		<div class="section">
			<h2>🔧 Configuration Reference</h2>
			
			<h3>Base Configuration</h3>
			<table class="param-table">
				<thead>
					<tr>
						<th>Property</th>
						<th>Type</th>
						<th>Default</th>
						<th>Description</th>
					</tr>
				</thead>
				<tbody>
					<tr>
						<td><code>type</code></td>
						<td>string</td>
						<td>-</td>
						<td>Chart type: "line", "area", "bar", "pie", "scatter"</td>
					</tr>
					<tr>
						<td><code>title</code></td>
						<td>string</td>
						<td>""</td>
						<td>Chart title</td>
					</tr>
					<tr>
						<td><code>width</code></td>
						<td>number</td>
						<td>800</td>
						<td>Width in pixels</td>
					</tr>
					<tr>
						<td><code>height</code></td>
						<td>number</td>
						<td>600</td>
						<td>Height in pixels</td>
					</tr>
					<tr>
						<td><code>data</code></td>
						<td>object</td>
						<td>-</td>
						<td>Chart-specific data configuration</td>
					</tr>
				</tbody>
			</table>

			<h3>Line/Area/Bar Chart Data</h3>
			<pre>{
  "xAxis": ["Label1", "Label2", "Label3"],
  "series": [
    {
      "name": "Series Name",
      "data": [100, 200, 150],
      "color": "#667eea"
    }
  ]
}</pre>

			<h3>Pie Chart Data</h3>
			<pre>{
  "data": [
    { "name": "Category A", "value": 300 },
    { "name": "Category B", "value": 150 },
    { "name": "Category C", "value": 450 }
  ]
}</pre>

			<h3>Scatter Chart Data</h3>
			<pre>{
  "series": [
    {
      "name": "Dataset 1",
      "data": [[10, 20], [15, 25], [20, 30]]
    }
  ]
}</pre>
		</div>

		<div class="section">
			<h2>💡 Features</h2>
			<ul class="feature-list">
				<li>Pure SVG output - no JavaScript required</li>
				<li>Multiple chart types supported</li>
				<li>Customizable colors and dimensions</li>
				<li>Lightweight and fast</li>
				<li>Easy to integrate with any application</li>
				<li>Cacheable responses</li>
				<li>Built with Go and chi router</li>
			</ul>
		</div>

		<div class="section">
			<h2>📝 Response Format</h2>
			<p><strong>Content-Type:</strong> <code>image/svg+xml</code></p>
			<p><strong>Cache-Control:</strong> <code>public, max-age=3600</code></p>
			<p style="margin-top: 1rem">All charts return pure SVG that can be embedded directly in HTML, documents, or downloaded as files.</p>
			
			<h3>Usage in HTML</h3>
			<pre>&lt;img src="http://localhost:8080/chart?data={base64_json}" alt="Chart" /&gt;</pre>
			
			<h3>Usage in Markdown</h3>
			<pre>![Chart](http://localhost:8080/chart?data={base64_json})</pre>
		</div>

		<div class="section">
			<h2>⚡ Try It</h2>
			<p>Test the API by clicking the links below:</p>
			<ul style="margin-top: 1rem;">
				<li style="margin: 0.5rem 0;"><a href="/health" target="_blank">Health Check</a></li>
				<li style="margin: 0.5rem 0;"><a href="/chart?data=eyJ0eXBlIjoibGluZSIsInRpdGxlIjoiU2FsZXMiLCJ3aWR0aCI6ODAwLCJoZWlnaHQiOjYwMCwiZGF0YSI6eyJ4QXhpcyI6WyJKYW4iLCJGZWIiLCJNYXIiXSwic2VyaWVzIjpbeyJuYW1lIjoiUmV2ZW51ZSIsImRhdGEiOlsxMDAsNTAsMTUwXSwiY29sb3IiOiIjNjY3ZWVhIn1dfX0" target="_blank">Example Line Chart</a></li>
			</ul>
		</div>
	</div>

	<div class="footer">
		<p>Powered by <strong>Go-Chart</strong> &amp; <strong>Chi Router</strong></p>
		<p style="margin-top: 0.5rem; font-size: 0.9rem;">Built with ❤️ using Go</p>
	</div>

	<script>
		const demos = {
			"demo-line": {
				type: "line",
				title: "Weekly Users",
				width: 700, height: 400,
				data: {
					xAxis: ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"],
					series: [
						{ name: "Active", data: [120, 200, 150, 80, 70, 110, 130], color: "#667eea" },
						{ name: "Inactive", data: [80, 100, 90, 120, 130, 90, 70], color: "#f56565" }
					]
				}
			},
			"demo-area": {
				type: "area",
				title: "Server Load",
				width: 700, height: 400,
				data: {
					xAxis: ["00:00", "04:00", "08:00", "12:00", "16:00", "20:00"],
					series: [
						{ name: "CPU", data: [15, 20, 45, 80, 65, 30], color: "#667eea" }
					]
				}
			},
			"demo-bar": {
				type: "bar",
				title: "Q4 Revenue",
				width: 700, height: 400,
				data: {
					xAxis: ["Oct", "Nov", "Dec"],
					series: [
						{ name: "Sales", data: [4200, 4800, 5600], color: "#667eea" }
					]
				}
			},
			"demo-pie": {
				type: "pie",
				title: "Storage Distribution",
				width: 700, height: 400,
				data: {
					data: [
						{ name: "Images", value: 1048 },
						{ name: "Video", value: 735 },
						{ name: "Docs", value: 580 }
					]
				}
			},
			"demo-scatter": {
				type: "scatter",
				title: "Data Distribution",
				width: 700, height: 400,
				data: {
					series: [{
						name: "Sample",
						data: [[10, 8], [8, 5], [12, 11], [7, 6], [11, 9], [14, 12], [6, 4], [4, 3]]
					}]
				}
			}
		};

		function loadDemoCharts() {
			for (const [id, config] of Object.entries(demos)) {
				const img = document.getElementById(id);
				if (img) {
					const jsonStr = JSON.stringify(config);
					const base64 = btoa(jsonStr);
					const urlSafeBase64 = base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
					img.src = "/chart?data=" + urlSafeBase64;
				}
			}
		}

		function copyExample() {
			const example = {
				type: "line",
				title: "Sales Data",
				width: 800,
				height: 600,
				data: {
					xAxis: ["Jan", "Feb", "Mar", "Apr", "May"],
					series: [
						{
							name: "Revenue",
							data: [1200, 1900, 1500, 2100, 2400],
							color: "#667eea"
						}
					]
				}
			};
			navigator.clipboard.writeText(JSON.stringify(example, null, 2));
			alert('Example copied to clipboard!');
		}

		window.addEventListener('DOMContentLoaded', loadDemoCharts);
	</script>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(apiDocsHTML))
}

func chartHandler(w http.ResponseWriter, r *http.Request) {
	// Get encoded data from query parameter
	encodedData := r.URL.Query().Get("data")
	if encodedData == "" {
		http.Error(w, "Missing 'data' parameter", http.StatusBadRequest)
		return
	}

	// Decode base64
	decodedBytes, err := base64.URLEncoding.DecodeString(encodedData)
	if err != nil {
		http.Error(w, "Invalid base64 encoding: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Parse chart config
	var config ChartConfig
	if err := json.Unmarshal(decodedBytes, &config); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Set defaults
	if config.Width == 0 {
		config.Width = 800
	}
	if config.Height == 0 {
		config.Height = 600
	}

	// Generate chart based on type
	var err2 error
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=3600")

	switch strings.ToLower(config.Type) {
	case "line":
		err2 = generateLineChart(w, config)
	case "area":
		err2 = generateAreaChart(w, config)
	case "bar":
		err2 = generateBarChart(w, config)
	case "pie":
		err2 = generatePieChart(w, config)
	case "scatter":
		err2 = generateScatterChart(w, config)
	default:
		http.Error(w, "Unsupported chart type: "+config.Type, http.StatusBadRequest)
		return
	}

	if err2 != nil {
		http.Error(w, "Error generating chart: "+err2.Error(), http.StatusInternalServerError)
		return
	}
}

func generateLineChart(w http.ResponseWriter, config ChartConfig) error {
	var data LineChartData
	if err := json.Unmarshal(config.Data, &data); err != nil {
		return err
	}

	graph := chart.Chart{
		Title:  config.Title,
		Width:  config.Width,
		Height: config.Height,
		XAxis: chart.XAxis{
			Style: chart.Style{
				FontSize: 10,
			},
			ValueFormatter: func(v interface{}) string {
				idx := int(v.(float64))
				if idx >= 0 && idx < len(data.XAxis) {
					return data.XAxis[idx]
				}
				return ""
			},
		},
		YAxis: chart.YAxis{
			Style: chart.Style{
				FontSize: 10,
			},
		},
	}

	colors := []drawing.Color{
		drawing.Color{R: 102, G: 126, B: 234, A: 255},
		drawing.Color{R: 237, G: 100, B: 166, A: 255},
		drawing.Color{R: 255, G: 159, B: 64, A: 255},
		drawing.Color{R: 75, G: 192, B: 192, A: 255},
	}

	for idx, series := range data.Series {
		xValues := make([]float64, len(series.Data))
		yValues := make([]float64, len(series.Data))

		for i, val := range series.Data {
			xValues[i] = float64(i)
			yValues[i] = toFloat64(val)
		}

		color := colors[idx%len(colors)]
		if series.Color != "" {
			if c, err := parseHexColor(series.Color); err == nil {
				color = c
			}
		}

		graph.Series = append(graph.Series, chart.ContinuousSeries{
			Name:    series.Name,
			XValues: xValues,
			YValues: yValues,
			Style: chart.Style{
				StrokeColor: color,
				StrokeWidth: 2,
			},
		})
	}

	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
	}

	return graph.Render(chart.SVG, w)
}

func generateAreaChart(w http.ResponseWriter, config ChartConfig) error {
	var data AreaChartData
	if err := json.Unmarshal(config.Data, &data); err != nil {
		return err
	}

	graph := chart.Chart{
		Title:  config.Title,
		Width:  config.Width,
		Height: config.Height,
		XAxis: chart.XAxis{
			Style: chart.Style{
				FontSize: 10,
			},
			ValueFormatter: func(v interface{}) string {
				idx := int(v.(float64))
				if idx >= 0 && idx < len(data.XAxis) {
					return data.XAxis[idx]
				}
				return ""
			},
		},
		YAxis: chart.YAxis{
			Style: chart.Style{
				FontSize: 10,
			},
		},
	}

	colors := []drawing.Color{
		drawing.Color{R: 102, G: 126, B: 234, A: 255},
		drawing.Color{R: 237, G: 100, B: 166, A: 255},
	}

	for idx, series := range data.Series {
		xValues := make([]float64, len(series.Data))
		yValues := make([]float64, len(series.Data))

		for i, val := range series.Data {
			xValues[i] = float64(i)
			yValues[i] = toFloat64(val)
		}

		color := colors[idx%len(colors)]
		if series.Color != "" {
			if c, err := parseHexColor(series.Color); err == nil {
				color = c
			}
		}

		fillColor := color
		fillColor.A = 100

		graph.Series = append(graph.Series, chart.ContinuousSeries{
			Name:    series.Name,
			XValues: xValues,
			YValues: yValues,
			Style: chart.Style{
				StrokeColor: color,
				StrokeWidth: 2,
				FillColor:   fillColor,
			},
		})
	}

	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
	}

	return graph.Render(chart.SVG, w)
}

func generateBarChart(w http.ResponseWriter, config ChartConfig) error {
	var data BarChartData
	if err := json.Unmarshal(config.Data, &data); err != nil {
		return err
	}

	if len(data.Series) == 0 {
		return nil
	}

	bars := []chart.Value{}
	for i, label := range data.XAxis {
		value := 0.0
		if i < len(data.Series[0].Data) {
			value = toFloat64(data.Series[0].Data[i])
		}
		bars = append(bars, chart.Value{
			Label: label,
			Value: value,
		})
	}

	graph := chart.BarChart{
		Title:  config.Title,
		Width:  config.Width,
		Height: config.Height,
		Bars:   bars,
		XAxis: chart.Style{
			FontSize: 10,
		},
		YAxis: chart.YAxis{
			Style: chart.Style{
				FontSize: 10,
			},
		},
	}

	return graph.Render(chart.SVG, w)
}

func generatePieChart(w http.ResponseWriter, config ChartConfig) error {
	var data PieChartData
	if err := json.Unmarshal(config.Data, &data); err != nil {
		return err
	}

	values := []chart.Value{}
	for _, item := range data.Data {
		values = append(values, chart.Value{
			Label: item.Name,
			Value: item.Value,
		})
	}

	graph := chart.PieChart{
		Title:  config.Title,
		Width:  config.Width,
		Height: config.Height,
		Values: values,
	}

	return graph.Render(chart.SVG, w)
}

func generateScatterChart(w http.ResponseWriter, config ChartConfig) error {
	var data ScatterChartData
	if err := json.Unmarshal(config.Data, &data); err != nil {
		return err
	}

	graph := chart.Chart{
		Title:  config.Title,
		Width:  config.Width,
		Height: config.Height,
		XAxis: chart.XAxis{
			Style: chart.Style{
				FontSize: 10,
			},
		},
		YAxis: chart.YAxis{
			Style: chart.Style{
				FontSize: 10,
			},
		},
	}

	colors := []drawing.Color{
		drawing.Color{R: 102, G: 126, B: 234, A: 255},
		drawing.Color{R: 237, G: 100, B: 166, A: 255},
	}

	for idx, series := range data.Series {
		xValues := make([]float64, len(series.Data))
		yValues := make([]float64, len(series.Data))

		for i, point := range series.Data {
			if len(point) >= 2 {
				xValues[i] = point[0]
				yValues[i] = point[1]
			}
		}

		color := colors[idx%len(colors)]

		graph.Series = append(graph.Series, chart.ContinuousSeries{
			Name:    series.Name,
			XValues: xValues,
			YValues: yValues,
			Style: chart.Style{
				StrokeWidth: chart.Disabled,
				DotWidth:    5,
				DotColor:    color,
			},
		})
	}

	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
	}

	return graph.Render(chart.SVG, w)
}

// Helper functions
func toFloat64(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0
}

func parseHexColor(hex string) (drawing.Color, error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return drawing.Color{}, nil
	}

	var r, g, b uint8
	_, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return drawing.Color{}, err
	}
	r64, _ := strconv.ParseUint(hex[0:2], 16, 8)
	r = uint8(r64)

	g64, _ := strconv.ParseUint(hex[2:4], 16, 8)
	g = uint8(g64)

	b64, _ := strconv.ParseUint(hex[4:6], 16, 8)
	b = uint8(b64)

	return drawing.Color{R: r, G: g, B: b, A: 255}, nil
}
