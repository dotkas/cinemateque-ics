package helpers

import (
	"testing"
)

func TestDate_Parsing(t *testing.T) {
	tests := []struct {
		description string
		input       string
		expected    string
		shouldFail  bool
	}{
		{
			"Successfully parses a day with non-latin letters",
			"Søndag 30. december21:30",
			"2018-12-30 21:30:00",
			false,
		},
		{
			"Successfully parses a day with one digit in day of month",
			"Søndag 1. december21:30",
			"2018-12-01 21:30:00",
			false,
		},
		{
			"Successfully parses a day with one digit in day hour",
			"Søndag 12. december9:30",
			"2018-12-12 09:30:00",
			false,
		},
		{
			"Fails on unknown month",
			"Søndag 12. hestember:30",
			"",
			true,
		},
		{
			"Fails on unknown timestamp",
			"Søndag 12. december80",
			"",

			true,
		},
		{
			"Fails on totally wrong string",
			"Claatu baratu naktuu ",
			"",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			res, err := ParseDate(tt.input)
			if err != nil {
				if tt.shouldFail == false {
					t.Errorf("%s should not fail, error was: %v", tt.description, err)
				}
				return
			}

			formatted := res.Format("2006-01-02 15:04:05")
			if tt.expected != formatted {
				t.Errorf("%s: %s != %s", tt.description, tt.expected, formatted)
				return
			}
		})
	}
}
