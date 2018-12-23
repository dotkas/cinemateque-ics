package ical

// https://www.ietf.org/rfc/rfc2445.txt

import (
	"bufio"
	"github.com/google/uuid"
	"io"
	"time"
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
	var b = bufio.NewWriter(w)

	if _, err := b.WriteString("BEGIN:VCALENDAR\r\n"); err != nil {
		return err
	}

	// use a slice map to preserve order during for range
	attrs := []map[string]string{
		{"VERSION:": c.Version},
		{"CALSCALE:": c.Calscale},
	}

	for _, item := range attrs {
		for k, v := range item {
			if len(v) == 0 {
				continue
			}

			if _, err := b.WriteString(k + v + "\r\n"); err != nil {
				return err
			}
		}
	}

	for _, component := range c.VComponent {
		if err := component.EncodeIcal(b); err != nil {
			return err
		}
	}

	if _, err := b.WriteString("END:VCALENDAR\r\n"); err != nil {
		return err
	}

	return b.Flush()
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
			timeStampLayout = timeStampLayout + "Z"
		}
	}

	if len(e.Tzid) != 0 && e.Tzid != "UTC" {
		tzidTxt = "TZID=" + e.Tzid + ";"
	}

	b := bufio.NewWriter(w)
	if _, err := b.WriteString("BEGIN:VEVENT\r\n"); err != nil {
		return err
	}

	if _, err := b.WriteString("DTSTAMP:" + time.Now().UTC().Format(stampLayout) + "\r\n"); err != nil {
		return err
	}

	uid, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	if _, err := b.WriteString("UID:" + uid.String() + "\r\n"); err != nil {
		return err
	}

	if len(e.Tzid) != 0 && e.Tzid != "UTC" {
		if _, err := b.WriteString("TZID:" + e.Tzid + "\r\n"); err != nil {
			return err
		}
	}

	if _, err := b.WriteString("SUMMARY:" + e.Summary + "\r\n"); err != nil {
		return err
	}

	if _, err := b.WriteString("URL:" + e.Url + "\r\n"); err != nil {
		return err
	}

	if _, err := b.WriteString("LOCATION:" + e.Location + "\r\n"); err != nil {
		return err
	}

	if e.Description != "" {
		if _, err := b.WriteString("DESCRIPTION:" + e.Description + "\r\n"); err != nil {
			return err
		}
	}

	if _, err := b.WriteString("DTSTART;" + tzidTxt + "VALUE=" + timeStampType + ":" + e.Start.Format(timeStampLayout) + "\r\n"); err != nil {
		return err
	}

	if _, err := b.WriteString("DTEND;" + tzidTxt + "VALUE=" + timeStampType + ":" + e.End.Format(timeStampLayout) + "\r\n"); err != nil {
		return err
	}

	if _, err := b.WriteString("END:VEVENT\r\n"); err != nil {
		return err
	}

	return b.Flush()
}
