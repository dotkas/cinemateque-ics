package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dotkas/cinemateque-ics/helpers"
	"github.com/dotkas/cinemateque-ics/ical"

	"github.com/PuerkitoBio/goquery"
)

const (
	LOCATION = "Cinemateket, Lønporten 2, 1121 København K, Denmark"

	SELECT_TITLE = "body > div.layout > div > div.layout__top > header > div.header__wrapper.js-header-body > div > div > div > div > div > p.header__hero__title"

	SELECT_SCHEDULE_NORMAL = "#block-dfifilmpageblockdfi-cinemateket-film-page-block > div.supplementary > div > div.supplementary__aside > div.supplementary__list > div > div > p"
	SELECT_SCHEDULE_EVENT  = "#block-dficinemateketeventpageblockdfi-cinemateket-event-page-block > div.supplementary > div > div.supplementary__aside > div.supplementary__list > div > div > p"

	SELECT_RUNTIME_NORMAL = "#block-dfifilmpageblockdfi-cinemateket-film-page-block > div.supplementary > div > div.supplementary__content > div.text.layout__unit > p:nth-child(3)"
	SELECT_RUNTIME_EVENT  = "#block-dficinemateketeventpageblockdfi-cinemateket-event-page-block > div.supplementary > div > div.supplementary__content > div:nth-child(2) > p:nth-child(6)"

	SELECT_DESCRIPTION_NORMAL = "#block-dfifilmpageblockdfi-cinemateket-film-page-block > div.supplementary > div > div.supplementary__content > div.text.layout__unit > p:nth-child(1)"
	SELECT_DESCRIPTION_EVENT  = "#block-dficinemateketeventpageblockdfi-cinemateket-event-page-block > div.supplementary > div > div.supplementary__content > div:nth-child(2) > p:nth-child(4)"
)

func getRuntime(doc *goquery.Document) (int, error) {
	attempts := []string{SELECT_RUNTIME_NORMAL, SELECT_RUNTIME_EVENT}
	for _, attempt := range attempts {
		blockWithRuntime := doc.Find(attempt).First().Text()
		r := regexp.MustCompile("(?P<runtime>\\d*)\\smin\\.")
		parsed := r.FindStringSubmatch(blockWithRuntime)
		if len(parsed) < 2 {
			continue
		}

		runtime, err := strconv.Atoi(parsed[1])
		if err != nil {
			return 0, fmt.Errorf("could not typecast %s to int", parsed)
		}

		return runtime, nil
	}

	return 0, fmt.Errorf("could not localize runtime in provided HTML file")
}

func getDescription(doc *goquery.Document) (string, error) {
	attempts := []string{SELECT_DESCRIPTION_NORMAL, SELECT_DESCRIPTION_EVENT}
	for _, attempt := range attempts {
		description := doc.Find(attempt).First().Text()

		if description == "" {
			continue
		}

		return description, nil

	}
	return "", fmt.Errorf("no description found in document")
}

func getTitle(doc *goquery.Document) (string, error) {
	title := doc.Find(SELECT_TITLE).First().Text()

	if title == "" {
		return title, fmt.Errorf("no title found in document")
	}

	return title, nil
}

type ErrInvalidStatusCode int

func (e ErrInvalidStatusCode) Error() string {
	return fmt.Sprintf("invalid status code; expected 200, got: %v", int(e))
}

func getEvents(url string) ([]ical.VEvent, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, ErrInvalidStatusCode(res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	title, err := getTitle(doc)
	if err != nil {
		return nil, err
	}
	log.Printf("Located title: %s", title)

	runtime, err := getRuntime(doc)
	if err != nil {
		return nil, err
	}

	desc, err := getDescription(doc)
	if err != nil {
		return nil, err
	}

	// Finds the available times of the event
	events := make([]ical.VEvent, 0)
	attempts := []string{SELECT_SCHEDULE_NORMAL, SELECT_SCHEDULE_EVENT}
	for _, attempt := range attempts {
		doc.Find(attempt).Each(func(i int, s *goquery.Selection) {
			log.Printf("%d: Located time: %s\n", i, s.Text())

			parsed, err := helpers.ParseDate(s.Text())
			if err != nil {
				panic(err)
			}

			e := ical.VEvent{
				Summary:     helpers.MustSerialize(title),
				Start:       parsed,
				End:         parsed.Add(time.Minute * time.Duration(runtime)),
				Description: helpers.MustSerialize(desc),
				Location:    helpers.MustSerialize(LOCATION),
				Url:         url,
				AllDay:      false,
				Tzid:        "Europe/Copenhagen",
			}
			events = append(events, e)
		})
	}

	return events, nil
}


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
	calendar := ical.NewBasicVCalendar()
	for _, event := range events {
		e := event // Avoid memory re-use (https://golang.org/ref/spec#For_range)
		calendar.VComponent = append(calendar.VComponent, &e)
	}

	f, err := os.Create("events.ics")
	if err != nil {
		return fmt.Errorf("couldn't open destination file: %v", err)
	}
	defer f.Close()

	if err := calendar.Encode(f); err != nil {
		log.Fatal(err)
	}

	return nil
}


func convert(events []ical.VEvent, w io.Writer) error {
	calendar := ical.NewBasicVCalendar()
	for _, event := range events {
		e := event // Avoid memory re-use (https://golang.org/ref/spec#For_range)
		calendar.VComponent = append(calendar.VComponent, &e)
	}

	return calendar.Encode(w)
}


func startServer(addr string) error {

	s := &http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 30 * time.Second,
		Handler:      http.HandlerFunc(eventHandler),
	}

	return s.ListenAndServe()

}


func eventHandler(rw http.ResponseWriter, req *http.Request) {

	rw.Header().Set("content-type", "text/calendar")

	eventURL := req.URL
	eventURL.Scheme = "https"
	eventURL.Host = "www.dfi.dk"

	if !strings.HasPrefix(eventURL.Path, "/cinemateket/biograf/") {
		http.Error(rw, "Invalid URL", http.StatusNotFound)
		return
	}

	if err := convert(eventURL.String(), rw); err != nil {

		switch typedErr := err.(type) {

		case ErrInvalidStatusCode:
			log.Printf("invalid status code from url (%v), relaying: %v", eventURL, int(typedErr))
			rw.WriteHeader(int(typedErr))
			return

		}

		log.Printf("error while converting event from url (%v): %v", eventURL, err)
		http.Error(rw, "Error fetching event", http.StatusInternalServerError)
		return
	}

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
