package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

const (
	appName = "bb"

	KeyDefaultWorkspace = "default_workspace"
	KeyAPIBaseURL       = "api_base_url"
	KeyGitProtocol      = "git_protocol"
)

// Config wraps Viper for bb-specific configuration.
type Config struct {
	v *viper.Viper
}

// Load reads the config file and binds environment variables.
// Safe to call even if no config file exists — defaults are used.
func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(ConfigDir())
	v.AddConfigPath(".") // per-repo .bb.yaml override

	// Environment variable bindings
	v.SetEnvPrefix("BB")
	v.AutomaticEnv()
	v.BindEnv(KeyDefaultWorkspace, "BITBUCKET_WORKSPACE")

	// Defaults
	v.SetDefault(KeyAPIBaseURL, "https://api.bitbucket.org/2.0")
	v.SetDefault(KeyGitProtocol, "https")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config: %w", err)
		}
		// No config file is fine — continue with defaults and env vars
	}

	return &Config{v: v}, nil
}

// Get returns a config value by key.
func (c *Config) Get(key string) string {
	return c.v.GetString(key)
}

// Set writes a config value and persists it to disk.
func (c *Config) Set(key, value string) error {
	c.v.Set(key, value)
	return c.writeConfig()
}

// DefaultWorkspace returns the configured default workspace slug.
// Resolution order: BITBUCKET_WORKSPACE env → config file → empty string.
func (c *Config) DefaultWorkspace() string {
	return c.v.GetString(KeyDefaultWorkspace)
}

// APIBaseURL returns the Bitbucket API base URL (allows overriding for testing).
func (c *Config) APIBaseURL() string {
	return c.v.GetString(KeyAPIBaseURL)
}

func (c *Config) writeConfig() error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	path := filepath.Join(dir, "config.yaml")
	return c.v.WriteConfigAs(path)
}

// ConfigDir returns the OS-appropriate configuration directory for bb.
//   - Unix/macOS: ~/.config/bb/
//   - Windows:    %APPDATA%\bb\
func ConfigDir() string {
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, appName)
		}
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, appName)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", appName)
}

// TokensFile returns the path to the stored OAuth tokens file.
func TokensFile() string {
	return filepath.Join(ConfigDir(), "tokens.json")
}
