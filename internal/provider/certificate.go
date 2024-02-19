package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go/api"
	"github.com/paloaltonetworks/cloud-ngfw-aws-go/api/certificate"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Data source.
func dataSourceCertificate() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for retrieving certificate information.",

		ReadContext: readCertificateDataSource,

		Schema: certificateSchema(false, nil),
	}
}

func readCertificateDataSource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)

	style := d.Get(ConfigTypeName).(string)
	d.Set(ConfigTypeName, style)
	scope := d.Get(ScopeName).(string)
	d.Set(ScopeName, scope)

	stack := d.Get(RulestackName).(string)
	name := d.Get("name").(string)

	id := configTypeId(style, buildCertificateId(scope, stack, name))

	req := certificate.ReadInput{
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
		ctx, "read certificate",
		map[string]interface{}{
			"ds":           true,
			ConfigTypeName: style,
			RulestackName:  req.Rulestack,
			ScopeName:      scope,
			"name":         req.Name,
		},
	)

	res, err := svc.ReadCertificate(ctx, req)
	if err != nil {
		if isObjectNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(id)

	var info *certificate.Info
	switch style {
	case CandidateConfig:
		info = res.Response.Candidate
	case RunningConfig:
		info = res.Response.Running
	}
	saveCertificate(d, stack, name, *info)

	return nil
}

// Resource.
func resourceCertificate() *schema.Resource {
	return &schema.Resource{
		Description: "Resource for certificate manipulation.",

		CreateContext: createCertificate,
		ReadContext:   readCertificate,
		UpdateContext: updateCertificate,
		DeleteContext: deleteCertificate,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: certificateSchema(true, []string{ConfigTypeName}),
	}
}

func createCertificate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)
	o := loadCertificate(d)
	tflog.Info(
		ctx, "create certificate",
		map[string]interface{}{
			RulestackName: o.Rulestack,
			"name":        o.Name,
			ScopeName:     o.Scope,
		},
	)

	if err := svc.CreateCertificate(ctx, o); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(buildCertificateId(o.Scope, o.Rulestack, o.Name))

	return readCertificate(ctx, d, meta)
}

func readCertificate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)
	scope, stack, name, err := parseCertificateId(d.Id())
	if err != nil {
		return diag.Errorf("Error in parsing ID %q: %s", d.Id(), err)
	}

	req := certificate.ReadInput{
		Rulestack: stack,
		Scope:     scope,
		Name:      name,
		Candidate: true,
	}
	tflog.Info(
		ctx, "read certificate",
		map[string]interface{}{
			RulestackName: req.Rulestack,
			"name":        name,
			ScopeName:     scope,
		},
	)

	res, err := svc.ReadCertificate(ctx, req)
	if err != nil {
		if isObjectNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.Set(ScopeName, scope)
	saveCertificate(d, stack, name, *res.Response.Candidate)

	return nil
}

func updateCertificate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)
	o := loadCertificate(d)
	tflog.Info(
		ctx, "update certificate",
		map[string]interface{}{
			RulestackName: o.Rulestack,
			"name":        o.Name,
			ScopeName:     o.Scope,
		},
	)

	if err := svc.UpdateCertificate(ctx, o); err != nil {
		return diag.FromErr(err)
	}

	return readCertificate(ctx, d, meta)
}

func deleteCertificate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)
	scope, stack, name, err := parseCertificateId(d.Id())
	if err != nil {
		return diag.Errorf("Error in parsing ID %q: %s", d.Id(), err)
	}

	tflog.Info(
		ctx, "delete certificate",
		map[string]interface{}{
			RulestackName: stack,
			"name":        name,
			ScopeName:     scope,
		},
	)

	input := certificate.DeleteInput{
		Rulestack: stack,
		Scope:     scope,
		Name:      name,
	}
	if err := svc.DeleteCertificate(ctx, input); err != nil && !isObjectNotFound(err) {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

// Schema handling.
func certificateSchema(isResource bool, rmKeys []string) map[string]*schema.Schema {
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
		"signer_arn": {
			Type:          schema.TypeString,
			Optional:      true,
			Description:   "The certificate signer ARN.",
			ConflictsWith: []string{"self_signed"},
		},
		"self_signed": {
			Type:          schema.TypeBool,
			Optional:      true,
			Description:   "Set to true if certificate is self-signed.",
			ConflictsWith: []string{"signer_arn"},
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

func loadCertificate(d *schema.ResourceData) certificate.Info {
	return certificate.Info{
		Rulestack:    d.Get(RulestackName).(string),
		Scope:        d.Get(ScopeName).(string),
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
		SignerArn:    d.Get("signer_arn").(string),
		SelfSigned:   d.Get("self_signed").(bool),
		AuditComment: d.Get("audit_comment").(string),
	}
}

func saveCertificate(d *schema.ResourceData, stack, name string, o certificate.Info) {
	d.Set(RulestackName, stack)
	d.Set("name", name)
	d.Set("description", o.Description)
	d.Set("signer_arn", o.SignerArn)
	d.Set("self_signed", o.SelfSigned)
	d.Set("audit_comment", o.AuditComment)
	d.Set("update_token", o.UpdateToken)
}

// Id functions.
func buildCertificateId(a, b, c string) string {
	return strings.Join([]string{a, b, c}, IdSeparator)
}

func parseCertificateId(v string) (string, string, string, error) {
	tok := strings.Split(v, IdSeparator)
	if len(tok) != 3 {
		return "", "", "", fmt.Errorf("Expecting 2 tokens, got %d", len(tok))
	}

	return tok[0], tok[1], tok[2], nil
}
