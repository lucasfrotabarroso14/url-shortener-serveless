package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"strings"
	"time"
)

var dbClient *dynamodb.Client
var s3Client *s3.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	dbClient = dynamodb.NewFromConfig(cfg)

	s3Client = s3.NewFromConfig(cfg)
}

func Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	code := req.PathParameters["code"]
	if code == "" {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 400,
			Body:       "Code parameter is required",
		}, nil
	}

	out, err := dbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String("url_shortener"),
		Key: map[string]types.AttributeValue{
			"short_code": &types.AttributeValueMemberS{Value: code},
		},
	})
	fmt.Println("Item found:", out)
	fmt.Println("PathParameters:", req.PathParameters)

	if err != nil || out.Item == nil {
		fmt.Println("Item not found:", err)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 404,
			Body:       "URL not found",
		}, nil
	}

	_, _ = dbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String("url_shortener"),
		Key: map[string]types.AttributeValue{
			"short_code": &types.AttributeValueMemberS{Value: code},
		},
		UpdateExpression: aws.String("SET hit_count = hit_count + :inc"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":inc": &types.AttributeValueMemberN{Value: "1"},
		},
	})

	attr, exists := out.Item["original_url"]
	if !exists {
		return events.APIGatewayV2HTTPResponse{StatusCode: 500, Body: "Campo original_url ausente"}, nil
	}

	urlAttr, ok := attr.(*types.AttributeValueMemberS)
	if !ok {
		return events.APIGatewayV2HTTPResponse{StatusCode: 500, Body: "Campo original_url não é uma string"}, nil
	}

	logContent := fmt.Sprintf("code = %s, url = %s, ip= %s, timestamp = %s \n",
		code, urlAttr.Value, req.Headers["x-fowarded-for"], time.Now().String())

	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String("url-shortener-access-logs"),
		Key:    aws.String(fmt.Sprintf("logs/%s.txt", code)),
		Body:   strings.NewReader(logContent),
	})
	if err != nil {
		fmt.Println("Failed to upload log:", err)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Body:       "Failed to upload log",
		}, nil
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 301,
		Headers: map[string]string{
			"Location": urlAttr.Value,
		},
	}, nil

}

func main() {
	lambda.Start(Handler)
}
