package skills

import "context"

// Skill represents a discoverable agent skill with frontmatter metadata.
type Skill struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Path        string   `json:"path"`
	Category    string   `json:"category,omitempty"`
	Triggers    []string `json:"triggers,omitempty"`
	IsManaged   bool     `json:"isManaged"`
}

// Service provides management and discovery of skills across AO.
type Service interface {
	List(ctx context.Context) ([]Skill, error)
	Get(ctx context.Context, name string) (Skill, error)
	Import(ctx context.Context, srcPath string) ([]Skill, error)
}
