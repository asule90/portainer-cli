package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

// PrintFormatted prints data based on the chosen output format (table, json, yaml).
func PrintFormatted(w io.Writer, format string, data any, renderTable func() error) error {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	case "yaml", "yml":
		enc := yaml.NewEncoder(w)
		return enc.Encode(data)
	case "table", "":
		if renderTable != nil {
			return renderTable()
		}
		// Fallback to JSON if table renderer is not provided
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	default:
		return fmt.Errorf("unsupported output format '%s' (use table, json, or yaml)", format)
	}
}
