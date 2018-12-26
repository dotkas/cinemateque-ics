package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

func HandleRequest(ctx context.Context) (string, error) {
	lc, _ := lambdacontext.FromContext(ctx)
	fmt.Print(lc)
	return "Hello World!", nil
}

func Handler() {
	lambda.Start(HandleRequest)

}
