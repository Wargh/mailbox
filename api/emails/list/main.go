package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/harryzcy/mailbox/internal/email"
	"github.com/harryzcy/mailbox/internal/util/apiutil"
)

// AWS Region
var region = os.Getenv("REGION")

func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (apiutil.Response, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	fmt.Println("request received")

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		fmt.Printf("unable to load SDK config, %v\n", err)
		return apiutil.NewErrorResponse(http.StatusInternalServerError, "internal error"), nil
	}

	emailType := req.QueryStringParameters["type"]
	year := req.QueryStringParameters["year"]
	month := req.QueryStringParameters["month"]
	order := req.QueryStringParameters["order"]
	pageSizeStr := req.QueryStringParameters["pageSize"]
	nextCursor := req.QueryStringParameters["nextCursor"]

	pageSize := email.DEFAULT_PAGE_SIZE
	if pageSizeStr != "" {
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil {
			return apiutil.NewErrorResponse(http.StatusBadRequest, "invalid input"), nil
		}
	}

	cursor := &email.Cursor{}
	err = cursor.BindString(nextCursor)
	if err != nil {
		return apiutil.NewErrorResponse(http.StatusBadRequest, "invalid input"), nil
	}

	fmt.Printf("request query: type: %s, year: %s, month: %s, order: %s, pageSize: %s, nextCursor: %s\n",
		emailType, year, month, order, pageSizeStr, nextCursor)

	result, err := email.List(ctx, dynamodb.NewFromConfig(cfg), email.ListInput{
		Type:       emailType,
		Year:       year,
		Month:      month,
		Order:      order,
		PageSize:   pageSize,
		NextCursor: cursor,
	})
	if err != nil {
		if err == email.ErrInvalidInput {
			return apiutil.NewErrorResponse(http.StatusBadRequest, "invalid input"), nil
		}
		if err == email.ErrTooManyRequests {
			fmt.Println("too many requests")
			return apiutil.NewErrorResponse(http.StatusTooManyRequests, "too many requests"), nil
		}
		fmt.Printf("email list failed: %v\n", err)
		return apiutil.NewErrorResponse(http.StatusInternalServerError, "internal error"), nil
	}

	body, err := json.Marshal(result)
	if err != nil {
		fmt.Printf("marshal failed: %v\n", err)
		return apiutil.NewErrorResponse(http.StatusInternalServerError, "internal error"), nil
	}
	return apiutil.NewSuccessJSONResponse(string(body)), nil
}

func main() {
	lambda.Start(handler)
}
