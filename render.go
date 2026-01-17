package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func collectData() (*ViewModel, error) {
	rawData := make(map[string]interface{})

	// Collect from JSON files
	for _, filePath := range config.DataSources.JSONFiles {
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Warning: Failed to read JSON file %s: %v", filePath, err)
			continue
		}
		var fileData map[string]interface{}
		if err := json.Unmarshal(data, &fileData); err != nil {
			log.Printf("Warning: Failed to parse JSON file %s: %v", filePath, err)
			continue
		}
		// Merge data
		for k, v := range fileData {
			rawData[k] = v
		}
	}

	// Collect from API endpoints
	for _, endpoint := range config.DataSources.APIEndpoints {
		resp, err := http.Get(endpoint)
		if err != nil {
			log.Printf("Warning: Failed to fetch from API %s: %v", endpoint, err)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			var apiData map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&apiData); err == nil {
				for k, v := range apiData {
					rawData[k] = v
				}
			}
		}
	}

	// Normalize into ViewModel
	return normalizeData(rawData), nil
}

func normalizeData(rawData map[string]interface{}) *ViewModel {
	vm := &ViewModel{
		Title:     "TRMNL Dashboard",
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Cards:     []Card{},
	}

	// Extract title and timestamp
	if title, ok := rawData["title"].(string); ok {
		vm.Title = title
	}
	if ts, ok := rawData["timestamp"].(string); ok {
		vm.Timestamp = ts
	}

	// Extract cards array if present
	if cardsArr, ok := rawData["cards"].([]interface{}); ok {
		for _, cardRaw := range cardsArr {
			if cardMap, ok := cardRaw.(map[string]interface{}); ok {
				card := Card{
					Label: getString(cardMap, "label", "N/A"),
					Value: cardMap["value"],
					Unit:  getString(cardMap, "unit", ""),
					Trend: getString(cardMap, "trend", "neutral"),
				}
				vm.Cards = append(vm.Cards, card)
			}
		}
	}

	// If no cards, extract individual fields
	if len(vm.Cards) == 0 {
		for key, value := range rawData {
			if key == "title" || key == "timestamp" || key == "cards" {
				continue
			}
			if len(vm.Cards) >= 4 {
				break
			}
			vm.Cards = append(vm.Cards, Card{
				Label: formatLabel(key),
				Value: formatValue(value),
				Unit:  "",
				Trend: "neutral",
			})
		}
	}

	// Ensure at least 2 cards
	for len(vm.Cards) < 2 {
		vm.Cards = append(vm.Cards, Card{
			Label: "Placeholder",
			Value: "N/A",
			Unit:  "",
			Trend: "neutral",
		})
	}

	// Limit to 4 cards
	if len(vm.Cards) > 4 {
		vm.Cards = vm.Cards[:4]
	}

	return vm
}

func getString(m map[string]interface{}, key, defaultValue string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return defaultValue
}

func formatLabel(key string) string {
	// Simple camelCase/snake_case to Title Case conversion
	parts := strings.FieldsFunc(key, func(r rune) bool {
		return r == '_' || r >= 'A' && r <= 'Z'
	})
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += " "
		}
		if len(part) > 0 {
			result += strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	if result == "" {
		return key
	}
	return result
}

func formatValue(value interface{}) interface{} {
	if value == nil {
		return "N/A"
	}
	switch v := value.(type) {
	case float64:
		return v
	case int:
		return v
	case bool:
		if v {
			return "Yes"
		}
		return "No"
	case string:
		if len(v) > 20 {
			return v[:17] + "..."
		}
		return v
	default:
		str := fmt.Sprintf("%v", v)
		if len(str) > 20 {
			return str[:17] + "..."
		}
		return str
	}
}

