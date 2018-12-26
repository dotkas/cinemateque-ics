package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type ErrInvalidStatusCode int

func (e ErrInvalidStatusCode) Error() string {
	return fmt.Sprintf("invalid status code; expected 200, got: %v", int(e))
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

	events, err := getEvents(eventURL.String())
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	if err := convert(events, rw); err != nil {
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

func startServer(addr string) error {

	s := &http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 30 * time.Second,
		Handler:      http.HandlerFunc(eventHandler),
	}

	return s.ListenAndServe()

}
