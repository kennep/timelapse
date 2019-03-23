package client

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

var refTime = time.Date(1977, 12, 6, 12, 04, 32, 12345, time.Local)

func verifyTime(input string, expected time.Time) {
	result, err := ParseTimeRef(input, refTime)
	if err != nil {
		panic(err)
	}

	r, err := result.MarshalText()
	if err != nil {
		panic(err)
	}
	e, err := expected.MarshalText()
	if err != nil {
		panic(err)
	}

	if bytes.Compare(r, e) != 0 {
		panic(fmt.Sprintf("Expected result for input '%s' to be %s, not %s", input, expected, result))
	}
}

func verifyOffset(input string, expected string) {
	expectedDuration, err := time.ParseDuration(expected)
	if err != nil {
		panic(err)
	}
	verifyTime(input, refTime.Add(expectedDuration))
}

func TestParseError(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Log("Got a panic as expected")
		} else {
			panic("Expected a panic")
		}
	}()
	verifyTime("error", refTime)
}

func TestParseNow(t *testing.T) {
	verifyTime("now", refTime)
}

func TestParseYesterday(t *testing.T) {
	verifyOffset("yesterday", "-24h")
}

func TestParseTomorrow(t *testing.T) {
	verifyOffset("tomorrow", "24h")
}

func TestParseSecondsAgo(t *testing.T) {
	verifyOffset("3 seconds ago", "-3s")
}

func TestParseMinutesAgo(t *testing.T) {
	verifyOffset("25 minutes ago", "-25m")
}

func TestParseHoursAgo(t *testing.T) {
	verifyOffset("5 hours ago", "-5h")
}

func TestParseDaysAgo(t *testing.T) {
	verifyOffset("1 day ago", "-24h")
}

func TestParseWeeksAgo(t *testing.T) {
	verifyOffset("1 week ago", "-168h") // 168 hours in 1 week
}

func TestParseMonthsAgo(t *testing.T) {
	var expTime = time.Date(1977, 10, 6, 12, 04, 32, 12345, time.Local)
	verifyTime("2 months ago", expTime)
}

func TestParseYearsAgo(t *testing.T) {
	var expTime = time.Date(1967, 12, 6, 12, 04, 32, 12345, time.Local)
	verifyTime("10 years ago", expTime)
}

func TestParseDuration(t *testing.T) {
	verifyOffset("1h10m", "1h10m")
}

func TestParseHours(t *testing.T) {
	var expTime = time.Date(1977, 12, 6, 13, 24, 0, 0, time.Local)
	verifyTime("13:24", expTime)
}

func TestParseHoursRelDate1(t *testing.T) {
	var expTime = time.Date(1977, 12, 5, 13, 24, 0, 0, time.Local)
	verifyTime("13:24 yesterday", expTime)
}

func TestParseHoursRelDate2(t *testing.T) {
	var expTime = time.Date(1977, 12, 5, 13, 24, 0, 0, time.Local)
	verifyTime("yesterday 13:24", expTime)
}

func TestParseHoursRelDate3(t *testing.T) {
	var expTime = time.Date(1977, 12, 5, 13, 24, 0, 0, time.Local)
	verifyTime("yesterday at 13:24", expTime)
}

func TestParseDayOfMonth(t *testing.T) {
	var expTime = time.Date(1977, 12, 18, 13, 24, 0, 0, time.Local)
	verifyTime("18. at 13:24", expTime)
}

func TestParseDayOfMonth2(t *testing.T) {
	var expTime = time.Date(1977, 12, 18, 13, 24, 0, 0, time.Local)
	verifyTime("18th at 13:24", expTime)
}

func TestParseMonthDay(t *testing.T) {
	var expTime = time.Date(1977, 6, 18, 13, 24, 0, 0, time.Local)
	verifyTime("18th jun at 13:24", expTime)
}

func TestParseMonthDay2(t *testing.T) {
	var expTime = time.Date(1977, 6, 18, 13, 24, 0, 0, time.Local)
	verifyTime("jun 18th at 13:24", expTime)
}

func TestParseRef3359(t *testing.T) {
	var expTime = time.Date(1978, 10, 9, 15, 16, 0, 0, time.FixedZone("-0200", -7200))
	verifyTime("1978-10-09T15:16:00-02:00", expTime)
}

func TestParseLocalDateTime(t *testing.T) {
	var expTime = time.Date(1978, 10, 9, 15, 16, 0, 0, time.Local)
	verifyTime("1978-10-09T15:16:00", expTime)
}

func TestParseLocalDateTime2(t *testing.T) {
	var expTime = time.Date(1978, 10, 9, 15, 16, 0, 0, time.Local)
	verifyTime("1978-10-09 15:16", expTime)
}

func TestParseLocalDateTime3(t *testing.T) {
	var expTime = time.Date(1978, 10, 9, 15, 16, 0, 0, time.Local)
	verifyTime("09.10.1978 15:16", expTime)
}

func TestParseLocalDateTime4(t *testing.T) {
	var expTime = time.Date(1977, 10, 9, 15, 16, 0, 0, time.Local)
	verifyTime("09.10 15:16", expTime)
}
