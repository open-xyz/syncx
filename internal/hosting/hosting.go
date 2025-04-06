package hosting

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

const ProjectsDir = "./projects"

// Initialize creates the projects directory if it doesn't exist
func Initialize() error {
	if _, err := os.Stat(ProjectsDir); os.IsNotExist(err) {
		return os.MkdirAll(ProjectsDir, 0755)
	}
	return nil
}

// CloneProject clones a git repository to the projects directory
func CloneProject(name, repoURL string) (string, error) {
	projectPath := filepath.Join(ProjectsDir, name)
	
	// Create project directory if it doesn't exist
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return "", err
	}
	
	// Clone the repository
	log.Printf("Cloning %s into %s", repoURL, projectPath)
	_, err := git.PlainClone(projectPath, false, &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
	})
	
	if err != nil {
		return "", err
	}
	
	return projectPath, nil
}

// UpdateProject pulls latest changes for a project
func UpdateProject(path string) error {
	r, err := git.PlainOpen(path)
	if err != nil {
		return err
	}
	
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	
	err = w.Pull(&git.PullOptions{})
	if err == git.NoErrAlreadyUpToDate {
		log.Println("Repository already up to date")
		return nil
	}
	
	return err
}

// CreateFileServer returns an HTTP handler that serves files from a project directory
func CreateFileServer(projectPath string) http.Handler {
	return http.FileServer(http.Dir(projectPath))
}

// ServeProject creates a simple HTTP server for a project
func ServeProject(projectPath string, port int) error {
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Serving project at %s on %s", projectPath, addr)
	
	fileServer := CreateFileServer(projectPath)
	return http.ListenAndServe(addr, fileServer)
} 