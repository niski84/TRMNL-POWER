//go:build windows
// +build windows

package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/getlantern/systray"
)

func startSystemTray(baseURL string) {
	serverQuitChan = make(chan bool, 1)
	systray.Run(func() {
		onReady(baseURL)
	}, onExit)
}

func onReady(baseURL string) {
	// Set icon (empty for now - can add icon resource later)
	systray.SetIcon([]byte{})
	systray.SetTooltip("TRMNL-POWER Server - " + baseURL)

	// Status menu item
	mStatus := systray.AddMenuItem("✓ Server Running", "TRMNL-POWER is running")
	mStatus.Disable()

	systray.AddSeparator()

	// View rendered image
	mViewImage := systray.AddMenuItem("View Rendered Image", "Open current screen.bmp in browser")
	
	// Open server in browser
	mOpenBrowser := systray.AddMenuItem("Open Server", "Open server dashboard in browser")

	systray.AddSeparator()

	// Configuration
	mEditConfig := systray.AddMenuItem("Edit Configuration", "Open config.json in Notepad")

	systray.AddSeparator()

	// Quit
	mQuit := systray.AddMenuItem("Exit", "Stop server and quit TRMNL-POWER")

	go func() {
		for {
			select {
			case <-mViewImage.ClickedCh:
				imageURL := baseURL + "/screen.bmp"
				openBrowser(imageURL)
			case <-mOpenBrowser.ClickedCh:
				openBrowser(baseURL)
			case <-mEditConfig.ClickedCh:
				editConfig()
			case <-mQuit.ClickedCh:
				log.Println("Exit requested from system tray - shutting down gracefully...")
				systray.Quit()
				// Signal server to shut down
				if serverQuitChan != nil {
					select {
					case serverQuitChan <- true:
					default:
					}
				}
				return
			}
		}
	}()
}

func onExit() {
	// Cleanup - program is exiting via tray quit
	log.Println("System tray exiting...")
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = exec.Command("xdg-open", url).Start()
	}
	if err != nil {
		log.Printf("Failed to open browser: %v", err)
	}
}

func editConfig() {
	configPath, err := filepath.Abs("config.json")
	if err != nil {
		configPath = "config.json"
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("Config file not found: %s", configPath)
		return
	}

	// Open in Notepad on Windows
	err = exec.Command("notepad.exe", configPath).Start()
	if err != nil {
		log.Printf("Failed to open config in Notepad: %v", err)
		// Fallback: try default editor
		openBrowser("file:///" + filepath.ToSlash(configPath))
	}
}

