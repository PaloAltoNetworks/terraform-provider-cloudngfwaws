package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go"
	"github.com/paloaltonetworks/cloud-ngfw-aws-go/tag"
	"github.com/paloaltonetworks/cloud-ngfw-aws-go/tag/firewall"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Data source.
func dataSourceNgfwTag() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for retrieving NGFW tag information.",

		ReadContext: readNgfwTagDataSource,

		Schema: ngfwTagSchema(false, nil),
	}
}

func readNgfwTagDataSource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := firewall.NewClient(meta.(*awsngfw.Client))

	req := firewall.ListInput{
		Firewall:   d.Get("ngfw").(string),
		AccountId:  d.Get("account_id").(string),
		MaxResults: 1000,
	}

	id := buildNgfwTagId(req.AccountId, req.Firewall)

	tflog.Info(
		ctx, "read ngfw tags",
		"ds", true,
		"ngfw", req.Firewall,
		"account_id", req.AccountId,
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

	saveNgfwTag(d, req.Firewall, req.AccountId, res.Response.Tags)

	return nil
}

// Resource.
func resourceNgfwTag() *schema.Resource {
	return &schema.Resource{
		Description: "Resource for NGFW tag manipulation.",

		CreateContext: createUpdateNgfwTag,
		ReadContext:   readNgfwTag,
		UpdateContext: createUpdateNgfwTag,
		DeleteContext: deleteNgfwTag,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: ngfwTagSchema(true, nil),
	}
}

func createUpdateNgfwTag(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := firewall.NewClient(meta.(*awsngfw.Client))
	o := loadNgfwTag(d)
	tflog.Info(
		ctx, "modify ngfw tags",
		"ngfw", o.Firewall,
		"account_id", o.AccountId,
	)

	if err := svc.Apply(ctx, o); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(buildNgfwTagId(o.AccountId, o.Firewall))

	return readNgfwTag(ctx, d, meta)
}

func readNgfwTag(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := firewall.NewClient(meta.(*awsngfw.Client))

	aid, ngfw, err := parseNgfwTagId(d.Id())
	if err != nil {
		return diag.Errorf("Error in parsing ID %q: %s", d.Id(), err)
	}

	req := firewall.ListInput{
		Firewall:   ngfw,
		AccountId:  aid,
		MaxResults: 100,
	}

	tflog.Info(
		ctx, "read ngfw tags",
		"ngfw", req.Firewall,
		"account_id", req.AccountId,
	)

	res, err := svc.List(ctx, req)
	if err != nil {
		if isObjectNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	saveNgfwTag(d, req.Firewall, req.AccountId, res.Response.Tags)

	return nil
}

func deleteNgfwTag(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := firewall.NewClient(meta.(*awsngfw.Client))

	aid, ngfw, err := parseNgfwTagId(d.Id())
	if err != nil {
		return diag.Errorf("Error in parsing ID %q: %s", d.Id(), err)
	}

	tflog.Info(
		ctx, "delete ngfw tags",
		"ngfw", ngfw,
		"account_id", aid,
	)

	if err := svc.Apply(ctx, firewall.Info{Firewall: ngfw, AccountId: aid}); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

// Schema handling.
func ngfwTagSchema(isResource bool, rmKeys []string) map[string]*schema.Schema {
	ans := map[string]*schema.Schema{
		"ngfw": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The NGFW name.",
			ForceNew:    true,
		},
		"account_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The account ID.",
		},
		"tags": tagsSchema(true, false),
	}

	for _, rmKey := range rmKeys {
		delete(ans, rmKey)
	}

	if !isResource {
		computed(ans, "", []string{"ngfw", "account_id"})
	}

	return ans
}

func loadNgfwTag(d *schema.ResourceData) firewall.Info {
	return firewall.Info{
		Firewall:  d.Get("ngfw").(string),
		AccountId: d.Get("account_id").(string),
		Tags:      loadTags(d.Get("tags")),
	}
}

func saveNgfwTag(d *schema.ResourceData, ngfw, aid string, o []tag.Details) {
	d.Set("ngfw", ngfw)
	d.Set("account_id", aid)
	d.Set("tags", dumpTags(o))
}

// Id functions.
func buildNgfwTagId(a, b string) string {
	return strings.Join([]string{a, b}, IdSeparator)
}

func parseNgfwTagId(v string) (string, string, error) {
	tok := strings.Split(v, IdSeparator)
	if len(tok) != 2 {
		return "", "", fmt.Errorf("Expecting 2 tokens, got %d", len(tok))
	}

	return tok[0], tok[1], nil
}
