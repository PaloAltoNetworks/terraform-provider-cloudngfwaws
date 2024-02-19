package provider

import (
	"context"
	"strconv"
	"strings"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go/api"
	"github.com/paloaltonetworks/cloud-ngfw-aws-go/api/country"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Data source (list countries).
func dataSourceCountry() *schema.Resource {
	return &schema.Resource{
		Description: "Data source get a list of countries and their country codes.",

		ReadContext: readCountry,

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
			"codes": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "The country code (as the key) and description (as the value).",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func readCountry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	input := country.ListInput{
		MaxResults: d.Get("max_results").(int),
		NextToken:  d.Get("token").(string),
	}

	d.Set("max_results", input.MaxResults)
	d.Set("token", input.NextToken)

	tflog.Info(
		ctx, "read countries",
		map[string]interface{}{
			"max_results": input.MaxResults,
			"token":       input.NextToken,
		},
	)

	svc := meta.(*api.ApiClient)

	ans, err := svc.ListCountry(ctx, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strings.Join(
		[]string{strconv.Itoa(input.MaxResults), input.NextToken},
		IdSeparator,
	))

	var codes map[string]interface{}
	if ans.Response != nil && len(ans.Response.Countries) > 0 {
		codes = make(map[string]interface{})
		for _, x := range ans.Response.Countries {
			codes[x.Code] = x.Description
		}
	}

	d.Set("next_token", ans.Response.NextToken)
	d.Set("codes", codes)

	return nil
}
