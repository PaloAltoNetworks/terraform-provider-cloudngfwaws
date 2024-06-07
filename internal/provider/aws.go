package provider

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	ststypes "github.com/aws/aws-sdk-go-v2/service/sts/types"
)

// Creds implements the aws.CredentialsProvider interface.
type Creds struct {
	Credentials *ststypes.Credentials
}

// Retrieve accepts sts credentials and returns aws.Credentials.
func (creds Creds) Retrieve(ctx context.Context) (aws.Credentials, error) {
	credentials := creds.Credentials
	return aws.Credentials{
		AccessKeyID:     *credentials.AccessKeyId,
		SecretAccessKey: *credentials.SecretAccessKey,
		SessionToken:    *credentials.SessionToken,
		Expires:         *credentials.Expiration,
		CanExpire:       true,
	}, nil
}
