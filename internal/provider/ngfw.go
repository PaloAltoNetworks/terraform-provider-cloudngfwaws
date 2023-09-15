package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go"
	ngfw "github.com/paloaltonetworks/cloud-ngfw-aws-go/firewall"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Data source (list NGFWs).
func dataSourceNgfws() *schema.Resource {
	return &schema.Resource{
		Description: "Data source get a list of NGFWs.",

		ReadContext: readNgfws,

		Schema: map[string]*schema.Schema{
			RulestackName: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The rulestack to filter on.",
			},
			"vpc_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of vpc ids to filter on.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"instances": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of NGFWs.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The NGFW name.",
						},
						"account_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The account id.",
						},
					},
				},
			},
		},
	}
}

func readNgfws(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := ngfw.NewClient(meta.(*awsngfw.Client))

	stack := d.Get(RulestackName).(string)
	vpcs := toStringSlice(d.Get("vpc_ids"))
	d.Set("vpc_ids", vpcs)

	tflog.Info(
		ctx, "read ngfws",
		"ds", true,
		"vpc_ids", vpcs,
		RulestackName, stack,
	)

	var nt string
	var listing []ngfw.ListFirewall
	for {
		input := ngfw.ListInput{
			Rulestack:  stack,
			MaxResults: 100,
			NextToken:  nt,
			VpcIds:     vpcs,
		}
		ans, err := svc.List(ctx, input)
		if err != nil {
			if isObjectNotFound(err) {
				break
			}
			return diag.FromErr(err)
		}

		listing = append(listing, ans.Response.Firewalls...)
		nt = ans.Response.NextToken
		if nt == "" {
			break
		}
	}

	d.SetId(strings.Join(
		append([]string{strconv.Itoa(len(listing))}, vpcs...),
		IdSeparator,
	))

	instances := make([]interface{}, 0, len(listing))
	for _, x := range listing {
		instances = append(instances, map[string]interface{}{
			"name":       x.Name,
			"account_id": x.AccountId,
		})
	}

	d.Set(RulestackName, stack)
	d.Set("instances", instances)

	return nil
}

// Data source for a single NGFW.
func dataSourceNgfw() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for retrieving NGFW information.",

		ReadContext: readNgfwDataSource,

		Schema: ngfwSchema(false, nil),
	}
}

func readNgfwDataSource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := ngfw.NewClient(meta.(*awsngfw.Client))

	name := d.Get("name").(string)
	account_id := d.Get("account_id").(string)

	req := ngfw.ReadInput{
		Name:      name,
		AccountId: account_id,
	}

	tflog.Info(
		ctx, "read ngfw",
		"ds", true,
		"name", name,
		"account_id", account_id,
	)

	res, err := svc.Read(ctx, req)
	if err != nil {
		if isObjectNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	id := buildNgfwId(res.Response.Firewall.AccountId, name)
	d.SetId(id)

	saveNgfw(d, res.Response)

	return nil
}

// Resource.
func resourceNgfw() *schema.Resource {
	return &schema.Resource{
		Description: "Resource for NGFW manipulation.",

		CreateContext: createNgfw,
		ReadContext:   readNgfw,
		UpdateContext: updateNgfw,
		DeleteContext: deleteNgfw,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: ngfwSchema(true, nil),
	}
}

func createNgfw(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := ngfw.NewClient(meta.(*awsngfw.Client))
	name := d.Get("name").(string)
	o := loadNgfw(d)

	tflog.Info(
		ctx, "create ngfw",
		"name", o.Name,
		"account_id", o.AccountId,
		"multi_vpc", o.MultiVpc,
		"link_id", o.LinkId,
	)

	res, err := svc.Create(ctx, o)
	if err != nil {
		return diag.FromErr(err)
	}

	account_id := res.Response.AccountId

	id := buildNgfwId(account_id, name)
	d.SetId(id)

	return readNgfw(ctx, d, meta)
}

