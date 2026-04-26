# TRMNL-POWER

Local-first rendering server for TRMNL e-ink displays.

## What It Does

TRMNL-POWER runs on your network and serves images to TRMNL e-ink devices, replacing the cloud service. Devices poll a local HTTP endpoint, get an image URL, and fetch a 1-bit BMP rendered from your own HTML templates and JSON data.

Templates live in `templates/`, data in `data/`. A background scheduler re-renders on a fixed interval and can rotate between multiple views automatically. Rendering is done by handing HTML to a Playwright (Node.js) script which produces a PNG; Go then converts it to e-ink-friendly 1-bit BMP at 800x480.

The setup flow walks you through putting a TRMNL device into captive-portal mode and pointing its base URL at the local server. No cloud account needed.

## Tech Stack

- Go (single binary)
- Node.js + Playwright for HTML-to-image rendering
- `getlantern/systray` for an optional Windows tray icon
- Go `html/template` for view rendering

## Installation

```bash
git clone https://github.com/niski84/TRMNL-POWER.git
cd TRMNL-POWER
go mod download
npm install
go build -o trmnl-renderer .
./trmnl-renderer
```

Windows users can grab a pre-built release and run `run.bat`.

CLI flags:
- `--generate-template <name>` scaffolds a new template + data file
- `--validate-templates` checks all templates for layout issues
- `--no-tray` runs without the system tray (Windows)

## Configuration

Edit `config.json`:

- `server.port` / `server.host` - HTTP listener
- `render.width` / `render.height` - display size (defaults 800x480)
- `render.refreshIntervalMinutes` - background re-render cadence
- `trmnl.apiKey` - access token devices send
- `trmnl.friendlyId` - identifier returned to the device
- `trmnl.refreshRateSeconds` - how often the device polls

Endpoints exposed: `/api/setup`, `/api/display`, `/screen.bmp`, and `POST /api/render` for manual refresh.

## License

MIT
