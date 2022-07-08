package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go"
	"github.com/paloaltonetworks/cloud-ngfw-aws-go/rule/stack"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Data source.
func dataSourceRulestack() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for retrieving rulestack information.",

		ReadContext: readRulestackDataSource,

		Schema: rulestackSchema(false, nil),
	}
}

func readRulestackDataSource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := stack.NewClient(meta.(*awsngfw.Client))

	style := d.Get(ConfigTypeName).(string)
	d.Set(ConfigTypeName, style)

	name := d.Get("name").(string)
	scope := d.Get(ScopeName).(string)

	id := configTypeId(style, buildRulestackId(scope, name))

	req := stack.ReadInput{
		Name:  name,
		Scope: scope,
	}
	switch style {
	case CandidateConfig:
		req.Candidate = true
	case RunningConfig:
		req.Running = true
	}

	tflog.Info(
		ctx, "read rulestack",
		"ds", true,
		"name", name,
		ScopeName, scope,
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

	var info *stack.Details
	switch style {
	case CandidateConfig:
		info = res.Response.Candidate
	case RunningConfig:
		info = res.Response.Running
	}
	saveRulestack(d, name, res.Response.State, *info)

	return nil
}

// Resource.
func resourceRulestack() *schema.Resource {
	return &schema.Resource{
		Description: "Resource for rulestack manipulation.",

		CreateContext: createRulestack,
		ReadContext:   readRulestack,
		UpdateContext: updateRulestack,
		DeleteContext: deleteRulestack,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: rulestackSchema(true, []string{ConfigTypeName}),
	}
}

func createRulestack(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := stack.NewClient(meta.(*awsngfw.Client))
	o, ti := loadRulestack(d)
	tflog.Info(
		ctx, "create rulestack",
		"name", o.Name,
		ScopeName, o.Entry.Scope,
	)

	if err := svc.Create(ctx, o); err != nil {
		return diag.FromErr(err)
	}

	tflog.Info(
		ctx, "apply rulestack tags",
		"name", ti.Rulestack,
		ScopeName, o.Entry.Scope,
	)

	if err := svc.ApplyTags(ctx, ti); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(buildRulestackId(o.Entry.Scope, o.Name))

	return readRulestack(ctx, d, meta)
}

func readRulestack(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := stack.NewClient(meta.(*awsngfw.Client))

	// Verify the correct number of tokens first.
	tok := strings.Split(d.Id(), IdSeparator)
	if len(tok) == 1 {
		d.SetId(buildRulestackId(d.Get(ScopeName).(string), tok[0]))
	}

	scope, name, err := parseRulestackId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	req := stack.ReadInput{
		Name:      name,
		Scope:     scope,
		Candidate: true,
	}
	tflog.Info(
		ctx, "read rulestack",
		"name", name,
		ScopeName, scope,
	)

	res, err := svc.Read(ctx, req)
	if err != nil {
		if isObjectNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	saveRulestack(d, res.Response.Name, res.Response.State, *res.Response.Candidate)

	return nil
}

func updateRulestack(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := stack.NewClient(meta.(*awsngfw.Client))
	o, ti := loadRulestack(d)
	tflog.Info(
		ctx, "update rulestack",
		"name", o.Name,
	)

	if err := svc.Update(ctx, o); err != nil {
		return diag.FromErr(err)
	}

	if err := svc.ApplyTags(ctx, ti); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(buildRulestackId(o.Entry.Scope, o.Name))
	return readRulestack(ctx, d, meta)
}

func deleteRulestack(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := stack.NewClient(meta.(*awsngfw.Client))
	scope, name, err := parseRulestackId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Info(
		ctx, "delete rulestack",
		"name", name,
		ScopeName, scope,
	)

	input := stack.SimpleInput{Name: name, Scope: scope}
	if err := svc.Delete(ctx, input); err != nil && !isObjectNotFound(err) {
		return diag.FromErr(err)
	}

	// Deleting a rulestack that has been committed requires a commit, and the API
	// doesn't do this because reasons...
	//
	// Polling the commit is unnecessary, as is checking the response from issuing
	// the commit.
	tflog.Info(
		ctx, "commit rulestack",
		"post-delete", true,
		"name", name,
		ScopeName, scope,
	)
	_ = svc.Commit(ctx, input)

	d.SetId("")
	return nil
}

// Schema handling.
func rulestackSchema(isResource bool, rmKeys []string) map[string]*schema.Schema {
	ans := map[string]*schema.Schema{
		ConfigTypeName: configTypeSchema(),
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
		ScopeName: scopeSchema(),
		"account_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The account ID.",
		},
		"account_group": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Account group.",
		},
		"minimum_app_id_version": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Minimum App-ID version number.",
		},
		"lookup_x_forwarded_for": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Lookup x forwarded for.",
		},
		TagsName: tagsSchema(true),
		"profile_config": {
			Type:     schema.TypeList,
			Required: true,
			MaxItems: 1,
			MinItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"anti_spyware": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Anti-spyware profile setting.",
						Default:     "BestPractice",
					},
					"anti_virus": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Anti-virus profile setting.",
						Default:     "BestPractice",
					},
					"vulnerability": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Vulnerability profile setting.",
						Default:     "BestPractice",
					},
					"url_filtering": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "URL filtering profile setting.",
						Default:     "None",
					},
					"file_blocking": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "File blocking profile setting.",
						Default:     "BestPractice",
					},
					"outbound_trust_certificate": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Outbound trust certificate.",
					},
					"outbound_untrust_certificate": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Outbound untrust certificate.",
					},
				},
			},
		},
		"state": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The rulestack state.",
		},
	}

	for _, rmKey := range rmKeys {
		delete(ans, rmKey)
	}

	if !isResource {
		computed(ans, "", []string{"name", ConfigTypeName, ScopeName})
	}

	return ans
}

