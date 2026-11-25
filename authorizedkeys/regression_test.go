package authorizedkeys

import (
	"reflect"
	"testing"
)

func TestParseLineEscapedQuotesInOptions(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		options []string
	}{
		{
			name:    "escaped quote with comma",
			line:    `command="echo \"x,y\"" ssh-rsa AAAA c`,
			options: []string{`command="echo \"x,y\""`},
		},
		{
			name:    "escaped backslash preserved",
			line:    `command="a\\b" ssh-rsa AAAA c`,
			options: []string{`command="a\\b"`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, ok := ParseLine(tt.line)
			if !ok {
				t.Fatalf("ParseLine(%q) not ok", tt.line)
			}
			if !reflect.DeepEqual(entry.Options, tt.options) {
				t.Fatalf("Options = %#v, want %#v", entry.Options, tt.options)
			}
		})
	}
}
