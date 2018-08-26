package handlers

import (
	"testing"
)

func TestValidTimeFormat(t *testing.T) {
	values := []struct {
		time     string
		expected bool
	}{
		{"11:59 PM", true},
		{"11:59 AM", true},
		{"12:00 AM", true},
		{"12:00 PM", true},
		{"01:00 AM", true},
		{"09:59 PM", true},
		{"12:00 am", true},
		{"12:00 pm", true},
		{"1:00 AM", false},
		{"00:00 PM", false},
		{"12:60 AM", false},
		{"11:59am", false},
		{"13:00 PM", false},
		{"1100 AM", false},
		{"-1:00 AM", false},
		{"12:-11 PM", false},
		{"12:00 ZM", false},
		{"01:00 PZ", false},
		{"11:59 MP", false},
		{"11:11 1M", false},
		{"04:00 A5", false},
		{"12345678", false},
		{"FA:ES TR", false},
		{"", false},
		{"        ", false},
	}

	for _, tt := range values {
		if result := validTimeFormat(tt.time); result != tt.expected {
			t.Errorf("validTimeFormat(%s) = got <%t> want <%t>", tt.time, result, tt.expected)
		}
	}
}

func TestCalculateTime(t *testing.T) {
	values := []struct {
		time     string
		change   int
		expected string
	}{
		{"12:00 AM", 0, "12:00 AM"},
		{"11:59 PM", 1, "12:00 AM"},
		{"12:00 AM", 1439, "11:59 PM"},
		{"12:00 AM", 720, "12:00 PM"},
		{"12:00 AM", 1440, "12:00 AM"},
		{"12:00 AM", 14400, "12:00 AM"},
		{"12:00 AM", -1, "11:59 PM"},
		{"11:59 PM", -1439, "12:00 AM"},
		{"12:00 AM", -720, "12:00 PM"},
		{"12:00 PM", -1441, "11:59 AM"},
	}

	for _, tt := range values {
		if result := calculateTime(tt.time, tt.change); result != tt.expected {
			t.Errorf("calculateTime(%s, %d) = got <%s> want <%s>", tt.time, tt.change, result, tt.expected)
		}
	}
}

func TestTimeToMinutes(t *testing.T) {
	values := []struct {
		time     string
		expected int
	}{
		{"12:00 AM", 0},
		{"12:59 AM", 59},
		{"01:00 AM", 60},
		{"11:59 AM", 719},
		{"12:00 PM", 720},
		{"12:59 PM", 779},
		{"01:00 PM", 780},
		{"11:59 PM", 1439},
	}

	for _, tt := range values {
		if result := timeToMinutes(tt.time); result != tt.expected {
			t.Errorf("timeToMinutes(%s) = got <%d> want <%d>", tt.time, result, tt.expected)
		}
	}
}

func TestMinutesToTime(t *testing.T) {
	values := []struct {
		minutes  int
		expected string
	}{
		{0, "12:00 AM"},
		{59, "12:59 AM"},
		{60, "01:00 AM"},
		{719, "11:59 AM"},
		{720, "12:00 PM"},
		{779, "12:59 PM"},
		{780, "01:00 PM"},
		{1439, "11:59 PM"},
	}

	for _, tt := range values {
		if result := minutesToTime(tt.minutes); result != tt.expected {
			t.Errorf("minutesToTime(%d) = got <%s> want <%s>", tt.minutes, result, tt.expected)
		}
	}
}
