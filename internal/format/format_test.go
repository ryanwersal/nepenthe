package format

import "testing"

func TestBytes(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0 B"},
		{1, "1 B"},
		{1023, "1023 B"},
		{1024, "1 KB"},
		{1536, "1 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024*1024 + 512*1024, "1.5 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{int64(1.5 * 1024 * 1024 * 1024), "1.5 GB"},
	}
	for _, tt := range tests {
		got := Bytes(tt.input)
		if got != tt.want {
			t.Errorf("Bytes(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCount(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{999, "999"},
		{1000, "1.0K"},
		{1500, "1.5K"},
		{999999, "1000.0K"},
		{1000000, "1.0M"},
		{2500000, "2.5M"},
	}
	for _, tt := range tests {
		got := Count(tt.input)
		if got != tt.want {
			t.Errorf("Count(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestInterval(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0s"},
		{30, "30s"},
		{59, "59s"},
		{60, "1m"},
		{120, "2m"},
		{3599, "59m"},
		{3600, "1h"},
		{7200, "2h"},
		{86399, "23h"},
		{86400, "1d"},
		{172800, "2d"},
	}
	for _, tt := range tests {
		got := Interval(tt.input)
		if got != tt.want {
			t.Errorf("Interval(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
