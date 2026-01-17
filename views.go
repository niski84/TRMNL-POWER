package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
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

// validateTemplate checks if a rendered template follows required structure and constraints
func validateTemplate(html string, viewName string) error {
	errors := []string{}
	
	// Check for required viewport meta tag
	viewportPattern := fmt.Sprintf(`width=%d, height=%d`, config.Render.Width, config.Render.Height)
	if !strings.Contains(html, viewportPattern) && 
	   !strings.Contains(html, "width=800, height=480") {
		errors = append(errors, fmt.Sprintf("template must include viewport meta tag: width=%d, height=%d", 
			config.Render.Width, config.Render.Height))
	}
	
	// Check for {{.Styles}} injection point (should be in rendered output as actual styles)
	// Note: We check for the CSS constant content as a proxy
	if !strings.Contains(html, "ENFORCED:") {
		log.Printf("Warning: Template '%s' may not include {{.Styles}} - base styles might not be applied", viewName)
	}
	
	// Check for body tag
	if !strings.Contains(strings.ToLower(html), "<body") {
		errors = append(errors, "template must include <body> tag")
	}
	
	// Check for dangerous body width/height overrides that might break layout
	if strings.Contains(html, "body {") {
		bodyStyles := extractBodyStyles(html)
		if strings.Contains(bodyStyles, "width:") && !strings.Contains(bodyStyles, "800px") {
			log.Printf("Warning: Template '%s' has custom body width - may break layout constraints", viewName)
		}
		if strings.Contains(bodyStyles, "height:") && !strings.Contains(bodyStyles, "480px") {
			log.Printf("Warning: Template '%s' has custom body height - may break layout constraints", viewName)
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("template validation failed for '%s': %s", viewName, strings.Join(errors, "; "))
	}
	
	return nil
}

// extractBodyStyles extracts CSS rules from body {} block
func extractBodyStyles(html string) string {
	// Simple extraction - find body { ... }
	start := strings.Index(html, "body {")
	if start == -1 {
		return ""
	}
	
	// Find the closing brace
	depth := 0
	end := start
	for i := start; i < len(html); i++ {
		if html[i] == '{' {
			depth++
		} else if html[i] == '}' {
			depth--
			if depth == 0 {
				end = i
				break
			}
		}
	}
	
	if end > start {
		return html[start:end+1]
	}
	return ""
}

func renderViewHTML(view View) (string, error) {
	tmpl, err := template.ParseFiles(view.Template)
	if err != nil {
		return "", fmt.Errorf("failed to parse template '%s': %w", view.Template, err)
	}

	viewData, err := loadViewData(view)
	if err != nil {
		return "", fmt.Errorf("failed to load data for view '%s': %w", view.Name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, viewData); err != nil {
		return "", fmt.Errorf("failed to execute template '%s': %w", view.Name, err)
	}

	html := buf.String()
	
	// Validate template structure and constraints
	if err := validateTemplate(html, view.Name); err != nil {
		return "", err
	}
	
	return html, nil
}

func getBool(m map[string]interface{}, key string, defaultValue bool) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return defaultValue
}

// validateViewBeforeRender checks template and data files before rendering
// Returns warnings (non-fatal) that can be logged
func validateViewBeforeRender(view View) []string {
	warnings := []string{}
	
	// Check template file exists
	if _, err := os.Stat(view.Template); os.IsNotExist(err) {
		warnings = append(warnings, fmt.Sprintf("Template file missing: %s", view.Template))
		return warnings // Return early if template doesn't exist
	}
	
	// Check data file exists
	if _, err := os.Stat(view.DataPath); os.IsNotExist(err) {
		warnings = append(warnings, fmt.Sprintf("Data file missing: %s", view.DataPath))
	}
	
	// Read template and check for common issues
	templateContent, err := os.ReadFile(view.Template)
	if err == nil {
		content := string(templateContent)
		
		// Check for viewport meta tag
		viewportPattern := fmt.Sprintf(`width=%d, height=%d`, config.Render.Width, config.Render.Height)
		if !strings.Contains(content, viewportPattern) && 
		   !strings.Contains(content, "width=800, height=480") {
			warnings = append(warnings, "Missing or incorrect viewport meta tag - should match render dimensions")
		}
		
		// Check for styles injection
		if !strings.Contains(content, "{{.Styles}}") {
			warnings = append(warnings, "Missing {{.Styles}} - base styles won't be applied automatically")
		}
		
		// Check for dangerous CSS patterns
		if strings.Contains(content, "body {") {
			bodyStyles := extractBodyStyles(content)
			if strings.Contains(bodyStyles, "width:") && !strings.Contains(bodyStyles, "800") {
				warnings = append(warnings, "Custom body width detected - may break layout constraints (use base styles instead)")
			}
			if strings.Contains(bodyStyles, "height:") && !strings.Contains(bodyStyles, "480") {
				warnings = append(warnings, "Custom body height detected - may break layout constraints (use base styles instead)")
			}
			if strings.Contains(bodyStyles, "overflow:") && strings.Contains(bodyStyles, "visible") {
				warnings = append(warnings, "body overflow:visible detected - content may exceed display bounds")
			}
		}
		
		// Check for required structure
		if !strings.Contains(content, "<body") {
			warnings = append(warnings, "Missing <body> tag - required for proper layout")
		}
		
		if !strings.Contains(content, "header") && !strings.Contains(content, ".header") {
			warnings = append(warnings, "Missing header structure - consider using <div class=\"header\">")
		}
		
		if !strings.Contains(content, "content") && !strings.Contains(content, ".content") {
			warnings = append(warnings, "Missing content container - consider using <div class=\"content\">")
		}
	} else {
		warnings = append(warnings, fmt.Sprintf("Could not read template file: %v", err))
	}
	
	return warnings
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

// generateTemplateBoilerplate creates a new template file with proper structure
func generateTemplateBoilerplate(templateName string) {
	// Load config if not already loaded
	if config.Render.Width == 0 {
		config.Render.Width = 800
		config.Render.Height = 480
	}
	
	boilerplate := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=%d, height=%d, initial-scale=1.0">
  <title>%s</title>
  <style>
{{.Styles}}
    /* Your custom styles here - they will be merged with base styles */
    /* Note: Base styles enforce size constraints, so avoid overriding:
       - body width/height (use base styles)
       - header height (fixed at 50px)
       - content max-height (max 420px)
    */
  </style>
</head>
<body>
  <div class="header">
    <div class="header-title">{{.Title}}</div>
    <div class="header-timestamp">{{.Timestamp}}</div>
  </div>
  <div class="content">
    <!-- Your content here -->
    <!-- 
      Available data:
      - {{.Title}} - Page title from JSON
      - {{.Timestamp}} - Timestamp from JSON
      - {{index .Fields "FieldName"}} - Access fields from JSON data
      
      Example dashboard card:
      <div class="card">
        <div class="card-label">My Metric</div>
        <div class="card-value-container">
          <div class="card-value">{{index .Fields "MyMetric"}}</div>
          {{if index .Fields "MyMetricUnit"}}<span class="card-unit">{{index .Fields "MyMetricUnit"}}</span>{{end}}
        </div>
      </div>
      
      Example todo item:
      {{range .Tasks}}
      <div class="todo-item {{if .Completed}}completed{{end}}">
        <span class="todo-checkbox">{{if .Completed}}✓{{else}}○{{end}}</span>
        <span class="todo-text">{{.Text}}</span>
        {{if .Category}}<span class="todo-category">{{.Category}}</span>{{end}}
      </div>
      {{end}}
    -->
  </div>
</body>
</html>`, config.Render.Width, config.Render.Height, templateName)
	
	// Create templates directory if it doesn't exist
	templateDir := "./templates"
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		log.Fatalf("Failed to create templates directory: %v", err)
	}
	
	// Write template file
	templatePath := filepath.Join(templateDir, templateName+".html")
	if _, err := os.Stat(templatePath); err == nil {
		log.Fatalf("Template file already exists: %s", templatePath)
	}
	
	if err := os.WriteFile(templatePath, []byte(boilerplate), 0644); err != nil {
		log.Fatalf("Failed to write template file: %v", err)
	}
	
	// Generate sample JSON data file
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}
	
	sampleJSON := fmt.Sprintf(`{
  "title": "%s",
  "timestamp": "2024-01-16 12:00:00",
  "fields": {
    "SampleField": 42,
    "SampleFieldUnit": "%%",
    "AnotherField": "Value",
    "AnotherFieldUnit": "unit"
  }
}`, templateName)
	
	dataPath := filepath.Join(dataDir, templateName+".json")
	if _, err := os.Stat(dataPath); err == nil {
		log.Printf("Note: Data file already exists: %s (not overwriting)", dataPath)
	} else {
		if err := os.WriteFile(dataPath, []byte(sampleJSON), 0644); err != nil {
			log.Printf("Warning: Failed to write data file: %v", err)
		} else {
			log.Printf("Created sample data file: %s", dataPath)
		}
	}
	
	log.Printf("✓ Template created: %s", templatePath)
	log.Printf("✓ Sample data created: %s", dataPath)
	log.Printf("\nNext steps:")
	log.Printf("1. Edit the template: %s", templatePath)
	log.Printf("2. Edit the data file: %s", dataPath)
	log.Printf("3. Add view to initViews() in views.go:")
	log.Printf("   {")
	log.Printf("       Name:     \"%s\",", templateName)
	log.Printf("       Template: \"./templates/%s.html\",", templateName)
	log.Printf("       DataPath: \"./data/%s.json\",", templateName)
	log.Printf("   }")
}

// validateAllTemplates validates all registered templates and reports issues
func validateAllTemplates() {
	initViews()
	
	if len(views) == 0 {
		log.Println("No views configured.")
		return
	}
	
	log.Printf("Validating %d template(s)...\n", len(views))
	
	allValid := true
	for _, view := range views {
		log.Printf("\n📋 Template: %s", view.Name)
		log.Printf("   Template: %s", view.Template)
		log.Printf("   Data:     %s", view.DataPath)
		
		warnings := validateViewBeforeRender(view)
		if len(warnings) > 0 {
			allValid = false
			log.Printf("   ⚠️  Warnings:")
			for _, warning := range warnings {
				log.Printf("      - %s", warning)
			}
		} else {
			log.Printf("   ✓ Valid")
		}
	}
	
	if allValid {
		log.Printf("\n✓ All templates are valid!")
	} else {
		log.Printf("\n⚠️  Some templates have warnings (non-fatal, but review recommended)")
	}
}

