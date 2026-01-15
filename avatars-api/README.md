# 🎨 Avatar Generator API

> Generate beautiful, unique avatars with a simple HTTP request.

The **Avatar Generator API** is a high-performance JSON/SVG service written in Go. It generates unique, deterministic SVG avatars based on a name input. It is perfect for user profiles, placeholders, and applications requiring consistent, personalized visuals without storing image files.

## ⚡ Features

* **Deterministic:** The same name always generates the exact same avatar.
* **SVG Format:** Scalable to any size without quality loss.
* **16 Unique Styles:** Ranging from classic initials to retro dithering and geometric patterns.
* **Fast & Lightweight:** Generated on-the-fly; no database or file storage required.
* **CORS Enabled:** Ready to use from any frontend domain.
* **Rate Limited:** Protects resources (default: 100 requests/minute per IP).
* **Accessible Colors:** Uses the OKLCH color space for high-contrast, perceptually uniform colors.

## 🚀 Quick Start

### Installation

```bash
# Clone the repository
git clone [https://github.com/yourusername/avatar-generator.git](https://github.com/yourusername/avatar-generator.git)

# Navigate to the directory
cd avatar-generator

# Run the server
go run main.go

```

The server typically starts on port `8080` (check your `main.go` configuration).

### Basic Usage

**Endpoint:** `GET /api/generate-avatar`

#### HTML Example

```html
<img src="http://localhost:8080/api/generate-avatar?name=John%20Doe&type=avatar&size=200" alt="John's Avatar">

```

#### JavaScript Example

```javascript
fetch('/api/generate-avatar?name=Bob%20Johnson&type=beam')
  .then(response => response.text())
  .then(svg => {
    document.getElementById('avatar-container').innerHTML = svg;
  });

```

---

## 📖 API Reference

### Request Parameters

| Parameter | Type | Required | Default | Description |
| --- | --- | --- | --- | --- |
| `name` | String | No | "User" | The seed for the generator. Same name = same avatar. |
| `type` | String | No | "avatar" | The visual style of the avatar (see list below). |
| `size` | Integer | No | 100 | Width and height of the SVG in pixels. |
| `color` | String | No | Auto | Hex code (e.g., `#FF5733`). If omitted, color is generated from the name. |

### Available Avatar Styles

Use these values for the `type` parameter:

1. `avatar` (Classic initials)
2. `gravatar` (Identicon)
3. `dither` (Retro plasma)
4. `ascii` (Procedural robot)
5. `dotmatrix` (LED display)
6. `terminal` (Block text)
7. `bauhaus` (Geometric art)
8. `ring` (Gradient rings)
9. `beam` (Network nodes)
10. `marble` (Fluid texture)
11. `glitch` (Cyberpunk effect)
12. `sunset` (Procedural scenery)
13. `smile` (Minimalist face)
14. `circuit` (PCB pattern)
15. `pixel` (Isometric cube)
16. `constellation` (Star map)

### Response Codes

* **200 OK:** Avatar generated successfully.
* **400 Bad Request:** Invalid avatar `type` specified.
* **429 Too Many Requests:** Rate limit exceeded.
* **500 Internal Server Error:** Server-side processing error.

---

## 🎨 Color Generation Logic

If a specific `color` is not provided, the API uses a hashing algorithm based on the input `name`.

It utilizes the **OKLCH color space** to ensure:

1. **Consistency:** The same name yields the same color.
2. **Accessibility:** Automatically calculates contrast for readability.
3. **Vibrancy:** Maintains perceptual uniformity across different hues.

## 📄 License

This project is licensed under the **MIT License**.

