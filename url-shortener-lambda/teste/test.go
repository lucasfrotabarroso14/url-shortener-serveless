package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
)

func generateFakeRequest(url string) events.APIGatewayV2HTTPRequest {
	return events.APIGatewayV2HTTPRequest{
		Body: fmt.Sprintf(`{"url":"%s"}`, url),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

func main() {
	req := generateFakeRequest("https://example.com")
	resp, err := HandlerShorten(context.TODO(), req)

}
