# Contributing to Avatar Generator API

First off, thanks for taking the time to contribute! 🎉

We welcome contributions of all kinds, whether it's fixing bugs, adding new avatar styles, improving documentation, or suggesting new features.

## 🛠 Development Prerequisites

You will need the following installed on your machine:
* [Go](https://golang.org/dl/) (version 1.18 or higher recommended)
* Git

## 📥 Getting Started

1.  **Fork** the repository on GitHub.
2.  **Clone** your fork locally:
    ```bash
    git clone [https://github.com/yourusername/avatar-generator.git](https://github.com/yourusername/avatar-generator.git)
    cd avatar-generator
    ```
3.  **Create a branch** for your feature or bug fix:
    ```bash
    git checkout -b feature/amazing-new-style
    ```

## 💻 Development Workflow

### Adding a New Avatar Style

If you want to add a new avatar style (e.g., `type=newstyle`), please follow these steps:
1.  Implement the SVG generation logic in the appropriate Go package.
2.  Ensure the style is deterministic (returns the same output for the same `name` input).
3.  Register the new type in the handler switch statement.
4.  Update the `documentationHandler` function in `main.go` (or wherever it resides) to include the new style in the HTML list.

### Code Style
We follow standard Go coding conventions. Before committing, please ensure your code is formatted:

```bash
go fmt ./...

```

### Running Tests

Please add tests for any new logic and run existing tests to ensure no regressions:

```bash
go test ./...

```

## 🚀 Submitting a Pull Request

1. **Commit** your changes with a clear message:
```bash
git commit -m "feat: add new 'origami' avatar style"

```


2. **Push** to your fork:
```bash
git push origin feature/amazing-new-style

```


3. Open a **Pull Request** against the `main` branch.
4. Describe your changes clearly in the PR description. If you added a visual style, please attach a screenshot!

## 📄 License

By contributing, you agree that your contributions will be licensed under its MIT License.

