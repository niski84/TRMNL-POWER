package main

import (
	"time"
)

func startViewRotation() {
	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()

	for range ticker.C {
		if shouldRotate() {
			rotateView()
		}
	}
}

