package db

import (
	"database/sql"
	"os"
	"testing"
)

func TestDatabaseOperations(t *testing.T) {
	// Use an in-memory database for testing
	tempFile := "test.db"
	defer os.Remove(tempFile)

	// Initialize the test database
	db, err := initTestDB(tempFile)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer db.Close()

	// Test AddProject
	projectName := "test-project"
	repoURL := "https://github.com/test/repo"
	path := "/tmp/test-project"

	id, err := AddProject(db, projectName, repoURL, path)
	if err != nil {
		t.Fatalf("Failed to add project: %v", err)
	}
	if id <= 0 {
		t.Fatalf("Expected positive project ID, got %d", id)
	}

	// Test GetProject
	project, err := GetProject(db, int(id))
	if err != nil {
		t.Fatalf("Failed to get project: %v", err)
	}
	if project.Name != projectName {
		t.Errorf("Expected project name %s, got %s", projectName, project.Name)
	}
	if project.RepoURL != repoURL {
		t.Errorf("Expected repo URL %s, got %s", repoURL, project.RepoURL)
	}
	if project.Path != path {
		t.Errorf("Expected path %s, got %s", path, project.Path)
	}

	// Test GetAllProjects
	projects, err := GetAllProjects(db)
	if err != nil {
		t.Fatalf("Failed to get all projects: %v", err)
	}
	if len(projects) != 1 {
		t.Errorf("Expected 1 project, got %d", len(projects))
	}

	// Test UpdateProject
	updatedName := "updated-project"
	project.Name = updatedName
	err = UpdateProject(db, project)
	if err != nil {
		t.Fatalf("Failed to update project: %v", err)
	}

	updatedProject, err := GetProject(db, project.ID)
	if err != nil {
		t.Fatalf("Failed to get updated project: %v", err)
	}
	if updatedProject.Name != updatedName {
		t.Errorf("Expected updated name %s, got %s", updatedName, updatedProject.Name)
	}

	// Test DeleteProject
	err = DeleteProject(db, project.ID)
	if err != nil {
		t.Fatalf("Failed to delete project: %v", err)
	}

	projects, err = GetAllProjects(db)
	if err != nil {
		t.Fatalf("Failed to get all projects after deletion: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("Expected 0 projects after deletion, got %d", len(projects))
	}
}

// initTestDB initializes a test database with the schema
func initTestDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			repo_url TEXT NOT NULL,
			path TEXT NOT NULL
		)`)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
} 