func renderHTML(vm *ViewModel) string {
	cardsHTML := ""
	for _, card := range vm.Cards {
		valueStr := formatValueString(card.Value)
		unitHTML := ""
		if card.Unit != "" {
			unitHTML = fmt.Sprintf(`<span class="card-unit">%s</span>`, escapeHTML(card.Unit))
		}
		trendHTML := ""
		if card.Trend != "neutral" {
			trendHTML = fmt.Sprintf(`<div class="card-trend trend-%s"></div>`, card.Trend)
		}
		cardsHTML += fmt.Sprintf(`
    <div class="card">
      <div class="card-label">%s</div>
      <div class="card-value-container">
        <div class="card-value">%s</div>
        %s
      </div>
      %s
    </div>`, escapeHTML(card.Label), escapeHTML(valueStr), unitHTML, trendHTML)
	}

	// Determine grid layout class
	gridClass := "content"
	if len(vm.Cards) == 1 {
		gridClass += " single"
	} else if len(vm.Cards) == 3 {
		gridClass += " three"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=%d, height=%d, initial-scale=1.0">
  <title>TRMNL Dashboard</title>
  <style>
%s
  </style>
</head>
<body>
  <div class="header">
    <div class="header-title">%s</div>
    <div class="header-timestamp">%s</div>
  </div>
  <div class="%s" id="content">
%s
  </div>
</body>
</html>`, config.Render.Width, config.Render.Height, tailwindCSS, escapeHTML(vm.Title), escapeHTML(vm.Timestamp), gridClass, cardsHTML)
}

func formatValueString(value interface{}) string {
	switch v := value.(type) {
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func escapeHTML(text string) string {
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	text = strings.ReplaceAll(text, `"`, "&quot;")
	text = strings.ReplaceAll(text, "'", "&#039;")
	return text
}

func renderToImage(html string, outputPath string) error {
	// Use Playwright via Node.js script for HTML to PNG conversion
	scriptPath := filepath.Join(".", "scripts", "playwright-render.js")
	
	// Check if script exists
	if _, err := os.Stat(scriptPath); err != nil {
		return fmt.Errorf("playwright render script not found at %s: %w. Make sure the script exists and Node.js is installed", scriptPath, err)
	}

	// Create temporary HTML file
	tempHTML, err := os.CreateTemp("", "trmnl-render-*.html")
	if err != nil {
		return fmt.Errorf("failed to create temp HTML file: %w", err)
	}
	defer os.Remove(tempHTML.Name())
	defer tempHTML.Close()

	// Write HTML to temp file
	if _, err := tempHTML.WriteString(html); err != nil {
		return fmt.Errorf("failed to write HTML to temp file: %w", err)
	}
	if err := tempHTML.Close(); err != nil {
		return fmt.Errorf("failed to close temp HTML file: %w", err)
	}

	// Save intermediate PNG path
	tempPNG := config.Render.TempPath + ".png"
	
	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(tempPNG), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Call Playwright script
	cmd := exec.Command("node", scriptPath,
		tempHTML.Name(),
		tempPNG,
		strconv.Itoa(config.Render.Width),
		strconv.Itoa(config.Render.Height),
	)
	
	// Capture stderr for better error messages
	var stderr strings.Builder
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		errMsg := stderr.String()
		if errMsg == "" {
			errMsg = err.Error()
		}
		return fmt.Errorf("playwright render failed: %s. Make sure Node.js and Playwright are installed (npm install playwright)", errMsg)
	}

	// Check if PNG was created
	if _, err := os.Stat(tempPNG); err != nil {
		return fmt.Errorf("playwright render completed but PNG file not found at %s: %w", tempPNG, err)
	}

	// Convert to 1-bit monochrome
	return convertToMonochrome(tempPNG, outputPath)
}

