//go:build windows
// +build windows

package main

import (
	"os/exec"
	"runtime"

	"github.com/getlantern/systray"
)

var serverQuitChan chan bool

func startSystemTray(baseURL string) {
	serverQuitChan = make(chan bool)
	systray.Run(func() {
		onReady(baseURL)
	}, onExit)
}

func onReady(baseURL string) {
	// Set icon (empty for now - can add icon resource later)
	systray.SetIcon([]byte{})
	systray.SetTooltip("TRMNL-POWER Server - " + baseURL)

	// Status menu item
	mStatus := systray.AddMenuItem("✓ Server Running", "")
	mStatus.Disable()

	systray.AddSeparator()

	// Open in browser
	mOpenBrowser := systray.AddMenuItem("Open in Browser", "Open server in default browser")

	// Show server URL
	mURL := systray.AddMenuItem("Server: "+baseURL, "")
	mURL.Disable()

	systray.AddSeparator()

	// Quit
	mQuit := systray.AddMenuItem("Quit TRMNL-POWER", "Stop server and exit")

	go func() {
		for {
			select {
			case <-mOpenBrowser.ClickedCh:
				openBrowser(baseURL)
			case <-mQuit.ClickedCh:
				log.Println("Quit requested from system tray")
				systray.Quit()
				close(serverQuitChan)
				return
			}
		}
	}()
}

func onExit() {
	// Signal server to stop
	if serverQuitChan != nil {
		select {
		case serverQuitChan <- true:
		default:
		}
	}
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

