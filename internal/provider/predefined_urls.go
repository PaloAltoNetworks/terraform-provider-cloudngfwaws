package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go/v2/api"
	url "github.com/paloaltonetworks/cloud-ngfw-aws-go/v2/api/predefinedurl"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Data source (list predefined url categories).
func dataSourcePredefinedUrlCategories() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for retrieving the predefined URL categories.",

		ReadContext: readPredefinedUrlCategories,

		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Pagination token.",
			},
			"max_results": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Max results.",
				Default:     100,
			},
			"next_token": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Next pagination token.",
			},
			"categories": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of predefined URL categories.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func readPredefinedUrlCategories(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)

	input := url.ListInput{
		NextToken:  d.Get("token").(string),
		MaxResults: d.Get("max_results").(int),
	}

	d.Set("token", input.NextToken)
	d.Set("max_results", input.MaxResults)

	tflog.Info(
		ctx, "read predefined url categories",
		map[string]interface{}{
			"ds":          true,
			"token":       input.NextToken,
			"max_results": input.MaxResults,
		},
	)

	ans, err := svc.ListUrlPredefinedCategories(ctx, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strings.Join(
		[]string{input.NextToken, strconv.Itoa(input.MaxResults)}, IdSeparator,
	))
	d.Set("next_token", ans.Response.NextToken)
	d.Set("categories", ans.Response.Categories)

	return nil
}

// Data source (predefined url category override).
func dataSourcePredefinedUrlCategoryOverride() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for retrieving a predefined URL category override.",

		ReadContext: readDataSourcePredefinedUrlCategoryOverride,

		Schema: predefinedUrlCategoryOverrideSchema(false, nil),
	}
}

func readDataSourcePredefinedUrlCategoryOverride(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)

	input := url.GetOverrideInput{
		Rulestack: d.Get(RulestackName).(string),
		Name:      d.Get("name").(string),
	}
	style := d.Get(ConfigTypeName).(string)
	switch style {
	case CandidateConfig:
		input.Candidate = true
	case RunningConfig:
		input.Running = true
	}

	d.Set(RulestackName, input.Rulestack)
	d.Set(ConfigTypeName, style)
	id := configTypeId(style, buildPredefinedUrlCategoryOverrideId(input.Rulestack, input.Name))

	tflog.Info(
		ctx, "read predefined url category override",
		map[string]interface{}{
			"ds":           true,
			RulestackName:  input.Rulestack,
			"name":         input.Name,
			ConfigTypeName: style,
		},
	)

	ans, err := svc.DescribeUrlCategoryActionOverride(ctx, input)
	if err != nil {
		if isObjectNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(id)

	var info url.OverrideDetails
	switch style {
	case CandidateConfig:
		info = ans.Response.Candidate
	case RunningConfig:
		info = ans.Response.Running
	}
	savePredefinedUrlCategoryOverride(d, input.Name, info)

	return nil
}

// Resource (predefined url category override).
func resourcePredefinedUrlCategoryOverride() *schema.Resource {
	return &schema.Resource{
		Description: "Resource for predefined URL category override management.",

		CreateContext: createUpdatePredefinedUrlCategoryOverride,
		ReadContext:   readPredefinedUrlCategoryOverride,
		UpdateContext: createUpdatePredefinedUrlCategoryOverride,
		DeleteContext: deletePredefinedUrlCategoryOverride,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: predefinedUrlCategoryOverrideSchema(true, []string{ConfigTypeName}),
	}
}

func createUpdatePredefinedUrlCategoryOverride(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)

	input := url.OverrideInput{
		Rulestack:    d.Get(RulestackName).(string),
		Name:         d.Get("name").(string),
		Action:       d.Get("action").(string),
		AuditComment: d.Get("audit_comment").(string),
	}

	id := buildPredefinedUrlCategoryOverrideId(input.Rulestack, input.Name)

	tflog.Info(
		ctx, "modify predefined url category override",
		map[string]interface{}{
			RulestackName:   input.Rulestack,
			"name":          input.Name,
			"action":        input.Action,
			"audit_comment": input.AuditComment,
		},
	)

	if err := svc.UpdateUrlCategoryActionOverride(ctx, input); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)
	return readPredefinedUrlCategoryOverride(ctx, d, meta)
}

func readPredefinedUrlCategoryOverride(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)

	stack, name, err := parsePredefinedUrlCategoryOverrideId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := url.GetOverrideInput{
		Rulestack: stack,
		Name:      name,
		Candidate: true,
	}

	d.Set(RulestackName, stack)

	tflog.Info(
		ctx, "read predefined url category override",
		map[string]interface{}{
			RulestackName: stack,
			"name":        name,
		},
	)

	ans, err := svc.DescribeUrlCategoryActionOverride(ctx, input)
	if err != nil {
		if isObjectNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	savePredefinedUrlCategoryOverride(d, name, ans.Response.Candidate)
	return nil
}

func deletePredefinedUrlCategoryOverride(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)

	stack, name, err := parsePredefinedUrlCategoryOverrideId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := url.OverrideInput{
		Rulestack: stack,
		Name:      name,
		Action:    "none",
	}

	if err := svc.UpdateUrlCategoryActionOverride(ctx, input); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

// Schema handling.
func predefinedUrlCategoryOverrideSchema(isResource bool, rmKeys []string) map[string]*schema.Schema {
	ao := []string{"none", "allow", "alert", "block"}

	ans := map[string]*schema.Schema{
		ConfigTypeName: configTypeSchema(),
		RulestackName:  rsSchema(),
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name.",
		},
		"action": {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  addStringInSliceValidation("The action to take.", ao),
			Default:      ao[0],
			ValidateFunc: validation.StringInSlice(ao, false),
		},
		"audit_comment": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The audit comment.",
		},
		"update_token": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Update token.",
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

func savePredefinedUrlCategoryOverride(d *schema.ResourceData, name string, o url.OverrideDetails) {
	d.Set("name", name)
	d.Set("action", o.Action)
	d.Set("audit_comment", o.AuditComment)
	d.Set("update_token", o.UpdateToken)
}

// Id functions.
func buildPredefinedUrlCategoryOverrideId(a, b string) string {
	return strings.Join([]string{a, b}, IdSeparator)
}

func parsePredefinedUrlCategoryOverrideId(v string) (string, string, error) {
	tok := strings.Split(v, IdSeparator)
	if len(tok) != 2 {
		return "", "", fmt.Errorf("Expecting 2 tokens, got %d", len(tok))
	}

	return tok[0], tok[1], nil
}
