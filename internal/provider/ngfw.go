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
			"max_results": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Max number of results.",
				Default:     100,
			},
			"next_token": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Token for the next page of results.",
			},
			"vpc_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of vpc ids.",
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
							Description: "The instance name.",
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
	)

	res, err := svc.Read(ctx, req)
	if err != nil {
		if isObjectNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	account_id = res.Response.Firewall.AccountId
	id := buildNgfwId(account_id, name)
	d.SetId(id)

	saveNgfw(ctx, d, name, *res.Response)

	return nil
}

func readNgfws(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := ngfw.NewClient(meta.(*awsngfw.Client))

	vpc_ids := make([]string, len(d.Get("vpc_ids").([]interface{})), len(d.Get("vpc_ids").([]interface{})))
	for i, id := range d.Get("vpc_ids").([]interface{}) {
		vpc_ids[i] = id.(string)
	}

	input := ngfw.ListInput{
		MaxResults: d.Get("max_results").(int),
		NextToken:  d.Get("next_token").(string),
		VpcIds:     vpc_ids,
	}

	d.Set("max_results", input.MaxResults)
	d.Set("next_token", input.NextToken)

	tflog.Info(
		ctx, "read ngfws",
		"ds", true,
		"vpc_ids", vpc_ids,
	)

	ans, err := svc.List(ctx, input)
	if err != nil {
		if isObjectNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(strings.Join(
		append([]string{strconv.Itoa(input.MaxResults), input.NextToken}, input.VpcIds...),
		IdSeparator,
	))

	d.Set("next_token", ans.Response.NextToken)

	instances := make([]interface{}, len(ans.Response.Firewalls), len(ans.Response.Firewalls))
	for i, instance := range ans.Response.Firewalls {

		instances[i] = map[string]interface{}{
			"name":       instance.Name,
			"account_id": instance.AccountId,
		}
	}

	d.Set("instances", instances)

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

		Schema: ngfwSchema(true, []string{"status", "endpoint_service_name"}),
	}
}