func loadRulestack(d *schema.ResourceData) (stack.Info, stack.AddTagsInput) {
	p := configFolder(d.Get("profile_config"))

	ti := stack.AddTagsInput{
		Rulestack: d.Get("name").(string),
		Scope:     d.Get(ScopeName).(string),
		Tags:      loadTags(d.Get(TagsName)),
	}

	return stack.Info{
		Name: d.Get("name").(string),
		Entry: stack.Details{
			Description:         d.Get("description").(string),
			Scope:               d.Get(ScopeName).(string),
			AccountId:           d.Get("account_id").(string),
			AccountGroup:        d.Get("account_group").(string),
			MinimumAppIdVersion: d.Get("minimum_app_id_version").(string),
			LookupXForwardedFor: d.Get("lookup_x_forwarded_for").(string),
			Profile: stack.ProfileConfig{
				AntiSpyware:                p["anti_spyware"].(string),
				AntiVirus:                  p["anti_virus"].(string),
				Vulnerability:              p["vulnerability"].(string),
				UrlFiltering:               p["url_filtering"].(string),
				FileBlocking:               p["file_blocking"].(string),
				OutboundTrustCertificate:   p["outbound_trust_certificate"].(string),
				OutboundUntrustCertificate: p["outbound_untrust_certificate"].(string),
			},
		},
	}, ti
}

func saveRulestack(d *schema.ResourceData, name, state string, o stack.Details) {
	pc := map[string]interface{}{
		"anti_spyware":                 o.Profile.AntiSpyware,
		"anti_virus":                   o.Profile.AntiVirus,
		"vulnerability":                o.Profile.Vulnerability,
		"url_filtering":                o.Profile.UrlFiltering,
		"file_blocking":                o.Profile.FileBlocking,
		"outbound_trust_certificate":   o.Profile.OutboundTrustCertificate,
		"outbound_untrust_certificate": o.Profile.OutboundUntrustCertificate,
	}

	d.Set("name", name)
	d.Set("description", o.Description)
	d.Set(ScopeName, o.Scope)
	d.Set("account_id", o.AccountId)
	d.Set("account_group", o.AccountGroup)
	d.Set("minimum_app_id_version", o.MinimumAppIdVersion)
	d.Set("lookup_x_forwarded_for", o.LookupXForwardedFor)
	d.Set("profile_config", []interface{}{pc})
	d.Set(TagsName, dumpTags(o.Tags))
	d.Set("state", state)
}

// Id functions.
func buildRulestackId(a, b string) string {
	return strings.Join([]string{a, b}, IdSeparator)
}

func parseRulestackId(v string) (string, string, error) {
	tok := strings.Split(v, IdSeparator)
	if len(tok) != 2 {
		return "", "", fmt.Errorf("Expected 2 tokens from ID")
	}

	return tok[0], tok[1], nil
}
