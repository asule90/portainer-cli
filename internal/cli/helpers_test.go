package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseEnvPairs(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, "test.env")
	envContent := `
# Comment line
FOO=bar
BAZ="quoted val"
`
	if err := os.WriteFile(envPath, []byte(envContent), 0600); err != nil {
		t.Fatalf("failed to write test env file: %v", err)
	}

	pairs, err := parseEnvPairs([]string{"OVERRIDE=yes", "FOO=updated"}, envPath)
	if err != nil {
		t.Fatalf("parseEnvPairs error: %v", err)
	}

	pairMap := make(map[string]string)
	for _, p := range pairs {
		pairMap[p.Name] = p.Value
	}

	if pairMap["FOO"] != "updated" {
		t.Errorf("expected FOO=updated, got '%s'", pairMap["FOO"])
	}
	if pairMap["BAZ"] != "quoted val" {
		t.Errorf("expected BAZ='quoted val', got '%s'", pairMap["BAZ"])
	}
	if pairMap["OVERRIDE"] != "yes" {
		t.Errorf("expected OVERRIDE=yes, got '%s'", pairMap["OVERRIDE"])
	}
}
