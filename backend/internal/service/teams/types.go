package teams

import "context"

// RoleMember defines a specialized squad member with preferred agent harness and assigned skills.
type RoleMember struct {
	Role         string   `json:"role"`
	Harness      string   `json:"harness,omitempty"`
	Model        string   `json:"model,omitempty"`
	Skills       []string `json:"skills"`
	Instructions string   `json:"instructions,omitempty"`
}

// Team defines a squad of specialized roles (e.g. "zeus-team-ios").
type Team struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Roles       []RoleMember `json:"roles"`
}

// Service provides persistence and query operations for Teams and Roles.
type Service interface {
	List(ctx context.Context) ([]Team, error)
	Get(ctx context.Context, id string) (Team, error)
	Save(ctx context.Context, team Team) (Team, error)
	Delete(ctx context.Context, id string) error
	GetTeamPrompt(ctx context.Context, teamID string) (string, error)
}
