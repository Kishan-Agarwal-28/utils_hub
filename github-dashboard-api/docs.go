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
    <title>GitHub Retro Stats API</title>
    <style>
        :root {
            --bg: #151515;
            --card-bg: #1c1c1c;
            --text: #c5c8c6;
            --green: #b5bd68;
            --blue: #81a2be;
            --purple: #b294bb;
            --yellow: #f0c674;
            --orange: #de935f;
            --red: #cc6666;
            --dim: #373b41;
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
        h1, h2, h3 { color: var(--green); text-transform: uppercase; letter-spacing: 1px; }
        h1 { border-bottom: 2px solid var(--dim); padding-bottom: 10px; margin-bottom: 30px; }
        
        /* Code Block */
        .code-block {
            background: var(--card-bg);
            border: var(--border);
            padding: 15px;
            border-radius: 6px;
            overflow-x: auto;
            color: var(--orange);
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
        .copy-btn:hover { background: var(--green); color: var(--bg); }

        /* Tables */
        table { width: 100%; border-collapse: collapse; margin-bottom: 30px; }
        th, td { text-align: left; padding: 12px; border-bottom: var(--border); }
        th { color: var(--blue); text-transform: uppercase; font-size: 0.9em; }
        td code { color: var(--yellow); background: #2d2d2d; padding: 2px 6px; border-radius: 4px; }

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
        }
        .medal-title { color: var(--purple); font-weight: bold; margin-bottom: 5px; display: block; }
        .medal-desc { font-size: 0.9em; opacity: 0.8; }
        .color-dot { display: inline-block; width: 10px; height: 10px; border-radius: 50%; margin-right: 5px; }

        /* Footer */
        footer { margin-top: 50px; border-top: var(--border); padding-top: 20px; text-align: center; opacity: 0.5; font-size: 0.8em; }
        a { color: var(--blue); text-decoration: none; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="container">
        <h1>GitHub Retro Stats v1.0</h1>
        
        <p>A high-precision, retro-styled dashboard generator for your GitHub profile. Includes strict 24h activity tracking, timezone adjustments, and a military-style ribbon rack for achievements.</p>

        <h2>⚡ Quick Setup</h2>
        <div class="code-block">
            <span id="url">/api/github-stats?username=YOUR_NAME</span>
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
                    <td>Your GitHub username (case-insensitive).</td>
                </tr>
                <tr>
                    <td><code>timezone</code></td>
                    <td><code>UTC</code></td>
                    <td>Adjusts "Time of Day" graphs. Supports abbreviations (<code>IST</code>, <code>EST</code>) or IANA (<code>Asia/Kolkata</code>).</td>
                </tr>
                <tr>
                    <td><code>format</code></td>
                    <td><code>svg</code></td>
                    <td>Output format. Use <code>png</code> for social media sharing (LinkedIn/Twitter).</td>
                </tr>
            </tbody>
        </table>

        <h2>🎖️ Service Ribbons Guide</h2>
        <p>Ribbons are automatically awarded based on account milestones. Up to 9 are displayed.</p>

        <div class="medal-grid">
            <div class="medal-card">
                <span class="medal-title" style="color:#b5bd68">Years of Service</span>
                <div class="medal-desc">
                    Based on account age.<br>
                    <span style="color:#b5bd68">● Green</span>: 1+ Years<br>
                    <span style="color:#cc6666">● Red Stripes</span>: 5+ Years<br>
                    <span style="color:#b294bb">● Purple/Gold</span>: 10+ Years (Veteran)
                </div>
            </div>

            <div class="medal-card">
                <span class="medal-title" style="color:#f0c674">Stars (Popularity)</span>
                <div class="medal-desc">
                    Total stars earned across all repos.<br>
                    <span style="color:#f0c674">● Yellow</span>: 10+ Stars<br>
                    <span style="color:#f0c674">● Gold/Red</span>: 1,000+ Stars (General)
                </div>
            </div>

            <div class="medal-card">
                <span class="medal-title" style="color:#cc6666">Commits (Effort)</span>
                <div class="medal-desc">
                    Annual contribution volume.<br>
                    <span style="color:#cc6666">● Red</span>: 100+ Commits<br>
                    <span style="color:#cc6666">● Red/Black</span>: 1,000+ Commits (Machine)
                </div>
            </div>

            <div class="medal-card">
                <span class="medal-title" style="color:#b294bb">Followers (Influence)</span>
                <div class="medal-desc">
                    Community reach.<br>
                    <span style="color:#81a2be">● Blue</span>: 100+ Followers<br>
                    <span style="color:#b294bb">● Purple</span>: 1,000+ Followers (Celeb)
                </div>
            </div>
             <div class="medal-card">
                <span class="medal-title" style="color:#81a2be">Special Honors</span>
                <div class="medal-desc">
                    Unique achievements.<br>
                    <span style="color:#81a2be">● Polyglot</span>: 4+ Languages used<br>
                    <span style="color:#5d4037">● Ancient One</span>: Joined before 2010
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