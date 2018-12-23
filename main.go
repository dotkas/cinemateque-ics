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

	SELECT_TITLE       = "body > div.layout > div > div.layout__top > header > div.header__wrapper.js-header-body > div > div > div > div > div > p.header__hero__title"
	SELECT_EVENTS      = "#block-dfifilmpageblockdfi-cinemateket-film-page-block > div.supplementary > div > div.supplementary__aside > div.supplementary__list > div > div > p"
	SELECT_RUNTIME     = "#block-dfifilmpageblockdfi-cinemateket-film-page-block > div.supplementary > div > div.supplementary__content > div.text.layout__unit > p:nth-child(3)"
	SELECT_DESCRIPTION = "#block-dfifilmpageblockdfi-cinemateket-film-page-block > div.supplementary > div > div.supplementary__content > div.text.layout__unit > p:nth-child(1)"
)

func getRuntime(doc *goquery.Document) (int, error) {
	blockWithRuntime := doc.Find(SELECT_RUNTIME).First().Text()
	r := regexp.MustCompile("(?P<runtime>\\d*)\\smin\\.")
	parsed := r.FindStringSubmatch(blockWithRuntime)
	if len(parsed) < 2 {
		return 0, fmt.Errorf("could not parse blockWithRuntime: %v", parsed)
	}

	runtime, err := strconv.Atoi(parsed[1])
	if err != nil {
		return 0, fmt.Errorf("could not typecast %s to int", parsed)
	}

	return runtime, nil
}

func getDescription(doc *goquery.Document) (string, error) {
	description := doc.Find(SELECT_DESCRIPTION).First().Text()

	if description == "" {
		return description, fmt.Errorf("no description found in document")
	}

	return description, nil
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
	doc.Find(SELECT_EVENTS).Each(func(i int, s *goquery.Selection) {
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

	return events, nil
}

func main() {
	url := flag.String("url", "", "write the URL from dfi.dk you wish to convert to an ICS file")
	flag.Parse()

	if *url == "" {
		log.Fatal("Please define a URL")
	}

	//url := "https://www.dfi.dk/cinemateket/biograf/alle-film/film/big-blue"

	events, err := getEvents(*url)
	if err != nil {
		panic(err)
	}

	calendar := ical.NewBasicVCalendar()
	for _, event := range events {
		e := event // Avoid memory re-use (https://golang.org/ref/spec#For_range)
		calendar.VComponent = append(calendar.VComponent, &e)
	}

	f, err := os.Create("events.ics")
	if err != nil {
		log.Fatalf("couldn't open destination file: %v", err)
	}
	defer f.Close()

	if err := calendar.Encode(f); err != nil {
		log.Fatal(err)
	}

	return
}
