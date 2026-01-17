package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Simple test to verify Playwright can render HTML to image
func main() {
	logFile, err := os.OpenFile("test-render.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("=== Starting Playwright render test ===")

	// Test HTML
	html := `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Test</title>
  <style>
    body { width: 800px; height: 480px; margin: 0; padding: 20px; font-family: Arial; }
    h1 { color: #000; font-size: 32px; }
  </style>
</head>
<body>
  <h1>Test Render</h1>
  <p>If you see this, Playwright rendering works!</p>
</body>
</html>`

	log.Println("Step 1: Setting up test...")
	fmt.Println("Step 1: Setting up test...")

	// Create temporary HTML file
	tempHTML, err := os.CreateTemp("", "test-render-*.html")
	if err != nil {
		log.Fatalf("Failed to create temp HTML file: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to create temp HTML file: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(tempHTML.Name())
	defer tempHTML.Close()

	if _, err := tempHTML.WriteString(html); err != nil {
		log.Fatalf("Failed to write HTML: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to write HTML: %v\n", err)
		os.Exit(1)
	}
	tempHTML.Close()

	log.Println("Step 2: Calling Playwright script...")
	fmt.Println("Step 2: Calling Playwright script...")

	scriptPath := filepath.Join(".", "scripts", "playwright-render.js")
	outputPath := "test-output.png"
	width := "800"
	height := "480"

	cmd := exec.Command("node", scriptPath,
		tempHTML.Name(),
		outputPath,
		width,
		height,
	)

	var stderr strings.Builder
	cmd.Stderr = &stderr

	log.Println("Step 3: Executing Playwright (this may take 10-30 seconds)...")
	fmt.Println("Step 3: Executing Playwright (this may take 10-30 seconds)...")

	if err := cmd.Run(); err != nil {
		errMsg := stderr.String()
		if errMsg == "" {
			errMsg = err.Error()
		}
		log.Printf("ERROR: Playwright render failed: %s", errMsg)
		fmt.Fprintf(os.Stderr, "ERROR: Playwright render failed: %s\n", errMsg)
		fmt.Fprintf(os.Stderr, "Make sure Node.js and Playwright are installed (npm install playwright)\n")
		fmt.Fprintf(os.Stderr, "Check test-render.log for details\n")
		os.Exit(1)
	}

	// Check if output file exists
	if _, err := os.Stat(outputPath); err != nil {
		log.Printf("ERROR: Output file not found: %v", err)
		fmt.Fprintf(os.Stderr, "ERROR: Output file not found: %v\n", err)
		os.Exit(1)
	}

	// Get file size
	info, err := os.Stat(outputPath)
	if err != nil {
		log.Printf("ERROR: Failed to stat output file: %v", err)
		fmt.Fprintf(os.Stderr, "ERROR: Failed to stat output file: %v\n", err)
		os.Exit(1)
	}

	log.Printf("Success! Output written to %s (%d bytes)", outputPath, info.Size())
	fmt.Printf("Success! Output written to %s (%d bytes)\n", outputPath, info.Size())
	fmt.Printf("Check test-render.log for details\n")
}

