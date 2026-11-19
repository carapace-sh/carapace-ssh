package authorizedkeys

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Index contains parsed authorized_keys entries.
type Index struct {
	entries []Entry
}

// Entries returns a copy of parsed entries.
func (i *Index) Entries() []Entry {
	if i == nil || len(i.entries) == 0 {
		return []Entry{}
	}
	out := make([]Entry, len(i.entries))
	copy(out, i.entries)
	return out
}

// Comments returns non-empty comments from parsed entries in file order.
func (i *Index) Comments() []string {
	if i == nil {
		return []string{}
	}
	comments := make([]string, 0, len(i.entries))
	for _, e := range i.entries {
		if c := strings.TrimSpace(e.Comment); c != "" {
			comments = append(comments, c)
		}
	}
	return comments
}

// DefaultPath returns the default authorized_keys path for the given home
// directory.
func DefaultPath(home string) string {
	if strings.TrimSpace(home) == "" {
		return ""
	}
	return filepath.Join(home, ".ssh", "authorized_keys")
}

// IndexDefault parses the default authorized_keys file for the given home
// directory. Missing or unreadable files are treated as non-fatal.
func IndexDefault(home string) (*Index, error) {
	return IndexFile(DefaultPath(home))
}

// IndexFile parses an authorized_keys file. Missing or unreadable files are
// treated as non-fatal and return an empty index.
func IndexFile(path string) (*Index, error) {
	if strings.TrimSpace(path) == "" {
		return &Index{}, nil
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) || os.IsPermission(err) {
			return &Index{}, nil
		}
		return &Index{}, err
	}
	defer f.Close()
	return IndexReader(f)
}

// IndexReader parses authorized_keys entries from a reader.
func IndexReader(r io.Reader) (*Index, error) {
	idx := &Index{}
	if r == nil {
		return idx, nil
	}
	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 64*1024), 1024*1024)
	for s.Scan() {
		if e, ok := ParseLine(s.Text()); ok {
			idx.entries = append(idx.entries, e)
		}
	}
	if err := s.Err(); err != nil {
		return idx, err
	}
	return idx, nil
}

// ParseLine parses a single authorized_keys line. It returns ok=false for empty
// lines, comments, or lines that cannot be parsed into a key entry.
func ParseLine(line string) (Entry, bool) {
	line = strings.TrimSpace(stripCR(line))
	if line == "" || strings.HasPrefix(line, "#") {
		return Entry{}, false
	}

	fields := splitFields(line)
	if len(fields) < 2 {
		return Entry{}, false
	}

	keyIdx := findKeyTypeIndex(fields)
	if keyIdx < 0 || keyIdx+1 >= len(fields) {
		return Entry{}, false
	}

	entry := Entry{
		KeyType: fields[keyIdx],
		Key:     fields[keyIdx+1],
	}
	if keyIdx > 0 {
		entry.Options = parseOptions(strings.Join(fields[:keyIdx], " "))
	}
	if keyIdx+2 < len(fields) {
		entry.Comment = strings.Join(fields[keyIdx+2:], " ")
	}
	if !entry.HasKey() {
		return Entry{}, false
	}
	return entry, true
}

func stripCR(s string) string {
	return strings.TrimSuffix(s, "\r")
}

func splitFields(s string) []string {
	fields := make([]string, 0, 4)
	var cur strings.Builder
	inDouble := false
	escaped := false

	flush := func() {
		if cur.Len() == 0 {
			return
		}
		fields = append(fields, cur.String())
		cur.Reset()
	}

	for i := 0; i < len(s); i++ {
		c := s[i]
		if escaped {
			cur.WriteByte(c)
			escaped = false
			continue
		}
		if c == '\\' && inDouble {
			escaped = true
			continue
		}
		if c == '"' {
			inDouble = !inDouble
			cur.WriteByte(c)
			continue
		}
		if !inDouble && (c == ' ' || c == '\t') {
			flush()
			continue
		}
		cur.WriteByte(c)
	}
	if escaped {
		cur.WriteByte('\\')
	}
	flush()
	return fields
}

func findKeyTypeIndex(fields []string) int {
	for i, f := range fields {
		if looksLikeKeyType(f) {
			// key type followed by key data
			if i+1 < len(fields) {
				return i
			}
		}
	}
	return -1
}

func looksLikeKeyType(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	switch {
	case strings.HasPrefix(s, "ssh-"),
		strings.HasPrefix(s, "ecdsa-sha2-"),
		strings.HasPrefix(s, "sk-ssh-"),
		strings.HasPrefix(s, "sk-ecdsa-"),
		strings.HasPrefix(s, "webauthn-sk-ecdsa-"):
		return true
	}
	if strings.Contains(s, "@openssh.com") && !strings.ContainsAny(s, " \t=,") {
		return true
	}
	return false
}

func parseOptions(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	opts := make([]string, 0, 2)
	var cur strings.Builder
	inDouble := false
	escaped := false

	flush := func() {
		opt := strings.TrimSpace(cur.String())
		cur.Reset()
		if opt == "" {
			return
		}
		opts = append(opts, opt)
	}

	for i := 0; i < len(s); i++ {
		c := s[i]
		if escaped {
			cur.WriteByte(c)
			escaped = false
			continue
		}
		if c == '\\' && inDouble {
			escaped = true
			continue
		}
		if c == '"' {
			inDouble = !inDouble
			cur.WriteByte(c)
			continue
		}
		if c == ',' && !inDouble {
			flush()
			continue
		}
		cur.WriteByte(c)
	}
	if escaped {
		cur.WriteByte('\\')
	}
	flush()
	return opts
}
