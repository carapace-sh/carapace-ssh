package authorizedkeys

import "strings"

// Entry represents a parsed authorized_keys entry.
type Entry struct {
	Options []string
	KeyType string
	Key     string
	Comment string
}

// HasKey reports whether the entry contains a key type and key data.
func (e Entry) HasKey() bool {
	return strings.TrimSpace(e.KeyType) != "" && strings.TrimSpace(e.Key) != ""
}
