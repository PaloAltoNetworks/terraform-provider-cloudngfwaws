package provider

import (
	"github.com/paloaltonetworks/cloud-ngfw-aws-go/v2/api"
	"github.com/paloaltonetworks/cloud-ngfw-aws-go/v2/api/appid"

	"context"
	"strconv"
	"strings"

	// "github.com/paloaltonetworks/cloud-ngfw-aws-go/v2"
	// "github.com/paloaltonetworks/cloud-ngfw-aws-go/v2/appid"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Data source (list app-id versions).
func dataSourceAppIdVersions() *schema.Resource {
	return &schema.Resource{
		Description: "Data source get a list of AppId versions.",

		ReadContext: readAppIdVersions,

		Schema: map[string]*schema.Schema{
			"max_results": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Max number of results.",
				Default:     100,
			},
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Pagination token.",
			},
			"next_token": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Token for the next page of results.",
			},
			"versions": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of AppId versions.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func readAppIdVersions(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	input := appid.ListInput{
		MaxResults: d.Get("max_results").(int),
		NextToken:  d.Get("token").(string),
	}

	d.Set("max_results", input.MaxResults)
	d.Set("token", input.NextToken)

	tflog.Info(
		ctx, "read appid versions",
		map[string]interface{}{"max_results": input.MaxResults,
			"token": input.NextToken},
	)

	svc := meta.(*api.ApiClient)

	ans, err := svc.ListAppID(ctx, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strings.Join(
		[]string{strconv.Itoa(input.MaxResults), input.NextToken},
		IdSeparator,
	))

	d.Set("versions", ans.Response.Versions)
	d.Set("next_token", ans.Response.NextToken)
	return nil
}

// Data source (app-id version).
func dataSourceAppIdVersion() *schema.Resource {
	return &schema.Resource{
		Description: "Data source to retrieve information on a given AppId version.",

		ReadContext: readAppIdVersion,

		Schema: map[string]*schema.Schema{
			"version": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The AppId version.",
			},
			"max_results": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Max results.",
				Default:     100,
			},
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Pagination token.",
			},
			"next_token": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Token for the next page of results.",
			},
			"applications": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of applications.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func readAppIdVersion(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	input := appid.ReadInput{
		Version:    d.Get("version").(string),
		MaxResults: d.Get("max_results").(int),
		NextToken:  d.Get("token").(string),
	}

	d.Set("version", input.Version)
	d.Set("max_results", input.MaxResults)
	d.Set("token", input.NextToken)

	tflog.Info(
		ctx, "read appid version",
		map[string]interface{}{
			"version":     input.Version,
			"max_results": input.MaxResults,
			"token":       input.NextToken,
		},
	)

	svc := meta.(*api.ApiClient)

	ans, err := svc.ReadAppID(ctx, input)
	if err != nil {
		if isObjectNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(input.Version)
	d.Set("next_token", ans.Response.NextToken)
	d.Set("applications", ans.Response.Applications)

	return nil
}

// Data source (app-id application).
