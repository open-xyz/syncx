package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/open-xyz/syncx/internal/balancing"
	"github.com/open-xyz/syncx/internal/config"
	"github.com/open-xyz/syncx/internal/db"
	"github.com/open-xyz/syncx/internal/hosting"
	"github.com/open-xyz/syncx/internal/scanning"
)

// ProjectData represents the combined project data shown in the UI
type ProjectData struct {
	db.Project
	ScanResult *scanning.ScanResult `json:"scan_result,omitempty"`
}

// PageData contains data passed to HTML templates
type PageData struct {
	Projects     []ProjectData
	FlashMessage string
	FlashClass   string
}

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config/syncx.yaml", "Path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	database := db.InitDB()
	defer database.Close()

	// Initialize hosting service
	if err := hosting.Initialize(); err != nil {
		log.Fatalf("Failed to initialize hosting service: %v", err)
	}

	// Initialize load balancer
	balancer, err := balancing.NewBalancer(cfg.Balancer.Endpoints)
	if err != nil {
		log.Fatalf("Failed to initialize load balancer: %v", err)
	}

	// Create router
	router := mux.NewRouter()

	// Set up templates
	tmpl := template.New("").Funcs(template.FuncMap{
		"lower": strings.ToLower,
	})
	tmpl, err = tmpl.ParseGlob("web/templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	// Dashboard route
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		projects, err := db.GetAllProjects(database)
		if err != nil {
			http.Error(w, "Error fetching projects", http.StatusInternalServerError)
			return
		}

		// Create combined project data
		projectData := make([]ProjectData, len(projects))
		for i, p := range projects {
			projectData[i] = ProjectData{Project: p}
			// Add scan results if available (in a real app, these would be stored in the DB)
			// For now, we'll just create empty placeholders
		}

		data := PageData{
			Projects: projectData,
		}

		tmpl.ExecuteTemplate(w, "dashboard.html", data)
	}).Methods("GET")

	// Add project route
	router.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		name := r.FormValue("name")
		repoURL := r.FormValue("repo_url")

		if name == "" || repoURL == "" {
			http.Error(w, "Name and repository URL are required", http.StatusBadRequest)
			return
		}

		// Clone the repository
		projectPath, err := hosting.CloneProject(name, repoURL)
		if err != nil {
			log.Printf("Error cloning repository: %v", err)
			http.Redirect(w, r, "/?error=Failed+to+clone+repository:+"+url.QueryEscape(err.Error()), http.StatusSeeOther)
			return
		}

		// Add project to database
		projectID, err := db.AddProject(database, name, repoURL, projectPath)
		if err != nil {
			log.Printf("Error adding project to database: %v", err)
			http.Redirect(w, r, "/?error=Failed+to+add+project", http.StatusSeeOther)
			return
		}

		// Check if Trivy is installed before attempting auto-scan
		if cfg.Scanning.AutoScanOnAdd && scanning.IsTrivyInstalled() {
			go func() {
				// Run scan in background
				_, err := scanning.ScanProject(int(projectID), projectPath)
				if err != nil {
					log.Printf("Auto-scan failed for project %d: %v", projectID, err)
				}
			}()
		} else if cfg.Scanning.AutoScanOnAdd && !scanning.IsTrivyInstalled() {
			log.Printf("Auto-scan skipped for project %d: Trivy not installed", projectID)
		}

		// Redirect back to dashboard
		http.Redirect(w, r, "/?success=Project+added", http.StatusSeeOther)
	}).Methods("POST")

	// Scan project route
	router.HandleFunc("/scan/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid project ID", http.StatusBadRequest)
			return
		}

		// Get project from database
		project, err := db.GetProject(database, id)
		if err != nil {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}

		// Check if Trivy is installed
		if !scanning.IsTrivyInstalled() {
			data := map[string]interface{}{
				"error":        "Trivy is not installed",
				"instructions": scanning.GetTrivyInstallInstructions(),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(data)
			return
		}

		// Scan the project
		result, err := scanning.ScanProject(project.ID, project.Path)
		if err != nil {
			log.Printf("Scan failed for project %d: %v", project.ID, err)
			// Return a more descriptive error, but still return a result
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(result)
			return
		}

		// Return the scan results
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}).Methods("GET")

	// Update project route
	router.HandleFunc("/update/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid project ID", http.StatusBadRequest)
			return
		}

		// Get project from database
		project, err := db.GetProject(database, id)
		if err != nil {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}

		// Update the project
		if err := hosting.UpdateProject(project.Path); err != nil {
			log.Printf("Update failed for project %d: %v", project.ID, err)
			http.Redirect(w, r, "/?error=Update+failed", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/?success=Project+updated", http.StatusSeeOther)
	}).Methods("GET")

	// Delete project route
	router.HandleFunc("/delete/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid project ID", http.StatusBadRequest)
			return
		}

		// Delete from database
		if err := db.DeleteProject(database, id); err != nil {
			http.Error(w, "Failed to delete project", http.StatusInternalServerError)
			return
		}

		// Note: In a real app, you might want to also delete the project files

		http.Redirect(w, r, "/?success=Project+deleted", http.StatusSeeOther)
	}).Methods("GET")

	// View project route (redirects to the load balancer)
	router.HandleFunc("/view/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid project ID", http.StatusBadRequest)
			return
		}

		// Get project from database
		project, err := db.GetProject(database, id)
		if err != nil {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}

		// In a real app, you'd set up proper routing through the load balancer
		// For now, we'll just redirect to a simple file server
		http.Redirect(w, r, "/projects/"+filepath.Base(project.Path), http.StatusSeeOther)
	}).Methods("GET")

	// Set up the load balancer route
	router.PathPrefix("/lb/").Handler(http.StripPrefix("/lb", balancer))

	// Set up static file serving for projects
	router.PathPrefix("/projects/").Handler(
		http.StripPrefix("/projects/", http.FileServer(http.Dir(hosting.ProjectsDir))))

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting SyncX server on %s", addr)
	
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	log.Fatal(server.ListenAndServe())
} 