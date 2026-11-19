package config

import (
	"strings"
	"testing"
)

func FuzzIndexReader(f *testing.F) {
	seeds := []string{
		"",
		"Host example.com\n  User alice\n",
		"Host *\n  HostName ignored\n",
		"Match host foo,bar originalhost baz exec \"echo hi\"\n  Host qux\n",
		"Match sessiontype shell host sessioned\n  Host session-target\n",
		"Include /nonexistent/*.conf\nInclude ~/missing\nHost included\n",
		"Match=host = alpha originalhost= beta\nHost gamma\n",
		"Host a \\\n  b\n",
		"Match all\nHost all-host\n",
		"# comment\n  \nUnknownKeyword value\n",
		"Host \"name with space\"\nInclude \"path with space.conf\"\n",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data string) {
		// Must never panic and must never execute Match exec/command.
		_, _ = IndexReader(strings.NewReader(data), Source{Path: "fuzz.conf", BaseDir: "."})
	})
}

func FuzzParseDirectives(f *testing.F) {
	seeds := []string{
		"",
		"Host a b c",
		"Match host x exec \"cmd\" user y",
		"Include = /etc/ssh/config.d/*.conf",
		"key=value",
		"\xff\xfe",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data string) {
		_ = parseDirectives([]byte(data))
	})
}
