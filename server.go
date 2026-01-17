package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func setupServer() {
	// TRMNL /api/setup endpoint
	http.HandleFunc("/api/setup", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		
		baseURL := "http://" + r.Host
		imageURL := baseURL + "/screen.bmp"
		
		response := map[string]string{
			"api_key":    config.TRMNL.APIKey,
			"friendly_id": config.TRMNL.FriendlyID,
			"image_url":  imageURL,
		}
		
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})

	// TRMNL /api/display endpoint
	// Returns stable JSON structure - device polls this, then fetches image_url
	http.HandleFunc("/api/display", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		
		token := r.Header.Get("Access-Token")
		if token != config.TRMNL.APIKey {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid Access-Token"})
			return
		}
		
		baseURL := "http://" + r.Host
		imageURL := baseURL + "/screen.bmp"
		
		// TRMNL device expects this exact structure (stable across refreshes)
		response := map[string]interface{}{
			"status":        0,
			"image_url":     imageURL,
			"filename":      "current",
			"refresh_rate":  strconv.Itoa(config.TRMNL.RefreshRateSecs),
			"update_firmware": false,
			"reset_firmware": false,
		}
		
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})

	// Serve screen.bmp (or screen.png as fallback)
	http.HandleFunc("/screen.bmp", serveImage)
	http.HandleFunc("/screen.png", serveImage)

	// Manual render trigger
	http.HandleFunc("/api/render", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		
		if err := renderOnce(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"stats":   renderStats,
		})
	})

	// Status endpoint
	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		
		response := map[string]interface{}{
			"status": "running",
			"config": map[string]interface{}{
				"refreshIntervalMinutes": config.Render.RefreshIntervalMins,
				"outputPath":             config.Render.OutputPath,
			},
		}
		
		if !renderStats.LastRenderTime.IsZero() {
			response["lastRender"] = map[string]interface{}{
				"time":              renderStats.LastRenderTime.Format(time.RFC3339),
				"dataFetchDuration": renderStats.DataFetchDuration.String(),
				"renderDuration":    renderStats.RenderDuration.String(),
				"conversionDuration": renderStats.ConversionDuration.String(),
				"outputSize":        renderStats.OutputSize,
			}
		}
		
		json.NewEncoder(w).Encode(response)
	})

	// Root endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		baseURL := "http://" + r.Host
		response := map[string]interface{}{
			"service": "TRMNL Renderer",
			"endpoints": map[string]string{
				"setup":   baseURL + "/api/setup",
				"display": baseURL + "/api/display",
				"image":   baseURL + "/screen.bmp",
				"render":  baseURL + "/api/render (POST)",
				"status":  baseURL + "/api/status",
			},
		}
		json.NewEncoder(w).Encode(response)
	})
}

func serveImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// TRMNL device always fetches image_url on each poll - use no-cache
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	
	// Serve screen.bmp from configured output path (TRMNL expects stable URL)
	bmpPath := config.Render.OutputPath
	pngPath := ""
	
	// If outputPath is .bmp, check for .png fallback in same dir
	if filepath.Ext(bmpPath) == ".bmp" {
		pngPath = bmpPath[:len(bmpPath)-4] + ".png"
	} else {
		// If outputPath is .png, try .bmp in same dir
		bmpPath = config.Paths.OutputDir + "/screen.bmp"
		pngPath = config.Render.OutputPath
	}
	
	var finalPath string
	var contentType string
	
	// Try BMP first (preferred for TRMNL)
	if _, err := os.Stat(bmpPath); err == nil {
		finalPath = bmpPath
		contentType = "image/bmp"
	} else if _, err := os.Stat(pngPath); err == nil {
		// Fallback to PNG
		finalPath = pngPath
		contentType = "image/png"
	} else {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Image not yet generated"})
		return
	}
	
	w.Header().Set("Content-Type", contentType)
	
	file, err := os.Open(finalPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to open image"})
		return
	}
	defer file.Close()
	
	io.Copy(w, file)
}

