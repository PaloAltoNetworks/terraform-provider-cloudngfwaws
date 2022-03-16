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
	TestAccAccountIdNotDefined = "Missing environment variable: CLOUDNGFWAWS_ACCOUNT_ID"
)

func init() {
	// Uncomment if we need runtime info later.
	/*
	   var err error
	   ctx := context.TODO()

	   con := awsngfw.Client{
	       Host: os.Getenv("CLOUDNGFWAWS_HOST"),
	       Region: os.Getenv("CLOUDNGFWAWS_REGION"),
	       Arn: os.Getenv("CLOUDNGFWAWS_ARN"),
	       LfaArn: os.Getenv("CLOUDNGFWAWS_LFA_ARN"),
	       LraArn: os.Getenv("CLOUDNGFWAWS_LRA_ARN"),
	       Logging: awsngfw.LogQuiet,
	   }
	   if err = con.Setup(); err != nil {
	       return
	   } else if err = con.RefreshJwts(ctx); err != nil {
	       return
	   }
	*/

	testAccAccountId = os.Getenv("CLOUDNGFWAWS_ACCOUNT_ID")
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
		"CLOUDNGFWAWS_HOST",
		"CLOUDNGFWAWS_REGION",
	}
	for _, x := range opts {
		if os.Getenv(x) == "" {
			t.Fatal(fmt.Sprintf("%q must be set for acctests.", x))
		}
	}

	shared_arn := "CLOUDNGFWAWS_ARN"
	arns := []string{"CLOUDNGFWAWS_LFA_ARN", "CLOUDNGFWAWS_LRA_ARN"}
	for _, arn := range arns {
		if os.Getenv(shared_arn) == "" && os.Getenv(arn) == "" {
			t.Fatal(fmt.Sprintf("One of %q or %q must be specified", shared_arn, arn))
		}
	}
}
