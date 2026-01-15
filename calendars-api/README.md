
# 📅 Calendar API

> A production-ready API for generating ICS files and calendar links for all major platforms.

![Version](https://img.shields.io/badge/version-1.0.0-success) ![Status](https://img.shields.io/badge/status-production_ready-success)

The **Calendar API** simplifies the creation of calendar events. Built with **TypeScript**, **Hono**, and **Zod**, it provides robust validation and generates `.ics` files or direct "Add to Calendar" links for Google, Outlook, Yahoo, and Office365.

## ⚡ Features

* **Universal Compatibility:** Generates standard `.ics` files and platform-specific web links.
* **Recurrence Patterns:** Support for Daily, Weekly, Monthly, and Yearly repeating events with custom intervals.
* **Advanced Scheduling:** Handles timezones, alarms/reminders, and attendee RSVP status.
* **Security First:** Built-in HTML sanitization, strict Zod schema validation, and rate limiting.
* **Batch Processing:** Generate single calendar files containing up to 100 distinct events.
* **Rich Metadata:** Support for geo-location, organizers, categories, and HTML descriptions.

## 🚀 Quick Start

### Base URL
The API is typically deployed at `https://your-api-domain.com`.

### 1. Generate a Single ICS File
**Endpoint:** `POST /calendar/ics`

```bash
curl -X POST [https://api.example.com/calendar/ics](https://api.example.com/calendar/ics) \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Team Sync",
    "start": "2026-03-01T10:00:00Z",
    "end": "2026-03-01T11:00:00Z",
    "timezone": "America/New_York",
    "organizer": { "name": "Alice", "email": "alice@example.com" }
  }' > event.ics

```

### 2. Generate Web Links (Google, Outlook, etc.)

**Endpoint:** `POST /calendar/links`

```bash
curl -X POST [https://api.example.com/calendar/links](https://api.example.com/calendar/links) \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Lunch & Learn",
    "start": "2026-03-05T12:00:00Z",
    "end": "2026-03-05T13:00:00Z",
    "description": "Join us for free pizza!"
  }'

```

---

## 📖 API Reference

### Endpoints

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/health` | Check API status and uptime. |
| `POST` | `/calendar/validate` | Validate event payload without generating files. |
| `POST` | `/calendar/links` | Get "Add to Calendar" URLs for Google, Outlook, Yahoo, etc. |
| `POST` | `/calendar/ics` | Download a single `.ics` file. |
| `POST` | `/calendar/batch` | Download an `.ics` file with multiple events (Max 100). |
| `POST` | `/calendar/all` | Get both ICS content string and web links in one JSON response. |

### Event Schema

The API accepts a JSON object with the following properties:

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `title` | String | ✅ | Event title (Max 200 chars). |
| `start` | ISO 8601 | ✅ | Start time (e.g., `2023-10-10T10:00:00Z`). |
| `end` | ISO 8601 | ✅ | End time. Must be after start time. |
| `timezone` | String | ❌ | IANA Timezone (default: `UTC`). |
| `description` | String | ❌ | Plain text description (Max 5000 chars). |
| `location` | String | ❌ | Physical location or meeting link. |
| `geo` | Object | ❌ | Lat/Lon coordinates (e.g., `{ "lat": 40.71, "lon": -74.00 }`). |
| `organizer` | Object | ❌ | `{ "name": "...", "email": "..." }`. |
| `attendees` | Array | ❌ | List of participants with RSVP status. |
| `recurrence` | Object | ❌ | Repeat rules (`freq`, `interval`, `until`, `byDay`). |
| `alarms` | Array | ❌ | Reminders (e.g., `{ "minutes": 15, "type": "display" }`). |

### Rate Limits

* **Limit:** Configured (default: ~100) requests per 5 minutes per IP.
* **Response:** Returns `429 Too Many Requests` if exceeded.

---

## 💻 Integration Example (JavaScript)

```javascript
// Example: Fetching calendar links for a frontend button
async function getLinks() {
  const response = await fetch('[https://api.example.com/calendar/links](https://api.example.com/calendar/links)', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      title: "Product Launch",
      start: "2026-06-01T09:00:00Z",
      end: "2026-06-01T10:00:00Z",
      timezone: "Europe/London"
    })
  });

  const data = await response.json();
  // data.links.google -> Open this URL in a new tab
  // data.links.outlook -> Open this URL in a new tab
}

```

## 📄 License
This project is licensed under the **MIT License**.