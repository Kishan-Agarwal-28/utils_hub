# 🌍 City Search API

> A high-performance, low-latency API to search for global cities, states, and countries.

This API provides fast, cached search capabilities for geolocation data. Built with **Go**, **SQLite**, and **Chi**, it features in-memory caching, rate limiting, and compression for optimal performance.

## ⚡ Features

* **Fast Search:** SQL queries optimized with `LIKE` and `ORDER BY length` for relevant results.
* **In-Memory Caching:** Uses LRU caching (1000 items) to serve frequent requests instantly.
* **Rate Limiting:** Protects resources (100 requests/minute per IP).
* **SQLite Backend:** Lightweight and efficient with WAL mode enabled.
* **Developer Ready:** Includes built-in HTML documentation and health checks.

## 🚀 Quick Start

### Prerequisites
* [Go](https://go.dev/dl/) (version 1.18+)
* An SQLite database file (`locations.db`) in the root directory.

### Installation

1.  **Clone the repository:**
    ```bash
    git clone [https://github.com/yourusername/city-search-api.git](https://github.com/yourusername/city-search-api.git)
    cd city-search-api
    ```

2.  **Prepare the Database:**
    Ensure `locations.db` exists with `cities`, `states`, and `countries` tables (see Contributing guide for schema).

3.  **Run the Server:**
    ```bash
    go mod tidy
    go run main.go
    ```
    The server starts on port **8081** (default).

## 📖 API Reference

### Base URL
`http://localhost:8081`

### 1. Search Locations
**Endpoint:** `GET /api/locations`

Search for a location. You must provide at least one parameter (`city`, `state`, or `country`).

**Parameters:**

| Parameter | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `city` | string | No* | Search by city name (e.g., "Paris"). |
| `state` | string | No* | Search by state/province. |
| `country` | string | No* | Search by country name. |

*\*At least one parameter is required.*




**Example Request:**
```bash
curl "http://localhost:8081/api/locations?city=Ashkasham"

```

**Example Response:**

```json
[
  {
    "city": "Ashkasham",
    "state": "Badakhshan",
    "country": "Afghanistan"
  }
]

```

### 2. Health Check

**Endpoint:** `GET /health`

Returns `200 OK` if the server is running.

### 3. Documentation

**Endpoint:** `GET /`

Returns the interactive HTML documentation page.

## 🛠 Configuration

* **Port:** Defaults to `8081`. Set `PORT` env variable to override.
* **Cache:** Stores up to 1000 distinct queries in RAM.
* **DB:** Max 10 open connections, 5 idle.

## 📄 License

This project is licensed under the **MIT License**.