package mcp

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

type Manager struct {
	homeDir string
	mu      sync.RWMutex
}

// New creates a new MCP Manager. If homeDir is empty, os.UserHomeDir() is used.
func New(homeDir string) *Manager {
	if homeDir == "" {
		homeDir, _ = os.UserHomeDir()
	}
	return &Manager{homeDir: homeDir}
}

func normalizeHarness(h string) string {
	h = strings.ToLower(strings.TrimSpace(h))
	switch h {
	case "opencode", "opencode-1", "opencode1":
		return "opencode"
	case "opencode-v2", "opencode2", "opencodev2":
		return "opencode-v2"
	case "claude", "claude-code", "claudecode":
		return "claude-code"
	case "cursor":
		return "cursor"
	case "mage":
		return "mage"
	case "agy", "antigravity":
		return "agy"
	default:
		return h
	}
}

func (m *Manager) opencodeConfigPath() string {
	primary := filepath.Join(m.homeDir, ".config", "opencode", "opencode.json")
	if _, err := os.Stat(primary); err == nil {
		return primary
	}
	legacy := filepath.Join(m.homeDir, ".opencode", "opencode.json")
	if _, err := os.Stat(legacy); err == nil {
		return legacy
	}
	return primary
}

func (m *Manager) mageConfigPath() string {
	primary := filepath.Join(m.homeDir, ".mage", "mage.json")
	if _, err := os.Stat(primary); err == nil {
		return primary
	}
	legacy := filepath.Join(m.homeDir, ".config", "mage", "mage.json")
	if _, err := os.Stat(legacy); err == nil {
		return legacy
	}
	return primary
}

func (m *Manager) agyConfigPath() string {
	return filepath.Join(m.homeDir, ".gemini", "config", "mcp_config.json")
}

func (m *Manager) claudeConfigPath() string {
	return filepath.Join(m.homeDir, ".claude.json")
}

func (m *Manager) cursorConfigPath() string {
	return filepath.Join(m.homeDir, ".cursor", "mcp.json")
}

func (m *Manager) List(ctx context.Context, harness string) ([]ServerConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	target := normalizeHarness(harness)
	var results []ServerConfig

	if target == "" || target == "all" || target == "opencode" {
		items, err := m.listOpencodeCompatible(m.opencodeConfigPath(), "opencode")
		if err == nil {
			results = append(results, items...)
		} else if !errors.Is(err, os.ErrNotExist) && target == "opencode" {
			return nil, err
		}
	}

	if target == "all" || target == "opencode-v2" {
		items, err := m.listOpencodeCompatible(m.opencodeConfigPath(), "opencode-v2")
		if err == nil {
			results = append(results, items...)
		} else if !errors.Is(err, os.ErrNotExist) && target == "opencode-v2" {
			return nil, err
		}
	}

	if target == "all" || target == "mage" {
		items, err := m.listOpencodeCompatible(m.mageConfigPath(), "mage")
		if err == nil {
			results = append(results, items...)
		} else if !errors.Is(err, os.ErrNotExist) && target == "mage" {
			return nil, err
		}
	}

	if target == "all" || target == "agy" {
		items, err := m.listAgy()
		if err == nil {
			results = append(results, items...)
		} else if !errors.Is(err, os.ErrNotExist) && target == "agy" {
			return nil, err
		}
	}

	if target == "all" || target == "claude-code" {
		items, err := m.listClaude()
		if err == nil {
			results = append(results, items...)
		} else if !errors.Is(err, os.ErrNotExist) && target == "claude-code" {
			return nil, err
		}
	}

	if target == "all" || target == "cursor" {
		items, err := m.listMcpServersCompatible(m.cursorConfigPath(), "cursor")
		if err == nil {
			results = append(results, items...)
		} else if !errors.Is(err, os.ErrNotExist) && target == "cursor" {
			return nil, err
		}
	}

	// For specific non-built-in target, check ~/.config/<target>/mcp.json
	if target != "" && target != "all" && target != "opencode" && target != "opencode-v2" && target != "mage" && target != "agy" && target != "claude-code" && target != "cursor" {
		cfgPath := filepath.Join(m.homeDir, ".config", target, "mcp.json")
		items, _ := m.listMcpServersCompatible(cfgPath, target)
		results = append(results, items...)
	}

	return results, nil
}

