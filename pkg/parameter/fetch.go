package parameter

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type ParamFetch struct {
	Name       string
	Decryption bool
}

func FetchParameter(ssmClient *ssm.Client, ctx context.Context, p ParamFetch) (string, error) {
	s, err := ssmClient.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(p.Name),
		WithDecryption: aws.Bool(p.Decryption),
	})
	if err != nil {
		return "", fmt.Errorf("failed to fetch %q: %w", p.Name, err)
	}
	return aws.ToString(s.Parameter.Value), nil
}
