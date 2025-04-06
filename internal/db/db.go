package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Project struct {
	ID      int
	Name    string
	RepoURL string
	Path    string
}

func InitDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./syncx.db")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			repo_url TEXT NOT NULL,
			path TEXT NOT NULL
		)`)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// AddProject adds a new project to the database
func AddProject(db *sql.DB, name, repoURL, path string) (int64, error) {
	stmt, err := db.Prepare("INSERT INTO projects(name, repo_url, path) VALUES(?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	
	result, err := stmt.Exec(name, repoURL, path)
	if err != nil {
		return 0, err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	
	return id, nil
}

// GetProject retrieves a project by ID
func GetProject(db *sql.DB, id int) (Project, error) {
	var project Project
	err := db.QueryRow("SELECT id, name, repo_url, path FROM projects WHERE id = ?", id).
		Scan(&project.ID, &project.Name, &project.RepoURL, &project.Path)
	return project, err
}

// GetAllProjects retrieves all projects
func GetAllProjects(db *sql.DB) ([]Project, error) {
	rows, err := db.Query("SELECT id, name, repo_url, path FROM projects")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.RepoURL, &p.Path); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	
	return projects, nil
}

// UpdateProject updates an existing project
func UpdateProject(db *sql.DB, project Project) error {
	_, err := db.Exec("UPDATE projects SET name = ?, repo_url = ?, path = ? WHERE id = ?",
		project.Name, project.RepoURL, project.Path, project.ID)
	return err
}

// DeleteProject removes a project from the database
func DeleteProject(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM projects WHERE id = ?", id)
	return err
} 