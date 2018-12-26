package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
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

func getEvents(url string) ([]ical.VEvent, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	title, err := getTitle(doc)
	if err != nil {
		return nil, err
	}

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

func getEventsFromFile(file string) ([]ical.VEvent, error) {
	//urls, err := getUrlsFromFile(file)
	//if err != nil {
	//	return nil, err
	//}

	urls := []string{"hest", "hest"}

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

//func getEventsFromUrl() ([]ical.VEvent, error) {
//
//}

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

func main() {
	url := flag.String("url", "", "write the URL from dfi.dk you wish to convert to an ICS file")
	inputFile := flag.String("file", "", "a plain text file with newline-seperated URLs")
	flag.Parse()

	if *url == "" && *inputFile == "" {
		log.Fatal("Please define either a URL (--url) or a file with URLs (--file)")
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
