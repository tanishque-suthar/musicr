package ui

import (
	"strings"
	"testing"
)

func TestFormatTime(t *testing.T) {
	tests := []struct {
		secs float64
		want string
	}{
		{0, "00:00"},
		{61, "01:01"},
		{3661, "61:01"},
		{-5, "00:00"},
		{3599, "59:59"},
	}
	for _, tt := range tests {
		got := formatTime(tt.secs)
		if got != tt.want {
			t.Errorf("formatTime(%v) = %q, want %q", tt.secs, got, tt.want)
		}
	}
}

func TestRenderProgressBarMinimumWidth(t *testing.T) {
	result := renderProgressBar(0, 100, 1)
	if result[0] == '[' {
		t.Error("narrow terminal should return time only, no brackets")
	}
}

func TestRenderProgressBarFull(t *testing.T) {
	result := renderProgressBar(100, 100, 80)
	if !strings.Contains(result, "█") {
		t.Error("full progress bar should have filled blocks")
	}
}

func TestRenderProgressBarOverflow(t *testing.T) {
	result := renderProgressBar(200, 100, 80)
	if !strings.Contains(result, "█") {
		t.Error("overflow bar should be fully filled")
	}
}

func TestRenderProgressBarZeroDuration(t *testing.T) {
	result := renderProgressBar(0, 0, 80)
	if !strings.Contains(result, "[") {
		t.Error("zero duration should still produce a bar")
	}
}

func TestRenderProgressBarNegativePos(t *testing.T) {
	result := renderProgressBar(-10, 100, 80)
	if !strings.Contains(result, "00:00") {
		t.Error("negative pos should show 00:00")
	}
}
