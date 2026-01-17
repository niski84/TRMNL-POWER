# Installation Requirements

## Prerequisites

1. **Go 1.24+** - Already configured in `go.mod`

2. **Chromium/Chrome** - Required for HTML-to-image rendering

   Install Chromium on Ubuntu/Debian:
   ```bash
   sudo apt-get update
   sudo apt-get install chromium-browser
   ```

   Or install Google Chrome and ensure `google-chrome` is in PATH.

3. **ImageMagick (optional)** - For true BMP3 1-bit output
   ```bash
   sudo apt-get install imagemagick
   ```
   If not installed, the renderer will use PNG format instead.

## Testing Installation

After installing Chromium, test the renderer:

```bash
# Build
./scripts/reload.sh

# Run test render
go run test-render.go

# If successful, you should see:
# - test-output.png created
# - test-render.log shows "Success!"

# Run the full server
./trmnl-renderer
```

## Troubleshooting

**Error: "executable file not found"**
- Chromium/Chrome is not installed or not in PATH
- Install with: `sudo apt-get install chromium-browser`
- Verify with: `which chromium-browser`

**Error: "chromedp render failed"**
- Check `server.log` for detailed error messages
- Ensure Chromium can run headless (no display needed)

**No images generated**
- Check that `output/` directory exists and is writable
- Check `server.log` for render errors
- Verify JSON data files exist in `data/` directory