func parseOpencodeServerEntry(out *[]ServerConfig, cfgPath, harnessName, name string, serverMap map[string]any) {
	srvType, _ := serverMap["type"].(string)
	if srvType == "" {
		srvType = "local"
	}

	enabled := true
	if val, exists := serverMap["enabled"]; exists {
		if b, isBool := val.(bool); isBool {
			enabled = b
		}
	}

	cfg := ServerConfig{
		Name:       name,
		Harness:    harnessName,
		Type:       srvType,
		ConfigPath: cfgPath,
		Enabled:    enabled,
	}

	if srvType == "remote" {
		cfg.URL, _ = serverMap["url"].(string)
	} else {
		if cmdArr, ok := serverMap["command"].([]any); ok && len(cmdArr) > 0 {
			cfg.Command = fmt.Sprint(cmdArr[0])
			for _, a := range cmdArr[1:] {
				cfg.Args = append(cfg.Args, fmt.Sprint(a))
			}
		} else if cmdStr, ok := serverMap["command"].(string); ok {
			cfg.Command = cmdStr
		}

		if envMap, ok := serverMap["environment"].(map[string]any); ok {
			cfg.Env = make(map[string]string, len(envMap))
			for k, v := range envMap {
				cfg.Env[k] = fmt.Sprint(v)
			}
		}
	}

	*out = append(*out, cfg)
}

func (m *Manager) listOpencodeCompatible(cfgPath, harnessName string) ([]ServerConfig, error) {
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, err
	}

	mcpVal, ok := root["mcp"].(map[string]any)
	if !ok {
		return nil, nil
	}

	var out []ServerConfig
	for name, item := range mcpVal {
		serverMap, ok := item.(map[string]any)
		if !ok {
			continue
		}

		if name == "servers" {
			if _, hasCmd := serverMap["command"]; !hasCmd {
				if _, hasURL := serverMap["url"]; !hasURL {
					for subName, subItem := range serverMap {
						if subMap, isMap := subItem.(map[string]any); isMap {
							parseOpencodeServerEntry(&out, cfgPath, harnessName, subName, subMap)
						}
					}
					continue
				}
			}
		}

		parseOpencodeServerEntry(&out, cfgPath, harnessName, name, serverMap)
	}
	return out, nil
}

func (m *Manager) listAgy() ([]ServerConfig, error) {
	cfgPath := m.agyConfigPath()
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, err
	}

	mcpServers, ok := root["mcpServers"].(map[string]any)
	if !ok {
		return nil, nil
	}

	var out []ServerConfig
	for name, item := range mcpServers {
		serverMap, ok := item.(map[string]any)
		if !ok {
			continue
		}

		url, _ := serverMap["url"].(string)
		srvType := "local"
		if url != "" {
			srvType = "remote"
		}

		cfg := ServerConfig{
			Name:       name,
			Harness:    "agy",
			Type:       srvType,
			URL:        url,
			ConfigPath: cfgPath,
			Enabled:    true,
		}

		if srvType == "local" {
			cfg.Command, _ = serverMap["command"].(string)
			if args, ok := serverMap["args"].([]any); ok {
				for _, a := range args {
					cfg.Args = append(cfg.Args, fmt.Sprint(a))
				}
			}
			if envMap, ok := serverMap["env"].(map[string]any); ok {
				cfg.Env = make(map[string]string, len(envMap))
				for k, v := range envMap {
					cfg.Env[k] = fmt.Sprint(v)
				}
			}
		}

		out = append(out, cfg)
	}
	return out, nil
}

func (m *Manager) listClaude() ([]ServerConfig, error) {
	return m.listMcpServersCompatible(m.claudeConfigPath(), "claude-code")
}

