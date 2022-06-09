package provider

import (
	"context"
	"strings"
	"time"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go"
	"github.com/paloaltonetworks/cloud-ngfw-aws-go/rule/stack"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Data source.
func dataSourceValidateRulestack() *schema.Resource {
	return &schema.Resource{
		Description: "Data source to validate the rulestack config.",

		ReadContext: readValidateRulestack,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			RulestackName: rsSchema(),
			"state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The rulestack state.",
			},
			ScopeName: scopeSchema(),
			"commit_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The commit status.",
			},
			"validation_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The validation status.",
			},
			"commit_errors": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Commit error messages.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"validation_errors": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Validation error messages.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func readValidateRulestack(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var err error
	var ans stack.CommitStatus
	pending := "Pending"

	svc := stack.NewClient(meta.(*awsngfw.Client))
	name := d.Get(RulestackName).(string)
	scope := d.Get(ScopeName).(string)

	id := strings.Join([]string{scope, name}, IdSeparator)

	req := stack.ReadInput{
		Scope: scope,
		Name:  name,
	}

	tflog.Info(
		ctx, "read rulestack",
		RulestackName, name,
		"for_validation", true,
	)

	res, err := svc.Read(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Info(
		ctx, "validate rulestack",
		RulestackName, name,
	)

	// Perform the validation.
	vin := stack.SimpleInput{
		Name:  name,
		Scope: scope,
	}
	if err = svc.Validate(ctx, vin); err != nil {
		return diag.FromErr(err)
	}

	// Wait until the status is not Pending.
	for {
		tflog.Info(
			ctx, "getting validation status",
			RulestackName, name,
		)

		ans, err = svc.CommitStatus(ctx, vin)
		if err != nil {
			return diag.FromErr(err)
		}

		if ans.Response.ValidationStatus != pending {
			break
		}

		time.Sleep(1 * time.Second)
	}

	d.SetId(id)
	d.Set(RulestackName, name)
	d.Set("state", res.Response.State)
	d.Set(ScopeName, scope)
	d.Set("commit_status", ans.Response.CommitStatus)
	d.Set("validation_status", ans.Response.ValidationStatus)
	d.Set("commit_errors", ans.Response.CommitMessages)
	d.Set("validation_errors", ans.Response.ValidationMessages)

	return nil
}
