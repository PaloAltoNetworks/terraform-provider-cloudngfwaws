package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go"
	lp "github.com/paloaltonetworks/cloud-ngfw-aws-go/firewall/logprofile"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Data source.
func dataSourceNgfwLogProfile() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for retrieving log profile information.",

		ReadContext: readNgfwLogProfileDataSource,

		Schema: ngfwLogProfileSchema(false, nil),
	}
}

func readNgfwLogProfileDataSource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := lp.NewClient(meta.(*awsngfw.Client))

	aid := d.Get("account_id").(string)
	ngfw := d.Get("ngfw").(string)

	req := lp.ReadInput{
		Firewall:  ngfw,
		AccountId: aid,
	}

	tflog.Info(
		ctx, "read ngfw log profile",
		"ds", true,
		"ngfw", ngfw,
		"account_id", aid,
	)

	res, err := svc.Read(ctx, req)
	if err != nil {
		if isObjectNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(buildNgfwLogProfileId(aid, ngfw))

	saveNgfwLogProfile(d, *res.Response)

	return nil
}

// Resource.
func resourceNgfwLogProfile() *schema.Resource {
	return &schema.Resource{
		Description: "Resource for NGFW log profile manipulation.",

		CreateContext: createUpdateNgfwLogProfile,
		ReadContext:   readNgfwLogProfile,
		UpdateContext: createUpdateNgfwLogProfile,
		DeleteContext: deleteNgfwLogProfile,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: ngfwLogProfileSchema(true, nil),
	}
}

func createUpdateNgfwLogProfile(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := lp.NewClient(meta.(*awsngfw.Client))
	o := loadNgfwLogProfile(d)

	tflog.Info(
		ctx, "modify ngfw log profile",
		"ngfw", o.Firewall,
		"account_id", o.AccountId,
	)

	if err := svc.Update(ctx, o); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(buildNgfwLogProfileId(o.AccountId, o.Firewall))

	return readNgfwLogProfile(ctx, d, meta)
}

func readNgfwLogProfile(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := lp.NewClient(meta.(*awsngfw.Client))
	aid, ngfw, err := parseNgfwLogProfileId(d.Id())
	if err != nil {
		return diag.Errorf("Error in parsing ID %q: %s", d.Id(), err)
	}

	req := lp.ReadInput{
		Firewall:  ngfw,
		AccountId: aid,
	}
	tflog.Info(
		ctx, "read ngfw log profile",
		"ngfw", ngfw,
		"account_id", aid,
	)

	res, err := svc.Read(ctx, req)
	if err != nil {
		if isObjectNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	saveNgfwLogProfile(d, *res.Response)

	return nil
}

func deleteNgfwLogProfile(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}

// Schema handling.
func ngfwLogProfileSchema(isResource bool, rmKeys []string) map[string]*schema.Schema {
	destinationTypes := []string{"S3", "CloudWatchLogs", "KinesisDataFirehose"}
	logTypes := []string{"TRAFFIC", "THREAT", "DECRYPTION"}

	ans := map[string]*schema.Schema{
		"ngfw": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the NGFW.",
			ForceNew:    true,
		},
		"account_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique ID of the account.",
		},
		"cloud_watch_metric_namespace": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The CloudWatch metric namespace.",
		},
		"log_destination": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "List of log destinations.",
			MinItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"destination": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The log destination details.",
					},
					"destination_type": {
						Type:         schema.TypeString,
						Optional:     true,
						Description:  addStringInSliceValidation("The log destination type.", destinationTypes),
						ValidateFunc: validation.StringInSlice(destinationTypes, false),
					},
					"log_type": {
						Type:         schema.TypeString,
						Optional:     true,
						Description:  addStringInSliceValidation("The type of logs.", logTypes),
						ValidateFunc: validation.StringInSlice(logTypes, false),
					},
				},
			},
		},
	}

	for _, rmKey := range rmKeys {
		delete(ans, rmKey)
	}

	if !isResource {
		computed(ans, "", []string{"ngfw", "account_id"})
	}

	return ans
}

func loadNgfwLogProfile(d *schema.ResourceData) lp.Info {
	var dests []lp.LogDestination

	list := d.Get("log_destination").([]interface{})
	if len(list) > 0 {
		dests = make([]lp.LogDestination, 0, len(list))
		for i := range list {
			x := list[i].(map[string]interface{})
			dests = append(dests, lp.LogDestination{
				Destination:     x["destination"].(string),
				DestinationType: x["destination_type"].(string),
				LogType:         x["log_type"].(string),
			})
		}
	}

	return lp.Info{
		AccountId:                 d.Get("account_id").(string),
		Firewall:                  d.Get("ngfw").(string),
		LogDestinations:           dests,
		CloudWatchMetricNamespace: d.Get("cloud_watch_metric_namespace").(string),
	}
}

func saveNgfwLogProfile(d *schema.ResourceData, o lp.Info) {
	var dests []interface{}
	if len(o.LogDestinations) > 0 {
		dests = make([]interface{}, 0, len(o.LogDestinations))
		for _, x := range o.LogDestinations {
			dests = append(dests, map[string]interface{}{
				"destination":      x.Destination,
				"destination_type": x.DestinationType,
				"log_type":         x.LogType,
			})
		}
	}

	d.Set("account_id", o.AccountId)
	d.Set("ngfw", o.Firewall)
	d.Set("log_destination", dests)
	d.Set("cloud_watch_metric_namespace", o.CloudWatchMetricNamespace)
}

// Id functions.
func buildNgfwLogProfileId(a, b string) string {
	return strings.Join([]string{a, b}, IdSeparator)
}

func parseNgfwLogProfileId(v string) (string, string, error) {
	tok := strings.Split(v, IdSeparator)
	if len(tok) != 2 {
		return "", "", fmt.Errorf("Expecting 2 tokens, got %d", len(tok))
	}

	return tok[0], tok[1], nil
}
