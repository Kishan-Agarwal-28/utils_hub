package main

import "net/http"

func documentationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	html := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>NPM Retro Stats API</title>
    <style>
        :root {
            --bg: #151515;
            --card-bg: #1c1c1c;
            --text: #e0e0e0;
            --npm-red: #cb3837;
            --white: #ffffff;
            --grey: #333333;
            --dim: #282828;
            --gold: #ffb300;
            --purple: #7b1fa2;
            --border: 1px solid var(--dim);
        }
        body {
            background-color: var(--bg);
            color: var(--text);
            font-family: 'Courier New', Courier, monospace;
            line-height: 1.6;
            margin: 0;
            padding: 20px;
            display: flex;
            justify-content: center;
        }
        .container {
            max-width: 900px;
            width: 100%;
        }
        h1, h2, h3 { color: var(--white); text-transform: uppercase; letter-spacing: 1px; }
        h1 { border-bottom: 2px solid var(--npm-red); padding-bottom: 10px; margin-bottom: 30px; }
        h1 span { color: var(--npm-red); }

        /* Code Block */
        .code-block {
            background: var(--card-bg);
            border: var(--border);
            padding: 15px;
            border-radius: 6px;
            overflow-x: auto;
            color: var(--npm-red);
            font-weight: bold;
            margin-bottom: 20px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .copy-btn {
            background: var(--dim);
            color: var(--text);
            border: none;
            padding: 5px 10px;
            cursor: pointer;
            font-family: inherit;
            border-radius: 4px;
        }
        .copy-btn:hover { background: var(--npm-red); color: var(--white); }

        /* Tables */
        table { width: 100%; border-collapse: collapse; margin-bottom: 30px; }
        th, td { text-align: left; padding: 12px; border-bottom: var(--border); }
        th { color: var(--npm-red); text-transform: uppercase; font-size: 0.9em; }
        td code { color: var(--white); background: #2d2d2d; padding: 2px 6px; border-radius: 4px; }

        /* Medal Grid */
        .medal-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
            gap: 20px;
            margin-top: 20px;
        }
        .medal-card {
            background: var(--card-bg);
            border: var(--border);
            padding: 15px;
            border-radius: 6px;
            position: relative;
            overflow: hidden;
        }
        .medal-card::before {
            content: "";
            position: absolute;
            top: 0; left: 0; width: 4px; height: 100%;
            background: var(--dim);
        }
        .medal-card.red::before { background: var(--npm-red); }
        .medal-card.gold::before { background: var(--gold); }
        .medal-card.purple::before { background: var(--purple); }

        .medal-title { color: var(--white); font-weight: bold; margin-bottom: 5px; display: block; }
        .medal-desc { font-size: 0.9em; opacity: 0.8; color: #aaa; }
        
        /* Footer */
        footer { margin-top: 50px; border-top: var(--border); padding-top: 20px; text-align: center; opacity: 0.5; font-size: 0.8em; }
        a { color: var(--npm-red); text-decoration: none; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="container">
        <h1>NPM <span>Retro Stats</span> v1.0</h1>
        
        <p>A high-precision, retro-styled dashboard for NPM maintainers. Visualizes annual downloads, package popularity, and publish activity frequency in a "Mission Control" aesthetic.</p>

        <h2>⚡ Quick Setup</h2>
        <div class="code-block">
            <span id="url">https://your-domain.com/api/npm-stats?username=YOUR_NPM_NAME</span>
            <button class="copy-btn" onclick="copyUrl()">COPY</button>
        </div>

        <h2>⚙️ Configuration</h2>
        <table>
            <thead>
                <tr>
                    <th>Parameter</th>
                    <th>Default</th>
                    <th>Description</th>
                </tr>
            </thead>
            <tbody>
                <tr>
                    <td><code>username</code></td>
                    <td>Required</td>
                    <td>Your exact NPM username (e.g., <code>shadcn</code>, <code>chalk</code>).</td>
                </tr>
                <tr>
                    <td><code>format</code></td>
                    <td><code>svg</code></td>
                    <td>Output format. Use <code>png</code> for embedding in LinkedIn/Twitter posts.</td>
                </tr>
            </tbody>
        </table>

        <h2>🎖️ Registry Honors System</h2>
        <p>Ribbons are awarded based on aggregate download volume and publishing history.</p>

        <div class="medal-grid">
            <div class="medal-card red">
                <span class="medal-title">Traffic Volume</span>
                <div class="medal-desc">
                    Based on total yearly downloads.<br>
                    <span style="color:#cb3837">● Red</span>: 10k+ Downloads<br>
                    <span style="color:#cb3837">● Red/White</span>: 1M+ Downloads (The 1M Club)<br>
                    <span style="color:#ffb300">● Gold/Red</span>: 10M+ Downloads (Legend)
                </div>
            </div>

            <div class="medal-card gold">
                <span class="medal-title">Prolific Publisher</span>
                <div class="medal-desc">
                    Based on total packages maintained.<br>
                    <span style="color:#616161">● Grey</span>: 10+ Packages<br>
                    <span style="color:#ffffff">● Black/White</span>: 50+ Packages<br>
                    <span style="color:#ffb300">● Black/Gold</span>: 100+ Packages (Registry God)
                </div>
            </div>

            <div class="medal-card purple">
                <span class="medal-title">Service & Tenancy</span>
                <div class="medal-desc">
                    Years since first package publish.<br>
                    <span style="color:#81a2be">● Blue</span>: 5+ Years (Senior)<br>
                    <span style="color:#7b1fa2">● Purple/Gold</span>: 10+ Years (NPM Veteran)
                </div>
            </div>
        </div>

        <footer>
            Built with Go, Chi & SVG. No external tracking. <br>
            <a href="https://github.com/kishan-agarwal-28">Created by Kishan Agarwal</a>
        </footer>
    </div>

    <script>
        function copyUrl() {
            const url = document.getElementById('url').innerText;
            navigator.clipboard.writeText(url);
            alert('Copied to clipboard!');
        }
    </script>
</body>
</html>
	`
	w.Write([]byte(html))
}