package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	testAccAccountId string
)

const (
	TestAccAccountIdNotDefined = "Missing environment variable: CLOUD_NGFW_ACCOUNT_ID"
)

func init() {
	// Uncomment if we need runtime info later.
	/*
	   var err error
	   ctx := context.TODO()

	   con := awsngfw.Client{
	       Host: os.Getenv("CLOUD_NGFW_HOST"),
	       Region: os.Getenv("CLOUD_NGFW_REGION"),
	       Arn: os.Getenv("CLOUD_NGFW_ARN"),
	       LfaArn: os.Getenv("CLOUD_NGFW_LFA_ARN"),
	       LraArn: os.Getenv("CLOUD_NGFW_LRA_ARN"),
	       Logging: awsngfw.LogQuiet,
	   }
	   if err = con.Setup(); err != nil {
	       return
	   } else if err = con.RefreshJwts(ctx); err != nil {
	       return
	   }
	*/

	testAccAccountId = os.Getenv("CLOUD_NGFW_ACCOUNT_ID")
}

var providerFactories = map[string]func() (*schema.Provider, error){
	"cloudngfwaws": func() (*schema.Provider, error) {
		return New("dev")(), nil
	},
}

func TestProvider(t *testing.T) {
	if err := New("dev")().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = New("dev")()
}

func testAccPreCheck(t *testing.T) {
	opts := []string{
		"CLOUD_NGFW_HOST",
		"CLOUD_NGFW_REGION",
	}
	for _, x := range opts {
		if os.Getenv(x) == "" {
			t.Fatal(fmt.Sprintf("%q must be set for acctests.", x))
		}
	}

	shared_arn := "CLOUD_NGFW_ARN"
	arns := []string{"CLOUD_NGFW_LFA_ARN", "CLOUD_NGFW_LRA_ARN"}
	for _, arn := range arns {
		if os.Getenv(shared_arn) == "" && os.Getenv(arn) == "" {
			t.Fatal(fmt.Sprintf("One of %q or %q must be specified", shared_arn, arn))
		}
	}
}
