package content

import (
	"encoding/json"
	"os"
)

// Resume represents the portfolio resume data
type Resume struct {
	Name    string `json:"name"`
	Title   string `json:"title"`
	Tagline string `json:"tagline"`
	Contact struct {
		Email    string `json:"email"`
		Website  string `json:"website"`
		Github   string `json:"github"`
		LinkedIn string `json:"linkedin"`
		Twitter  string `json:"twitter"`
	} `json:"contact"`
	Summary    string       `json:"summary"`
	Experience []Experience `json:"experience"`
	Skills     struct {
		Languages []string `json:"languages"`
		Frontend  []string `json:"frontend"`
		Backend   []string `json:"backend"`
		Databases []string `json:"databases"`
		DevOps    []string `json:"devops"`
		Tools     []string `json:"tools"`
		Mobile    []string `json:"mobile"`
	} `json:"skills"`
	Education []struct {
		Institution string `json:"institution"`
		Degree      string `json:"degree"`
		Location    string `json:"location"`
		Period      string `json:"period"`
		Score       string `json:"score"`
	} `json:"education"`
	Achievements []string `json:"achievements"`
}

// Experience represents work experience
type Experience struct {
	Company    string   `json:"company"`
	Role       string   `json:"role"`
	Period     string   `json:"period"`
	Highlights []string `json:"highlights"`
}

// Project represents a single project
type Project struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tech        []string `json:"tech"`
	Status      string   `json:"status"`
	Links       struct {
		Demo   string `json:"demo,omitempty"`
		Github string `json:"github,omitempty"`
	} `json:"links"`
}

// Projects container
type Projects struct {
	Projects []Project `json:"projects"`
}

// Loader handles loading content from files
type Loader struct {
	basePath string
}

// NewLoader creates a content loader
func NewLoader(basePath string) *Loader {
	return &Loader{basePath: basePath}
}

// LoadResume reads and parses the resume JSON
func (l *Loader) LoadResume() (*Resume, error) {
	path := l.basePath + "/resume.json"
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var resume Resume
	if err := json.Unmarshal(data, &resume); err != nil {
		return nil, err
	}

	return &resume, nil
}

// LoadProjects reads and parses the projects JSON
func (l *Loader) LoadProjects() (*Projects, error) {
	path := l.basePath + "/projects.json"
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var projects Projects
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, err
	}

	return &projects, nil
}

// LoadBio reads the bio markdown file
func (l *Loader) LoadBio() (string, error) {
	path := l.basePath + "/bio.md"
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// GetProjectByID finds a project by its ID
func (p *Projects) GetProjectByID(id string) *Project {
	for _, project := range p.Projects {
		if project.ID == id {
			return &project
		}
	}
	return nil
}
