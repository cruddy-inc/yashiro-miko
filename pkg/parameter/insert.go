package parameter

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

type ParamInsert struct {
	Name      string
	Value     string
	Secure    bool
	Overwrite bool
}

func PutParameter(ssmClient *ssm.Client, ctx context.Context, params []ParamInsert) error {
	for i := 0; i < len(params); i++ {
		p := params[i]
		s := types.ParameterTypeString
		if p.Secure {
			s = types.ParameterTypeSecureString
		}
		_, err := ssmClient.PutParameter(
			ctx,
			&ssm.PutParameterInput{
				Name:      aws.String(p.Name),
				Value:     aws.String(p.Value),
				Type:      s,
				Overwrite: aws.Bool(p.Overwrite),
			},
		)
		if err != nil {
			return fmt.Errorf("failed to insert %q: %w", p.Name, err)
		}
	}
	return nil
}
