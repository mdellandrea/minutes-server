package handlers

import (
	"fmt"
	"regexp"
	"strconv"
)

var timeFormatValidator = regexp.MustCompile(`(\d{2}):(\d{2})\s([AaPp][Mm])`)

func validTimeFormat(timeStr string) bool {
	if !timeFormatValidator.MatchString(timeStr) {
		return false
	}
	matches := timeFormatValidator.FindStringSubmatch(timeStr)
	h, _ := strconv.Atoi(matches[1]) // Hours
	m, _ := strconv.Atoi(matches[2]) // Minutes
	if h > 12 || h < 1 {
		return false
	}
	if m > 59 || m < 0 {
		return false
	}

	return true
}

func calculateTime(timeStr string, change int) string {
	start := timeToMinutes(timeStr)
	ch := change % 1440
	diff := (start + ch) % 1440

	if diff < 0 {
		diff = diff + 1440
	}

	return minutesToTime(diff)
}

func timeToMinutes(timeStr string) int {
	matches := timeFormatValidator.FindStringSubmatch(timeStr)
	h, _ := strconv.Atoi(matches[1]) // Hours
	m, _ := strconv.Atoi(matches[2]) // Minutes

	if h == 12 {
		h = 0
	}

	var mm int
	switch v := matches[3]; v {
	case "AM":
		mm = 0
	case "PM":
		mm = 12
	}

	return ((h + mm) * 60) + m
}

func minutesToTime(minutes int) string {
	var x = minutes
	var mm = "AM"

	if minutes >= 720 {
		mm = "PM"
		x = minutes - 720
	}

	h := x / 60
	if h == 0 {
		h = 12
	}
	m := x % 60

	return fmt.Sprintf("%02d:%02d %s", h, m, mm)
}
