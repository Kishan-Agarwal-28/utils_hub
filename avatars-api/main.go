package main

import (
	"crypto/md5"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	lru "github.com/hashicorp/golang-lru/v2"
)

var cache *lru.Cache[string, string]

func main() {
	var err error
	cache, err = lru.New[string, string](1000)
	if err != nil {
		log.Fatal(err)
	}
	loadGeoJSON()
	if len(constellationData.Features) == 0 {
		panic("CRITICAL ERROR: No constellation features found in JSON!")
	}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/health"))
	r.Use(middleware.Compress(5))
	r.Use(httprate.Limit(
		100,
		1*time.Minute,
		httprate.WithKeyFuncs(httprate.KeyByIP),
	))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
	}))

	r.Get("/", documentationHandler)
	r.Get("/api/generate-avatar", generateAvatarHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	log.Printf("🚀 Server running on port %s", port)
	http.ListenAndServe(":"+port, r)
}
func documentationHandler(w http.ResponseWriter, r *http.Request) {
	const apiDocsHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Avatar Generator API - Documentation</title>
	<style>
		:root {
			--bg-color: #f8f9fa;
			--text-color: #212529;
			--accent-color: #007bff;
			--code-bg: #e9ecef;
			--card-bg: #ffffff;
			--border-color: #dee2e6;
			--shadow: 0 4px 6px rgba(0,0,0,0.1);
		}
		@media (prefers-color-scheme: dark) {
			:root {
				--bg-color: #0d1117;
				--text-color: #e6edf3;
				--accent-color: #58a6ff;
				--code-bg: #161b22;
				--card-bg: #161b22;
				--border-color: #30363d;
				--shadow: 0 4px 6px rgba(0,0,0,0.4);
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
			background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
			color: white;
			padding: 3rem 2rem;
			text-align: center;
		}
		.header h1 { font-size: 2.5rem; margin-bottom: 0.5rem; }
		.header p { font-size: 1.2rem; opacity: 0.9; }
		.container { 
			max-width: 1200px; 
			margin: 0 auto; 
			padding: 2rem; 
		}
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
			padding-bottom: 0.5rem;
			border-bottom: 2px solid var(--accent-color);
		}
		h3 { margin: 1.5rem 0 1rem; color: var(--text-color); }
		.endpoint { 
			background: var(--code-bg); 
			padding: 1rem; 
			border-radius: 5px; 
			font-family: monospace; 
			font-size: 1rem; 
			margin: 1rem 0;
			border: 1px solid var(--border-color);
			overflow-x: auto;
		}
		.method { 
			color: #fff; 
			background: #28a745; 
			padding: 4px 10px; 
			border-radius: 4px; 
			margin-right: 10px; 
			font-weight: bold; 
			font-size: 0.9rem;
		}
		code { 
			background: var(--code-bg); 
			padding: 3px 6px; 
			border-radius: 4px; 
			font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace; 
			font-size: 0.9em;
		}
		pre { 
			background: var(--code-bg); 
			padding: 1rem; 
			border-radius: 5px; 
			overflow-x: auto; 
			border: 1px solid var(--border-color);
			margin: 1rem 0;
		}
		table { 
			width: 100%; 
			border-collapse: collapse; 
			margin: 1rem 0; 
		}
		th, td { 
			text-align: left; 
			padding: 12px; 
			border-bottom: 1px solid var(--border-color); 
		}
		th { 
			background: var(--code-bg); 
			font-weight: 600;
		}
		.badge { 
			display: inline-block; 
			padding: 3px 10px; 
			border-radius: 12px; 
			font-size: 0.85em; 
			font-weight: bold; 
		}
		.required { background: #dc3545; color: white; }
		.optional { background: #6c757d; color: white; }
		.avatar-grid {
			display: grid;
			grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
			gap: 1.5rem;
			margin: 2rem 0;
		}
		.avatar-card {
			background: var(--card-bg);
			border: 1px solid var(--border-color);
			border-radius: 8px;
			padding: 1rem;
			text-align: center;
			transition: transform 0.2s, box-shadow 0.2s;
		}
		.avatar-card:hover {
			transform: translateY(-4px);
			box-shadow: 0 8px 12px rgba(0,0,0,0.15);
		}
		.avatar-card img {
			width: 150px;
			height: 150px;
			border-radius: 8px;
			margin-bottom: 0.5rem;
			background: var(--code-bg);
		}
		.avatar-card h4 {
			margin: 0.5rem 0;
			color: var(--accent-color);
		}
		.avatar-card code {
			font-size: 0.8rem;
			word-break: break-all;
		}
		.footer { 
			text-align: center; 
			padding: 2rem; 
			color: #6c757d; 
			font-size: 0.9rem;
		}
		@media (max-width: 768px) {
			.header h1 { font-size: 2rem; }
			.header p { font-size: 1rem; }
			.container { padding: 1rem; }
			.section { padding: 1rem; }
			.avatar-grid {
				grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
				gap: 1rem;
			}
			.avatar-card img {
				width: 120px;
				height: 120px;
			}
		}
	</style>
</head>
<body>
	<div class="header">
		<h1>🎨 Avatar Generator API</h1>
		<p>Generate beautiful, unique avatars with a simple HTTP request</p>
	</div>

	<div class="container">
		<div class="section">
			<h2>📖 Overview</h2>
			<p>A high-performance JSON API that generates unique, deterministic SVG avatars based on a name input. Perfect for user profiles, comments, and any application needing consistent, personalized avatars.</p>
		</div>

		<div class="section">
			<h2>🚀 Quick Start</h2>
			<h3>Base URL</h3>
			<div class="endpoint">
				<span class="method">GET</span> /api/generate-avatar
			</div>

			<h3>Example Request</h3>
			<pre><code>GET /api/generate-avatar?name=John%20Doe&type=avatar&size=200</code></pre>

			<h3>Parameters</h3>
			<table>
				<thead>
					<tr>
						<th>Parameter</th>
						<th>Type</th>
						<th>Required</th>
						<th>Default</th>
						<th>Description</th>
					</tr>
				</thead>
				<tbody>
					<tr>
						<td><code>name</code></td>
						<td>String</td>
						<td><span class="badge optional">Optional</span></td>
						<td>"User"</td>
						<td>Name used to generate the avatar. Same name always produces the same avatar.</td>
					</tr>
					<tr>
						<td><code>type</code></td>
						<td>String</td>
						<td><span class="badge optional">Optional</span></td>
						<td>"avatar"</td>
						<td>Avatar style. See available types below.</td>
					</tr>
					<tr>
						<td><code>size</code></td>
						<td>Integer</td>
						<td><span class="badge optional">Optional</span></td>
						<td>100</td>
						<td>Size of the avatar in pixels (width and height).</td>
					</tr>
					<tr>
						<td><code>color</code></td>
						<td>String</td>
						<td><span class="badge optional">Optional</span></td>
						<td>Auto</td>
						<td>Hex color code (e.g., #FF5733). If not provided, color is generated from name.</td>
					</tr>
				</tbody>
			</table>
		</div>

		<div class="section">
			<h2>🎭 Avatar Types</h2>
			<p>All examples use the name "Alex Morgan" for consistency. Each type generates a unique, deterministic design.</p>
			
			<div class="avatar-grid">
				<div class="avatar-card">
					<img src="/api/generate-avatar?name=Alex%20Morgan&type=avatar&size=150" alt="Avatar">
					<h4>avatar</h4>
					<p>Classic circular avatar with initials</p>
					<code>type=avatar</code>
				</div>

				<div class="avatar-card">
					<img src="/api/generate-avatar?name=Alex%20Morgan&type=gravatar&size=150" alt="Gravatar">
					<h4>gravatar</h4>
					<p>GitHub-style identicon</p>
					<code>type=gravatar</code>
				</div>

				<div class="avatar-card">
					<img src="/api/generate-avatar?name=Alex%20Morgan&type=dither&size=150" alt="Dither">
					<h4>dither</h4>
					<p>Retro dithered plasma effect</p>
					<code>type=dither</code>
				</div>

				<div class="avatar-card">
					<img src="/api/generate-avatar?name=Alex%20Morgan&type=ascii&size=150" alt="ASCII">
					<h4>ascii</h4>
					<p>Procedural ASCII robot</p>
					<code>type=ascii</code>
				</div>

				<div class="avatar-card">
					<img src="/api/generate-avatar?name=Alex%20Morgan&type=dotmatrix&size=150" alt="Dot Matrix">
					<h4>dotmatrix</h4>
					<p>LED dot matrix display</p>
					<code>type=dotmatrix</code>
				</div>

				<div class="avatar-card">
					<img src="/api/generate-avatar?name=Alex%20Morgan&type=terminal&size=150" alt="Terminal">
					<h4>terminal</h4>
					<p>Retro terminal block text</p>
					<code>type=terminal</code>
				</div>

				<div class="avatar-card">
					<img src="/api/generate-avatar?name=Alex%20Morgan&type=bauhaus&size=150" alt="Bauhaus">
					<h4>bauhaus</h4>
					<p>Geometric Bauhaus design</p>
					<code>type=bauhaus</code>
				</div>

				<div class="avatar-card">
					<img src="/api/generate-avatar?name=Alex%20Morgan&type=ring&size=150" alt="Ring">
					<h4>ring</h4>
					<p>Gradient ring pattern</p>
					<code>type=ring</code>
				</div>

				<div class="avatar-card">
					<img src="/api/generate-avatar?name=Alex%20Morgan&type=beam&size=150" alt="Beam">
					<h4>beam</h4>
					<p>Connected network nodes</p>
					<code>type=beam</code>
				</div>

				<div class="avatar-card">
					<img src="/api/generate-avatar?name=Alex%20Morgan&type=marble&size=150" alt="Marble">
					<h4>marble</h4>
					<p>Marble texture effect</p>
					<code>type=marble</code>
				</div>

				<div class="avatar-card">
					<img src="/api/generate-avatar?name=Alex%20Morgan&type=glitch&size=150" alt="Glitch">
					<h4>glitch</h4>
					<p>Cyberpunk glitch effect</p>
					<code>type=glitch</code>
				</div>

				<div class="avatar-card">
					<img src="/api/generate-avatar?name=Alex%20Morgan&type=sunset&size=150" alt="Sunset">
					<h4>sunset</h4>
					<p>Procedural sunset scene</p>
					<code>type=sunset</code>
				</div>

				<div class="avatar-card">
					<img src="/api/generate-avatar?name=Alex%20Morgan&type=smile&size=150" alt="Smile">
					<h4>smile</h4>
					<p>Minimalist face with expressions</p>
					<code>type=smile</code>
				</div>

				<div class="avatar-card">
					<img src="/api/generate-avatar?name=Alex%20Morgan&type=circuit&size=150" alt="Circuit">
					<h4>circuit</h4>
					<p>Circuit board pattern</p>
					<code>type=circuit</code>
				</div>

				<div class="avatar-card">
					<img src="/api/generate-avatar?name=Alex%20Morgan&type=pixel&size=150" alt="Pixel">
					<h4>pixel</h4>
					<p>Isometric pixel art cube</p>
					<code>type=pixel</code>
				</div>

			<div class="avatar-card">
				<img src="/api/generate-avatar?name=Alex%20Morgan&type=constellation&size=150" alt="Constellation">
				<h4>constellation</h4>
				<p>Real constellation star maps</p>
				<code>type=constellation</code>
			</div>
	</div>
</div>

<div class="section">
	<h2>💡 Usage Examples</h2>
	
	<h3>HTML Image Tag</h3>
	<pre><code>&lt;img src="/api/generate-avatar?name=Jane%20Smith&type=avatar&size=100" alt="Avatar"&gt;</code></pre>

	<h3>With Custom Color</h3>
	<pre><code>&lt;img src="/api/generate-avatar?name=John&type=gravatar&color=%23FF5733" alt="Avatar"&gt;</code></pre>

	<h3>Large Size</h3>
	<pre><code>&lt;img src="/api/generate-avatar?name=Sarah&type=dotmatrix&size=500" alt="Avatar"&gt;</code></pre>

	<h3>JavaScript Fetch</h3>
	<pre><code>fetch('/api/generate-avatar?name=Bob%20Johnson&type=beam')
  .then(response => response.text())
  .then(svg => {
    document.getElementById('avatar').innerHTML = svg;
  });</code></pre>
</div>
<div class="section">
			<h2>⚡ Features</h2>
			<ul style="list-style: none; padding: 0;">
				<li style="padding: 0.5rem 0;">✅ <strong>Deterministic</strong> - Same name always generates the same avatar</li>
				<li style="padding: 0.5rem 0;">✅ <strong>SVG Format</strong> - Scalable to any size without quality loss</li>
				<li style="padding: 0.5rem 0;">✅ <strong>16 Unique Styles</strong> - From classic to creative designs</li>
				<li style="padding: 0.5rem 0;">✅ <strong>Fast & Lightweight</strong> - Generated on-the-fly, no storage needed</li>
				<li style="padding: 0.5rem 0;">✅ <strong>CORS Enabled</strong> - Use from any domain</li>
				<li style="padding: 0.5rem 0;">✅ <strong>Rate Limited</strong> - 100 requests per minute per IP</li>
				<li style="padding: 0.5rem 0;">✅ <strong>Dark Mode Support</strong> - Responsive documentation</li>
			</ul>
		</div>

		<div class="section">
			<h2>📝 Response Format</h2>
			<p>All avatars are returned as SVG (Scalable Vector Graphics) with the content type:</p>
			<div class="endpoint">Content-Type: image/svg+xml; charset=utf-8</div>
			
			<h3>Status Codes</h3>
			<table>
				<thead>
					<tr>
						<th>Code</th>
						<th>Description</th>
					</tr>
				</thead>
				<tbody>
					<tr>
						<td><code>200 OK</code></td>
						<td>Request successful, avatar returned</td>
					</tr>
					<tr>
						<td><code>400 Bad Request</code></td>
						<td>Invalid avatar type specified</td>
					</tr>
					<tr>
						<td><code>429 Too Many Requests</code></td>
						<td>Rate limit exceeded (100 req/min)</td>
					</tr>
					<tr>
						<td><code>500 Internal Server Error</code></td>
						<td>Server-side error occurred</td>
					</tr>
				</tbody>
			</table>
		</div>

		<div class="section">
			<h2>🎨 Color Generation</h2>
			<p>When no color is specified, avatars use the <strong>OKLCH color space</strong> for perceptually uniform, accessible colors. The algorithm:</p>
			<ul style="margin-left: 2rem; margin-top: 1rem;">
				<li>Generates a deterministic hash from the input name</li>
				<li>Maps hash values to hue, chroma, and lightness</li>
				<li>Ensures sufficient contrast for readability</li>
				<li>Produces consistent colors across all sessions</li>
			</ul>
		</div>
	</div>

	<div class="footer">
		<p>Avatar Generator API v1.0 • Built with Go & SVG • MIT License</p>
	</div>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(apiDocsHTML))
}

func clamp(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

func toByte(f float64) uint8 {

	val := f
	if val <= 0.0031308 {
		val = 12.92 * val
	} else {
		val = 1.055*math.Pow(val, 1.0/2.4) - 0.055
	}
	return uint8(clamp(val) * 255)
}

func oklchToHex(l, c, h float64) string {

	hRad := h * (math.Pi / 180.0)
	a := c * math.Cos(hRad)
	b := c * math.Sin(hRad)

	l_ := l + 0.3963377774*a + 0.2158037573*b
	m_ := l - 0.1055613458*a - 0.0638541728*b
	s_ := l - 0.0894841775*a - 1.2914855480*b

	l_ = math.Pow(l_, 3)
	m_ = math.Pow(m_, 3)
	s_ = math.Pow(s_, 3)

	red := 4.0767416621*l_ - 3.3077115913*m_ + 0.2309699292*s_
	green := -1.2684380046*l_ + 2.6097574011*m_ - 0.3413193965*s_
	blue := -0.0041960863*l_ - 0.7034186147*m_ + 1.7076147010*s_

	return fmt.Sprintf("#%02x%02x%02x", toByte(red), toByte(green), toByte(blue))
}

func generateColor(s string) (string, [16]byte) {
	hash := md5.Sum([]byte(s))
	hue := float64(uint16(hash[0])<<8|uint16(hash[1])) / 65535.0 * 360.0

	chroma := 0.10 + (float64(hash[2])/255.0)*0.06

	lightness := 0.55 + (float64(hash[3])/255.0)*0.13

	hex := oklchToHex(lightness, chroma, hue)
	return hex, hash
}
func getInitials(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "?"
	}
	parts := strings.Fields(name)
	initials := string([]rune(parts[0])[0])
	if len(parts) > 1 {
		initials += string([]rune(parts[1])[0])
	}
	return strings.ToUpper(initials)
}
func generateIdenticon(input string) string {
	color, hash := generateColor(input)
	var rects strings.Builder
	gridSize := 5
	cellSize := 50
	for col := range 3 {
		for row := range 5 {
			byteIndex := (col * 5) + row
			if hash[byteIndex%16]%2 == 0 {
				x := col * cellSize
				y := row * cellSize
				fmt.Fprintf(&rects, `<rect x="%d" y="%d" width="%d" height="%d" fill="%s" />`, x, y, cellSize, cellSize, color)
				if col < 2 {
					mirrorX := (gridSize - 1 - col) * cellSize
					fmt.Fprintf(&rects, `<rect x="%d" y="%d" width="%d" height="%d" fill="%s" />`, mirrorX, y, cellSize, cellSize, color)
				}
			}
		}
	}

	return rects.String()
}

var bayerMatrix4x4 = [4][4]float64{
	{0, 8, 2, 10},
	{12, 4, 14, 6},
	{3, 11, 1, 9},
	{15, 7, 13, 5},
}

func getPlasmaValue(x, y int, width, height float64, hash [16]byte) float64 {

	u := float64(x) / width
	v := float64(y) / height

	seedX := float64(hash[0]) / 10.0
	seedY := float64(hash[1]) / 10.0
	freq := 3.0 + (float64(hash[2] % 5))

	value := math.Sin(u*freq+seedX) + math.Cos(v*freq+seedY) + math.Sin((u+v)*freq)

	return (value + 3.0) / 6.0
}

func generateDitheredAvatar(name string) string {
	primaryColor, hash := generateColor(name)
	secondaryColor := "#11011D"
	cols, rows := 32, 32
	pixelSize := 10

	var rects strings.Builder

	for y := 0; y < rows; y++ {
		for x := range cols {
			luminance := getPlasmaValue(x, y, float64(cols), float64(rows), hash)

			threshold := bayerMatrix4x4[y%4][x%4] / 17.0

			fill := secondaryColor

			if (luminance + 0.1) < threshold {
				fill = primaryColor
			}

			if fill == primaryColor {
				rects.WriteString(fmt.Sprintf(
					`<rect x="%d" y="%d" width="%d" height="%d" fill="%s" />`,
					x*pixelSize, y*pixelSize, pixelSize, pixelSize, fill,
				))
			}
		}
	}

	return fmt.Sprintf(`
	<svg xmlns="http://www.w3.org/2000/svg" width="100%%" height="100%%" viewBox="0 0 320 320">
		<rect width="100%%" height="100%%" fill="%s" />
		<g transform="translate(16, 16) scale(0.9)">
			%s
		</g>
	</svg>`, secondaryColor, rects.String())
}

var robotHeads = []string{
	" /_\\ ",  // 0: Cone
	" [~] ",   // 1: Boxy
	" (o) ",   // 2: Round
	" <_> ",   // 3: V-shape
	" {^} ",   // 4: Spiked
	" [..] ",  // 5: Monitor
	" .__. ",  // 6: Flat
	" /MM\\ ", // 7: Crown
	" (**) ",  // 8: Goggles
	" d[ ]b ", // 9: Headphones
	" @__@ ",  // 10: Princess
	" <oo> ",  // 11: Owl-like
}

var robotEyes = []string{
	"|o_o|", // 0: Normal
	"|-.-|", // 1: Sleepy
	"|0_0|", // 2: Wide
	"|X_X|", // 3: Dead
	"|>_<|", // 4: Angry
	"|@_@|", // 5: Dizzy
	"|$_$|", // 6: Money
	"|~_~|", // 7: Winking
	"|O_O|", // 8: Stare
	"|=_|=", // 9: Laser
	"|9_6|", // 10: Crazy
	"|+.+|", // 11: System Mode
}

var robotBodies = []string{
	"/[_]\\", // 0: Trapezoid
	" |-| ",  // 1: Thin
	" [=] ",  // 2: Box
	" /#\\ ", // 3: Pyramid
	" (•) ",  // 4: Round
	" |%| ",  // 5: Vents
	" <_> ",  // 6: Hourglass
	" /|\\ ", // 7: Stick arms
	"-[_]-",  // 8: Wide arms
	" (|) ",  // 9: Oval
	"=[_]=",  // 10: Heavy Arms
	" /B\\ ", // 11: Button
}

var robotLegs = []string{
	" d b ",  // 0: Feet
	" / \\ ", // 1: Stance
	" ||| ",  // 2: Tracks
	" _| |_", // 3: Wide
	" (@) ",  // 4: Wheel
	" /_\\ ", // 5: Skirt
	" | | ",  // 6: Sticks
	" <_> ",  // 7: Point
	" _A_ ",  // 8: Tripod
	" ( ) ",  // 9: Hover
	" J L ",  // 10: Boots
	" V V ",  // 11: Sharp
}

func pickPart(list []string, b byte) string {
	return list[int(b)%len(list)]
}

func generateAsciiRobot(name string, size string) string {

	bgColor, hash := generateColor(name)

	head := pickPart(robotHeads, hash[0])
	eyes := pickPart(robotEyes, hash[1])
	body := pickPart(robotBodies, hash[2])
	legs := pickPart(robotLegs, hash[3])

	asciiArt := []string{
		head,
		eyes,
		body,
		legs,
	}

	var textBlock strings.Builder
	yPos := 60
	lineHeight := 24

	for _, line := range asciiArt {
		safeLine := strings.ReplaceAll(line, " ", "\u00A0")

		fmt.Fprintf(&textBlock, `<tspan x="50%%" dy="%d" text-anchor="middle">%s</tspan>`,
			lineHeight, safeLine)
	}

	return fmt.Sprintf(`
	<svg xmlns="http://www.w3.org/2000/svg" width="%[1]s" height="%[1]s" viewBox="0 0 200 200">
		<rect width="100%%" height="100%%" fill="%[2]s" />
		<g transform="translate(10, 10) scale(0.9)">
			<text x="50%%" y="%[3]d" font-family="monospace" font-weight="bold" font-size="28" fill="white" letter-spacing="2">
				%[4]s
			</text>
		</g>
	</svg>`, size, bgColor, yPos, textBlock.String())
}

var dotFont = map[rune][]string{
	'A': {"01110", "10001", "10001", "11111", "10001", "10001", "10001"},
	'B': {"11110", "10001", "10001", "11110", "10001", "10001", "11111"},
	'C': {"01111", "10000", "10000", "10000", "10000", "10000", "01111"},
	'D': {"11110", "10001", "10001", "10001", "10001", "10001", "11110"},
	'E': {"11111", "10000", "10000", "11110", "10000", "10000", "11111"},
	'F': {"11111", "10000", "10000", "11110", "10000", "10000", "10000"},
	'G': {"01111", "10000", "10000", "10111", "10001", "10001", "01111"},
	'H': {"10001", "10001", "10001", "11111", "10001", "10001", "10001"},
	'I': {"01110", "00100", "00100", "00100", "00100", "00100", "01110"},
	'J': {"00111", "00001", "00001", "00001", "00001", "10001", "01110"},
	'K': {"10001", "10010", "10100", "11000", "10100", "10010", "10001"},
	'L': {"10000", "10000", "10000", "10000", "10000", "10000", "11111"},
	'M': {"10001", "11011", "10101", "10101", "10001", "10001", "10001"},
	'N': {"10001", "11001", "10101", "10011", "10001", "10001", "10001"},
	'O': {"01110", "10001", "10001", "10001", "10001", "10001", "01110"},
	'P': {"11110", "10001", "10001", "11110", "10000", "10000", "10000"},
	'Q': {"01110", "10001", "10001", "10001", "10001", "10010", "01101"},
	'R': {"11110", "10001", "10001", "11110", "10100", "10010", "10001"},
	'S': {"01111", "10000", "10000", "01110", "00001", "00001", "11110"},
	'T': {"11111", "00100", "00100", "00100", "00100", "00100", "00100"},
	'U': {"10001", "10001", "10001", "10001", "10001", "10001", "01110"},
	'V': {"10001", "10001", "10001", "10001", "10001", "01010", "00100"},
	'W': {"10001", "10001", "10001", "10101", "10101", "11011", "10001"},
	'X': {"10001", "10001", "01010", "00100", "01010", "10001", "10001"},
	'Y': {"10001", "10001", "10001", "01010", "00100", "00100", "00100"},
	'Z': {"11111", "00001", "00010", "00100", "01000", "10000", "11111"},
	'?': {"01110", "10001", "00010", "00100", "00100", "00000", "00100"},
}

func generateDotMatrix(name string, size string) string {

	mainColor, _ := generateColor(name)

	initials := getInitials(name)
	if len(initials) > 2 {
		initials = initials[:2]
	}

	var svgContent strings.Builder

	dotRadius := 4
	spacing := 12
	letterSpacing := 10
	startX := 20
	startY := 35

	currentX := startX

	for _, char := range initials {
		grid, ok := dotFont[char]
		if !ok {
			grid = dotFont['?']
		}

		for row := range 7 {
			for col := range 5 {
				cx := currentX + (col * spacing)
				cy := startY + (row * spacing)

				isLit := grid[row][col] == '1'

				if isLit {

					fmt.Fprintf(&svgContent, `<circle cx="%d" cy="%d" r="%d" fill="%s" opacity="0.4" />`,
						cx, cy, dotRadius+2, mainColor)
					fmt.Fprintf(&svgContent, "<circle cx=\"%d\" cy=\"%d\" r=\"%d\" fill=\"%s\" />\n", cx, cy, dotRadius, mainColor)
				} else {
					fmt.Fprintf(&svgContent, `<circle cx="%d" cy="%d" r="%d" fill="#333" opacity="0.3" />`,
						cx, cy, dotRadius)
				}
			}
		}
		currentX += (5 * spacing) + letterSpacing
	}

	return fmt.Sprintf(`
	<svg xmlns="http://www.w3.org/2000/svg" width="%[1]s" height="%[1]s" viewBox="0 0 170 170">
		<rect width="100%%" height="100%%" fill="#111111" />
		<g transform="translate(8.5, 8.5) scale(0.9)">
			%[2]s
		</g>
	</svg>`, size, svgContent.String())
}

var blockFont = map[rune][]string{
	'A': {"01110", "10001", "11111", "10001", "10001"},
	'B': {"11110", "10001", "11110", "10001", "11110"},
	'C': {"01111", "10000", "10000", "10000", "01111"},
	'D': {"11110", "10001", "10001", "10001", "11110"},
	'E': {"11111", "10000", "11110", "10000", "11111"},
	'F': {"11111", "10000", "11110", "10000", "10000"},
	'G': {"01111", "10000", "10011", "10001", "01111"},
	'H': {"10001", "10001", "11111", "10001", "10001"},
	'I': {"01110", "00100", "00100", "00100", "01110"},
	'J': {"00111", "00001", "00001", "10001", "01110"},
	'K': {"10001", "10010", "11000", "10010", "10001"},
	'L': {"10000", "10000", "10000", "10000", "11111"},
	'M': {"10001", "11011", "10101", "10001", "10001"},
	'N': {"10001", "11001", "10101", "10011", "10001"},
	'O': {"01110", "10001", "10001", "10001", "01110"},
	'P': {"11110", "10001", "11110", "10000", "10000"},
	'Q': {"01110", "10001", "10001", "10010", "01101"},
	'R': {"11110", "10001", "11110", "10010", "10001"},
	'S': {"01111", "10000", "01110", "00001", "11110"},
	'T': {"11111", "00100", "00100", "00100", "00100"},
	'U': {"10001", "10001", "10001", "10001", "01110"},
	'V': {"10001", "10001", "10001", "01010", "00100"},
	'W': {"10001", "10001", "10101", "11011", "10001"},
	'X': {"10001", "01010", "00100", "01010", "10001"},
	'Y': {"10001", "01010", "00100", "00100", "00100"},
	'Z': {"11111", "00010", "00100", "01000", "11111"},
	'?': {"01110", "10001", "00100", "00000", "00100"},
}

func generateTerminalBlock(name string, size string) string {
	textColor, _ := generateColor(name)
	initials := getInitials(name)
	if len(initials) > 2 {
		initials = initials[:2]
	}

	var blocks strings.Builder

	blockSize := 20
	gap := 2

	startX := 40
	startY := 60
	letterSpacing := 20

	currentX := startX

	for _, char := range initials {
		grid, ok := blockFont[char]
		if !ok {
			grid = blockFont['?']
		}

		for row := range 5 {
			for col := range 5 {
				if grid[row][col] == '1' {
					x := currentX + (col * (blockSize + gap))
					y := startY + (row * (blockSize + gap))

					fmt.Fprintf(&blocks, `<rect x="%d" y="%d" width="%d" height="%d" fill="#000" opacity="0.5" />`,
						x+4, y+4, blockSize, blockSize)

					fmt.Fprintf(&blocks, `<rect x="%d" y="%d" width="%d" height="%d" fill="%s" />`,
						x, y, blockSize, blockSize, textColor)
				}
			}
		}
		currentX += (5 * (blockSize + gap)) + letterSpacing
	}
	cursorX := currentX
	fmt.Fprintf(&blocks, `<rect x="%d" y="%d" width="%d" height="%d" fill="%s" opacity="0.7" />`,
		cursorX, startY+(4*(blockSize+gap)), blockSize, blockSize, textColor)

	scanlines := `
	<defs>
		<pattern id="scanlines" patternUnits="userSpaceOnUse" width="10" height="4">
			<rect width="10" height="2" fill="#000" opacity="0.3" />
		</pattern>
	</defs>
	<rect width="100%" height="100%" fill="url(#scanlines)" />`

	return fmt.Sprintf(`
	<svg xmlns="http://www.w3.org/2000/svg" width="%[1]s" height="%[1]s" viewBox="0 0 350 350">
		<rect width="100%%" height="100%%" fill="#1a1b26" />
		%[3]s
		<g transform="translate(17.5, 17.5) scale(0.9)">
			%[2]s
		</g>
	</svg>`, size, blocks.String(), scanlines)
}

var bauhausPalette = []string{"#FFB900", "#E74856", "#0078D7", "#0099BC", "#7A7574", "#FF4343", "#00CC6A", "#8E8CD8"}

func generateBauhaus(name string, size string) string {
	hash := md5.Sum([]byte(name))
	bgIndex := int(hash[0]) % len(bauhausPalette)
	bgColor := bauhausPalette[bgIndex]

	var shapes strings.Builder

	numShapes := 3 + (int(hash[1]) % 3)

	for i := 0; i < numShapes; i++ {
		h1 := int(hash[i+2])
		h2 := int(hash[i+5])
		h3 := int(hash[i+8])

		color := bauhausPalette[(bgIndex+i+1)%len(bauhausPalette)]

		shapeType := h1 % 3

		x := h2 % 100
		y := h3 % 100
		w := 20 + (h1 % 60)

		opacity := 0.5 + (float64(h2%5) / 10.0)

		switch shapeType {
		case 0:
			fmt.Fprintf(&shapes, `<circle cx="%d" cy="%d" r="%d" fill="%s" opacity="%.2f" />`,
				x, y, w/2, color, opacity)
		case 1:
			rotation := 0
			if h1%2 == 0 {
				rotation = 45
			}
			fmt.Fprintf(&shapes, `<rect x="%d" y="%d" width="%d" height="%d" fill="%s" opacity="%.2f" transform="rotate(%d %d %d)" />`,
				x-w/2, y-w/2, w, w, color, opacity, rotation, x, y)
		case 2:
			p1 := fmt.Sprintf("%d,%d", x, y-(w/2))
			p2 := fmt.Sprintf("%d,%d", x-(w/2), y+(w/2))
			p3 := fmt.Sprintf("%d,%d", x+(w/2), y+(w/2))
			fmt.Fprintf(&shapes, `<polygon points="%s %s %s" fill="%s" opacity="%.2f" />`,
				p1, p2, p3, color, opacity)
		}
	}

	return fmt.Sprintf(`
	<svg xmlns="http://www.w3.org/2000/svg" width="%[1]s" height="%[1]s" viewBox="0 0 100 100">
		<rect width="100%%" height="100%%" fill="%[2]s" />
		<g transform="translate(5, 5) scale(0.9)">
			%[3]s
		</g>
	</svg>`, size, bgColor, shapes.String())
}

func generateRing(name string, size string) string {
	hash := md5.Sum([]byte(name))
	c1, _ := generateColor(name)
	c2, _ := generateColor(name + "2")
	c3, _ := generateColor(name + "3")

	angle := int(hash[0]) % 360

	gradID := fmt.Sprintf("grad-%x", hash[0:3])

	return fmt.Sprintf(`
	<svg xmlns="http://www.w3.org/2000/svg" width="%[1]s" height="%[1]s" viewBox="0 0 100 100">
		<defs>
			<linearGradient id="%[5]s" x1="0%%" y1="0%%" x2="100%%" y2="100%%" gradientTransform="rotate(%[6]d .5 .5)">
				<stop offset="0%%" stop-color="%[2]s" />
				<stop offset="50%%" stop-color="%[3]s" />
				<stop offset="100%%" stop-color="%[4]s" />
			</linearGradient>
		</defs>
		<g transform="translate(5, 5) scale(0.9)">
			<circle cx="50" cy="50" r="50" fill="url(#%[5]s)" />
		</g>
	</svg>`, size, c1, c2, c3, gradID, angle)
}
func generateBeam(name string, size string) string {
	hash := md5.Sum([]byte(name))
	bgColor := "#0a0a0a"
	accentColor, _ := generateColor(name)

	var svgContent strings.Builder

	type Point struct{ X, Y int }
	points := make([]Point, 6)

	for i := range 6 {
		points[i] = Point{
			X: 10 + int(hash[i])%80,
			Y: 10 + int(hash[i+6])%80,
		}
		fmt.Fprintf(&svgContent, `<circle cx="%d" cy="%d" r="3" fill="%s" />`, points[i].X, points[i].Y, accentColor)
	}
	for i := range 6 {
		for j := i + 1; j < 6; j++ {
			p1 := points[i]
			p2 := points[j]
			distSq := (p1.X-p2.X)*(p1.X-p2.X) + (p1.Y-p2.Y)*(p1.Y-p2.Y)

			if distSq < 3600 {
				opacity := 1.0 - (float64(distSq) / 3600.0)
				fmt.Fprintf(&svgContent, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="1" opacity="%.2f" />`, p1.X, p1.Y, p2.X, p2.Y, accentColor, opacity)
			}
		}
	}

	return fmt.Sprintf(`
	<svg xmlns="http://www.w3.org/2000/svg" width="%[1]s" height="%[1]s" viewBox="0 0 100 100">
		<rect width="100%%" height="100%%" fill="%[2]s" />
		<g transform="translate(5, 5) scale(0.9)">
			%[3]s
		</g>
	</svg>`, size, bgColor, svgContent.String())
}

func generateMarble(name string, size string) string {
	hash := md5.Sum([]byte(name))
	c1, _ := generateColor(name)
	c2, _ := generateColor(name + "x")

	freq := 0.005 + (float64(hash[0])/255.0)*0.02

	octaves := 1 + (int(hash[1]) % 4)

	return fmt.Sprintf(`
	<svg xmlns="http://www.w3.org/2000/svg" width="%[1]s" height="%[1]s" viewBox="0 0 100 100">
		<defs>
			<filter id="liquid">
				<feTurbulence type="fractalNoise" baseFrequency="%[4].4f" numOctaves="%[5]d" result="noise" />
				<feDiffuseLighting in="noise" lighting-color="white" surfaceScale="2">
					<feDistantLight azimuth="45" elevation="60" />
				</feDiffuseLighting>
			</filter>
			<linearGradient id="grad" x1="0%%" y1="0%%" x2="100%%" y2="100%%">
				<stop offset="0%%" stop-color="%[2]s" />
				<stop offset="100%%" stop-color="%[3]s" />
			</linearGradient>
		</defs>
		<g transform="translate(5, 5) scale(0.9)">
			<rect width="100%%" height="100%%" fill="url(#grad)" />
			<rect width="100%%" height="100%%" fill="transparent" filter="url(#liquid)" opacity="0.5" style="mix-blend-mode: overlay;" />
		</g>
	</svg>`, size, c1, c2, freq, octaves)
}

func generateGlitch(name string, size string) string {
	hash := md5.Sum([]byte(name))
	initials := getInitials(name)

	bgColor := "#0f0f0f"

	var glitchLines strings.Builder
	for i := 0; i < 5; i++ {
		y := int(hash[i]) % 100
		h := int(hash[i+5])%5 + 1
		w := int(hash[i+10])%50 + 20
		x := int(hash[i+2]) % 80

		glitchLines.WriteString(fmt.Sprintf(
			`<rect x="%d" y="%d" width="%d" height="%d" fill="white" opacity="0.1" />`,
			x, y, w, h,
		))
	}

	return fmt.Sprintf(`
	<svg xmlns="http://www.w3.org/2000/svg" width="%[1]s" height="%[1]s" viewBox="0 0 100 100">
		<rect width="100%%" height="100%%" fill="%[2]s" />
		<g transform="translate(5, 5) scale(0.9)">
			<text x="48" y="55" dominant-baseline="middle" text-anchor="middle" font-family="Arial Black, sans-serif" font-weight="900" font-size="50" fill="#00ffff" opacity="0.8" style="mix-blend-mode: screen;">
				%[3]s
			</text>
			
			<text x="52" y="55" dominant-baseline="middle" text-anchor="middle" font-family="Arial Black, sans-serif" font-weight="900" font-size="50" fill="#ff0000" opacity="0.8" style="mix-blend-mode: screen;">
				%[3]s
			</text>
			
			<text x="50" y="55" dominant-baseline="middle" text-anchor="middle" font-family="Arial Black, sans-serif" font-weight="900" font-size="50" fill="#ffffff">
				%[3]s
			</text>
			
			%[4]s
		</g>
	</svg>`, size, bgColor, initials, glitchLines.String())
}

func generateSunset(name string, size string) string {
	hash := md5.Sum([]byte(name))

	var skyTop, skyBot string
	mood := hash[0] % 3
	switch mood {
	case 0:
		skyTop, skyBot = "#3e1c6b", "#ff8a5c"
	case 1:
		skyTop, skyBot = "#29b6f6", "#fff9c4"
	default:
		skyTop, skyBot = "#0d1b2a", "#415a77"
	}
	sunX := 20 + (int(hash[1]) % 60)
	sunY := 20 + (int(hash[2]) % 30)
	sunColor := "#ffffff"
	if mood == 0 {
		sunColor = "#ffeb3b"
	}
	var mountains strings.Builder
	mountains.WriteString("M 0 100 L 0 60 ")

	for x := 0; x <= 100; x += 5 {
		y := 60.0 + math.Sin(float64(x)*0.1+float64(hash[3]))*15.0
		if x%10 == 0 {
			y -= 5
		}

		fmt.Fprintf(&mountains, "L %d %.2f ", x, y)
	}
	mountains.WriteString("L 100 100 Z")

	mountColor := "#1a1a1a"
	if mood == 1 {
		mountColor = "#4caf50"
	}

	return fmt.Sprintf(`
	<svg xmlns="http://www.w3.org/2000/svg" width="%[1]s" height="%[1]s" viewBox="0 0 100 100">
		<defs>
			<linearGradient id="sky" x1="0%%" y1="0%%" x2="0%%" y2="100%%">
				<stop offset="0%%" stop-color="%[2]s" />
				<stop offset="100%%" stop-color="%[3]s" />
			</linearGradient>
		</defs>
		<rect width="100%%" height="100%%" fill="url(#sky)" />
		<g transform="translate(5, 5) scale(0.9)">
			<circle cx="%d" cy="%d" r="8" fill="%s" opacity="0.9" />
			
			<path d="%s" fill="%s" opacity="0.9" />
		</g>
	</svg>`, size, skyTop, skyBot, sunX, sunY, sunColor, mountains.String(), mountColor)
}
func generateSmile(name string, size string) string {
	hash := md5.Sum([]byte(name))
	skinTones := []string{"#FFDFC4", "#F0C8C9", "#E5B99F", "#8D5524", "#C68642", "#FFDCB1", "#E0AC69", "#B9D2B1", "#A8C8E8"}
	skinColor := skinTones[int(hash[0])%len(skinTones)]
	eyeType := int(hash[1]) % 3
	mouthType := int(hash[2]) % 4
	hasBlush := int(hash[3])%2 == 0

	var features strings.Builder

	switch eyeType {
	case 0:
		features.WriteString(`<circle cx="35" cy="45" r="5" fill="#333" /><circle cx="65" cy="45" r="5" fill="#333" />`)
	case 1:
		features.WriteString(`<path d="M 30 45 Q 35 40 40 45" stroke="#333" stroke-width="3" fill="none" />`)
		features.WriteString(`<path d="M 60 45 Q 65 40 70 45" stroke="#333" stroke-width="3" fill="none" />`)
	case 2:
		features.WriteString(`<circle cx="35" cy="45" r="5" fill="#333" />`)
		features.WriteString(`<rect x="60" y="44" width="10" height="2" fill="#333" />`)
	}

	switch mouthType {
	case 0:
		features.WriteString(`<path d="M 35 65 Q 50 75 65 65" stroke="#333" stroke-width="3" fill="none" stroke-linecap="round" />`)
	case 1:
		features.WriteString(`<path d="M 35 65 Q 50 80 65 65 Z" fill="#fff" stroke="#333" stroke-width="2" />`)
	case 2:
		features.WriteString(`<line x1="40" y1="70" x2="60" y2="70" stroke="#333" stroke-width="3" stroke-linecap="round" />`)
	case 3:
		features.WriteString(`<circle cx="50" cy="70" r="6" stroke="#333" stroke-width="3" fill="none" />`)
	}

	if hasBlush {
		features.WriteString(`<circle cx="30" cy="55" r="5" fill="#ff0000" opacity="0.2" />`)
		features.WriteString(`<circle cx="70" cy="55" r="5" fill="#ff0000" opacity="0.2" />`)
	}

	return fmt.Sprintf(`
	<svg xmlns="http://www.w3.org/2000/svg" width="%[1]s" height="%[1]s" viewBox="0 0 100 100">
		<g transform="translate(5, 5) scale(0.9)">
			<circle cx="50" cy="50" r="45" fill="%[2]s" />
			%[3]s
		</g>
	</svg>`, size, skinColor, features.String())
}

func generateCircuit(name string, size string) string {
	hash := md5.Sum([]byte(name))

	boardColors := []string{"#004d40", "#1a237e", "#212121", "#1b5e20"}
	bgColor := boardColors[int(hash[0])%len(boardColors)]
	traceColor := "#ffd700"

	var traces strings.Builder

	for i := range 5 {

		x1 := 10 + (int(hash[i]) % 80)
		y1 := 10 + (int(hash[i+5]) % 80)
		x2 := 10 + (int(hash[i+2]) % 80)
		y2 := 10 + (int(hash[i+7]) % 80)

		fmt.Fprintf(&traces, `<path d="M %d %d L %d %d L %d %d" stroke="%s" stroke-width="2" fill="none" opacity="0.8" />`,
			x1, y1, x2, y1, x2, y2, traceColor)
		traces.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="3" fill="%s" />`, x1, y1, traceColor))
		traces.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="3" fill="%s" />`, x2, y2, traceColor))
	}

	return fmt.Sprintf(`
	<svg xmlns="http://www.w3.org/2000/svg" width="%[1]s" height="%[1]s" viewBox="0 0 100 100">
		<rect width="100%%" height="100%%" fill="%[2]s" />
		<g transform="translate(5, 5) scale(0.9)">
			%[3]s
		</g>
	</svg>`, size, bgColor, traces.String())
}

func generatePixel(name string, size string) string {
	baseHex, hash := generateColor(name)
	var topPattern strings.Builder
	if hash[0]%2 == 0 {
		topPattern.WriteString(`<path d="M 50 30 L 70 40 L 50 50 L 30 40 Z" fill="rgba(255,255,255,0.3)" />`)
	}

	return fmt.Sprintf(`
	<svg xmlns="http://www.w3.org/2000/svg" width="%[1]s" height="%[1]s" viewBox="0 0 100 100">
		<rect width="100%%" height="100%%" fill="#f0f0f0" />
		<g transform="translate(5, 5) scale(0.9)">
			<path d="M 20 35 L 50 50 L 50 80 L 20 65 Z" fill="%[2]s" />
			
			<path d="M 50 50 L 80 35 L 80 65 L 50 80 Z" fill="%[2]s" />
			<path d="M 50 50 L 80 35 L 80 65 L 50 80 Z" fill="black" opacity="0.2" />
			
			<path d="M 50 20 L 80 35 L 50 50 L 20 35 Z" fill="%[2]s" />
			<path d="M 50 20 L 80 35 L 50 50 L 20 35 Z" fill="white" opacity="0.3" />
			
			%[3]s
		</g>
	</svg>`, size, baseHex, topPattern.String())
}

type GeoJSON struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

type Feature struct {
	ID         string     `json:"id"`
	Properties Properties `json:"properties"`
	Geometry   Geometry   `json:"geometry"`
}

type Properties struct {
	Name string `json:"n"`
}

type Geometry struct {
	Type        string        `json:"type"`
	Coordinates [][][]float64 `json:"coordinates"`
}

func normalizeGeoJSON(coords [][][]float64) [][][]float64 {
	minX, maxX := math.MaxFloat64, -math.MaxFloat64
	minY, maxY := math.MaxFloat64, -math.MaxFloat64

	for _, line := range coords {
		for _, point := range line {
			x, y := point[0], point[1]
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
			if y < minY {
				minY = y
			}
			if y > maxY {
				maxY = y
			}
		}
	}

	width := maxX - minX
	height := maxY - minY
	if width == 0 {
		width = 1
	}
	if height == 0 {
		height = 1
	}
	scaleX := 60.0 / width
	scaleY := 60.0 / height
	scale := math.Min(scaleX, scaleY)

	offsetX := (100.0 - (width * scale)) / 2.0
	offsetY := (100.0 - (height * scale)) / 2.0

	normalized := make([][][]float64, len(coords))
	for i, line := range coords {
		normalized[i] = make([][]float64, len(line))
		for j, point := range line {
			newX := ((point[0] - minX) * scale) + offsetX
			newY := 100 - (((point[1] - minY) * scale) + offsetY)
			normalized[i][j] = []float64{newX, newY}
		}
	}
	return normalized
}

var constellationData GeoJSON

//go:embed constellations.json
var constellationFile embed.FS

func loadGeoJSON() {
	data, err := constellationFile.ReadFile("constellations.json")
	if err != nil {
		log.Fatalf("Failed to read embedded json: %v", err)
	}

	err = json.Unmarshal(data, &constellationData)
	if err != nil {
		log.Fatalf("Failed to parse json: %v", err)
	}

	fmt.Printf("✅ Loaded %d constellations\n", len(constellationData.Features))
}

func generateGeoJSONAvatar(name string, size string) string {
	hash := md5.Sum([]byte(name))
	idx := int(hash[0]) % len(constellationData.Features)
	feature := constellationData.Features[idx]
	lines := normalizeGeoJSON(feature.Geometry.Coordinates)

	bgStart := "#1e1b4b"
	bgEnd := "#020617"
	var svgContent strings.Builder

	for i := range 40 {
		x := (int(hash[i%16]) * (i + 3)) % 100
		y := (int(hash[(i+2)%16]) * (i + 5)) % 100
		op := float64((int(hash[i%16])%5)+1) / 10.0
		fmt.Fprintf(&svgContent, `<circle cx="%d" cy="%d" r="0.4" fill="white" opacity="%.1f" />`, x, y, op)
	}
	for _, line := range lines {
		for i := 0; i < len(line)-1; i++ {
			p1 := line[i]
			p2 := line[i+1]

			fmt.Fprintf(&svgContent, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#93c5fd" stroke-width="0.5" opacity="0.8" />`,
				p1[0], p1[1], p2[0], p2[1])

			fmt.Fprintf(&svgContent, `<circle cx="%.1f" cy="%.1f" r="1.5" fill="white" />`, p1[0], p1[1])
			fmt.Fprintf(&svgContent, `<circle cx="%.1f" cy="%.1f" r="3" fill="#38bdf8" opacity="0.2" />`, p1[0], p1[1])
		}

		lastP := line[len(line)-1]
		fmt.Fprintf(&svgContent, `<circle cx="%.1f" cy="%.1f" r="1.5" fill="white" />`, lastP[0], lastP[1])
	}

	return fmt.Sprintf(`
	<svg xmlns="http://www.w3.org/2000/svg" width="%[1]s" height="%[1]s" viewBox="0 0 100 100">
		<defs>
			<radialGradient id="grad" cx="50%%" cy="50%%" r="80%%">
				<stop offset="0%%" stop-color="%[2]s" />
				<stop offset="100%%" stop-color="%[3]s" />
			</radialGradient>
		</defs>
		<rect width="100%%" height="100%%" fill="url(#grad)" />
		%[4]s
		<text x="50" y="90" text-anchor="middle" font-family="Times New Roman" font-weight="bold" font-size="6" fill="#7dd3fc" letter-spacing="0.5">
			%[5]s
		</text>
	</svg>`, size, bgStart, bgEnd, svgContent.String(), strings.ToUpper(feature.Properties.Name))
}

func generateAvatarHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	avatarType := query.Get("type")
	size := query.Get("size")
	color := query.Get("color")
	name := query.Get("name")
	if size == "" {
		size = "100"
	}
	if avatarType == "" {
		avatarType = "avatar"
	}
	if name == "" {
		name = "User"
	}

	if color == "" {
		color, _ = generateColor(name)
	}

	cacheKey := fmt.Sprintf("%s:%s:%s:%s", name, avatarType, size, color)

	if cachedSVG, found := cache.Get(cacheKey); found {
		w.Header().Set("X-Cache", "HIT")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(cachedSVG))
		return
	}

	var avatarContent string

	switch avatarType {
	case "avatar":
		initials := getInitials(name)
		avatarContent = fmt.Sprintf(`
		<svg xmlns="http://www.w3.org/2000/svg" width="%[1]s" height="%[1]s" viewBox="0 0 100 100">
			<g transform="translate(5, 5) scale(0.9)">
				<circle cx="50" cy="50" r="50" fill="%[2]s" />
				<text x="50" y="55" dominant-baseline="middle" text-anchor="middle" font-family="Arial, sans-serif" font-size="40" fill="#ffffff">%[3]s</text>
			</g>
		</svg>`, size, color, initials)
	case "gravatar":
		identiconRects := generateIdenticon(name)
		avatarContent = fmt.Sprintf(`
		<svg xmlns="http://www.w3.org/2000/svg" width="%[1]s" height="%[1]s" viewBox="0 0 250 250">
			<rect width="100%%" height="100%%" fill="#11011D" />
			<g transform="translate(20, 10) scale(0.8)">
				%[3]s
			</g>	
		</svg>`, size, color, identiconRects)
	case "dither":
		svgBody := generateDitheredAvatar(name)
		avatarContent = strings.Replace(svgBody, `width="100%" height="100%"`, fmt.Sprintf(`width="%s" height="%s"`, size, size), 1)
	case "ascii":
		avatarContent = generateAsciiRobot(name, size)
	case "dotmatrix":
		avatarContent = generateDotMatrix(name, size)
	case "terminal":
		avatarContent = generateTerminalBlock(name, size)
	case "bauhaus":
		avatarContent = generateBauhaus(name, size)
	case "ring":
		avatarContent = generateRing(name, size)
	case "beam":
		avatarContent = generateBeam(name, size)
	case "marble":
		avatarContent = generateMarble(name, size)
	case "glitch":
		avatarContent = generateGlitch(name, size)
	case "sunset":
		avatarContent = generateSunset(name, size)
	case "smile":
		avatarContent = generateSmile(name, size)
	case "circuit":
		avatarContent = generateCircuit(name, size)
	case "pixel":
		avatarContent = generatePixel(name, size)
	case "constellation":
		loadGeoJSON()
		avatarContent = generateGeoJSONAvatar(name, size)
	default:
		http.Error(w, "Invalid avatar type. Use 'avatar' or 'gravatar'.", http.StatusBadRequest)
		return
	}

	// Store in cache
	cache.Add(cacheKey, avatarContent)

	w.Header().Set("X-Cache", "MISS")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(avatarContent))
}
