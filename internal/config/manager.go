package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"
)

// Manager handles configuration and context management
type Manager struct {
	configDir   string
	contextsDir string
}

// NewManager creates a new configuration manager
func NewManager() (*Manager, error) {
	// Create XDG-compliant config directory
	configDir := filepath.Join(xdg.ConfigHome, "flint")
	contextsDir := filepath.Join(configDir, "contexts")

	// Ensure directories exist
	if err := os.MkdirAll(contextsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directories: %w", err)
	}

	return &Manager{
		configDir:   configDir,
		contextsDir: contextsDir,
	}, nil
}

// GetConfigDir returns the main configuration directory path
func (m *Manager) GetConfigDir() string {
	return m.configDir
}

// GetContextsDir returns the contexts directory path
func (m *Manager) GetContextsDir() string {
	return m.contextsDir
}

// GetGlobalConfigPath returns the path to the global config file
func (m *Manager) GetGlobalConfigPath() string {
	return filepath.Join(m.configDir, "config.yaml")
}

// GetContextPath returns the path to a specific context file
func (m *Manager) GetContextPath(name string) string {
	return filepath.Join(m.contextsDir, fmt.Sprintf("%s.yaml", name))
}

// LoadGlobalConfig loads the global configuration
func (m *Manager) LoadGlobalConfig() (*GlobalConfig, error) {
	configPath := m.GetGlobalConfigPath()
	
	// Create default config if file doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := &GlobalConfig{
			ActiveContext:      "",
			OutputFormat:       "json",
			ColorsEnabled:      true,
			PaginationSize:     30,
			OrganizationDisplay: true,
		}
		
		if err := m.SaveGlobalConfig(defaultConfig); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		
		return defaultConfig, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config GlobalConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveGlobalConfig saves the global configuration
func (m *Manager) SaveGlobalConfig(config *GlobalConfig) error {
	configPath := m.GetGlobalConfigPath()
	
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadContext loads a specific context configuration
func (m *Manager) LoadContext(name string) (*Context, error) {
	if name == "" {
		return nil, fmt.Errorf("context name cannot be empty")
	}

	contextPath := m.GetContextPath(name)
	
	if _, err := os.Stat(contextPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("context '%s' not found", name)
	}

	data, err := os.ReadFile(contextPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read context file: %w", err)
	}

	var context Context
	if err := yaml.Unmarshal(data, &context); err != nil {
		return nil, fmt.Errorf("failed to parse context file: %w", err)
	}

	return &context, nil
}

// SaveContext saves a context configuration
func (m *Manager) SaveContext(context *Context) error {
	if context.Name == "" {
		return fmt.Errorf("context name cannot be empty")
	}

	contextPath := m.GetContextPath(context.Name)
	
	data, err := yaml.Marshal(context)
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}

	if err := os.WriteFile(contextPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write context file: %w", err)
	}

	return nil
}

// ListContexts returns all available context names
func (m *Manager) ListContexts() ([]string, error) {
	files, err := os.ReadDir(m.contextsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read contexts directory: %w", err)
	}

	var contexts []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".yaml" {
			name := file.Name()[:len(file.Name())-5] // Remove .yaml extension
			contexts = append(contexts, name)
		}
	}

	return contexts, nil
}

// DeleteContext removes a context configuration
func (m *Manager) DeleteContext(name string) error {
	if name == "" {
		return fmt.Errorf("context name cannot be empty")
	}

	contextPath := m.GetContextPath(name)
	
	if _, err := os.Stat(contextPath); os.IsNotExist(err) {
		return fmt.Errorf("context '%s' not found", name)
	}

	if err := os.Remove(contextPath); err != nil {
		return fmt.Errorf("failed to delete context file: %w", err)
	}

	return nil
}

// GetActiveContext returns the currently active context
func (m *Manager) GetActiveContext() (*Context, error) {
	globalConfig, err := m.LoadGlobalConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load global config: %w", err)
	}

	if globalConfig.ActiveContext == "" {
		return nil, fmt.Errorf("no active context set")
	}

	return m.LoadContext(globalConfig.ActiveContext)
}

// SetActiveContext sets the active context
func (m *Manager) SetActiveContext(name string) error {
	// Verify context exists
	if _, err := m.LoadContext(name); err != nil {
		return err
	}

	globalConfig, err := m.LoadGlobalConfig()
	if err != nil {
		return fmt.Errorf("failed to load global config: %w", err)
	}

	globalConfig.ActiveContext = name

	return m.SaveGlobalConfig(globalConfig)
}
