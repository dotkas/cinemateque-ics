package main

import (
	"github.com/dotkas/cinemateque-ics/ical"
	"io"
)

const (
	LOCATION = "Cinemateket, Lønporten 2, 1121 København K, Denmark"
)

func convert(events []ical.VEvent, w io.Writer) error {
	calendar := ical.NewBasicVCalendar()
	for _, event := range events {
		e := event // Avoid memory re-use (https://golang.org/ref/spec#For_range)
		calendar.VComponent = append(calendar.VComponent, &e)
	}

	return calendar.Encode(w)
}
