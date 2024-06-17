package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go/api"
	"github.com/paloaltonetworks/cloud-ngfw-aws-go/api/account"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Resource.
func resourceAccountOnboarding() *schema.Resource {
	return &schema.Resource{
		Description: "Resource for Account Onboarding.",

		CreateContext: createAccountOnboarding,
		DeleteContext: deleteAccountOnboarding,
		ReadContext:   readAccountOnboarding,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: accountOnboardingSchema(true, nil),
	}
}

// createAccountOnboarding resource waits for the account onbarding to complete.
func createAccountOnboarding(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)
	accountId := d.Get("account_id").(string)
	res, err := Wait4AccountOnboardingCompletion(ctx, svc, accountId)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(accountId)
	d.Set("account_id", accountId)
	SaveAccount(d, res.Response)
	return nil
}

func deleteAccountOnboarding(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}

func readAccountOnboarding(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)
	accountId := d.Get("account_id").(string)
	req := account.ReadInput{
		AccountId: accountId,
	}
	res, err := svc.ReadAccount(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(accountId)
	d.Set("account_id", accountId)
	d.Set("onboarding_status", res.Response.OnboardingStatus)
	return nil
}

func Wait4AccountOnboardingCompletion(ctx context.Context, svc *api.ApiClient, accountId string) (account.ReadOutput, error) {
	res := account.ReadOutput{}
	var err error
	for i := 0; i < 10; i++ {
		req := account.ReadInput{
			AccountId: accountId,
		}
		res, err = svc.ReadAccount(ctx, req)
		if err != nil {
			return res, err
		}
		switch res.Response.OnboardingStatus {
		case "Success":
			return res, nil
		default:
			tflog.Info(ctx, "read account",
				map[string]interface{}{
					"status": res.Response.OnboardingStatus,
				})
		}
		time.Sleep(30 * time.Second)
	}
	return res, fmt.Errorf("timed out waiting for onboarding status")
}

// Schema handling.
func accountOnboardingSchema(isResource bool, rmKeys []string) map[string]*schema.Schema {
	ans := map[string]*schema.Schema{
		"account_id": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The account IDs",
		},
		"onboarding_status": {
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			Description: "Onboarding status of the account",
		},
	}

	for _, rmKey := range rmKeys {
		delete(ans, rmKey)
	}
	return ans
}
