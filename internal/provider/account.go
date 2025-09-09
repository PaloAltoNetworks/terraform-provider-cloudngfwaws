package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go/v2/api"
	"github.com/paloaltonetworks/cloud-ngfw-aws-go/v2/api/account"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Data source (list Accounts).
func dataSourceAccounts() *schema.Resource {
	return &schema.Resource{
		Description: "Data source get a list of Accounts.",

		ReadContext: readAccounts,

		Schema: map[string]*schema.Schema{
			"describe": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Flag to include account details while listing accounts.",
				Default:     false,
			},
			"account_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "List of account ids.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"account_details": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "List of account details.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The account id.",
						},
						"onboarding_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Onboarding status of the account.",
						},
						"external_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "External Id of the onboarded account",
						},
					},
				},
			},
		},
	}
}

func readAccounts(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)

	var listing []string
	var listingDescribe []account.AccountDetail
	var nt string
	describe := d.Get("describe").(bool)
	for {
		input := account.ListInput{
			Describe:   describe,
			NextToken:  nt,
			MaxResults: 50,
		}
		ans, err := svc.ListAccounts(ctx, input)
		if err != nil {
			if isObjectNotFound(err) {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
		listing = append(listing, ans.Response.AccountIds...)
		if describe {
			listingDescribe = append(listingDescribe, ans.Response.AccountDetails...)
		}
		nt = ans.Response.NextToken
		if nt == "" {
			break
		}
	}
	d.SetId(strconv.Itoa(len(listing)))
	d.Set("account_ids", listing)
	if describe {
		account_details := make([]interface{}, 0, len(listingDescribe))
		for _, x := range listingDescribe {
			account_details = append(account_details, map[string]interface{}{
				"account_id":        x.AccountId,
				"onboarding_status": x.OnboardingStatus,
				"external_id":       x.ExternalId,
			})
		}
		d.Set("account_details", account_details)
	}
	return nil
}

// Data source for a single Account.
func dataSourceAccount() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for retrieving account information.",

		ReadContext: readAccountDataSource,

		Schema: accountSchema(false, nil),
	}
}

func readAccountDataSource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)

	accountId := d.Get("account_id").(string)

	tflog.Info(
		ctx, "read account",
		map[string]interface{}{
			"ds":         true,
			"account_id": accountId,
		},
	)
	req := account.ReadInput{
		AccountId: accountId,
	}
	res, err := svc.ReadAccount(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(accountId)
	SaveAccount(d, res.Response)
	return nil
}

// Resource.
func resourceAccount() *schema.Resource {
	return &schema.Resource{
		Description: "Resource for Account manipulation.",

		CreateContext: createAccount,
		ReadContext:   readAccount,
		DeleteContext: deleteAccount,
		UpdateContext: updateAccount,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: accountSchema(true, nil),
	}
}

func createAccount(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)

	accountId := d.Get("account_id").(string)
	tflog.Info(
		ctx, "create account",
		map[string]interface{}{
			"account_id": accountId,
			"Origin":     "ProgrammaticAccess",
		},
	)

	o := account.CreateInput{
		AccountId: accountId,
		Origin:    "ProgrammaticAccess",
	}

	res, err := svc.CreateAccount(ctx, o)
	if err != nil {
		err = fmt.Errorf("failed to create account %s, err: %s", accountId, err)
		return diag.FromErr(err)
	}
	d.SetId(accountId)
	saveCreatedAccount(d, res.Response)
	return nil
}

func readAccount(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)

	accountId := d.Get("account_id").(string)
	tflog.Info(
		ctx, "read account",
		map[string]interface{}{
			"account_id": accountId,
		},
	)

	req := account.ReadInput{
		AccountId: accountId,
	}

	res, err := svc.ReadAccount(ctx, req)
	if err != nil {
		if isObjectNotFound(err) {
			d.SetId("")
			return nil
		}
		err = fmt.Errorf("failed to read account %s, err: %s", accountId, err)
		return diag.FromErr(err)
	}
	d.SetId(accountId)
	SaveAccount(d, res.Response)

	return nil
}

func deleteAccount(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)

	accountId := d.Get("account_id").(string)

	tflog.Info(
		ctx, "delete account",
		map[string]interface{}{
			"account_id": accountId,
		},
	)

	account := account.DeleteInput{
		AccountId: accountId,
	}

	if err := svc.DeleteAccount(ctx, account); err != nil && !isObjectNotFound(err) {
		err = fmt.Errorf("failed to delete account %s, err: %s", accountId, err)
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func updateAccount(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Errorf("account update is not supported")
}

func SaveAccount(d *schema.ResourceData, o account.ReadResponse) {
	d.Set("account_id", o.AccountId)
	d.Set("cft_url", o.CloudFormationTemplateURL)
	d.Set("onboarding_status", o.OnboardingStatus)
	d.Set("external_id", o.ExternalId)
	d.Set("service_account_id", o.ServiceAccountId)
	d.Set("sns_topic_arn", o.SNSTopicArn)
	d.Set("update_token", o.UpdateToken)
}

func saveCreatedAccount(d *schema.ResourceData, o account.Info) {
	d.Set("trusted_account", o.TrustedAccount)
	d.Set("external_id", o.ExternalId)
	d.Set("sns_topic_arn", o.SNSTopicArn)
	d.Set("origin", o.Origin)
}

// Schema handling.
func accountSchema(isResource bool, rmKeys []string) map[string]*schema.Schema {
	ans := map[string]*schema.Schema{
		"account_id": {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "The account ID",
		},
		"cft_url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The CFT URL.",
			Computed:    true,
		},
		"update_token": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The update token.",
		},
		"onboarding_status": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The Account onboarding status",
		},
		"external_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The external ID of the account",
		},
		"service_account_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The account ID of cloud NGFW service",
		},
		"sns_topic_arn": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The SNS topic ARN",
		},
		"trusted_account": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The trusted account ID",
		},
		"origin": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Origin of account onboarding",
		},
	}

	for _, rmKey := range rmKeys {
		delete(ans, rmKey)
	}
	return ans
}