func (m *Manager) listMcpServersCompatible(cfgPath, harnessName string) ([]ServerConfig, error) {
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, err
	}

	mcpServers, ok := root["mcpServers"].(map[string]any)
	if !ok {
		return nil, nil
	}

	var out []ServerConfig
	for name, item := range mcpServers {
		serverMap, ok := item.(map[string]any)
		if !ok {
			continue
		}

		url, _ := serverMap["url"].(string)
		srvType := "local"
		if url != "" {
			srvType = "remote"
		}

		cfg := ServerConfig{
			Name:       name,
			Harness:    harnessName,
			Type:       srvType,
			URL:        url,
			ConfigPath: cfgPath,
			Enabled:    true,
		}

		if srvType == "local" {
			cfg.Command, _ = serverMap["command"].(string)
			if args, ok := serverMap["args"].([]any); ok {
				for _, a := range args {
					cfg.Args = append(cfg.Args, fmt.Sprint(a))
				}
			}
			if envMap, ok := serverMap["env"].(map[string]any); ok {
				cfg.Env = make(map[string]string, len(envMap))
				for k, v := range envMap {
					cfg.Env[k] = fmt.Sprint(v)
				}
			}
		}

		out = append(out, cfg)
	}
	return out, nil
}

func (m *Manager) Add(ctx context.Context, req AddServerRequest) (ServerConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return ServerConfig{}, errors.New("server name is required")
	}

	harness := normalizeHarness(req.Harness)
	if harness == "" {
		harness = "opencode"
	}

	srvType := strings.ToLower(strings.TrimSpace(req.Type))
	if srvType == "" || srvType == "stdio" {
		srvType = "local"
	} else if srvType == "http" || srvType == "sse" {
		srvType = "remote"
	}

	switch harness {
	case "opencode":
		return m.addOpencodeCompatible(m.opencodeConfigPath(), "opencode", name, srvType, req)
	case "opencode-v2":
		return m.addOpencodeCompatible(m.opencodeConfigPath(), "opencode-v2", name, srvType, req)
	case "mage":
		return m.addOpencodeCompatible(m.mageConfigPath(), "mage", name, srvType, req)
	case "agy":
		return m.addAgy(name, srvType, req)
	case "claude-code":
		return m.addMcpServersCompatible(m.claudeConfigPath(), "claude-code", name, srvType, req)
	case "cursor":
		return m.addMcpServersCompatible(m.cursorConfigPath(), "cursor", name, srvType, req)
	default:
		cfgPath := filepath.Join(m.homeDir, ".config", harness, "mcp.json")
		return m.addMcpServersCompatible(cfgPath, harness, name, srvType, req)
	}
}

func (m *Manager) addOpencodeCompatible(cfgPath, harnessName, name, srvType string, req AddServerRequest) (ServerConfig, error) {
	var root map[string]any

	data, err := os.ReadFile(cfgPath)
	if err == nil {
		_ = json.Unmarshal(data, &root)
	}
	if root == nil {
		root = make(map[string]any)
	}

	mcpVal, ok := root["mcp"].(map[string]any)
	if !ok {
		mcpVal = make(map[string]any)
		root["mcp"] = mcpVal
	}

	serverEntry := map[string]any{
		"type":    srvType,
		"enabled": true,
	}

	if srvType == "remote" {
		serverEntry["url"] = req.URL
	} else {
		cmdList := []string{req.Command}
		cmdList = append(cmdList, req.Args...)
		serverEntry["command"] = cmdList
		if len(req.Env) > 0 {
			serverEntry["environment"] = req.Env
		} else {
			serverEntry["environment"] = map[string]string{}
		}
	}

	targetMap := mcpVal
	if nested, ok := mcpVal["servers"].(map[string]any); ok {
		targetMap = nested
	}
	targetMap[name] = serverEntry

	if err := writeJSONPretty(cfgPath, root); err != nil {
		return ServerConfig{}, err
	}

	return ServerConfig{
		Name:       name,
		Harness:    harnessName,
		Type:       srvType,
		Command:    req.Command,
		Args:       req.Args,
		URL:        req.URL,
		Env:        req.Env,
		ConfigPath: cfgPath,
		Enabled:    true,
	}, nil
}

