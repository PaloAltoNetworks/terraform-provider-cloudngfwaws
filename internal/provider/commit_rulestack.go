package provider

import (
	"context"
	"time"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go"
	"github.com/paloaltonetworks/cloud-ngfw-aws-go/rule/stack"

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
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
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
	svc := stack.NewClient(meta.(*awsngfw.Client))
	name := d.Get(RulestackName).(string)
	scope := d.Get(ScopeName).(string)
	pending := "Pending"
	input := stack.SimpleInput{
		Name:  name,
		Scope: scope,
	}

	d.Set(ScopeName, scope)

	tflog.Info(
		ctx, "commit rulestack",
		RulestackName, name,
		ScopeName, scope,
	)

	// Perform the commit.
	if err := svc.Commit(ctx, input); err != nil {
		return diag.FromErr(err)
	}

	// Wait until the status is not Pending.
	for {
		tflog.Info(
			ctx, "getting commit status",
			RulestackName, name,
			ScopeName, scope,
		)

		ans, err := svc.CommitStatus(ctx, input)
		if err != nil {
			return diag.FromErr(err)
		}

		if ans.Response.CommitStatus != pending {
			break
		}

		time.Sleep(1 * time.Second)
	}

	d.SetId(buildRulestackId(scope, name))

	return readCommitRulestack(ctx, d, meta)
}

func readCommitRulestack(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := stack.NewClient(meta.(*awsngfw.Client))
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
		RulestackName, name,
		"for_commit", true,
		ScopeName, scope,
	)

	res, err := svc.Read(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	cs, err := svc.CommitStatus(ctx, stack.SimpleInput{Name: name, Scope: scope})
	if err != nil {
		return diag.FromErr(err)
	}

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
