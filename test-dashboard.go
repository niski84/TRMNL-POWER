package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// TestDashboard - Simple test to render dashboard view to test-output.png
func TestDashboard() {
	// Load config
	configData, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Failed to read config.json: %v", err)
	}
	if err := json.Unmarshal(configData, &config); err != nil {
		log.Fatalf("Failed to parse config.json: %v", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir("test-output.png"), 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Initialize views
	initViews()

	// Find dashboard view
	var dashboardView View
	found := false
	for _, view := range views {
		if view.Name == "dashboard" {
			dashboardView = view
			found = true
			break
		}
	}

	if !found {
		log.Fatalf("Dashboard view not found. Available views: %v", getViewNames())
	}

	fmt.Printf("Rendering dashboard view...\n")
	fmt.Printf("  Template: %s\n", dashboardView.Template)
	fmt.Printf("  Data: %s\n", dashboardView.DataPath)
	fmt.Printf("  Output: test-output.png\n")

	// Render HTML template
	html, err := renderViewHTML(dashboardView)
	if err != nil {
		log.Fatalf("Failed to render HTML: %v", err)
	}

	fmt.Printf("HTML rendered successfully (%d bytes)\n", len(html))

	// Render to image
	if err := renderToImage(html, "test-output.png"); err != nil {
		log.Fatalf("Failed to render image: %v", err)
	}

	// Check if file was created
	info, err := os.Stat("test-output.png")
	if err != nil {
		log.Fatalf("Output file not found: %v", err)
	}

	fmt.Printf("Success! Dashboard rendered to test-output.png (%.2f KB)\n", float64(info.Size())/1024)
}

func getViewNames() []string {
	names := make([]string, len(views))
	for i, v := range views {
		names[i] = v.Name
	}
	return names
}

// renderViewToTest - Renders any view to {view-name}-render.png
func renderViewToTest(viewName string) {
	outputFile := fmt.Sprintf("%s-render.png", viewName)
	
	// Find the specified view
	var targetView View
	found := false
	for _, view := range views {
		if view.Name == viewName {
			targetView = view
			found = true
			break
		}
	}

	if !found {
		log.Fatalf("View '%s' not found. Available views: %v", viewName, getViewNames())
	}

	fmt.Printf("Rendering view '%s'...\n", viewName)
	fmt.Printf("  Template: %s\n", targetView.Template)
	fmt.Printf("  Data: %s\n", targetView.DataPath)
	fmt.Printf("  Output: %s\n", outputFile)

	// Render HTML template
	html, err := renderViewHTML(targetView)
	if err != nil {
		log.Fatalf("Failed to render HTML: %v", err)
	}

	fmt.Printf("HTML rendered successfully (%d bytes)\n", len(html))

	// Render to image
	if err := renderToImage(html, outputFile); err != nil {
		log.Fatalf("Failed to render image: %v", err)
	}

	// Check if file was created
	info, err := os.Stat(outputFile)
	if err != nil {
		log.Fatalf("Output file not found: %v", err)
	}

	fmt.Printf("Success! View '%s' rendered to %s (%.2f KB)\n", viewName, outputFile, float64(info.Size())/1024)
}

