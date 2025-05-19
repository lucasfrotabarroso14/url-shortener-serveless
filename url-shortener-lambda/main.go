package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"math/rand"
	"time"
)

var dbClient *dynamodb.Client

//func init() {
//	// Resolver personalizado para apontar para o DynamoDB local
//	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
//		return aws.Endpoint{
//			URL:           "http://localhost:8000", // Docker local
//			SigningRegion: "us-east-1",
//		}, nil
//	})
//
//	// Carregar config com credenciais dummy e resolver local
//	cfg, err := config.LoadDefaultConfig(context.TODO(),
//		config.WithRegion("us-east-1"),
//		config.WithCredentialsProvider(
//			credentials.NewStaticCredentialsProvider("dummy", "dummy", ""),
//		),
//		config.WithEndpointResolverWithOptions(customResolver),
//	)
//	if err != nil {
//		panic("failed to load config: " + err.Error())
//	}
//
//	dbClient = dynamodb.NewFromConfig(cfg)
//}

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	dbClient = dynamodb.NewFromConfig(cfg)
}

type ShortenRequest struct {
	Url string `json:"url"`
}

type ShortenItem struct {
	ShortCode   string `json:"short_code"`
	OriginalUrl string `json:"original_url"`
	CreatedAt   string `json:"created_at"`
	HitCount    int    `json:"hit_count"`
}

func generateShortCode(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
func Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var input ShortenRequest
	err := json.Unmarshal([]byte(req.Body), &input)
	if err != nil || input.Url == "" {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 400,
			Body:       "Invalid request",
		}, nil
	}

	shortCode := generateShortCode(6)
	shortItem := map[string]types.AttributeValue{
		"short_code":   &types.AttributeValueMemberS{Value: shortCode},
		"original_url": &types.AttributeValueMemberS{Value: input.Url},
		"created_at":   &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		"hit_count":    &types.AttributeValueMemberN{Value: "0"},
	}
	_, err = dbClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("url_shortener"),
		Item:      shortItem,
	})
	if err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Body:       err.Error(),
		}, nil
	}
	response := fmt.Sprintf(`{"short_url": "https://seuencurtador.com/%s"}`, shortCode)
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Body:       response,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil

}

func main() {

	rand.Seed(time.Now().UnixNano())
	lambda.Start(Handler)

	//req := generateFakeRequest("https://example.com")
	//resp, err := Handler(context.Background(), req)
	//if err != nil {
	//	fmt.Println("Erro:", err)
	//}
	//fmt.Println("StatusCode:", resp.StatusCode)
	//fmt.Println("Body:", resp.Body)

}

//
//func generateFakeRequest(url string) events.APIGatewayV2HTTPRequest {
//	return events.APIGatewayV2HTTPRequest{
//		Body: fmt.Sprintf(`{"url":"%s"}`, url),
//		Headers: map[string]string{
//			"Content-Type": "application/json",
//		},
//	}
//}
