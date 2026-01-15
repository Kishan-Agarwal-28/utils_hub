
# GitHub Retro Stats API

> A high-precision, retro-styled dashboard generator for your GitHub profile.

**GitHub Retro Stats** generates dynamic SVG analytics for your profile, featuring strict 24-hour activity tracking, timezone adjustments, and a military-style ribbon rack for your coding achievements. It is built with **Go**, **Chi**, and pure **SVG** generation.

## ⚡ Features

* **Retro Aesthetic:** A clean, dark-mode terminal look using a curated color palette (`#151515` background, `#c5c8c6` text).
* **Service Ribbons:** Automatic achievement badges displayed as a military ribbon rack.
* **Precision Tracking:** Strict 24h activity graphs.
* **Timezone Aware:** Adjusts "Time of Day" stats to your local time.
* **Privacy Focused:** No external tracking or database storage.

## 🚀 Quick Setup

Add this markdown snippet to your GitHub profile `README.md`:

```markdown
![My Retro Stats](https://utils.koyeb.app/github/api/github-stats?username=YOUR_USERNAME&timezone=UTC)

```

---

### Configuration Parameters

| Parameter | Default | Description |
| --- | --- | --- |
| `username` | **Required** | Your GitHub username (case-insensitive). |
| `timezone` | `UTC` | Adjusts graphs. Supports abbreviations (`IST`, `EST`) or IANA (`Asia/Kolkata`). |
| `format` | `svg` | Output format. Use `png` for sharing on LinkedIn/Twitter. |

---

## 🎖️ Service Ribbons Guide

The API automatically awards up to **9 ribbons** based on your account history.

| Ribbon Type | Tier | Requirement | Color Scheme |
| --- | --- | --- | --- |
| **Years of Service** | Recruit | 1+ Years | <span style="color:#b5bd68">● Green</span> |
|  | Veteran | 5+ Years | <span style="color:#cc6666">● Red Stripes</span> |
|  | Elite | 10+ Years | <span style="color:#b294bb">● Purple/Gold</span> |
| **Stars (Popularity)** | Rising | 10+ Stars | <span style="color:#f0c674">● Yellow</span> |
|  | General | 1,000+ Stars | <span style="color:#f0c674">● Gold/Red</span> |
| **Commits (Effort)** | Active | 100+ Commits | <span style="color:#cc6666">● Red</span> |
|  | Machine | 1,000+ Commits | <span style="color:#cc6666">● Red/Black</span> |
| **Followers (Influence)** | Known | 100+ Followers | <span style="color:#81a2be">● Blue</span> |
|  | Celeb | 1,000+ Followers | <span style="color:#b294bb">● Purple</span> |
| **Special Honors** | Polyglot | 4+ Languages | <span style="color:#81a2be">● Mixed</span> |
|  | Ancient One | Joined pre-2010 | <span style="color:#5d4037">● Brown/Gold</span> |

## 🛠️ Deployment

This service is designed to run on serverless platforms or containers.

1. **Clone the repo:**
```bash
git clone [https://github.com/yourusername/github-retro-stats.git](https://github.com/yourusername/github-retro-stats.git)

```


2. **Run locally:**
```bash
go run main.go

```


3. **Access:**
Open `http://localhost:8080` to see the documentation and test your stats.

## 📄 License

MIT License. Created by [Kishan Agarwal](https://github.com/kishan-agarwal-28).
