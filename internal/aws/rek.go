// internal/aws/rek.go
package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	rek "github.com/aws/aws-sdk-go-v2/service/rekognition"
)

type RekClient interface {
	CreateFaceLivenessSession(ctx context.Context, params *rek.CreateFaceLivenessSessionInput, optFns ...func(*rek.Options)) (*rek.CreateFaceLivenessSessionOutput, error)
	GetFaceLivenessSessionResults(ctx context.Context, params *rek.GetFaceLivenessSessionResultsInput, optFns ...func(*rek.Options)) (*rek.GetFaceLivenessSessionResultsOutput, error)
	CompareFaces(ctx context.Context, params *rek.CompareFacesInput, optFns ...func(*rek.Options)) (*rek.CompareFacesOutput, error)
}

type Rek struct {
	Client *rek.Client
	Region string
}

func NewRek(ctx context.Context, awsRegion string) (*Rek, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsRegion))
	if err != nil {
		return nil, err
	}

	return &Rek{
		Client: rek.NewFromConfig(cfg),
		Region: awsRegion,
	}, nil
}
