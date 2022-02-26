package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go"
	"github.com/paloaltonetworks/cloud-ngfw-aws-go/object/fqdn"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Data source.
func dataSourceFqdnList() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for retrieving fqdn list information.",

		ReadContext: readFqdnListDataSource,

		Schema: fqdnListSchema(false, nil),
	}
}

func readFqdnListDataSource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := fqdn.NewClient(meta.(*awsngfw.Client))

	style := d.Get(ConfigTypeName).(string)
	d.Set(ConfigTypeName, style)

	stack := d.Get(RulestackName).(string)
	name := d.Get("name").(string)

	id := configTypeId(style, buildFqdnListId(stack, name))

	req := fqdn.ReadInput{
		Rulestack: stack,
		Name:      name,
	}
	switch style {
	case CandidateConfig:
		req.Candidate = true
	case RunningConfig:
		req.Running = true
	}

	tflog.Info(
		ctx, "read fqdn list",
		"ds", true,
		ConfigTypeName, style,
		RulestackName, req.Rulestack,
		"name", req.Name,
	)

	res, err := svc.Read(ctx, req)
	if err != nil {
		if isObjectNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(id)

	var info *fqdn.Info
	switch style {
	case CandidateConfig:
		info = res.Response.Candidate
	case RunningConfig:
		info = res.Response.Running
	}
	saveFqdnList(d, stack, name, *info)

	return nil
}

// Resource.
func resourceFqdnList() *schema.Resource {
	return &schema.Resource{
		Description: "Resource for fqdn list manipulation.",

		CreateContext: createFqdnList,
		ReadContext:   readFqdnList,
		UpdateContext: updateFqdnList,
		DeleteContext: deleteFqdnList,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: fqdnListSchema(true, []string{ConfigTypeName}),
	}
}

func createFqdnList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := fqdn.NewClient(meta.(*awsngfw.Client))
	o := loadFqdnList(d)
	tflog.Info(
		ctx, "create fqdn list",
		RulestackName, o.Rulestack,
		"name", o.Name,
	)

	if err := svc.Create(ctx, o); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(buildFqdnListId(o.Rulestack, o.Name))

	return readFqdnList(ctx, d, meta)
}

func readFqdnList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := fqdn.NewClient(meta.(*awsngfw.Client))
	stack, name, err := parseFqdnListId(d.Id())
	if err != nil {
		return diag.Errorf("Error in parsing ID %q: %s", d.Id(), err)
	}

	req := fqdn.ReadInput{
		Rulestack: stack,
		Name:      name,
		Candidate: true,
	}
	tflog.Info(
		ctx, "read fqdn list",
		RulestackName, req.Rulestack,
		"name", name,
	)

	res, err := svc.Read(ctx, req)
	if err != nil {
		if isObjectNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	saveFqdnList(d, stack, name, *res.Response.Candidate)

	return nil
}

func updateFqdnList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := fqdn.NewClient(meta.(*awsngfw.Client))
	o := loadFqdnList(d)
	tflog.Info(
		ctx, "update fqdn list",
		RulestackName, o.Rulestack,
		"name", o.Name,
	)

	if err := svc.Update(ctx, o); err != nil {
		return diag.FromErr(err)
	}

	return readFqdnList(ctx, d, meta)
}

func deleteFqdnList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := fqdn.NewClient(meta.(*awsngfw.Client))
	stack, name, err := parseFqdnListId(d.Id())
	if err != nil {
		return diag.Errorf("Error in parsing ID %q: %s", d.Id(), err)
	}

	tflog.Info(
		ctx, "delete fqdn list",
		RulestackName, stack,
		"name", name,
	)

	if err := svc.Delete(ctx, stack, name); err != nil && !isObjectNotFound(err) {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

// Schema handling.
func fqdnListSchema(isResource bool, rmKeys []string) map[string]*schema.Schema {
	ans := map[string]*schema.Schema{
		ConfigTypeName: configTypeSchema(),
		RulestackName:  rsSchema(),
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name.",
			ForceNew:    true,
		},
		"description": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The description.",
		},
		"fqdn_list": {
			Type:        schema.TypeSet,
			Required:    true,
			Description: "The fqdn list.",
			MinItems:    1,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"audit_comment": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The audit comment.",
		},
		"update_token": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The update token.",
		},
	}

	for _, rmKey := range rmKeys {
		delete(ans, rmKey)
	}

	if !isResource {
		computed(ans, "", []string{ConfigTypeName, RulestackName, "name"})
	}

	return ans
}

func loadFqdnList(d *schema.ResourceData) fqdn.Info {
	return fqdn.Info{
		Rulestack:    d.Get(RulestackName).(string),
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
		FqdnList:     setToSlice(d.Get("fqdn_list")),
		AuditComment: d.Get("audit_comment").(string),
	}
}

func saveFqdnList(d *schema.ResourceData, stack, name string, o fqdn.Info) {
	d.Set(RulestackName, stack)
	d.Set("name", name)
	d.Set("description", o.Description)
	d.Set("fqdn_list", sliceToSet(o.FqdnList))
	d.Set("audit_comment", o.AuditComment)
	d.Set("update_token", o.UpdateToken)
}

// Id functions.
func buildFqdnListId(a, b string) string {
	return strings.Join([]string{a, b}, IdSeparator)
}

func parseFqdnListId(v string) (string, string, error) {
	tok := strings.Split(v, IdSeparator)
	if len(tok) != 2 {
		return "", "", fmt.Errorf("Expecting 2 tokens, got %d", len(tok))
	}

	return tok[0], tok[1], nil
}
