
# 🛠️ Utils Project

> A collection of high-performance, single-purpose microservices and utilities built for modern developers.

The **Utils Project** is a suite of lightweight APIs designed to solve specific problems—from generating assets like avatars and charts to handling complex data logic like calendar events and geolocation search. Each service is built to be stateless, fast, and easy to deploy on serverless platforms (like Koyeb, Fly.io, or Vercel).

## 📂 Service Directory

Each utility is a standalone module. Click the **Documentation** link for specific API usage, installation instructions, and configuration details.

| Service | Tech Stack | Description | Docs |
| --- | --- | --- | --- |
| **Avatar Generator** | Go | Generate deterministic, unique SVG avatars (Identicons, retro, etc.) from any string. | **[View Docs →](./avatars-api/README.md)** |
| **Calendar API** | TypeScript | Create valid `.ics` files and "Add to Calendar" links for Google, Outlook, & Office365. | **[View Docs →](./calendars-api/README.md)** |
| **Charts API** | Go | Render server-side SVG charts (Line, Bar, Pie) via simple Base64-encoded URLs. | **[View Docs →](./charts-api/README.md)** |
| **City Search** | Go | Ultra-fast, cached autocomplete API for world cities, states, and countries. | **[View Docs →](./locations-api/README.md)** |
| **GitHub Stats** | Go | Retro "Mission Control" style dashboard for your GitHub profile README. | **[View Docs →](./github-dashboard-api/README.md)** |
| **NPM Stats** | Go | Visualize NPM download stats and package history with achievement ribbons. | **[View Docs →](./npm-dashboard-api/README.md)** |

---

## 🚀 Getting Started

To work on the entire suite locally:

1. **Clone the Monorepo:**
```bash
git clone https://github.com/yourusername/utils.git
cd utils

```


2. **Navigate to a Service:**
Each service handles its own dependencies.
```bash
cd avatars-api
go run main.go

```


3. **Environment Variables:**
Check the individual `README.md` files for required `.env` variables (e.g., `GITHUB_TOKEN` for the Stats API).

## 🛠️ Tech Stack Overview

We prioritize performance and low overhead.

* **Backend Languages:** Go (Golang) and TypeScript.
* **Web Frameworks:** [Chi](https://github.com/go-chi/chi) (Go) and [Hono](https://hono.dev) (TypeScript).
* **Data:** SQLite (for City Search) and stateless processing for others.
* **Output:** Primarily **SVG** for visual tools and **JSON** for data tools.

## 🤝 Contributing

We welcome contributions to any of the sub-projects!

1. **Check the Guide:** Please read our individual `CONTRIBUTING.md` for code style and PR guidelines.
2. **Pick a Service:** Navigate to the specific service directory you want to improve.
3. **Run Tests:** Ensure you run the specific test suite for that service before submitting.

## 📄 License

This entire repository is licensed under the **MIT License**. See the `LICENSE` file for details.