# Backup Image

## About

A desktop application using Wails v2 and Nuxt 4. The Go entry point is
`main.go`, and the frontend is in `frontend/app`.

## Requirements

- Go 1.26 or later
- Node.js compatible with Nuxt 4 and npm
- Wails CLI v2.15.0: `go install github.com/wailsapp/wails/v2/cmd/wails@v2.15.0`

On Ubuntu, install the native build dependencies:

```bash
sudo apt update
sudo apt install build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev
```

The WebKit 4.1 build tag is configured in `wails.json`, so no extra command-line
flags are needed.

## Live Development

Run from the project root:

```bash
wails dev
```

Wails installs frontend dependencies, generates the Go bindings and static
assets, starts the Nuxt development server, and opens the desktop window.
Changes to the frontend reload through Nuxt; changes to Go rebuild the app.

## Building

```bash
wails build
```

Nuxt runs in client-side mode. Wails uses `npm run generate` to create the static
frontend in `frontend/.output/public` and embeds it into the application.
The executable is written to `build/bin`.
