package email

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// GetResult represents the result of get method
type GetResult struct {
	TimeIndex
	Subject     string   `json:"subject"`
	DateSent    string   `json:"dateSent"`
	Source      string   `json:"source"`
	Destination []string `json:"destination"`
	From        []string `json:"from"`
	To          []string `json:"to"`
	ReturnPath  string   `json:"returnPath"`
	Text        string   `json:"text"`
	HTML        string   `json:"html"`
}

// Get returns the email
func Get(ctx context.Context, api GetItemAPI, messageID string) (*GetResult, error) {
	resp, err := api.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"MessageID": &types.AttributeValueMemberS{Value: messageID},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Item) == 0 {
		return nil, ErrNotFound
	}
	result := new(GetResult)
	err = attributevalue.UnmarshalMap(resp.Item, result)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	result.Type, result.TimeReceived, err = unmarshalGSI(resp.Item)
	if err != nil {
		return nil, err
	}

	fmt.Println("get method finished successfully")
	return result, nil
}