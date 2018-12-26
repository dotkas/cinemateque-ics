package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/dotkas/cinemateque-ics/ical"
	"io"
	"log"
	"os"
)

const (
	LOCATION = "Cinemateket, Lønporten 2, 1121 København K, Denmark"
)

func getEventsFromFile(path string) ([]ical.VEvent, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}

	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	events := make([]ical.VEvent, 0)
	for _, u := range urls {
		subEvents, err := getEvents(u)
		if err != nil {
			log.Fatal(err)
		}

		for _, event := range subEvents {
			events = append(events, event)
		}
	}

	return events, nil
}

func generateIcalFile(events []ical.VEvent) error {
	f, err := os.Create("events.ics")
	if err != nil {
		return fmt.Errorf("couldn't open destination file: %v", err)
	}
	defer f.Close()

	return convert(events, f)
}

func convert(events []ical.VEvent, w io.Writer) error {
	calendar := ical.NewBasicVCalendar()
	for _, event := range events {
		e := event // Avoid memory re-use (https://golang.org/ref/spec#For_range)
		calendar.VComponent = append(calendar.VComponent, &e)
	}

	return calendar.Encode(w)
}

func main() {
	url := flag.String("url", "", "write the URL from dfi.dk you wish to convert to an ICS file")
	inputFile := flag.String("file", "", "a plain text file with newline-seperated URLs")
	listen := flag.String("listen", "", "Listen on this address for incoming web requests to convert to ICS")

	flag.Parse()

	if *url == "" && *inputFile == "" && *listen == "" {
		log.Fatal("Please define either a URL (--url), a file with URLs (--file) or a port to listen on (--listen)")
	}

	if *listen != "" {
		log.Fatal(startServer(*listen))
		return
	}

	if *inputFile != "" {
		events, err := getEventsFromFile(*inputFile)
		if err != nil {
			log.Fatal(err)
			return
		}

		err = generateIcalFile(events)
		if err != nil {
			log.Fatal(err)
		}

		return
	}

	events, err := getEvents(*url)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	err = generateIcalFile(events)
	if err != nil {
		log.Fatal(err)
	}

	return
}
