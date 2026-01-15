
# Contributing to Calendar API

Thank you for your interest in improving the Calendar API! This document guides you through the setup, development, and contribution process.

## 🛠 Tech Stack

* **Runtime:** Node.js / Bun (depending on your specific deployment)
* **Framework:** [Hono](https://hono.dev/)
* **Language:** TypeScript
* **Validation:** [Zod](https://zod.dev/)
* **Utilities:** `dayjs` (date manipulation), `string-strip-html` (sanitization)

## 📥 Local Setup

1.  **Clone the repository:**
    ```bash
    git clone [https://github.com/your-org/calendar-api.git](https://github.com/your-org/calendar-api.git)
    cd calendar-api
    ```

2.  **Install dependencies:**
    ```bash
    npm install
    # or
    bun install
    ```

3.  **Run the development server:**
    ```bash
    npm run dev
    ```
    The server typically starts on `http://localhost:3000`.

## 🧪 Development Workflow

### Modifying the Schema
The validation logic is centralized in the `eventSchema` Zod object.
* **Adding Fields:** If you add a new property (e.g., `conferenceData`), ensure you update the schema in `src/schema.ts` (or equivalent file).
* **Validation Rules:** Keep validation strict. For example, use `.min(1)` for required strings and `.datetime()` for ISO dates.

### Testing
We encourage writing tests for new features.
1.  **Run tests:**
    ```bash
    npm test
    ```
2.  **Manual Testing:**
    Use the `/calendar/validate` endpoint to check your JSON payloads without generating files. This is faster than downloading ICS files repeatedly.

### Code Style
* Use **TypeScript** strict mode.
* Ensure all dates are handled using **ISO 8601** strings.
* Run the linter before committing:
    ```bash
    npm run lint
    ```

## 🚀 Submitting a Pull Request

1.  Create a new branch: `git checkout -b feature/my-new-feature`.
2.  Make your changes and verify them locally.
3.  Commit your changes with clear messages.
4.  Push to the repository and open a Pull Request.
5.  Ensure the "Health Check" endpoint (`GET /health`) still returns 200 OK.

## ⚠️ Important Notes

* **Sanitization:** The API uses `string-strip-html` to prevent XSS in descriptions. Do not bypass this for `htmlDescription` fields unless necessary and reviewed.
* **Timezones:** Always test recurrence rules across timezone boundaries to ensure `dayjs` logic holds up.

We look forward to your contributions!