func readNgfw(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := ngfw.NewClient(meta.(*awsngfw.Client))

	account_id, name, err := parseNgfwId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	req := ngfw.ReadInput{
		Name:      name,
		AccountId: account_id,
	}

	tflog.Info(
		ctx, "read ngfw",
		"name", name,
		"account_id", account_id,
	)

	res, err := svc.Read(ctx, req)
	if err != nil {
		if isObjectNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	saveNgfw(d, res.Response)

	return nil
}

func updateNgfw(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := ngfw.NewClient(meta.(*awsngfw.Client))
	o := loadNgfw(d)

	tflog.Info(
		ctx, "update ngfw",
		"name", o.Name,
		"account_id", o.AccountId,
		"multi_vpc", o.MultiVpc,
		"link_id", o.LinkId,
	)

	if err := svc.Modify(ctx, o); err != nil {
		return diag.FromErr(err)
	}

	return readNgfw(ctx, d, meta)
}

func deleteNgfw(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := ngfw.NewClient(meta.(*awsngfw.Client))

	account_id, name, err := parseNgfwId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Info(
		ctx, "delete ngfw",
		"name", name,
		"account_id", account_id,
	)

	fw := ngfw.DeleteInput{
		Name:      name,
		AccountId: account_id,
	}

	if err = svc.Delete(ctx, fw); err != nil && !isObjectNotFound(err) {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

// Schema handling.
func ngfwSchema(isResource bool, rmKeys []string) map[string]*schema.Schema {
	endpoint_mode_opts := []string{"ServiceManaged", "CustomerManaged"}

	ans := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The NGFW name.",
			ForceNew:    true,
		},
		"vpc_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The vpc id.",
			ForceNew:    true,
		},
		"account_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The account ID. This field is mandatory if using multiple accounts.",
			ForceNew:    true,
		},
		"description": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The description.",
		},
		"endpoint_mode": {
			Type:         schema.TypeString,
			Required:     true,
			Description:  addStringInSliceValidation("Set endpoint mode from the following options.", endpoint_mode_opts),
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice(endpoint_mode_opts, false),
		},
		"endpoint_service_name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The endpoint service name.",
		},
		"subnet_mapping": {
			Type:        schema.TypeList,
			Required:    true,
			MinItems:    1,
			Description: "Subnet mappings.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"subnet_id": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The subnet id, for when the endpoint mode is service managed.",
					},
					"availability_zone": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The availability zone, for when the endpoint mode is customer managed.",
					},
					// Future use
					// "az_id": {
					// 	Type:        schema.TypeString,
					// 	Optional:    true,
					// 	Description: "The availability zone id.",
					// },
				},
			},
		},
		"app_id_version": {
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			Description: "App-ID version number.",
		},
		"automatic_upgrade_app_id_version": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Automatic App-ID upgrade version number.",
			Default:     true,
		},
		RulestackName: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The rulestack for this NGFW.",
		},
		"global_rulestack": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The global rulestack for this NGFW.",
			ForceNew:    true,
		},
		"multi_vpc": {
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
			Description: "Share NGFW with Multiple VPCs. This feature can be enabled only if the endpoint_mode is CustomerManaged.",
		},
		"link_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The link ID.",
		},
		"link_status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The link status.",
		},
		TagsName: tagsSchema(true),
		"update_token": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The update token.",
		},
		"status": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"firewall_status": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The firewall status.",
					},
					"failure_reason": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The firewall failure reason.",
					},
					"rulestack_status": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The rulestack status.",
					},
					"attachment": {
						Type:        schema.TypeList,
						Computed:    true,
						Description: "The firewall attachments.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"endpoint_id": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "The endpoint id.",
								},
								"status": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "The attachment status.",
								},
								"rejected_reason": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "The reject reason.",
								},
								"subnet_id": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "The subnet id.",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, rmKey := range rmKeys {
		delete(ans, rmKey)
	}

	if !isResource {
		computed(ans, "", []string{"name", "account_id"})
	}

	return ans
}

