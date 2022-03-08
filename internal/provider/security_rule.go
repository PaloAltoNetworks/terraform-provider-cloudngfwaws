package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go"
	"github.com/paloaltonetworks/cloud-ngfw-aws-go/rule/security"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Data source.
func dataSourceSecurityRule() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for retrieving security rule information.",

		ReadContext: readSecurityRuleDataSource,

		Schema: securityRuleSchema(false, nil),
	}
}

func readSecurityRuleDataSource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := security.NewClient(meta.(*awsngfw.Client))

	style := d.Get(ConfigTypeName).(string)
	d.Set(ConfigTypeName, style)

	stack := d.Get(RulestackName).(string)
	rlist := d.Get(RuleListName).(string)
	priority := d.Get("priority").(int)

	id := configTypeId(style, buildSecurityRuleId(stack, rlist, priority))

	req := security.ReadInput{
		Rulestack: stack,
		RuleList:  rlist,
		Priority:  priority,
	}
	switch style {
	case CandidateConfig:
		req.Candidate = true
	case RunningConfig:
		req.Running = true
	}

	tflog.Info(
		ctx, "read security rule",
		"ds", true,
		ConfigTypeName, style,
		RulestackName, req.Rulestack,
		RuleListName, req.RuleList,
		"priority", req.Priority,
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

	var info *security.Details
	switch style {
	case CandidateConfig:
		info = res.Response.Candidate
	case RunningConfig:
		info = res.Response.Running
	}

	saveSecurityRule(d, stack, rlist, priority, *info)

	return nil
}

// Resource.
func resourceSecurityRule() *schema.Resource {
	return &schema.Resource{
		Description: "Resource for security rule manipulation.",

		CreateContext: createSecurityRule,
		ReadContext:   readSecurityRule,
		UpdateContext: updateSecurityRule,
		DeleteContext: deleteSecurityRule,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: securityRuleSchema(true, []string{ConfigTypeName}),
	}
}

func createSecurityRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := security.NewClient(meta.(*awsngfw.Client))
	o := loadSecurityRule(d)
	tflog.Info(
		ctx, "create security rule",
		RulestackName, o.Rulestack,
		RuleListName, o.RuleList,
		"priority", o.Priority,
		"name", o.Entry.Name,
	)

	if err := svc.Create(ctx, o); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(buildSecurityRuleId(o.Rulestack, o.RuleList, o.Priority))

	return readSecurityRule(ctx, d, meta)
}

func readSecurityRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := security.NewClient(meta.(*awsngfw.Client))
	stack, rlist, priority, err := parseSecurityRuleId(d.Id())
	if err != nil {
		return diag.Errorf("Error in parsing ID %q: %s", d.Id(), err)
	}

	req := security.ReadInput{
		Rulestack: stack,
		RuleList:  rlist,
		Priority:  priority,
		Candidate: true,
	}
	tflog.Info(
		ctx, "read security rule",
		RulestackName, req.Rulestack,
		RuleListName, req.RuleList,
		"priority", req.Priority,
	)

	res, err := svc.Read(ctx, req)
	if err != nil {
		if isObjectNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	saveSecurityRule(d, stack, rlist, priority, *res.Response.Candidate)

	return nil
}

func updateSecurityRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := security.NewClient(meta.(*awsngfw.Client))
	o := loadSecurityRule(d)
	tflog.Info(
		ctx, "update security rule",
		RulestackName, o.Rulestack,
		RuleListName, o.RuleList,
		"priority", o.Priority,
	)

	if err := svc.Update(ctx, o); err != nil {
		return diag.FromErr(err)
	}

	return readSecurityRule(ctx, d, meta)
}

func deleteSecurityRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := security.NewClient(meta.(*awsngfw.Client))
	stack, rlist, priority, err := parseSecurityRuleId(d.Id())
	if err != nil {
		return diag.Errorf("Error in parsing ID %q: %s", d.Id(), err)
	}

	tflog.Info(
		ctx, "delete security rule",
		RulestackName, stack,
		RuleListName, rlist,
		"priority", priority,
	)

	if err := svc.Delete(ctx, stack, rlist, priority); err != nil && !isObjectNotFound(err) {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

// Schema handling.
func securityRuleSchema(isResource bool, rmKeys []string) map[string]*schema.Schema {
	action_values := []string{"Allow", "DenySilent", "DenyResetServer", "DenyResetBoth"}
	decryption_values := []string{"", "SSLOutboundInspection"}

	ans := map[string]*schema.Schema{
		ConfigTypeName: configTypeSchema(),
		RulestackName:  rsSchema(),
		RuleListName:   ruleListSchema(),
		"priority": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "The rule priority.",
			ForceNew:    true,
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name.",
		},
		"description": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The description.",
		},
		"enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Set to false to disable this rule.",
			Default:     true,
		},
		"source": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "The source spec.",
			MinItems:    1,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"cidrs": {
						Type:        schema.TypeSet,
						Optional:    true,
						Description: "List of CIDRs.",
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"countries": {
						Type:        schema.TypeSet,
						Optional:    true,
						Description: "List of countries.",
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"feeds": {
						Type:        schema.TypeSet,
						Optional:    true,
						Description: "List of feeds.",
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"prefix_lists": {
						Type:        schema.TypeSet,
						Optional:    true,
						Description: "List of prefix list.",
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				},
			},
		},
		"negate_source": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Negate the source definition.",
		},
		"destination": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "The destination spec.",
			MinItems:    1,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"cidrs": {
						Type:        schema.TypeSet,
						Optional:    true,
						Description: "List of CIDRs.",
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"countries": {
						Type:        schema.TypeSet,
						Optional:    true,
						Description: "List of countries.",
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"feeds": {
						Type:        schema.TypeSet,
						Optional:    true,
						Description: "List of feeds.",
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"prefix_lists": {
						Type:        schema.TypeSet,
						Optional:    true,
						Description: "List of prefix list.",
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"fqdn_lists": {
						Type:        schema.TypeSet,
						Optional:    true,
						Description: "List of FQDN lists.",
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				},
			},
		},
		"negate_destination": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Negate the destination definition.",
		},
		"applications": {
			Type:        schema.TypeSet,
			Required:    true,
			Description: "The list of applications.",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"category": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "The category spec.",
			MinItems:    1,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"url_category_names": {
						Type:        schema.TypeSet,
						Optional:    true,
						Description: "List of URL category names.",
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"feeds": {
						Type:        schema.TypeSet,
						Optional:    true,
						Description: "List of feeds.",
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				},
			},
		},
		"protocol": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The protocol.",
			Default:     "application-default",
		},
		"audit_comment": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The audit comment.",
		},
		"action": {
			Type:         schema.TypeString,
			Required:     true,
			Description:  addStringInSliceValidation("The action to take.", action_values),
			ValidateFunc: validation.StringInSlice(action_values, false),
		},
		"logging": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Enable logging at end.",
			Default:     true,
		},
		"decryption_rule_type": {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  addStringInSliceValidation("Decryption rule type.", decryption_values),
			ValidateFunc: validation.StringInSlice(decryption_values, false),
		},
		TagsName: tagsSchema(false, false),
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
		computed(ans, "", []string{RulestackName, RuleListName, "priority"})
	}

	return ans
}