func convertToMonochrome(inputPath, outputPath string) error {
	// Read PNG
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	// Resize and convert to 1-bit
	bounds := img.Bounds()
	resized := image.NewRGBA(image.Rect(0, 0, config.Render.Width, config.Render.Height))
	
	// Simple resize (nearest neighbor)
	for y := 0; y < config.Render.Height; y++ {
		for x := 0; x < config.Render.Width; x++ {
			srcX := x * bounds.Dx() / config.Render.Width
			srcY := y * bounds.Dy() / config.Render.Height
			r, g, b, _ := img.At(srcX, srcY).RGBA()
			gray := uint8((r + g + b) / 3 / 256)
			// Threshold at 128
			if gray < 128 {
				resized.Set(x, y, color.RGBA{0, 0, 0, 255})
			} else {
				resized.Set(x, y, color.RGBA{255, 255, 255, 255})
			}
		}
	}

	// Atomic file replacement: write to temp file first, then rename
	// This prevents partial reads when TRMNL device fetches the image
	var tempPath string
	var finalPath string
	
	if filepath.Ext(outputPath) == ".bmp" {
		// For BMP output, write to .tmp first
		tempPath = outputPath + ".tmp"
		finalPath = outputPath
	} else {
		// For PNG or other formats, just write directly
		tempPath = outputPath
		finalPath = outputPath
	}

	// Write as PNG (Go's image/png handles 2-color PNGs well)
	outputFile, err := os.Create(tempPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	encoder := &png.Encoder{CompressionLevel: png.BestCompression}
	if err := encoder.Encode(outputFile, resized); err != nil {
		os.Remove(tempPath) // Clean up on error
		return err
	}
	outputFile.Close() // Close before rename

	// Try ImageMagick for true BMP3 1-bit if available
	tryImageMagickBMP(tempPath)

	// Atomic rename: temp -> final (atomic on POSIX systems)
	if tempPath != finalPath {
		if err := os.Rename(tempPath, finalPath); err != nil {
			os.Remove(tempPath) // Clean up on error
			return fmt.Errorf("failed to atomically replace %s: %w", finalPath, err)
		}
	}

	return nil
}

func tryImageMagickBMP(pngPath string) {
	// This will be implemented via exec if ImageMagick is available
	// For now, we'll use PNG which works fine
	bmpPath := config.Render.OutputPath
	if bmpPath == "" {
		return
	}
	// Copy PNG as fallback
	if bmpPath != pngPath {
		src, err := os.Open(pngPath)
		if err != nil {
			return
		}
		defer src.Close()
		dst, err := os.Create(bmpPath)
		if err != nil {
			return
		}
		defer dst.Close()
		io.Copy(dst, src)
	}
}

func renderAllViews() error {
	initViews()
	if len(views) == 0 {
		return fmt.Errorf("no views configured")
	}

	start := time.Now()
	totalDataDuration := time.Duration(0)
	totalHTMLDuration := time.Duration(0)
	totalImageDuration := time.Duration(0)
	var lastSize int64

	// Render each view to its own image file
	for _, view := range views {
		// Load view data
		dataStart := time.Now()
		_, err := loadViewData(view)
		if err != nil {
			log.Printf("Warning: Failed to load data for view %s: %v", view.Name, err)
			continue
		}
		totalDataDuration += time.Since(dataStart)

		// Render HTML template
		htmlStart := time.Now()
		html, err := renderViewHTML(view)
		if err != nil {
			log.Printf("Warning: Failed to render HTML for view %s: %v", view.Name, err)
			continue
		}
		totalHTMLDuration += time.Since(htmlStart)

		// Render to image
		imgStart := time.Now()
		outputPath := filepath.Join(config.Paths.OutputDir, view.Name+".png")
		if err := renderToImage(html, outputPath); err != nil {
			log.Printf("Warning: Failed to render image for view %s: %v", view.Name, err)
			continue
		}
		imgDuration := time.Since(imgStart)
		totalImageDuration += imgDuration

		// Get file size
		info, _ := os.Stat(outputPath)
		if info != nil {
			lastSize = info.Size()
		}

		log.Printf("Rendered view %s: html=%v, image=%v, size=%.2f KB",
			view.Name, time.Since(htmlStart), imgDuration, float64(lastSize)/1024)
	}

	renderStats = RenderStats{
		DataFetchDuration:  totalDataDuration,
		RenderDuration:     totalHTMLDuration,
		ConversionDuration: totalImageDuration,
		OutputSize:         lastSize,
		LastRenderTime:     time.Now(),
	}

	lastRenderTime = time.Now()
	log.Printf("All views rendered: total=%v, views=%d",
		time.Since(start), len(views))

	// Render current view to screen.bmp for TRMNL device (atomic replacement)
	if err := renderCurrentViewToTRMNL(); err != nil {
		log.Printf("Warning: Failed to render current view to screen.bmp: %v", err)
		// Non-fatal - per-view renders still succeeded
	}

	return nil
}

// renderCurrentViewToTRMNL renders the current rotating view to screen.bmp
// This is what the TRMNL device fetches via /screen.bmp
func renderCurrentViewToTRMNL() error {
	view := getCurrentView()
	if view.Name == "" {
		return fmt.Errorf("no current view available")
	}

	// Render HTML template
	html, err := renderViewHTML(view)
	if err != nil {
		return fmt.Errorf("failed to render HTML: %w", err)
	}

	// Render to the configured output path (screen.bmp) with atomic replacement
	if err := renderToImage(html, config.Render.OutputPath); err != nil {
		return fmt.Errorf("failed to render image: %w", err)
	}

	log.Printf("Rendered current view '%s' to %s for TRMNL", view.Name, config.Render.OutputPath)
	return nil
}

func renderOnce() error {
	// Legacy function - now renders all views
	return renderAllViews()
}

func startScheduler() {
	interval := time.Duration(config.Render.RefreshIntervalMins) * time.Minute
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("Scheduled render triggered")
		if err := renderAllViews(); err != nil {
			log.Printf("Scheduled render failed: %v", err)
		}
	}
}

