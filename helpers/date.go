package helpers

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

func getMonthNumber(name string) (string, error) {
	switch strings.TrimSpace(name) {
	case "januar":
		return "01", nil
	case "februar":
		return "02", nil
	case "marts":
		return "03", nil
	case "april":
		return "04", nil
	case "maj":
		return "05", nil
	case "juni":
		return "06", nil
	case "juli":
		return "07", nil
	case "august":
		return "08", nil
	case "september":
		return "09", nil
	case "oktober":
		return "10", nil
	case "november":
		return "11", nil
	case "december":
		return "12", nil
	}
	return "", fmt.Errorf("did not understand the name called %s", name)
}

func ParseDate(datetime string) (time.Time, error) {
	// Format is Fredag 21. december16:30
	r := regexp.MustCompile("(?P<Number>\\d*)\\.\\s(?P<month>\\D*)(?P<time>\\d*:\\d*)")

	parsed := r.FindStringSubmatch(datetime)
	if len(parsed) < 3 {
		return time.Now(), fmt.Errorf("could not find all expected matches, matched string was: %v", parsed)
	}

	n := parsed[1]

	// Ensure zero prefix for single digit days
	if len(n) == 1 {
		n = fmt.Sprintf("0%s", n)
	}

	m, err := getMonthNumber(parsed[2])
	if err != nil {
		return time.Now(), err
	}
	t := parsed[3]

	layout := "2006-01-02T15:04:05.000Z"
	str := fmt.Sprintf("%d-%s-%sT%s:00.000Z", time.Now().Year(), m, n, t)
	timestamp, err := time.Parse(layout, str)
	if err != nil {
		return time.Now(), fmt.Errorf("could not parse date: %s", err)
	}

	return timestamp, nil
}
