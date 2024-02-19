package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go/api"
	"github.com/paloaltonetworks/cloud-ngfw-aws-go/api/feed"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Data source.
func dataSourceIntelligentFeed() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for retrieving intelligent feed information.",

		ReadContext: readIntelligentFeedDataSource,

		Schema: intelligentFeedSchema(false, nil),
	}
}

func readIntelligentFeedDataSource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)

	style := d.Get(ConfigTypeName).(string)
	d.Set(ConfigTypeName, style)

	scope := d.Get(ScopeName).(string)
	d.Set(ScopeName, scope)

	stack := d.Get(RulestackName).(string)
	name := d.Get("name").(string)

	id := configTypeId(style, buildIntelligentFeedId(scope, stack, name))

	req := feed.ReadInput{
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
		ctx, "read intelligent feed",
		map[string]interface{}{
			"ds":           true,
			ConfigTypeName: style,
			RulestackName:  req.Rulestack,
			ScopeName:      scope,
			"name":         req.Name,
		},
	)

	res, err := svc.ReadFeed(ctx, req)
	if err != nil {
		if isObjectNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(id)

	var info *feed.Info
	switch style {
	case CandidateConfig:
		info = res.Response.Candidate
	case RunningConfig:
		info = res.Response.Running
	}
	saveIntelligentFeed(d, stack, name, *info)

	return nil
}

// Resource.
func resourceIntelligentFeed() *schema.Resource {
	return &schema.Resource{
		Description: "Resource for intelligent feed manipulation.",

		CreateContext: createIntelligentFeed,
		ReadContext:   readIntelligentFeed,
		UpdateContext: updateIntelligentFeed,
		DeleteContext: deleteIntelligentFeed,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: intelligentFeedSchema(true, []string{ConfigTypeName}),
	}
}

func createIntelligentFeed(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)
	o := loadIntelligentFeed(d)
	tflog.Info(
		ctx, "create intelligent feed",
		map[string]interface{}{
			RulestackName: o.Rulestack,
			"name":        o.Name,
			ScopeName:     o.Scope,
		},
	)

	if err := svc.CreateFeed(ctx, o); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(buildIntelligentFeedId(o.Scope, o.Rulestack, o.Name))

	return readIntelligentFeed(ctx, d, meta)
}

func readIntelligentFeed(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)
	scope, stack, name, err := parseIntelligentFeedId(d.Id())
	if err != nil {
		return diag.Errorf("Error in parsing ID %q: %s", d.Id(), err)
	}

	req := feed.ReadInput{
		Rulestack: stack,
		Scope:     scope,
		Name:      name,
		Candidate: true,
	}
	tflog.Info(
		ctx, "read intelligent feed",
		map[string]interface{}{
			RulestackName: req.Rulestack,
			"name":        name,
			ScopeName:     scope,
		},
	)

	res, err := svc.ReadFeed(ctx, req)
	if err != nil {
		if isObjectNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.Set(ScopeName, scope)
	saveIntelligentFeed(d, stack, name, *res.Response.Candidate)

	return nil
}

func updateIntelligentFeed(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)
	o := loadIntelligentFeed(d)
	tflog.Info(
		ctx, "update intelligent feed",
		map[string]interface{}{
			RulestackName: o.Rulestack,
			"name":        o.Name,
			ScopeName:     o.Scope,
		},
	)

	if err := svc.UpdateFeed(ctx, o); err != nil {
		return diag.FromErr(err)
	}

	return readIntelligentFeed(ctx, d, meta)
}

func deleteIntelligentFeed(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)
	scope, stack, name, err := parseIntelligentFeedId(d.Id())
	if err != nil {
		return diag.Errorf("Error in parsing ID %q: %s", d.Id(), err)
	}

	tflog.Info(
		ctx, "delete intelligent feed",
		map[string]interface{}{
			RulestackName: stack,
			"name":        name,
			ScopeName:     scope,
		},
	)

	input := feed.DeleteInput{
		Rulestack: stack,
		Scope:     scope,
		Name:      name,
	}
	if err := svc.DeleteFeed(ctx, input); err != nil && !isObjectNotFound(err) {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

// Schema handling.
func intelligentFeedSchema(isResource bool, rmKeys []string) map[string]*schema.Schema {
	type_values := []string{"IP_LIST", "URL_LIST"}
	frequency_values := []string{"HOURLY", "DAILY"}
	time_low := 0
	time_high := 23

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
		"certificate": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The certificate profile.",
		},
		"url": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The intelligent feed source.",
		},
		"type": {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  addStringInSliceValidation("The intelligent feed type.", type_values),
			Default:      "IP_LIST",
			ValidateFunc: validation.StringInSlice(type_values, false),
		},
		"frequency": {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  addStringInSliceValidation("Update frequency.", frequency_values),
			Default:      "HOURLY",
			ValidateFunc: validation.StringInSlice(frequency_values, false),
		},
		"time": {
			Type:         schema.TypeInt,
			Optional:     true,
			Description:  addIntBetweenValidation("The time to poll for updates if frequency is daily.", time_low, time_high),
			ValidateFunc: validation.IntBetween(time_low, time_high),
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

func loadIntelligentFeed(d *schema.ResourceData) feed.Info {
	return feed.Info{
		Rulestack:    d.Get(RulestackName).(string),
		Scope:        d.Get(ScopeName).(string),
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
		Certificate:  d.Get("certificate").(string),
		Url:          d.Get("url").(string),
		Type:         d.Get("type").(string),
		Frequency:    d.Get("frequency").(string),
		Time:         d.Get("time").(int),
		AuditComment: d.Get("audit_comment").(string),
	}
}

func saveIntelligentFeed(d *schema.ResourceData, stack, name string, o feed.Info) {
	d.Set(RulestackName, stack)
	d.Set("name", name)
	d.Set("description", o.Description)
	d.Set("certificate", o.Certificate)
	d.Set("url", o.Url)
	d.Set("type", o.Type)
	d.Set("frequency", o.Frequency)
	d.Set("time", o.Time)
	d.Set("audit_comment", o.AuditComment)
	d.Set("update_token", o.UpdateToken)
}

// Id functions.
func buildIntelligentFeedId(a, b, c string) string {
	return strings.Join([]string{a, b, c}, IdSeparator)
}

func parseIntelligentFeedId(v string) (string, string, string, error) {
	tok := strings.Split(v, IdSeparator)
	if len(tok) != 3 {
		return "", "", "", fmt.Errorf("Expecting 2 tokens, got %d", len(tok))
	}

	return tok[0], tok[1], tok[2], nil
}