func loadSecurityRule(d *schema.ResourceData) security.Info {
	src := configFolder(d.Get("source"))
	dst := configFolder(d.Get("destination"))
	cat := configFolder(d.Get("category"))

	return security.Info{
		Rulestack: d.Get(RulestackName).(string),
		RuleList:  d.Get(RuleListName).(string),
		Priority:  d.Get("priority").(int),
		Entry: security.Details{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
			Enabled:     d.Get("enabled").(bool),
			Source: security.SourceDetails{
				Cidrs:       setToSlice(src["cidrs"]),
				Countries:   setToSlice(src["countries"]),
				Feeds:       setToSlice(src["feeds"]),
				PrefixLists: setToSlice(src["prefix_lists"]),
			},
			NegateSource: d.Get("negate_source").(bool),
			Destination: security.DestinationDetails{
				Cidrs:       setToSlice(dst["cidrs"]),
				Countries:   setToSlice(dst["countries"]),
				Feeds:       setToSlice(dst["feeds"]),
				PrefixLists: setToSlice(dst["prefix_lists"]),
				FqdnLists:   setToSlice(dst["fqdn_lists"]),
			},
			NegateDestination: d.Get("negate_destination").(bool),
			Applications:      setToSlice(d.Get("applications")),
			Category: security.CategoryDetails{
				UrlCategoryNames: setToSlice(cat["url_category_names"]),
				Feeds:            setToSlice(cat["feeds"]),
			},
			Protocol:           d.Get("protocol").(string),
			AuditComment:       d.Get("audit_comment").(string),
			Action:             d.Get("action").(string),
			Logging:            d.Get("logging").(bool),
			DecryptionRuleType: d.Get("decryption_rule_type").(string),
			//Tags: tlist,
			//UpdateToken: d.Get("update_token").(string),
		},
	}
}

func saveSecurityRule(d *schema.ResourceData, stack, rlist string, priority int, o security.Details) {
	src := map[string]interface{}{
		"cidrs":        sliceToSet(o.Source.Cidrs),
		"countries":    sliceToSet(o.Source.Countries),
		"feeds":        sliceToSet(o.Source.Feeds),
		"prefix_lists": sliceToSet(o.Source.PrefixLists),
	}
	dst := map[string]interface{}{
		"cidrs":        sliceToSet(o.Destination.Cidrs),
		"countries":    sliceToSet(o.Destination.Countries),
		"feeds":        sliceToSet(o.Destination.Feeds),
		"prefix_lists": sliceToSet(o.Destination.PrefixLists),
		"fqdn_lists":   sliceToSet(o.Destination.FqdnLists),
	}
	cat := map[string]interface{}{
		"url_category_names": sliceToSet(o.Category.UrlCategoryNames),
		"feeds":              sliceToSet(o.Category.Feeds),
	}

	d.Set(RulestackName, stack)
	d.Set(RuleListName, rlist)
	d.Set("priority", priority)
	d.Set("name", o.Name)
	d.Set("description", o.Description)
	d.Set("enabled", o.Enabled)
	d.Set("source", []interface{}{src})
	d.Set("negate_source", o.NegateSource)
	d.Set("destination", []interface{}{dst})
	d.Set("negate_destination", o.NegateDestination)
	d.Set("applications", sliceToSet(o.Applications))
	d.Set("category", []interface{}{cat})
	d.Set("protocol", o.Protocol)
	d.Set("audit_comment", o.AuditComment)
	d.Set("action", o.Action)
	d.Set("logging", o.Logging)
	d.Set("decryption_rule_type", o.DecryptionRuleType)
	d.Set(TagsName, dumpTags(o.Tags))
	d.Set("update_token", o.UpdateToken)
}

// Id functions.
func buildSecurityRuleId(a, b string, c int) string {
	return strings.Join([]string{a, b, strconv.Itoa(c)}, IdSeparator)
}

func parseSecurityRuleId(v string) (string, string, int, error) {
	tok := strings.Split(v, IdSeparator)
	if len(tok) != 3 {
		return "", "", 0, fmt.Errorf("Expecting 3 tokens, got %d", len(tok))
	}

	priority, err := strconv.Atoi(tok[2])
	if err != nil {
		return "", "", 0, err
	}

	return tok[0], tok[1], priority, nil
}
