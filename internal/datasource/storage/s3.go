package storage

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/harryzcy/mailbox/internal/env"
	"github.com/harryzcy/mailbox/internal/model"
	"github.com/jhillyerd/enmime/v2"
)

var (
	ErrorInvalidDisposition = errors.New("invalid disposition")

	DispositionAttachments = "attachments"
	DispositionInlines     = "inlines"
	DispositionOthers      = "others"
)

type GetEmailResult struct {
	Text        string
	HTML        string
	Attachments model.Files
	Inlines     model.Files
	OtherParts  model.Files
}

// S3Storage is an interface that defines required S3 functions
type S3Storage interface {
	GetEmail(ctx context.Context, api S3GetObjectAPI, messageID string) (*GetEmailResult, error)
	DeleteEmail(ctx context.Context, api S3DeleteObjectAPI, messageID string) error
	GetEmailRaw(ctx context.Context, api S3GetObjectAPI, messageID string) ([]byte, error)
	GetEmailContent(ctx context.Context, api S3GetObjectAPI, messageID, disposition, contentID string) (*GetEmailContentResult, error)
}

type s3Storage struct{}

// S3 holds functions that handles S3 related operations
var S3 S3Storage = s3Storage{}

// readEmailEnvelope is used in GetEmail will be mocked in unit testing
var readEmailEnvelope = enmime.ReadEnvelope

// S3GetObjectAPI defines set of API required by GetEmail functions
type S3GetObjectAPI interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

// GetEmail retrieves an email from s3 bucket
func (s s3Storage) GetEmail(ctx context.Context, api S3GetObjectAPI, messageID string) (*GetEmailResult, error) {
	object, err := api.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &env.S3Bucket,
		Key:    &messageID,
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		err = object.Body.Close()
		fmt.Println("error closing object body", err)
	}()

	env, err := readEmailEnvelope(object.Body)
	if err != nil {
		return nil, err
	}
	return &GetEmailResult{
		Text:        env.Text,
		HTML:        env.HTML,
		Attachments: ParseFiles(env.Attachments),
		Inlines:     ParseFiles(env.Inlines),
		OtherParts:  ParseFiles(env.OtherParts),
	}, nil
}

// GetEmailRaw retrieves raw MIME email from s3 bucket
func (s s3Storage) GetEmailRaw(ctx context.Context, api S3GetObjectAPI, messageID string) ([]byte, error) {
	object, err := api.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &env.S3Bucket,
		Key:    &messageID,
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		err = object.Body.Close()
		fmt.Println("error closing object body", err)
	}()

	raw, err := io.ReadAll(object.Body)
	return raw, err
}

type GetEmailContentResult struct {
	model.File
	Content []byte
}

// GetEmailContent retrieved the attachment of inline of an email from s3 bucket
func (s s3Storage) GetEmailContent(ctx context.Context, api S3GetObjectAPI, messageID, disposition, contentID string) (*GetEmailContentResult, error) {
	object, err := api.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &env.S3Bucket,
		Key:    &messageID,
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		err = object.Body.Close()
		fmt.Println("error closing object body", err)
	}()

	env, err := readEmailEnvelope(object.Body)
	if err != nil {
		return nil, err
	}

	var parts []*enmime.Part

	switch disposition {
	case DispositionAttachments:
		parts = env.Attachments
	case DispositionInlines:
		parts = env.Inlines
	case DispositionOthers:
		parts = env.OtherParts
	default:
		return nil, ErrorInvalidDisposition
	}

	// find the part with the correct contentID
	for _, part := range parts {
		if part.ContentID == contentID {
			return &GetEmailContentResult{
				File: model.File{
					ContentID:         part.ContentID,
					ContentType:       part.ContentType,
					ContentTypeParams: part.ContentTypeParams,
					Filename:          part.FileName,
				},
				Content: part.Content,
			}, nil
		}
	}
	return nil, nil
}

// S3DeleteObjectAPI defines set of API required by DeleteEmail functions
type S3DeleteObjectAPI interface {
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

// DeleteEmail deletes an email from S3 bucket
func (s s3Storage) DeleteEmail(ctx context.Context, api S3DeleteObjectAPI, messageID string) error {
	_, err := api.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &env.S3Bucket,
		Key:    &messageID,
	})
	if err != nil {
		return err
	}

	return nil
}

// ParseFiles parses enmime parts into File slice
func ParseFiles(parts []*enmime.Part) model.Files {
	files := make([]model.File, len(parts))
	for i, part := range parts {
		files[i] = model.File{
			ContentID:         part.ContentID,
			ContentType:       part.ContentType,
			ContentTypeParams: part.ContentTypeParams,
			Filename:          part.FileName,
		}
	}
	return files
}
