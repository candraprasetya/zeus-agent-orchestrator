package teams

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	ErrTeamNotFound = errors.New("team not found")
	ErrInvalidTeam  = errors.New("team ID and name are required")
)

// Store provides JSON file-backed persistence for teams under <dataDir>/teams.json.
type Store struct {
	path string
	mu   sync.RWMutex
}

// NewStore initializes a new teams store.
func NewStore(dataDir string) *Store {
	return &Store{
		path: filepath.Join(dataDir, "teams.json"),
	}
}

func (s *Store) load() (map[string]Team, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		// Provide default Zeus squads out of the box if no teams file exists yet
		defaults := defaultTeams()
		_ = s.saveUnlocked(defaults)
		return defaults, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read teams: %w", err)
	}
	var list []Team
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("decode teams: %w", err)
	}
	res := make(map[string]Team, len(list))
	for _, t := range list {
		res[t.ID] = t
	}
	return res, nil
}

func (s *Store) saveUnlocked(teams map[string]Team) error {
	list := make([]Team, 0, len(teams))
	for _, t := range teams {
		list = append(list, t)
	}
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o750); err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}

func defaultTeams() map[string]Team {
	return map[string]Team{
		"zeus-team-ios": {
			ID:          "zeus-team-ios",
			Name:        "Zeus iOS Squad",
			Description: "Specialized iOS team for CloveUI, Swift Domain, and Swinject Data layer.",
			Roles: []RoleMember{
				{
					Role:         "ui-engineer",
					Harness:      "mage",
					Skills:       []string{"zeus-ios-presentation", "zeus-ios-design-system"},
					Instructions: "Specialized in iOS CloveUI, MVVM-C, ViewBinding, ScreenV2, and UI components.",
				},
				{
					Role:         "domain-engineer",
					Harness:      "opencode",
					Skills:       []string{"zeus-ios-domain"},
					Instructions: "Specialized in Pure Swift Domain layer: Entities, UseCases, and Repository Protocols.",
				},
				{
					Role:         "data-engineer",
					Harness:      "mage",
					Skills:       []string{"zeus-ios-data"},
					Instructions: "Specialized in iOS Data layer: Swinject DI, RequestEndpoints, DTOs, and Repositories.",
				},
			},
		},
		"zeus-team-android": {
			ID:          "zeus-team-android",
			Name:        "Zeus Android Squad",
			Description: "Specialized Android team for Jetpack Compose CloveUI, Kotlin Domain, and Hilt/Room Data layer.",
			Roles: []RoleMember{
				{
					Role:         "ui-engineer",
					Harness:      "mage",
					Skills:       []string{"zeus-android-presentation", "zeus-android-design-system"},
					Instructions: "Specialized in Android Jetpack Compose, CloveTheme tokens, and Presentation MVI/MVVM.",
				},
				{
					Role:         "domain-engineer",
					Harness:      "opencode",
					Skills:       []string{"zeus-android-domain"},
					Instructions: "Specialized in Pure Kotlin Domain: Entities, UseCases, and Repository Interfaces.",
				},
				{
					Role:         "data-engineer",
					Harness:      "mage",
					Skills:       []string{"zeus-android-data"},
					Instructions: "Specialized in Android Data layer: Hilt DI, Room DAOs, Retrofit DTOs, and Repositories.",
				},
			},
		},
	}
}

// Manager implements the Service interface.
type Manager struct {
	store *Store
}

// New creates a new Teams manager.
func New(dataDir string) *Manager {
	return &Manager{store: NewStore(dataDir)}
}

func (m *Manager) List(ctx context.Context) ([]Team, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	m.store.mu.RLock()
	defer m.store.mu.RUnlock()

	teams, err := m.store.load()
	if err != nil {
		return nil, err
	}
	res := make([]Team, 0, len(teams))
	for _, t := range teams {
		res = append(res, t)
	}
	return res, nil
}

func (m *Manager) Get(ctx context.Context, id string) (Team, error) {
	if err := ctx.Err(); err != nil {
		return Team{}, err
	}
	m.store.mu.RLock()
	defer m.store.mu.RUnlock()

	teams, err := m.store.load()
	if err != nil {
		return Team{}, err
	}
	team, ok := teams[id]
	if !ok {
		return Team{}, ErrTeamNotFound
	}
	return team, nil
}

func (m *Manager) Save(ctx context.Context, team Team) (Team, error) {
	if err := ctx.Err(); err != nil {
		return Team{}, err
	}
	if strings.TrimSpace(team.ID) == "" || strings.TrimSpace(team.Name) == "" {
		return Team{}, ErrInvalidTeam
	}
	m.store.mu.Lock()
	defer m.store.mu.Unlock()

	teams, err := m.store.load()
	if err != nil {
		teams = make(map[string]Team)
	}
	teams[team.ID] = team
	if err := m.store.saveUnlocked(teams); err != nil {
		return Team{}, err
	}
	return team, nil
}

func (m *Manager) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.store.mu.Lock()
	defer m.store.mu.Unlock()

	teams, err := m.store.load()
	if err != nil {
		return err
	}
	if _, ok := teams[id]; !ok {
		return ErrTeamNotFound
	}
	delete(teams, id)
	return m.store.saveUnlocked(teams)
}

// GetTeamPrompt generates a prompt section instructing the orchestrator how to delegate tasks to this team's roles.
func (m *Manager) GetTeamPrompt(ctx context.Context, teamID string) (string, error) {
	team, err := m.Get(ctx, teamID)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Assigned Squad: %s\n", team.Name))
	if team.Description != "" {
		sb.WriteString(team.Description + "\n\n")
	}
	sb.WriteString("When implementing an implementation plan or delegating tasks, delegate each task to the corresponding specialized role using `ao spawn`:\n\n")

	for _, role := range team.Roles {
		harness := role.Harness
		if harness == "" {
			harness = "mage"
		}
		skillsList := strings.Join(role.Skills, ", ")
		sb.WriteString(fmt.Sprintf("### Role: `%s`\n", role.Role))
		sb.WriteString(fmt.Sprintf("- **Harness**: `%s`\n", harness))
		if len(role.Skills) > 0 {
			sb.WriteString(fmt.Sprintf("- **Assigned Skills**: `%s`\n", skillsList))
		}
		if role.Instructions != "" {
			sb.WriteString(fmt.Sprintf("- **Scope**: %s\n", role.Instructions))
		}
		sb.WriteString(fmt.Sprintf("- **Spawn Example**: `ao spawn --harness %s --name \"<task-name>\" --prompt \"You are the %s. Use skills [%s]. Task: ...\"`\n\n", harness, role.Role, skillsList))
	}

	return sb.String(), nil
}
