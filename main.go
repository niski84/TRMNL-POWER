package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type Config struct {
	Server struct {
		Port string `json:"port"`
		Host string `json:"host"`
	} `json:"server"`
	Render struct {
		Width               int    `json:"width"`
		Height              int    `json:"height"`
		RefreshIntervalMins int    `json:"refreshIntervalMinutes"`
		OutputPath          string `json:"outputPath"`
		TempPath            string `json:"tempPath"`
	} `json:"render"`
	DataSources struct {
		JSONFiles     []string `json:"jsonFiles"`
		APIEndpoints  []string `json:"apiEndpoints"`
		Scripts       []string `json:"scripts"`
	} `json:"dataSources"`
	TRMNL struct {
		APIKey          string `json:"apiKey"`
		FriendlyID      string `json:"friendlyId"`
		RefreshRateSecs int    `json:"refreshRateSeconds"`
	} `json:"trmnl"`
	Paths struct {
		Template  string `json:"template"`
		OutputDir string `json:"outputDir"`
	} `json:"paths"`
}

type ViewModel struct {
	Title     string `json:"title"`
	Timestamp string `json:"timestamp"`
	Cards     []Card `json:"cards"`
}

type Card struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
	Unit  string      `json:"unit"`
	Trend string      `json:"trend"`
}

var config Config
var lastRenderTime time.Time
var renderStats RenderStats

type RenderStats struct {
	DataFetchDuration   time.Duration
	RenderDuration      time.Duration
	ConversionDuration  time.Duration
	OutputSize          int64
	LastRenderTime      time.Time
}

// Shared variables for graceful shutdown
var httpServer *http.Server
var serverQuitChan chan bool

func main() {
	// Load config
	configData, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Failed to read config.json: %v", err)
	}
	if err := json.Unmarshal(configData, &config); err != nil {
		log.Fatalf("Failed to parse config.json: %v", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(config.Paths.OutputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Initialize views
	initViews()
	lastRotationTime = time.Now()

	// Check if we should render a specific view for testing
	// Note: This requires test-render.go to be included in the build
	if len(os.Args) > 1 && os.Args[1] == "--test-render" {
		log.Fatal("Test render feature requires test-render.go to be included in build")
	}
	
	// Generate template boilerplate
	if len(os.Args) > 1 && os.Args[1] == "--generate-template" {
		if len(os.Args) < 3 {
			log.Fatal("Usage: trmnl-renderer --generate-template <template-name>")
		}
		templateName := os.Args[2]
		generateTemplateBoilerplate(templateName)
		return
	}
	
	// Validate all templates
	if len(os.Args) > 1 && os.Args[1] == "--validate-templates" {
		validateAllTemplates()
		return
	}

	// Perform initial render of all views
	log.Println("Performing initial render of all views...")
	if err := renderAllViews(); err != nil {
		log.Printf("Initial render failed: %v", err)
	}

	// Start scheduled rendering
	go startScheduler()

	// Start view rotation
	go startViewRotation()

	// Setup HTTP server
	setupServer()

	addr := fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port)
	
	// Get local IP address for help text
	localIP, err := getLocalIP()
	if err != nil {
		log.Printf("Warning: Could not determine local IP: %v", err)
		localIP = "LAN-IP"
	}
	baseURL := fmt.Sprintf("http://%s:%s", localIP, config.Server.Port)
	
	// Check for --no-tray flag to run in console mode
	useTray := true
	for _, arg := range os.Args[1:] {
		if arg == "--no-tray" {
			useTray = false
			break
		}
	}

	// Display startup help text
	displayStartupHelp(baseURL, localIP, config.Server.Port)
	
	log.Printf("\nTRMNL Renderer service listening on http://%s", addr)
	log.Printf("Endpoints:")
	log.Printf("  - TRMNL Setup: http://%s/api/setup", baseURL)
	log.Printf("  - TRMNL Display: http://%s/api/display", baseURL)
	log.Printf("  - Image: http://%s/screen.bmp", baseURL)
	log.Printf("  - Manual Render: POST http://%s/api/render", baseURL)

	// Create HTTP server with graceful shutdown support
	httpServer = &http.Server{
		Addr:    addr,
		Handler: nil, // Uses http.DefaultServeMux from setupServer()
	}

	// Start system tray if on Windows and not disabled
	if useTray {
		go func() {
			log.Println("Starting system tray...")
			log.Println("Server running in system tray. Right-click icon for menu.")
			log.Println("Use --no-tray flag for console-only mode.")
			startSystemTray(baseURL)
		}()
	}

	// Run server with graceful shutdown support
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			if serverQuitChan != nil {
				select {
				case <-serverQuitChan:
					// Normal shutdown
					return
				default:
				}
			}
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for shutdown signal (from tray or console)
	if useTray && serverQuitChan != nil {
		// Wait for quit signal from tray
		<-serverQuitChan
		log.Println("Shutting down server gracefully...")
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		} else {
			log.Println("Server stopped successfully")
		}
	} else {
		// Console mode - wait for Ctrl+C or keep running
		// Note: On Windows, console can be minimized when tray is active
		select {} // Block forever (server runs in goroutine above)
	}
}

