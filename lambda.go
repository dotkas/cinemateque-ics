package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if !strings.HasPrefix(request.Path, "/cinemateket/biograf/") {
		return events.APIGatewayProxyResponse{Body: "Invalid URL", StatusCode: http.StatusNotFound}, nil
	}

	path := url.URL{
		Scheme: "https",
		Host:   "www.dfi.dk",
		Path:   request.Path,
	}

	e, err := getEvents(path.String())
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	w := new(bytes.Buffer)
	if err := convert(e, w); err != nil {
		switch typedErr := err.(type) {

		case ErrInvalidStatusCode:
			return events.APIGatewayProxyResponse{
				Body:       fmt.Sprintf("invalid status code from url (%v), relaying: %v", path.String(), int(typedErr)),
				StatusCode: http.StatusInternalServerError,
			}, nil
		}

		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("error while converting event from url (%v): %v", path.String(), err),
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		Headers:    map[string]string{"content-type": "text/calendar"},
		Body:       w.String(),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handleRequest)
}