func createNgfw(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := ngfw.NewClient(meta.(*awsngfw.Client))
	name := d.Get("name").(string)
	o := loadNgfw(ctx, d)

	tflog.Info(
		ctx, "create ngfw",
		"name", o.Name,
		"payload", o,
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
	)

	res, err := svc.Read(ctx, req)
	if err != nil {
		if isObjectNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	saveNgfw(ctx, d, name, *res.Response)

	return nil
}

func updateNgfw(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := ngfw.NewClient(meta.(*awsngfw.Client))
	o := loadNgfw(ctx, d)

	tflog.Info(
		ctx, "update ngfw",
		"name", o.Name,
	)

	req := ngfw.ReadInput{
		Name:      o.Name,
		AccountId: o.AccountId,
	}

	res, err := svc.Read(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("description") {
		input := ngfw.Info{
			Name:        o.Name,
			Description: o.Description,
			AccountId:   o.AccountId,
		}
		if err := svc.UpdateDescription(ctx, input); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("app_id_version") || d.HasChange("automatic_upgrade_app_id_version") {
		input := ngfw.Info{
			Name:                         o.Name,
			AccountId:                    o.AccountId,
			AppIdVersion:                 o.AppIdVersion,
			AutomaticUpgradeAppIdVersion: o.AutomaticUpgradeAppIdVersion,
		}
		if err := svc.UpdateNGFirewallContentVersion(ctx, input); err != nil {
			return diag.FromErr(err)
		}
	}

	assoc := make([]ngfw.SubnetMapping, 0, len(o.SubnetMappings))
	disassoc := make([]ngfw.SubnetMapping, 0, len(res.Response.Firewall.SubnetMappings))

	for _, x := range o.SubnetMappings {
		found := false

		for _, y := range res.Response.Firewall.SubnetMappings {
			if x.SubnetId == y.SubnetId {
				found = true
				break
			}
		}

		if !found {
			assoc = append(assoc, ngfw.SubnetMapping{
				SubnetId: x.SubnetId,
			})
		}
	}

	for _, x := range res.Response.Firewall.SubnetMappings {
		found := false

		for _, y := range o.SubnetMappings {
			if x.SubnetId == y.SubnetId {
				found = true
				break
			}
		}

		if !found {
			disassoc = append(disassoc, ngfw.SubnetMapping{
				SubnetId: x.SubnetId,
			})
		}
	}

	if len(assoc) != 0 || len(disassoc) != 0 {
		if len(assoc) == 0 {
			assoc = nil
		}
		if len(disassoc) == 0 {
			disassoc = nil
		}

		input := ngfw.Info{
			Name:                       o.Name,
			AccountId:                  o.AccountId,
			AssociateSubnetMappings:    assoc,
			DisassociateSubnetMappings: disassoc,
		}
		if err := svc.UpdateSubnetMappings(ctx, input); err != nil {
			return diag.FromErr(err)
		}
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

	fw := ngfw.ReadInput{
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
			Description: "The name.",
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
			Description: "The account ID.",
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
			Description:  addStringInSliceValidation("Set endpoint mode from the following options", endpoint_mode_opts),
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
		RulestackName:       rsSchema(),
		GlobalRulestackName: gRsSchema(),
		TagsName:            tagsSchema(true, true),
		"update_token": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The update token.",
		},
		"status": {
			Type:     schema.TypeList,
			MaxItems: 1,
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
					"attachments": {
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
		computed(ans, "", []string{"name"})
	}

	return ans
}

func loadNgfw(ctx context.Context, d *schema.ResourceData) ngfw.Info {

	return ngfw.Info{
		Name:                         d.Get("name").(string),
		VpcId:                        d.Get("vpc_id").(string),
		AccountId:                    d.Get("account_id").(string),
		Description:                  d.Get("description").(string),
		EndpointMode:                 d.Get("endpoint_mode").(string),
		SubnetMappings:               loadSubnetMappings(ctx, d.Get("subnet_mapping").([]interface{})),
		AppIdVersion:                 d.Get("app_id_version").(string),
		AutomaticUpgradeAppIdVersion: d.Get("automatic_upgrade_app_id_version").(bool),
		RuleStackName:                d.Get(RulestackName).(string),
		GlobalRuleStackName:          d.Get(GlobalRulestackName).(string),
		Tags:                         loadTags(d.Get(TagsName)),
	}
}

func saveNgfw(ctx context.Context, d *schema.ResourceData, name string, o ngfw.ReadResponse) {

	d.Set("name", name)
	d.Set("vpc_id", o.Firewall.VpcId)
	d.Set("account_id", o.Firewall.AccountId)
	d.Set("description", o.Firewall.Description)
	d.Set("endpoint_mode", o.Firewall.EndpointMode)
	d.Set("endpoint_service_name", o.Firewall.EndpointServiceName)
	d.Set("subnet_mapping", saveSubnetMappings(ctx, o.Firewall.SubnetMappings))
	d.Set("app_id_version", o.Firewall.AppIdVersion)
	d.Set("automatic_upgrade_app_id_version", o.Firewall.AutomaticUpgradeAppIdVersion)
	d.Set(RulestackName, o.Firewall.RuleStackName)
	d.Set(GlobalRulestackName, o.Firewall.GlobalRuleStackName)
	d.Set(TagsName, dumpTags(o.Firewall.Tags))
	d.Set("update_token", o.Firewall.UpdateToken)
	if o.Status != nil {
		d.Set("status", saveStatus(ctx, *o.Status))
	}

}

func saveSubnetMappings(ctx context.Context, subnetMappings []ngfw.SubnetMapping) []interface{} {
	if subnetMappings != nil {
		mappings := make([]interface{}, len(subnetMappings), len(subnetMappings))

		for i, sm := range subnetMappings {
			_sm := make(map[string]interface{})
			_sm["subnet_id"] = sm.SubnetId
			_sm["availability_zone"] = sm.AvailabilityZone
			// _sm["az_id"] = sm.AvailabilityZoneId
			mappings[i] = _sm
		}
		return mappings
	}

	return make([]interface{}, 0)
}

func loadSubnetMappings(ctx context.Context, subnetMappings []interface{}) []ngfw.SubnetMapping {
	if subnetMappings != nil {
		mappings := make([]ngfw.SubnetMapping, len(subnetMappings), len(subnetMappings))

		for i, sm := range subnetMappings {
			_smi := sm.(map[string]interface{})
			_sm := ngfw.SubnetMapping{
				SubnetId:         _smi["subnet_id"].(string),
				AvailabilityZone: _smi["availability_zone"].(string),
			}
			mappings[i] = _sm
		}

		return mappings
	}

	return make([]ngfw.SubnetMapping, 0)
}

func saveStatus(ctx context.Context, status ngfw.FirewallStatus) []interface{} {

	s := make([]interface{}, 1, 1)

	attachments := make([]map[string]interface{}, len(status.Attachments), len(status.Attachments))

	for i, att := range status.Attachments {
		_att := make(map[string]interface{})
		_att["endpoint_id"] = att.EndpointId
		_att["status"] = att.Status
		_att["rejected_reason"] = att.RejectedReason
		_att["subnet_id"] = att.SubnetId
		attachments[i] = _att
	}

	_s := make(map[string]interface{})

	_s["firewall_status"] = status.FirewallStatus
	_s["failure_reason"] = status.FailureReason
	_s["rulestack_status"] = status.RuleStackStatus
	_s["attachments"] = attachments

	s[0] = _s

	return s

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
