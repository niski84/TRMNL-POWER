package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"time"
)

type View struct {
	Name     string
	Template string
	DataPath string
}

type ViewData struct {
	Title     string
	Timestamp string
	Styles    template.CSS
	Tasks     []Task                   `json:"tasks,omitempty"`
	Cards     []Card                   `json:"cards,omitempty"`
	Fields    map[string]interface{}   `json:"fields,omitempty"` // Flexible fields for templating
}

type Task struct {
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
	Category  string `json:"category,omitempty"`
}

var views []View
var currentViewIndex int
var lastRotationTime time.Time
var rotationInterval = 15 * time.Minute

func initViews() {
	views = []View{
		{
			Name:     "todo",
			Template: "./templates/todo.html",
			DataPath: "./data/todo.json",
		},
		{
			Name:     "dashboard",
			Template: "./templates/dashboard.html",
			DataPath: "./data/example.json",
		},
		{
			Name:     "chores",
			Template: "./templates/chores.html",
			DataPath: "./data/chores.json",
		},
		// Add more views here as we create them
	}
}

func loadViewData(view View) (*ViewData, error) {
	data, err := os.ReadFile(view.DataPath)
	if err != nil {
		return nil, err
	}

	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return nil, err
	}

	viewData := &ViewData{
		Title:     "TRMNL Dashboard",
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Styles:    template.CSS(tailwindCSS),
		Fields:    make(map[string]interface{}),
	}

	// Extract title and timestamp (these are special fields)
	if title, ok := rawData["title"].(string); ok {
		viewData.Title = title
	}
	if ts, ok := rawData["timestamp"].(string); ok {
		viewData.Timestamp = ts
	}

	// Extract tasks for todo view
	if tasksArr, ok := rawData["tasks"].([]interface{}); ok {
		for _, taskRaw := range tasksArr {
			if taskMap, ok := taskRaw.(map[string]interface{}); ok {
				task := Task{
					Text:      getString(taskMap, "text", ""),
					Completed: getBool(taskMap, "completed", false),
					Category:  getString(taskMap, "category", ""),
				}
				viewData.Tasks = append(viewData.Tasks, task)
			}
		}
	}

	// Extract cards for card-based views
	if cardsArr, ok := rawData["cards"].([]interface{}); ok {
		for _, cardRaw := range cardsArr {
			if cardMap, ok := cardRaw.(map[string]interface{}); ok {
				card := Card{
					Label: getString(cardMap, "label", "N/A"),
					Value: cardMap["value"],
					Unit:  getString(cardMap, "unit", ""),
					Trend: getString(cardMap, "trend", "neutral"),
				}
				viewData.Cards = append(viewData.Cards, card)
			}
		}
	}

	// Populate Fields map with all other fields from JSON (for flexible templating)
	// First check if there's a nested "fields" object, otherwise use root-level fields
	if fieldsObj, ok := rawData["fields"].(map[string]interface{}); ok {
		// Use nested fields object
		for key, value := range fieldsObj {
			viewData.Fields[key] = value
		}
	} else {
		// Use root-level fields (skip special fields that are handled above)
		skipFields := map[string]bool{
			"title":     true,
			"timestamp": true,
			"tasks":     true,
			"cards":     true,
			"fields":    true,
		}
		
		for key, value := range rawData {
			if !skipFields[key] {
				viewData.Fields[key] = value
			}
		}
	}

	return viewData, nil
}

func renderViewHTML(view View) (string, error) {
	tmpl, err := template.ParseFiles(view.Template)
	if err != nil {
		return "", err
	}

	viewData, err := loadViewData(view)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, viewData); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func getBool(m map[string]interface{}, key string, defaultValue bool) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return defaultValue
}

func getCurrentView() View {
	if len(views) == 0 {
		initViews()
	}
	if len(views) == 0 {
		return View{} // Empty view as fallback
	}
	return views[currentViewIndex]
}

func shouldRotate() bool {
	if len(views) <= 1 {
		return false
	}
	return time.Since(lastRotationTime) >= rotationInterval
}

func rotateView() {
	if len(views) <= 1 {
		return
	}
	currentViewIndex = (currentViewIndex + 1) % len(views)
	lastRotationTime = time.Now()
	log.Printf("Rotated to view: %s (index %d)", views[currentViewIndex].Name, currentViewIndex)
}

func getCurrentImagePath() string {
	view := getCurrentView()
	if view.Name == "" {
		return filepath.Join(config.Paths.OutputDir, "screen.png")
	}
	return filepath.Join(config.Paths.OutputDir, view.Name+".png")
}

