package provider

import (
	"context"
	"time"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go/v2/api"
	"github.com/paloaltonetworks/cloud-ngfw-aws-go/v2/api/stack"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Resource.
func resourceCommitRulestack() *schema.Resource {
	s := "Running"

	return &schema.Resource{
		Description: "Resource for committing the rulestack config.",

		CreateContext: createUpdateCommitRulestack,
		ReadContext:   readCommitRulestack,
		UpdateContext: createUpdateCommitRulestack,
		DeleteContext: deleteCommitRulestack,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Read:   schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			RulestackName: rsSchema(),
			ScopeName:     scopeSchema(),
			"state": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The rulestack state. This can only be the default value.",
				Default:      s,
				ValidateFunc: validation.StringInSlice([]string{s}, false),
			},
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

func createUpdateCommitRulestack(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)
	name := d.Get(RulestackName).(string)
	scope := d.Get(ScopeName).(string)
	input := stack.SimpleInput{
		Name:  name,
		Scope: scope,
	}

	tflog.Info(
		ctx, "commit rulestack",
		map[string]interface{}{
			RulestackName: name,
			ScopeName:     scope,
		},
	)

	// Perform the commit.
	if err := svc.CommitRuleStack(ctx, input); err != nil {
		return diag.FromErr(err)
	}

	// Wait until the status is not pending.
	if _, err := svc.PollCommitRulestack(ctx, input); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(buildRulestackId(scope, name))

	return readCommitRulestack(ctx, d, meta)
}

func readCommitRulestack(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)
	scope, name, err := parseRulestackId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	req := stack.ReadInput{
		Name:  name,
		Scope: scope,
	}
	tflog.Info(
		ctx, "read rulestack",
		map[string]interface{}{
			RulestackName: name,
			"for_commit":  true,
			ScopeName:     scope,
		},
	)

	res, err := svc.ReadRuleStack(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Info(
		ctx, "read rulestack commit status",
		map[string]interface{}{
			RulestackName: name,
			ScopeName:     scope,
			"for_commit":  true,
		},
	)
	cs, err := svc.CommitStatusRuleStack(ctx, stack.SimpleInput{Name: name, Scope: scope})
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set(ScopeName, scope)
	d.Set(RulestackName, name)
	d.Set("state", res.Response.State)
	d.Set("commit_status", cs.Response.CommitStatus)
	d.Set("validation_status", cs.Response.ValidationStatus)
	d.Set("commit_errors", cs.Response.CommitMessages)
	d.Set("validation_errors", cs.Response.ValidationMessages)

	return nil
}

func deleteCommitRulestack(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}
