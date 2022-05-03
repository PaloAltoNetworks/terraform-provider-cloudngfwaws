package provider

import (
	"context"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go"
	"github.com/paloaltonetworks/cloud-ngfw-aws-go/tag"
	"github.com/paloaltonetworks/cloud-ngfw-aws-go/tag/rulestack"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Data source.
func dataSourceRulestackTag() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for retrieving rulestack tag information.",

		ReadContext: readRulestackTagDataSource,

		Schema: rulestackTagSchema(false, nil),
	}
}

func readRulestackTagDataSource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := rulestack.NewClient(meta.(*awsngfw.Client))

	req := rulestack.ListInput{
		Rulestack:  d.Get(RulestackName).(string),
		MaxResults: 1000,
	}

	id := req.Rulestack

	tflog.Info(
		ctx, "read rulestack tags",
		"ds", true,
		RulestackName, req.Rulestack,
	)

	res, err := svc.List(ctx, req)
	if err != nil {
		if isObjectNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(id)

	saveRulestackTag(d, req.Rulestack, res.Response.Tags)

	return nil
}

// Resource.
func resourceRulestackTag() *schema.Resource {
	return &schema.Resource{
		Description: "Resource for rulestack tag manipulation.",

		CreateContext: createUpdateRulestackTag,
		ReadContext:   readRulestackTag,
		UpdateContext: createUpdateRulestackTag,
		DeleteContext: deleteRulestackTag,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: rulestackTagSchema(true, nil),
	}
}

func createUpdateRulestackTag(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := rulestack.NewClient(meta.(*awsngfw.Client))
	o := loadRulestackTag(d)
	tflog.Info(
		ctx, "modify rulestack tags",
		RulestackName, o.Rulestack,
	)

	if err := svc.Apply(ctx, o); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(o.Rulestack)

	return readRulestackTag(ctx, d, meta)
}

func readRulestackTag(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := rulestack.NewClient(meta.(*awsngfw.Client))

	rs := d.Id()

	req := rulestack.ListInput{
		Rulestack:  rs,
		MaxResults: 100,
	}

	tflog.Info(
		ctx, "read rulestack tags",
		RulestackName, req.Rulestack,
	)

	res, err := svc.List(ctx, req)
	if err != nil {
		if isObjectNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	saveRulestackTag(d, req.Rulestack, res.Response.Tags)

	return nil
}

func deleteRulestackTag(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := rulestack.NewClient(meta.(*awsngfw.Client))

	rs := d.Id()

	tflog.Info(
		ctx, "delete rulestack tags",
		RulestackName, rs,
	)

	if err := svc.Apply(ctx, rulestack.Info{Rulestack: rs}); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

// Schema handling.
func rulestackTagSchema(isResource bool, rmKeys []string) map[string]*schema.Schema {
	ans := map[string]*schema.Schema{
		RulestackName: rsSchema(),
		"tags":        tagsSchema(true, false),
	}

	for _, rmKey := range rmKeys {
		delete(ans, rmKey)
	}

	if !isResource {
		computed(ans, "", []string{RulestackName})
	}

	return ans
}

func loadRulestackTag(d *schema.ResourceData) rulestack.Info {
	return rulestack.Info{
		Rulestack: d.Get(RulestackName).(string),
		Tags:      loadTags(d.Get("tags")),
	}
}

func saveRulestackTag(d *schema.ResourceData, rs string, o []tag.Details) {
	d.Set(RulestackName, rs)
	d.Set("tags", dumpTags(o))
}
