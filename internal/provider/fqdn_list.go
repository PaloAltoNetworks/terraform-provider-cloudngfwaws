package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go/api"
	"github.com/paloaltonetworks/cloud-ngfw-aws-go/api/fqdn"

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
	svc := meta.(*api.ApiClient)

	style := d.Get(ConfigTypeName).(string)
	d.Set(ConfigTypeName, style)

	stack := d.Get(RulestackName).(string)
	name := d.Get("name").(string)

	scope := d.Get(ScopeName).(string)
	d.Set(ScopeName, scope)

	id := configTypeId(style, buildFqdnListId(scope, stack, name))

	req := fqdn.ReadInput{
		Rulestack: stack,
		Scope:     scope,
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
		map[string]interface{}{
			"ds":           true,
			ConfigTypeName: style,
			RulestackName:  req.Rulestack,
			ScopeName:      scope,
			"name":         req.Name,
		},
	)

	res, err := svc.ReadFqdn(ctx, req)
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
	svc := meta.(*api.ApiClient)
	o := loadFqdnList(d)
	tflog.Info(
		ctx, "create fqdn list",
		map[string]interface{}{
			RulestackName: o.Rulestack,
			"name":        o.Name,
			ScopeName:     o.Scope,
		},
	)

	if err := svc.CreateFqdn(ctx, o); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(buildFqdnListId(o.Scope, o.Rulestack, o.Name))

	return readFqdnList(ctx, d, meta)
}

func readFqdnList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)
	scope, stack, name, err := parseFqdnListId(d.Id())
	if err != nil {
		return diag.Errorf("Error in parsing ID %q: %s", d.Id(), err)
	}

	req := fqdn.ReadInput{
		Rulestack: stack,
		Name:      name,
		Candidate: true,
		Scope:     scope,
	}
	tflog.Info(
		ctx, "read fqdn list",
		map[string]interface{}{
			RulestackName: req.Rulestack,
			ScopeName:     scope,
			"name":        name,
		},
	)

	res, err := svc.ReadFqdn(ctx, req)
	if err != nil {
		if isObjectNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.Set(ScopeName, scope)
	saveFqdnList(d, stack, name, *res.Response.Candidate)

	return nil
}

func updateFqdnList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)
	o := loadFqdnList(d)
	tflog.Info(
		ctx, "update fqdn list",
		map[string]interface{}{
			RulestackName: o.Rulestack,
			ScopeName:     o.Scope,
			"name":        o.Name,
		},
	)

	if err := svc.UpdateFqdn(ctx, o); err != nil {
		return diag.FromErr(err)
	}

	return readFqdnList(ctx, d, meta)
}

func deleteFqdnList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)
	scope, stack, name, err := parseFqdnListId(d.Id())
	if err != nil {
		return diag.Errorf("Error in parsing ID %q: %s", d.Id(), err)
	}

	tflog.Info(
		ctx, "delete fqdn list",
		map[string]interface{}{
			RulestackName: stack,
			ScopeName:     scope,
			"name":        name,
		},
	)

	input := fqdn.DeleteInput{
		Rulestack: stack,
		Scope:     scope,
		Name:      name,
	}
	if err := svc.DeleteFqdn(ctx, input); err != nil && !isObjectNotFound(err) {
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
		ScopeName:      scopeSchema(),
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
		computed(ans, "", []string{ConfigTypeName, RulestackName, ScopeName, "name"})
	}

	return ans
}

func loadFqdnList(d *schema.ResourceData) fqdn.Info {
	return fqdn.Info{
		Rulestack:    d.Get(RulestackName).(string),
		Scope:        d.Get(ScopeName).(string),
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
func buildFqdnListId(a, b, c string) string {
	return strings.Join([]string{a, b, c}, IdSeparator)
}

func parseFqdnListId(v string) (string, string, string, error) {
	tok := strings.Split(v, IdSeparator)
	if len(tok) != 3 {
		return "", "", "", fmt.Errorf("Expecting 2 tokens, got %d", len(tok))
	}

	return tok[0], tok[1], tok[2], nil
}
