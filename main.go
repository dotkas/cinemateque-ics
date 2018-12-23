package main

import (
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

const LOCATION = "Cinemateket, Lønporten 2, 1121 København K, Denmark"

func getRuntime(doc *goquery.Document) (int, error) {
	blockWithRuntime := doc.Find("#block-dfifilmpageblockdfi-cinemateket-film-page-block > div.supplementary > div > div.supplementary__content > div.text.layout__unit > p:nth-child(3)").First().Text()
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
	description := doc.Find("#block-dfifilmpageblockdfi-cinemateket-film-page-block > div.supplementary > div > div.supplementary__content > div.text.layout__unit > p:nth-child(1)").First().Text()

	if description == "" {
		return description, fmt.Errorf("no description found in document")
	}

	return description, nil
}

func getTitle(doc *goquery.Document) (string, error) {
	title := doc.Find("body > div.layout > div > div.layout__top > header > div.header__wrapper.js-header-body > div > div > div > div > div > p.header__hero__title").First().Text()

	if title == "" {
		return title, fmt.Errorf("no title found in document")
	}

	return title, nil
}

func getEvents(url string) ([]ical.VEvent, error) {
	// Request the HTML page.
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	events := make([]ical.VEvent, 0)

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
	doc.Find("#block-dfifilmpageblockdfi-cinemateket-film-page-block > div.supplementary > div > div.supplementary__aside > div.supplementary__list > div > div > p").Each(func(i int, s *goquery.Selection) {
		fmt.Printf("Located time: %d: %s\n", i, s.Text())

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
	url := "https://www.dfi.dk/cinemateket/biograf/alle-film/film/big-blue"

	events, err := getEvents(url)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Events: %v\n", events)

	calendar := ical.NewBasicVCalendar()
	for _, event := range events {
		event := event
		calendar.VComponent = append(calendar.VComponent, &event)
	}

	f, err := os.Create("events.ics")
	if err != nil {
		log.Fatal("couldn't open destination file: %v", err)
	}

	defer f.Close()

	if err := calendar.Encode(f); err != nil {
		log.Fatal(err)
	}

	return
}
