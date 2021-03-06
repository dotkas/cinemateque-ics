package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/dotkas/cinemateque-ics/helpers"
	"github.com/dotkas/cinemateque-ics/ical"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

const (
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

	if len(events) == 0 {
		return nil, fmt.Errorf("no runtimes found for title %s", title)
	}

	return events, nil
}