func loadNgfw(d *schema.ResourceData) ngfw.Info {
	var sm []ngfw.SubnetMapping
	list := d.Get("subnet_mapping").([]interface{})
	if len(list) > 0 {
		sm = make([]ngfw.SubnetMapping, 0, len(list))
		for i := range list {
			x := list[i].(map[string]interface{})
			sm = append(sm, ngfw.SubnetMapping{
				SubnetId:         x["subnet_id"].(string),
				AvailabilityZone: x["availability_zone"].(string),
			})
		}
	}

	return ngfw.Info{
		Name:                         d.Get("name").(string),
		VpcId:                        d.Get("vpc_id").(string),
		AccountId:                    d.Get("account_id").(string),
		SubnetMappings:               sm,
		AppIdVersion:                 d.Get("app_id_version").(string),
		Description:                  d.Get("description").(string),
		Rulestack:                    d.Get(RulestackName).(string),
		GlobalRulestack:              d.Get("global_rulestack").(string),
		MultiVpc:                     d.Get("multi_vpc").(bool),
		EndpointMode:                 d.Get("endpoint_mode").(string),
		AutomaticUpgradeAppIdVersion: d.Get("automatic_upgrade_app_id_version").(bool),
		LinkId:                       d.Get("link_id").(string),
		LinkStatus:                   d.Get("link_status").(string),
		Tags:                         loadTags(d.Get(TagsName)),
	}
}

func saveNgfw(d *schema.ResourceData, o ngfw.ReadResponse) {
	var sm []interface{}
	if len(o.Firewall.SubnetMappings) > 0 {
		sm = make([]interface{}, 0, len(o.Firewall.SubnetMappings))
		for _, x := range o.Firewall.SubnetMappings {
			sm = append(sm, map[string]interface{}{
				"subnet_id":         x.SubnetId,
				"availability_zone": x.AvailabilityZone,
			})
		}
	}

	var att []interface{}
	if len(o.Status.Attachments) > 0 {
		att = make([]interface{}, 0, len(o.Status.Attachments))
		for _, x := range o.Status.Attachments {
			att = append(att, map[string]interface{}{
				"endpoint_id":     x.EndpointId,
				"status":          x.Status,
				"rejected_reason": x.RejectedReason,
				"subnet_id":       x.SubnetId,
			})
		}
	}

	stat := []interface{}{
		map[string]interface{}{
			"firewall_status":  o.Status.FirewallStatus,
			"failure_reason":   o.Status.FailureReason,
			"rulestack_status": o.Status.RulestackStatus,
			"attachment":       att,
		},
	}

	d.Set("name", o.Firewall.Name)
	d.Set("account_id", o.Firewall.AccountId)
	d.Set("subnet_mapping", sm)
	d.Set("vpc_id", o.Firewall.VpcId)
	d.Set("app_id_version", o.Firewall.AppIdVersion)
	d.Set("description", o.Firewall.Description)
	d.Set(RulestackName, o.Firewall.Rulestack)
	d.Set("global_rulestack", o.Firewall.GlobalRulestack)
	d.Set("multi_vpc", o.Firewall.MultiVpc)
	d.Set("endpoint_service_name", o.Firewall.EndpointServiceName)
	d.Set("endpoint_mode", o.Firewall.EndpointMode)
	d.Set("automatic_upgrade_app_id_version", o.Firewall.AutomaticUpgradeAppIdVersion)
	d.Set("link_id", o.Firewall.LinkId)
	d.Set("link_status", o.Firewall.LinkStatus)
	d.Set(TagsName, dumpTags(o.Firewall.Tags))
	d.Set("update_token", o.Firewall.UpdateToken)
	d.Set("status", stat)
}

// Id functions.
func buildNgfwId(a, b string) string {
	return strings.Join([]string{a, b}, IdSeparator)
}

func parseNgfwId(v string) (string, string, error) {
	tok := strings.Split(v, IdSeparator)
	if len(tok) != 2 {
		return "", "", fmt.Errorf("Expecting 2 tokens, got %d", len(tok))
	}

	return tok[0], tok[1], nil
}
