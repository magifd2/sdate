package main

import (
	"testing"
	"time"
)

func TestParseInput(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantSnap  string
		wantRel   string
		wantErr   bool
	}{
		{"Empty input", "", "", "", false},
		{"Snap only", "@d", "@d", "", false},
		{"Relative only", "-1h", "", "-1h", false},
		{"Relative and snap", "-1d@d", "@d", "-1d", false},
		{"Snap and relative", "@h+2h", "@h", "+2h", false},
		{"Invalid input", "invalid", "", "", true},
		{"Invalid quantity", "-xh@d", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := parseInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if spec.Snap != tt.wantSnap {
					t.Errorf("parseInput() snap = %v, want %v", spec.Snap, tt.wantSnap)
				}
				if spec.Relative != tt.wantRel {
					t.Errorf("parseInput() relative = %v, want %v", spec.Relative, tt.wantRel)
				}
			}
		})
	}
}

func TestApplyOperation(t *testing.T) {
	baseTime := time.Date(2023, 10, 27, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name    string
		spec    *SplunkLikeTimeSpec
		want    time.Time
		wantErr bool
	}{
		{
			name: "Snap to day",
			spec: &SplunkLikeTimeSpec{Snap: "@d"},
			want: time.Date(2023, 10, 27, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Add 2 hours",
			spec: &SplunkLikeTimeSpec{Relative: "+2h"},
			want: time.Date(2023, 10, 27, 12, 30, 0, 0, time.UTC),
		},
		{
			name: "Subtract 1 day and snap to day",
			spec: &SplunkLikeTimeSpec{Relative: "-1d", Snap: "@d"},
			want: time.Date(2023, 10, 26, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Unknown snap unit",
			spec: &SplunkLikeTimeSpec{Snap: "@x"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applyOperation(baseTime, tt.spec)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyOperation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !got.Equal(tt.want) {
				t.Errorf("applyOperation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertFormat(t *testing.T) {
	tests := []struct {
		name       string
		userFormat string
		want       string
	}{
		{"YYYY/MM/DD hh:mm:ss", "YYYY/MM/DD hh:mm:ss", "2006/01/02 15:04:05"},
		{"With milliseconds", "YYYY-MM-DD hh:mm:ss.SSS", "2006-01-02 15:04:05.000"},
		{"With timezone", "YYYY-MM-DD hh:mm:ss TZ", "2006-01-02 15:04:05 MST"},
		{"With timezone offset", "YYYY-MM-DD hh:mm:ss ZZ", "2006-01-02 15:04:05 -07:00"},
		{"With timezone offset no colon", "YYYY-MM-DD hh:mm:ss ZZZ", "2006-01-02 15:04:05 -0700"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertFormat(tt.userFormat); got != tt.want {
				t.Errorf("convertFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}
