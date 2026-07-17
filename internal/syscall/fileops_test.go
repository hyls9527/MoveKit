package winsyscall

import (
	"testing"
)

func TestToLongPath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`C:\Users\test`, `\\?\C:\Users\test`},
		{`\\?\C:\Users\test`, `\\?\C:\Users\test`}, // already prefixed
		{`\\server\share`, `\\?\UNC\server\share`},
		{`D:\foo\bar.txt`, `\\?\D:\foo\bar.txt`},
	}
	for _, tt := range tests {
		got := ToLongPath(tt.input)
		if got != tt.want {
			t.Errorf("ToLongPath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestIsCrossVolume(t *testing.T) {
	tests := []struct {
		src, tgt string
		want     bool
	}{
		{`C:\src`, `C:\dst`, false},
		{`C:\src`, `D:\dst`, true},
		{`C:\src`, `c:\dst`, false}, // case-insensitive
		{`D:\src`, `E:\dst`, true},
		{`relative`, `C:\dst`, false}, // no volume name
	}
	for _, tt := range tests {
		got := IsCrossVolume(tt.src, tt.tgt)
		if got != tt.want {
			t.Errorf("IsCrossVolume(%q, %q) = %v, want %v", tt.src, tt.tgt, got, tt.want)
		}
	}
}