// getLocalIP returns the first non-loopback local IP address
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}
	
	return "", fmt.Errorf("no local IP address found")
}

// displayStartupHelp displays instructions for connecting TRMNL devices
func displayStartupHelp(baseURL, localIP, port string) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("TRMNL Local Server - Device Connection Guide")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("\nServer Base URL: %s\n", baseURL)
	fmt.Printf("Local IP Address: %s\n", localIP)
	fmt.Printf("Port: %s\n\n", port)
	
	fmt.Println("How to switch a TRMNL device from cloud to local-only server:")
	fmt.Println()
	fmt.Println("Step 1: Trigger WiFi Captive Portal")
	fmt.Println("  - Hold the button on the back of your TRMNL device for 5-7 seconds, then release")
	fmt.Println("  - The device should broadcast a WiFi network named 'TRMNL'")
	fmt.Println("  - Connect to that SSID from your computer/phone")
	fmt.Println("  - The captive portal should open automatically")
	fmt.Println()
	fmt.Println("Step 2: Change Server Base URL")
	fmt.Printf("  - In the captive portal, set the 'host' or 'base URL' to: %s\n", baseURL)
	fmt.Println("  - Save/apply the changes")
	fmt.Println()
	fmt.Println("Step 3: Verify Connection")
	fmt.Println("  - Watch server logs for these requests:")
	fmt.Println("    • GET /api/setup (with header ID: <MAC>)")
	fmt.Println("    • GET /api/display (with Access-Token header)")
	fmt.Println("  - If you see these calls, the device is now connected locally")
	fmt.Println()
	fmt.Println("Troubleshooting:")
	fmt.Println("  • Captive portal not appearing?")
	fmt.Println("    - Forget 'TRMNL' WiFi network and reconnect")
	fmt.Println("    - Disable VPN during setup")
	fmt.Println("  • Device still showing cloud content?")
	fmt.Println("    - Base URL didn't save - repeat Step 2")
	fmt.Println("    - Verify the host field shows the correct URL")
	fmt.Println("  • No host/base URL field in captive portal?")
	fmt.Println("    - Device may need open-source firmware with BYOS/BYOD support")
	fmt.Println("    - Use TRMNL Flash Assistant to update firmware first")
	fmt.Println()
	fmt.Println("Verification Checklist:")
	fmt.Printf("  ✓ Device base URL is set to: %s\n", baseURL)
	fmt.Printf("  ✓ Server sees /api/setup and /api/display requests\n")
	fmt.Printf("  ✓ Opening %s/screen.bmp in browser shows current image\n", baseURL)
	fmt.Println("  ✓ Device refreshes to that image on its next poll")
	fmt.Println()
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
}

