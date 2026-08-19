package main

import (
	"testing"
	"time"
)

func TestParseInactiveHours(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    time.Duration
		wantErr bool
	}{
		{name: "未指定は既定値48時間", value: "", want: 48 * time.Hour},
		{name: "正の整数", value: "24", want: 24 * time.Hour},
		{name: "0は不正", value: "0", wantErr: true},
		{name: "負数は不正", value: "-1", wantErr: true},
		{name: "数値以外は不正", value: "abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseInactiveHours(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseInactiveHours(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseInactiveHours(%q) = %s, want %s", tt.value, got, tt.want)
			}
		})
	}
}

func TestParseBool(t *testing.T) {
	for value, want := range map[string]bool{"true": true, "1": true, "false": false, "": false, "yes": false} {
		if got := parseBool(value); got != want {
			t.Errorf("parseBool(%q) = %v, want %v", value, got, want)
		}
	}
}
