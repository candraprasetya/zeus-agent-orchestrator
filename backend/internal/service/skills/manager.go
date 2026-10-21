package skills

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var ErrSkillNotFound = errors.New("skill not found")

// frontmatter holds parsed YAML header fields from SKILL.md.
type frontmatter struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Triggers    []string `yaml:"triggers"`
}

// Manager manages installed and shared skills in the AO data directory.
type Manager struct {
	dataDir string
}

// New creates a new Manager backed by the given AO data directory.
func New(dataDir string) *Manager {
	return &Manager{dataDir: dataDir}
}

// SkillsDir returns the root directory where installed skills live.
func (m *Manager) SkillsDir() string {
	return filepath.Join(m.dataDir, "skills")
}

// List scans the skills directory and returns all discoverable skills.
func (m *Manager) List(ctx context.Context) ([]Skill, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	root := m.SkillsDir()
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return []Skill{}, nil
	}

	var results []Skill
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(d.Name(), "SKILL.md") {
			skillDir := filepath.Dir(p)
			skill, err := parseSkillFile(p, skillDir, root)
			if err == nil {
				results = append(results, skill)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan skills dir: %w", err)
	}
	return results, nil
}

// Get finds a skill by its name.
func (m *Manager) Get(ctx context.Context, name string) (Skill, error) {
	skills, err := m.List(ctx)
	if err != nil {
		return Skill{}, err
	}
	for _, s := range skills {
		if s.Name == name {
			return s, nil
		}
	}
	return Skill{}, ErrSkillNotFound
}

// Import copies or merges a directory containing skills into the AO skills store.
func (m *Manager) Import(ctx context.Context, srcPath string) ([]Skill, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	info, err := os.Stat(srcPath)
	if err != nil {
		return nil, fmt.Errorf("source path: %w", err)
	}
	if !info.IsDir() {
		return nil, errors.New("source path must be a directory")
	}

	destRoot := m.SkillsDir()
	if err := os.MkdirAll(destRoot, 0o750); err != nil {
		return nil, fmt.Errorf("create skills dest dir: %w", err)
	}

	// Walk srcPath to locate all SKILL.md files and copy each enclosing skill directory
	var imported []Skill
	err = filepath.WalkDir(srcPath, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Skip hidden dirs like .git
			if strings.HasPrefix(d.Name(), ".") && d.Name() != "." {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.EqualFold(d.Name(), "SKILL.md") {
			skillDir := filepath.Dir(p)
			// Read frontmatter to get skill name
			skill, err := parseSkillFile(p, skillDir, srcPath)
			if err != nil {
				return nil // skip malformed skill
			}
			name := skill.Name
			if name == "" {
				name = filepath.Base(skillDir)
			}

			destSkillDir := filepath.Join(destRoot, name)
			if err := copyDir(skillDir, destSkillDir); err != nil {
				return fmt.Errorf("copy skill %q: %w", name, err)
			}

			skill.Path = destSkillDir
			skill.IsManaged = true
			imported = append(imported, skill)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("import skills: %w", err)
	}

	return imported, nil
}

// parseSkillFile reads frontmatter from SKILL.md.
func parseSkillFile(filePath, skillDir, rootDir string) (Skill, error) {
	data, err := os.ReadFile(filePath) //nolint:gosec // controlled file scan
	if err != nil {
		return Skill{}, err
	}

	fm, err := extractFrontmatter(data)
	if err != nil {
		return Skill{}, err
	}

	name := fm.Name
	if name == "" {
		name = filepath.Base(skillDir)
	}

	// Derive category if in a subfolder relative to root
	rel, err := filepath.Rel(rootDir, skillDir)
	category := ""
	if err == nil && rel != "." {
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if len(parts) > 1 {
			category = parts[0]
		}
	}

	return Skill{
		Name:        name,
		Description: strings.TrimSpace(fm.Description),
		Path:        skillDir,
		Category:    category,
		Triggers:    fm.Triggers,
		IsManaged:   true,
	}, nil
}

func extractFrontmatter(data []byte) (frontmatter, error) {
	reader := bufio.NewReader(bytes.NewReader(data))
	firstLine, err := reader.ReadString('\n')
	if err != nil || strings.TrimSpace(firstLine) != "---" {
		return frontmatter{}, errors.New("no frontmatter start")
	}

	var header bytes.Buffer
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		if strings.TrimSpace(line) == "---" {
			break
		}
		header.WriteString(line)
	}

	var fm frontmatter
	if err := yaml.Unmarshal(header.Bytes(), &fm); err != nil {
		return frontmatter{}, err
	}
	return fm, nil
}

// copyDir recursively copies files and directories from src to dst.
func copyDir(src, dst string) error {
	if err := os.RemoveAll(dst); err != nil && !os.IsNotExist(err) {
		return err
	}
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o750)
		}
		return copyFile(p, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src) //nolint:gosec // file copy helper
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600) //nolint:gosec // file copy helper
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
