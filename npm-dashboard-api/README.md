# NPM Retro Stats API

> A high-precision, retro-styled dashboard for NPM maintainers.



**NPM Retro Stats** generates dynamic SVG analytics for your NPM profile. It visualizes annual download traffic, package popularity, and publishing activity in a "Mission Control" aesthetic, complete with a military-style ribbon rack for your open-source achievements.

## ⚡ Features

* **Mission Control Aesthetic:** Dark mode interface (`#151515`) with monospaced typography and high-contrast charts.
* **Registry Honors:** Automatic achievement ribbons based on download volume and tenancy.
* **SVG & PNG Support:** Vector graphics for profiles, raster images for social media sharing.
* **Privacy Focused:** No external tracking; acts as a stateless proxy to NPM public APIs.

## 🚀 Quick Setup

```markdown
![My NPM Stats](https://your-domain.com/api/npm-stats?username=YOUR_NPM_USERNAME)

```

### Configuration Parameters

| Parameter | Default | Description |
| --- | --- | --- |
| `username` | **Required** | Your exact NPM username (e.g., `react`, `lodash`). |
| `format` | `svg` | Output format. Use `png` for embedding in LinkedIn/Twitter posts. |

---

## 🎖️ Registry Honors System

The API automatically calculates and awards ribbons based on your registry history.

| Ribbon Category | Tier | Requirement | Color Scheme |
| --- | --- | --- | --- |
| **Traffic Volume** | Active | 10k+ Downloads | <span style="color:#cb3837">● Red</span> |
| *(Yearly Downloads)* | The 1M Club | 1M+ Downloads | <span style="color:#cb3837">● Red/White</span> |
|  | Legend | 10M+ Downloads | <span style="color:#ffb300">● Gold/Red</span> |
| **Prolific Publisher** | Maintainer | 10+ Packages | <span style="color:#616161">● Grey</span> |
| *(Total Packages)* | Architect | 50+ Packages | <span style="color:#ffffff">● Black/White</span> |
|  | Registry God | 100+ Packages | <span style="color:#ffb300">● Black/Gold</span> |
| **Service & Tenancy** | Senior | 5+ Years | <span style="color:#81a2be">● Blue</span> |
| *(First Publish)* | NPM Veteran | 10+ Years | <span style="color:#7b1fa2">● Purple/Gold</span> |

## 🛠️ Deployment

1. **Clone the repository:**
```bash
git clone [https://github.com/yourusername/npm-retro-stats.git](https://github.com/yourusername/npm-retro-stats.git)

```


2. **Run locally:**
```bash
go run main.go

```


3. **Test:**
Open `http://localhost:8080` (or your configured port) to view the documentation.

## 📄 License
This project is licensed under the **MIT License**.
