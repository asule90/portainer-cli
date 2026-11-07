package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/portainer/portainer-cli/pkg/stacks"
)

// parseEnvPairs converts KEY=VAL CLI flags and optional env file into []stacks.Pair.
func parseEnvPairs(envFlags []string, envFile string) ([]stacks.Pair, error) {
	pairsMap := make(map[string]string)
	var order []string

	if envFile != "" {
		file, err := os.Open(envFile)
		if err != nil {
			return nil, fmt.Errorf("failed opening env file '%s': %w", envFile, err)
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
				v := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
				if _, exists := pairsMap[k]; !exists {
					order = append(order, k)
				}
				pairsMap[k] = v
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("error reading env file: %w", err)
		}
	}

	for _, flag := range envFlags {
		parts := strings.SplitN(flag, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid env format '%s': expected KEY=VALUE", flag)
		}
		k := strings.TrimSpace(parts[0])
		v := parts[1]
		if _, exists := pairsMap[k]; !exists {
			order = append(order, k)
		}
		pairsMap[k] = v
	}

	res := make([]stacks.Pair, 0, len(order))
	for _, k := range order {
		res = append(res, stacks.Pair{
			Name:  k,
			Value: pairsMap[k],
		})
	}
	return res, nil
}
