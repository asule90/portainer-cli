package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config holds the application configuration parameters.
type Config struct {
	URL        string
	APIKey     string
	JWTToken   string
	Insecure   bool
	EndpointID int
	Output     string
}

// Load reads configuration from flags, environment variables, and any local .env file.
func Load() *Config {
	// 1. Try reading .env file if present
	loadDotEnv()

	cfg := &Config{
		URL:        getEnv("PORTAINER_URL", "https://localhost:9443"),
		APIKey:     getEnv("PORTAINER_API_KEY", getEnv("X_API_KEY", "")),
		JWTToken:   getEnv("PORTAINER_JWT", ""),
		Insecure:   getBoolEnv("PORTAINER_INSECURE", true),
		EndpointID: getIntEnv("PORTAINER_ENDPOINT_ID", 0),
		Output:     getEnv("PORTAINER_OUTPUT", "table"),
	}

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

// loadDotEnv searches for .env files up the directory tree, in child portainer-cli, or home config.
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

	if home, err := os.UserHomeDir(); err == nil {
		searchDirs = append(searchDirs,
			filepath.Join(home, ".config", "portainer-cli"),
			filepath.Join(home, ".portainer-cli"),
		)
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