func (m *Manager) addAgy(name, srvType string, req AddServerRequest) (ServerConfig, error) {
	cfgPath := m.agyConfigPath()
	var root map[string]any

	data, err := os.ReadFile(cfgPath)
	if err == nil {
		_ = json.Unmarshal(data, &root)
	}
	if root == nil {
		root = make(map[string]any)
	}

	mcpVal, ok := root["mcpServers"].(map[string]any)
	if !ok {
		mcpVal = make(map[string]any)
		root["mcpServers"] = mcpVal
	}

	serverEntry := make(map[string]any)
	if srvType == "remote" {
		serverEntry["url"] = req.URL
	} else {
		serverEntry["command"] = req.Command
		serverEntry["args"] = req.Args
		if len(req.Env) > 0 {
			serverEntry["env"] = req.Env
		} else {
			serverEntry["env"] = map[string]string{}
		}
	}

	mcpVal[name] = serverEntry

	if err := writeJSONPretty(cfgPath, root); err != nil {
		return ServerConfig{}, err
	}

	return ServerConfig{
		Name:       name,
		Harness:    "agy",
		Type:       srvType,
		Command:    req.Command,
		Args:       req.Args,
		URL:        req.URL,
		Env:        req.Env,
		ConfigPath: cfgPath,
		Enabled:    true,
	}, nil
}

func (m *Manager) addMcpServersCompatible(cfgPath, harnessName, name, srvType string, req AddServerRequest) (ServerConfig, error) {
	var root map[string]any

	data, err := os.ReadFile(cfgPath)
	if err == nil {
		_ = json.Unmarshal(data, &root)
	}
	if root == nil {
		root = make(map[string]any)
	}

	mcpVal, ok := root["mcpServers"].(map[string]any)
	if !ok {
		mcpVal = make(map[string]any)
		root["mcpServers"] = mcpVal
	}

	serverEntry := make(map[string]any)
	if srvType == "remote" {
		serverEntry["url"] = req.URL
	} else {
		serverEntry["command"] = req.Command
		serverEntry["args"] = req.Args
		if len(req.Env) > 0 {
			serverEntry["env"] = req.Env
		} else {
			serverEntry["env"] = map[string]string{}
		}
	}

	mcpVal[name] = serverEntry

	if err := writeJSONPretty(cfgPath, root); err != nil {
		return ServerConfig{}, err
	}

	return ServerConfig{
		Name:       name,
		Harness:    harnessName,
		Type:       srvType,
		Command:    req.Command,
		Args:       req.Args,
		URL:        req.URL,
		Env:        req.Env,
		ConfigPath: cfgPath,
		Enabled:    true,
	}, nil
}

func (m *Manager) Delete(ctx context.Context, harness, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("server name is required")
	}

	harness = normalizeHarness(harness)
	if harness == "" {
		harness = "opencode"
	}

	switch harness {
	case "opencode":
		return deleteFromOpencode(m.opencodeConfigPath(), name)
	case "opencode-v2":
		return deleteFromOpencode(m.opencodeConfigPath(), name)
	case "mage":
		return deleteFromOpencode(m.mageConfigPath(), name)
	case "agy":
		return deleteFromJSONKey(m.agyConfigPath(), "mcpServers", name)
	case "claude-code":
		return deleteFromJSONKey(m.claudeConfigPath(), "mcpServers", name)
	case "cursor":
		return deleteFromJSONKey(m.cursorConfigPath(), "mcpServers", name)
	default:
		cfgPath := filepath.Join(m.homeDir, ".config", harness, "mcp.json")
		return deleteFromJSONKey(cfgPath, "mcpServers", name)
	}
}

func deleteFromOpencode(cfgPath, itemKey string) error {
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return err
	}

	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return err
	}

	subMap, ok := root["mcp"].(map[string]any)
	if !ok {
		return nil
	}

	delete(subMap, itemKey)
	if nested, ok := subMap["servers"].(map[string]any); ok {
		delete(nested, itemKey)
	}

	return writeJSONPretty(cfgPath, root)
}

func deleteFromJSONKey(cfgPath, rootKey, itemKey string) error {
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return err
	}

	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return err
	}

	subMap, ok := root[rootKey].(map[string]any)
	if !ok {
		return nil
	}

	delete(subMap, itemKey)

	return writeJSONPretty(cfgPath, root)
}

func writeJSONPretty(path string, val any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(val, "", "  ")
	if err != nil {
		return err
	}
	bytes = append(bytes, '\n')

	return os.WriteFile(path, bytes, 0644)
}
