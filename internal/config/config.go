package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config holds the application configuration parameters.
type Config struct {
	URL        string `json:"url"`
	APIKey     string `json:"api_key"`
	JWTToken   string `json:"jwt_token,omitempty"`
	Insecure   bool   `json:"insecure"`
	EndpointID int    `json:"endpoint_id,omitempty"`
	Output     string `json:"output"`
}

// FileConfig represents the JSON structure saved in ~/.config/portainer-cli/config.json.
type FileConfig struct {
	URL        string `json:"url,omitempty"`
	APIKey     string `json:"api_key,omitempty"`
	JWTToken   string `json:"jwt_token,omitempty"`
	Insecure   *bool  `json:"insecure,omitempty"`
	EndpointID int    `json:"endpoint_id,omitempty"`
	Output     string `json:"output,omitempty"`
}

// ConfigFilePath returns the absolute path to ~/.config/portainer-cli/config.json.
func ConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "portainer-cli", "config.json"), nil
}

// LoadFileConfig loads the persistent JSON config file if it exists.
func LoadFileConfig() (*FileConfig, error) {
	path, err := ConfigFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &FileConfig{}, nil
		}
		return nil, err
	}

	var fc FileConfig
	if err := json.Unmarshal(data, &fc); err != nil {
		return nil, fmt.Errorf("invalid json in %s: %w", path, err)
	}
	return &fc, nil
}

// SaveFileConfig persists the given FileConfig to ~/.config/portainer-cli/config.json.
func SaveFileConfig(fc *FileConfig) error {
	path, err := ConfigFilePath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed creating config dir %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(fc, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, append(data, '\n'), 0600)
}

// UpdateFileKey updates a single key in the config file.
func UpdateFileKey(key, value string) error {
	fc, err := LoadFileConfig()
	if err != nil {
		return err
	}

	switch strings.ToLower(strings.TrimSpace(key)) {
	case "url":
		fc.URL = strings.TrimSpace(value)
	case "api-key", "api_key", "key":
		fc.APIKey = strings.TrimSpace(value)
	case "jwt", "jwt_token":
		fc.JWTToken = strings.TrimSpace(value)
	case "insecure":
		val := strings.ToLower(strings.TrimSpace(value))
		b := val == "true" || val == "1" || val == "yes"
		fc.Insecure = &b
	case "endpoint", "endpoint-id", "endpoint_id":
		id, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return fmt.Errorf("invalid endpoint-id: %w", err)
		}
		fc.EndpointID = id
	case "output":
		fc.Output = strings.TrimSpace(value)
	default:
		return fmt.Errorf("unknown config key '%s' (valid keys: url, api-key, jwt, insecure, endpoint-id, output)", key)
	}

	return SaveFileConfig(fc)
}

// DeleteFileConfig removes ~/.config/portainer-cli/config.json.
func DeleteFileConfig() error {
	path, err := ConfigFilePath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Load reads configuration by cascading:
// 1. Defaults
// 2. Persistent file (~/.config/portainer-cli/config.json)
// 3. .env file (if found in pwd or parents)
// 4. Environment variables (PORTAINER_*, X_API_KEY)
func Load() *Config {
	// 1. Base defaults
	cfg := &Config{
		URL:      "https://localhost:9443",
		Insecure: true,
		Output:   "table",
	}

	// 2. Persistent file (~/.config/portainer-cli/config.json)
	if fc, err := LoadFileConfig(); err == nil && fc != nil {
		if fc.URL != "" {
			cfg.URL = fc.URL
		}
		if fc.APIKey != "" {
			cfg.APIKey = fc.APIKey
		}
		if fc.JWTToken != "" {
			cfg.JWTToken = fc.JWTToken
		}
		if fc.Insecure != nil {
			cfg.Insecure = *fc.Insecure
		}
		if fc.EndpointID > 0 {
			cfg.EndpointID = fc.EndpointID
		}
		if fc.Output != "" {
			cfg.Output = fc.Output
		}
	}

	// 3. Local or directory .env file
	loadDotEnv()

	// 4. Environment variables
	cfg.URL = getEnv("PORTAINER_URL", cfg.URL)
	cfg.APIKey = getEnv("PORTAINER_API_KEY", getEnv("X_API_KEY", cfg.APIKey))
	cfg.JWTToken = getEnv("PORTAINER_JWT", cfg.JWTToken)
	cfg.Insecure = getBoolEnv("PORTAINER_INSECURE", cfg.Insecure)
	cfg.EndpointID = getIntEnv("PORTAINER_ENDPOINT_ID", cfg.EndpointID)
	cfg.Output = getEnv("PORTAINER_OUTPUT", cfg.Output)

	return cfg
}

func getEnv(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && strings.TrimSpace(val) != "" {
		return strings.TrimSpace(val)
	}
	return defaultVal
}

func getBoolEnv(key string, defaultVal bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		val = strings.ToLower(strings.TrimSpace(val))
		return val == "true" || val == "1" || val == "yes"
	}
	return defaultVal
}

func getIntEnv(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
			return i
		}
	}
	return defaultVal
}

// loadDotEnv searches for .env files up the directory tree or in child portainer-cli.
func loadDotEnv() {
	curr, err := os.Getwd()
	if err != nil {
		return
	}

	searchDirs := []string{curr, filepath.Join(curr, "portainer-cli")}
	scanCurr := curr
	for i := 0; i < 3; i++ {
		parent := filepath.Dir(scanCurr)
		if parent == scanCurr {
			break
		}
		searchDirs = append(searchDirs, parent)
		scanCurr = parent
	}

	for _, dir := range searchDirs {
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			parseEnvFile(envPath)
			return
		}
	}
}

func parseEnvFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			v = strings.Trim(v, `"'`)
			if _, exists := os.LookupEnv(k); !exists {
				_ = os.Setenv(k, v)
			}
		}
	}
}
