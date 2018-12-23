package ical

// https://www.ietf.org/rfc/rfc2445.txt

import (
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
)

const (
	stampLayout    = "20060102T150405Z"
	dateLayout     = "20060102"
	dateTimeLayout = "20060102T150405"
)

type VCalendar struct {
	Version  string // 2.0
	Calscale string // GREGORIAN

	VComponent []VComponent
}

func NewBasicVCalendar() *VCalendar {
	return &VCalendar{
		Version:  "2.0",
		Calscale: "GREGORIAN",
	}
}

func (c *VCalendar) Encode(w io.Writer) error {

	if _, err := fmt.Fprint(w, "BEGIN:VCALENDAR\r\n"); err != nil {
		return err
	}

	// use a slice map to preserve order during for range
	attrs := []map[string]string{
		{"VERSION": c.Version},
		{"CALSCALE": c.Calscale},
	}

	for _, item := range attrs {
		for k, v := range item {
			if len(v) == 0 {
				continue
			}

			if _, err := fmt.Fprintf(w, "%s:%s\r\n", k, v); err != nil {
				return err
			}
		}
	}

	for _, component := range c.VComponent {
		if err := component.EncodeIcal(w); err != nil {
			return err
		}
	}

	_, err := fmt.Fprint(w, "END:VCALENDAR\r\n")

	return err

}

type VComponent interface {
	EncodeIcal(w io.Writer) error
}

type VEvent struct {
	AllDay bool

	Start       time.Time
	End         time.Time
	Summary     string
	Description string
	Tzid        string
	Url         string
	Location    string
}

func (e *VEvent) EncodeIcal(w io.Writer) error {
	var timeStampLayout, timeStampType, tzidTxt string

	if e.AllDay {
		timeStampLayout = dateLayout
		timeStampType = "DATE"
	} else {
		timeStampLayout = dateTimeLayout
		timeStampType = "DATE-TIME"
		if len(e.Tzid) == 0 || e.Tzid == "UTC" {
			timeStampLayout = fmt.Sprintf("%sZ", timeStampLayout)
		}
	}

	if len(e.Tzid) != 0 && e.Tzid != "UTC" {
		tzidTxt = fmt.Sprintf("TZID=%s", e.Tzid)
	}

	if _, err := fmt.Fprint(w, "BEGIN:VEVENT\r\n"); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "DTSTAMP:%s\r\n", time.Now().UTC().Format(stampLayout)); err != nil {
		return err
	}

	uid, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "UID:%s\r\n", uid); err != nil {
		return err
	}

	if len(e.Tzid) != 0 && e.Tzid != "UTC" {
		if _, err := fmt.Fprintf(w, "TZID:%s\r\n", e.Tzid); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(w, "SUMMARY:%s\r\n", e.Summary); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "URL:%s\r\n", e.Url); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "LOCATION:%s\r\n", e.Location); err != nil {
		return err
	}

	if e.Description != "" {
		if _, err := fmt.Fprintf(w, "DESCRIPTION:%s\r\n", e.Description); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(w, "DTSTART;%s;VALUE=%s:%s\r\n", tzidTxt, timeStampType, e.Start.Format(timeStampLayout)); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "DTEND;%s;VALUE=%s:%s\r\n", tzidTxt, timeStampType, e.End.Format(timeStampLayout)); err != nil {
		return err
	}

	_, err = fmt.Fprint(w, "END:VEVENT\r\n")

	return err

}
