package config

import (
	"strings"
	"unicode"
)

type directive struct {
	keyword string
	args    []string
}

func parseDirectives(content []byte) []directive {
	text := string(content)
	// Best-effort handling of line continuations.
	text = joinContinuedLines(text)

	lines := strings.Split(text, "\n")
	directives := make([]directive, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		line = stripComment(line)
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		keyword, rest, ok := splitKeyword(line)
		if !ok {
			continue
		}
		args := splitArgs(rest)
		directives = append(directives, directive{keyword: strings.ToLower(keyword), args: args})
	}
	return directives
}

func joinContinuedLines(text string) string {
	var out strings.Builder
	out.Grow(len(text))
	inSingle := false
	inDouble := false
	inComment := false
	escaped := false
	for i := 0; i < len(text); i++ {
		c := text[i]
		if c == '\n' || c == '\r' {
			inComment = false
			escaped = false
			out.WriteByte(c)
			continue
		}
		if inComment {
			out.WriteByte(c)
			continue
		}
		if escaped {
			escaped = false
			out.WriteByte(c)
			continue
		}
		if c == '\\' && !inSingle {
			// Line continuation only outside comments and single quotes.
			if i+1 < len(text) {
				next := text[i+1]
				if next == '\n' {
					out.WriteByte(' ')
					i++
					continue
				}
				if next == '\r' && i+2 < len(text) && text[i+2] == '\n' {
					out.WriteByte(' ')
					i += 2
					continue
				}
			} else {
				// Trailing backslash at EOF: tolerated continuation.
				out.WriteByte(' ')
				continue
			}
			escaped = true
			out.WriteByte(c)
			continue
		}
		switch c {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '#':
			if !inSingle && !inDouble {
				inComment = true
			}
		}
		out.WriteByte(c)
	}
	return out.String()
}

func stripComment(line string) string {
	inSingle := false
	inDouble := false
	escaped := false
	for i, r := range line {
		if escaped {
			escaped = false
			continue
		}
		if r == '\\' && !inSingle {
			escaped = true
			continue
		}
		switch r {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '#':
			if !inSingle && !inDouble {
				return line[:i]
			}
		}
	}
	return line
}

func splitKeyword(line string) (keyword, rest string, ok bool) {
	i := 0
	for i < len(line) {
		r := rune(line[i])
		if unicode.IsSpace(r) || line[i] == '=' {
			break
		}
		i++
	}
	if i == 0 {
		return "", "", false
	}
	keyword = line[:i]
	rest = strings.TrimLeftFunc(line[i:], func(r rune) bool {
		return unicode.IsSpace(r) || r == '='
	})
	return keyword, rest, true
}

func splitArgs(s string) []string {
	args := make([]string, 0, 4)
	var cur strings.Builder
	inSingle := false
	inDouble := false
	escaped := false
	tokenStarted := false
	flush := func() {
		if !tokenStarted {
			return
		}
		args = append(args, cur.String())
		cur.Reset()
		tokenStarted = false
	}

	for i := 0; i < len(s); i++ {
		c := s[i]
		if escaped {
			cur.WriteByte(c)
			escaped = false
			tokenStarted = true
			continue
		}
		if c == '\\' && !inSingle {
			escaped = true
			tokenStarted = true
			continue
		}
		switch c {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
				tokenStarted = true
				continue
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
				tokenStarted = true
				continue
			}
		}
		if !inSingle && !inDouble && unicode.IsSpace(rune(c)) {
			flush()
			continue
		}
		cur.WriteByte(c)
		tokenStarted = true
	}
	if escaped {
		cur.WriteByte('\\')
	}
	flush()
	return args
}

func splitCommaList(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}
