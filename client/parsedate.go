package client

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var durations = map[string]time.Duration{
	"second": time.Second,
	"minute": time.Minute,
	"hour":   time.Hour,
}

var months = [...]string{
	"january",
	"february",
	"march",
	"april",
	"may",
	"june",
	"july",
	"august",
	"september",
	"october",
	"december",
}

var zonedTimeFormats = [...]string{
	time.RFC3339,
}

var localTimeFormats = [...]string{
	"2006-01-02T15:04:05",
	"2006-01-02 15:04",
	"02.01.2006 15:04",
	"02.01 15:04",
}

func ParseTimeRef(input string, refTime time.Time) (time.Time, error) {
	for _, f := range zonedTimeFormats {
		result, err := time.Parse(f, input)
		if err == nil {
			return fixResult(result, refTime), nil
		}
	}

	for _, f := range localTimeFormats {
		result, err := time.ParseInLocation(f, input, refTime.Location())
		if err == nil {
			return fixResult(result, refTime), nil
		}
	}

	input = strings.ToLower(input)

	outputTime := refTime
	var matchLen int
	for len(input) > 0 {
		input = strings.TrimLeft(input, " ")
		outputTime, matchLen = internalParseTimeRef(input, outputTime)
		if matchLen == 0 {
			return time.Time{}, errors.New("Could not parse date/time: " + input)
		}
		input = input[matchLen:]
	}

	return outputTime, nil
}

func fixResult(result time.Time, refTime time.Time) time.Time {
	if result.Year() == 0 {
		result = time.Date(refTime.Year(), result.Month(), result.Day(), result.Hour(), result.Minute(), result.Second(), result.Nanosecond(), result.Location())
	}
	return result
}

func internalParseTimeRef(input string, refTime time.Time) (time.Time, int) {

	if input == "now" {
		return refTime, len(input)
	}

	words := strings.SplitN(input, " ", 2)
	if len(words) > 0 {
		firstWord := words[0]
		duration, err := time.ParseDuration(firstWord)
		if err == nil {
			return refTime.Add(duration), len(firstWord)
		}

		if firstWord == "yesterday" {
			return refTime.Add(time.Duration(-24) * time.Hour), len(firstWord)
		}

		if firstWord == "tomorrow" {
			return refTime.Add(time.Duration(24) * time.Hour), len(firstWord)
		}

		if firstWord == "at" {
			return refTime, len(firstWord)
		}

		if len(firstWord) >= 3 {
			matchCount := 0
			matchIdx := -1
			for idx, m := range months {
				if strings.HasPrefix(m, firstWord) {
					matchCount++
					matchIdx = idx
				}
			}
			if matchCount == 1 {
				return time.Date(refTime.Year(), time.Month(matchIdx+1), refTime.Day(), refTime.Hour(), refTime.Minute(), refTime.Second(), refTime.Nanosecond(), refTime.Location()), len(firstWord)
			}
		}
	}

	exp := regexp.MustCompile("^([0-9]+):([0-9]+)")
	matches := exp.FindStringSubmatch(input)
	if matches != nil {
		hours, _ := strconv.Atoi(matches[1])
		minutes, _ := strconv.Atoi(matches[2])

		if hours < 0 || hours > 23 {
			return refTime, 0
		}
		if minutes < 0 || minutes > 60 {
			return refTime, 0
		}
		return time.Date(refTime.Year(), refTime.Month(), refTime.Day(), hours, minutes, 0, 0, refTime.Location()), len(matches[0])
	}

	exp = regexp.MustCompile("^([0-9]+)((\\.)|(th)|(nd))")
	matches = exp.FindStringSubmatch(input)
	if matches != nil {
		day, _ := strconv.Atoi(matches[1])
		if day < 1 || day > 31 {
			return refTime, 0
		}
		return time.Date(refTime.Year(), refTime.Month(), day, refTime.Hour(), refTime.Minute(), refTime.Second(), refTime.Nanosecond(), refTime.Location()), len(matches[0])
	}

	exp = regexp.MustCompile("^([0-9]+)\\s+((second)|(minute)|(hour)|(day)|(week)|(month)|(year))s?\\s+ago")
	matches = exp.FindStringSubmatch(input)
	if matches != nil {
		amount, _ := strconv.Atoi(matches[1])
		unit := matches[2]

		if durations[unit] > 0 {
			duration := time.Duration(amount) * durations[unit]
			return refTime.Add(-duration), len(matches[0])
		}

		days := 0
		months := 0
		years := 0
		switch unit {
		case "day":
			days = -amount
		case "week":
			days = -amount * 7
		case "month":
			months = -amount
		case "year":
			years = -amount
		}
		return refTime.AddDate(years, months, days), len(matches[0])

	}

	return refTime, 0
}
