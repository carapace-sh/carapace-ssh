package knownhosts

import (
	"path/filepath"
	"strings"
)

func expandTilde(path string, home string) string {
	if path == "" || path[0] != '~' {
		return path
	}
	if home == "" {
		return path
	}
	if path == "~" {
		return home
	}
	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\") {
		return filepath.Join(home, path[2:])
	}
	// ~user/... forms are not expanded without user lookup; return as-is.
	return path
}
