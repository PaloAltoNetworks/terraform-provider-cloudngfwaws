package provider

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go/v2/api"
	"github.com/paloaltonetworks/cloud-ngfw-aws-go/v2/api/firewall"
	ngfw "github.com/paloaltonetworks/cloud-ngfw-aws-go/v2/api/firewall"
	"github.com/paloaltonetworks/cloud-ngfw-aws-go/v2/api/tag"

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
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The region to filter on.",
			},
			"instances": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of NGFWs.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"region": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The region the NGFW is in.",
						},
						"firewall_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The NGFW ID.",
						},
					},
				},
			},
		},
	}
}

func readNgfws(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)

	stack := d.Get(RulestackName).(string)
	region := d.Get("region").(string)
	if region == "" {
		region = svc.GetRegion(ctx)
	}
	tflog.Info(
		ctx, "read ngfws",
		map[string]interface{}{
			"ds":          true,
			RulestackName: stack,
		},
	)

	var nt string
	var listing []ngfw.ListFirewall
	for {
		input := ngfw.ListInput{
			Rulestack:  stack,
			MaxResults: 100,
			NextToken:  nt,
			Region:     region,
		}
		ans, err := svc.ListFirewall(ctx, input)
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
		append([]string{strconv.Itoa(len(listing))}, region),
		IdSeparator,
	))

	instances := make([]interface{}, 0, len(listing))
	for _, x := range listing {
		instances = append(instances, map[string]interface{}{
			"firewall_id": x.FirewallId,
			"region":      x.Region,
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
	svc := meta.(*api.ApiClient)

	req := ngfw.ReadInput{
		FirewallId: d.Get("firewall_id").(string),
	}

	tflog.Info(
		ctx, "read ngfw",
		map[string]interface{}{
			"ds":         true,
			"FirewallId": d.Id(),
		},
	)

	res, err := svc.ReadFirewall(ctx, req)
	if err != nil {
		if isObjectNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(d.Id())

	if err := saveNgfw(ctx, d, res.Response); err != nil {
		return diag.FromErr(err)
	}

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
		Timeouts: &schema.ResourceTimeout{
			Create:  &resourceTimeout,
			Read:    &resourceTimeout,
			Update:  &resourceTimeout,
			Delete:  &resourceTimeout,
			Default: &resourceTimeout,
		},
	}
}

func createNgfw(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)
	o, err := loadNgfw(d, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Info(
		ctx, "create ngfw",
		map[string]interface{}{},
	)

	res, err := svc.CreateFirewallWithWait(ctx, o)
	if err != nil {
		return diag.FromErr(err)
	}
	id := res.Response.Id
	d.SetId(id)
	d.Set("firewall_id", id)
	return readNgfw(ctx, d, meta)
}

func readNgfw(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)

	req := ngfw.ReadInput{
		FirewallId: d.Get("firewall_id").(string),
	}

	tflog.Info(
		ctx, "read ngfw",
		map[string]interface{}{
			"FirewallId": d.Get("firewall_id").(string),
		},
	)

	res, err := svc.ReadFirewall(ctx, req)
	if err != nil {
		if isObjectNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if err := saveNgfw(ctx, d, res.Response); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func updateNgfw(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)
	req := ngfw.ReadInput{
		FirewallId: d.Get("firewall_id").(string),
	}
	res, err := svc.ReadFirewall(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}
	curEps := res.Response.Firewall.Endpoints
	o, err := loadNgfw(d, curEps)
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Info(
		ctx, "update ngfw",
		map[string]interface{}{},
	)

	err = svc.ModifyFirewallWithWait(ctx, o)

	if err != nil {
		return diag.FromErr(err)
	}

	return readNgfw(ctx, d, meta)
}

func deleteNgfw(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)

	tflog.Info(
		ctx, "delete ngfw",
		map[string]interface{}{
			"firewall_id": d.Get("firewall_id").(string),
		},
	)

	fw := ngfw.DeleteInput{
		FirewallId: d.Get("firewall_id").(string),
	}
	err := svc.DeleteFirewallWithWait(ctx, fw)
	if err != nil {
		if isObjectNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

// Schema handling.
func ngfwSchema(isResource bool, rmKeys []string) map[string]*schema.Schema {
	endpoint_mode_opts := []string{"ServiceManaged", "CustomerManaged"}
	ipPoolTypes := []string{"AWSService", "BYOIP"}
	ans := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The NGFW name.",
		},
		"firewall_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The Firewall ID.",
		},
		"description": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The NGFW description.",
		},
		"account_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The description.",
		},
		"vpc_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The VPC ID for the NGFW.",
		},
		"endpoint_mode": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: addStringInSliceValidation("Set endpoint mode from the following options.", endpoint_mode_opts),
		},
		"endpoint_service_name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The endpoint service name.",
		},
		"az_list": {
			Type: schema.TypeSet,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Required:    true,
			Description: "The list of availability zone IDs for this NGFW.",
		},
		"subnet_mapping": {
			Type:        schema.TypeList,
			Optional:    true,
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
						Computed:    true,
						Description: "The availability zone, for when the endpoint mode is customer managed.",
					},
					"availability_zone_id": {
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
						Description: "The availability zone ID, for when the endpoint mode is customer managed.",
					},
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
		"multi_vpc": {
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
			Description: "Share NGFW with Multiple VPCs. This feature can be enabled only if the endpoint_mode is CustomerManaged.",
		},
		"allowlist_accounts": {
			Type: schema.TypeSet,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Optional:    true,
			Description: "The list of allowed accounts for this NGFW.",
		},
		"change_protection": {
			Type: schema.TypeSet,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Optional:    true,
			Computed:    true,
			Description: "Enables or disables change protection for the NGFW.",
		},
		RulestackName: {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The rulestack for this NGFW.",
		},
		"global_rulestack": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The global rulestack for this NGFW.",
			ForceNew:    true,
		},
		"link_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The link ID.",
		},
		"link_status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The link status.",
		},
		"tags": {
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Description: "The tags.",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"update_token": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The update token.",
		},
		"deployment_update_token": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The update token.",
		},
		"private_access": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Type of Private Access",
					},
					"resource_id": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "AWS ResourceID",
					},
				},
			},
		},
		"user_id": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"enabled": {
						Type:        schema.TypeBool,
						Required:    true,
						Description: "Enable UserID Config",
					},
					"collector_name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The Collector Name",
					},
					"port": {
						Type:        schema.TypeInt,
						Required:    true,
						Description: "The Port",
					},
					"secret_key_arn": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "AWS Secret Key ARN",
					},
					"agent_name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Agent Name for UserID",
					},
					"user_id_status": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Status and State of UserID Configuration",
					},
					"custom_include_exclude_network": {
						Type:        schema.TypeList,
						Optional:    true,
						MinItems:    1,
						Description: "List of Custom Include Exclude Networks",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"enabled": {
									Type:        schema.TypeBool,
									Required:    true,
									Description: "Enable this specific custom include/exclude network",
								},
								"name": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "Name of subnet filter",
								},
								"network_address": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "Network IP address of the subnet filter",
								},
								"discovery_include": {
									Type:        schema.TypeBool,
									Required:    true,
									Description: "Include or exclude this subnet from user-id configuration",
								},
							},
						},
					},
				},
			},
		},
		"egress_nat": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"enabled": {
						Type:        schema.TypeBool,
						Required:    true,
						Description: "Enable egress NAT",
					},
					"settings": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"ip_pool_type": {
									Type:         schema.TypeString,
									Optional:     true,
									Description:  addStringInSliceValidation("Set ip pool type from the following options.", ipPoolTypes),
									ValidateFunc: validation.StringInSlice(ipPoolTypes, false),
								},
								"ipam_pool_id": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "The IP pool ID",
								},
							},
						},
					},
				},
			},
		},
		"endpoints": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     endpointsSchemaResource(),
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
					"device_rulestack_commit_status": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The device rulestack commit status.",
					},
				},
			},
		},
	}

	for _, rmKey := range rmKeys {
		delete(ans, rmKey)
	}

	if !isResource {
		computed(ans, "", make([]string, 0))
		ans["firewall_id"].Computed = false
		ans["firewall_id"].Required = true
	}

	return ans
}

func curEndpoints(ctx context.Context, d *schema.ResourceData, client *api.ApiClient) ([]firewall.EndpointConfig, error) {
	req := ngfw.ReadInput{
		FirewallId: d.Get("firewall_id").(string),
	}

	tflog.Info(
		ctx, "read ngfw",
		map[string]interface{}{
			"FirewallId": d.Get("firewall_id").(string),
		},
	)

	res, err := client.ReadFirewall(ctx, req)
	if err != nil {
		return nil, err
	}
	eps := res.Response.Firewall.Endpoints
	return eps, nil
}

func loadUserIdConfig(d *schema.ResourceData) (*ngfw.UserIDConfig, error) {
	firewallUserIdConfig := ngfw.UserIDConfig{}
	userId := d.Get("user_id").([]interface{})
	if len(userId) == 0 {
		return nil, nil
	}
	if len(userId) > 0 {
		userIdMap := userId[0].(map[string]interface{})
		firewallUserIdConfig.Port = userIdMap["port"].(int)
		firewallUserIdConfig.Enabled = userIdMap["enabled"].(bool)
		if !firewallUserIdConfig.Enabled {
			return &firewallUserIdConfig, nil
		}
		firewallUserIdConfig.CollectorName = userIdMap["collector_name"].(string)
		firewallUserIdConfig.SecretKeyARN = userIdMap["secret_key_arn"].(string)
		firewallUserIdConfig.UserIDStatus = userIdMap["user_id_status"].(string)
		firewallUserIdConfig.AgentName = userIdMap["agent_name"].(string)
		if firewallUserIdConfig.AgentName == "" {
			firewallUserIdConfig.AgentName = "agent-" + d.Get("firewall_id").(string)
		}
		rawCustomNetworks := userIdMap["custom_include_exclude_network"].([]interface{})
		customNetworks := make([]ngfw.UserIDCustomSubnetFilter, len(rawCustomNetworks))
		for i, raw := range rawCustomNetworks {
			data := raw.(map[string]interface{})
			customNetworks[i] = ngfw.UserIDCustomSubnetFilter{
				Enabled:          data["enabled"].(bool),
				Name:             data["name"].(string),
				NetworkAddress:   data["network_address"].(string),
				DiscoveryInclude: data["discovery_include"].(bool),
			}
		}
		firewallUserIdConfig.CustomIncludeExcludeNetwork = customNetworks
	}
	return &firewallUserIdConfig, nil
}

func loadPrivateAccessConfig(d *schema.ResourceData) (*ngfw.PrivateAccessConfig, error) {
	firewallPrivateAccessConfig := ngfw.PrivateAccessConfig{}
	privAccess := d.Get("private_access").([]interface{})
	if len(privAccess) == 0 {
		return nil, nil
	}
	if len(privAccess) > 0 {
		privAccessMap := privAccess[0].(map[string]interface{})
		firewallPrivateAccessConfig.Type = privAccessMap["type"].(string)
		firewallPrivateAccessConfig.ResourceID = privAccessMap["resource_id"].(string)
	}
	return &firewallPrivateAccessConfig, nil
}

func loadEndpoints(firewallEgressNat ngfw.EgressNATConfig, d *schema.ResourceData, curEps []firewall.EndpointConfig) []ngfw.EndpointConfig {
	defaultCidrs := make([]interface{}, 0)
	defaultCidrsStringMap := map[string]bool{
		"10.0.0.0/8":     true,
		"172.16.0.0/12":  true,
		"192.168.0.0/16": true,
	}
	for cidr, _ := range defaultCidrsStringMap {
		defaultCidrs = append(defaultCidrs, cidr)
	}
	endpoints := d.Get("endpoints").([]interface{})
	curEpMap := EpSubnetMap(curEps)
	resEps := make([]ngfw.EndpointConfig, 0)
	for _, ep := range endpoints {
		curEp := ep.(map[string]interface{})
		fwEgressNat := firewallEgressNat.Enabled
		epId := curEp["endpoint_id"].(string)
		subnetId := curEp["subnet_id"].(string)
		log.Printf("Subnet ID: %s", subnetId)
		if epId == "" {
			epId = GetEpId(curEpMap, subnetId)
			log.Printf("Endpoint ID: %s", epId)
		}
		firewallEp := ngfw.EndpointConfig{
			EgressNATEnabled: curEp["egress_nat_enabled"].(bool),
			EndpointId:       epId,
			VpcId:            curEp["vpc_id"].(string),
			AccountId:        curEp["account_id"].(string),
			ZoneId:           curEp["zone_id"].(string),
			SubnetId:         subnetId,
			Status:           curEp["status"].(string),
			RejectedReason:   curEp["rejected_reason"].(string),
			Mode:             curEp["mode"].(string),
		}
		if !fwEgressNat {
			firewallEp.EgressNATEnabled = fwEgressNat
		}
		curPrefixes := curEp["prefixes"].([]interface{})
		if len(curPrefixes) == 0 {
			prefixes := []interface{}{
				map[string]interface{}{
					"private_prefix": []interface{}{
						map[string]interface{}{
							"cidrs": schema.NewSet(schema.HashString, defaultCidrs),
						},
					},
				},
			}
			curEp["prefixes"] = prefixes
		}
		if firewallEp.Prefixes == nil {
			firewallEp.Prefixes = DefaultPrefixInfo()
		}
		if prefixes, ok := curEp["prefixes"]; ok {
			inputPrefixes := prefixes.([]interface{})
			if len(inputPrefixes) > 0 {
				inputPrefixesMap := inputPrefixes[0].(map[string]interface{})
				if privatePrefix, ok := inputPrefixesMap["private_prefix"]; ok {
					inputPrivatePrefix := privatePrefix.([]interface{})
					inputPrivatePrefixMap := inputPrivatePrefix[0].(map[string]interface{})
					if len(inputPrivatePrefix) > 0 {
						if cidrs, ok := inputPrivatePrefixMap["cidrs"]; ok {
							firewallCidrs := make([]string, 0)
							cidrsList := cidrs.(*schema.Set).List()
							cidrMap := make(map[string]bool)
							for _, cidr := range cidrsList {
								cidrMap[cidr.(string)] = true
							}
							for _, cidr := range cidrsList {
								firewallCidrs = append(firewallCidrs, cidr.(string))
							}
							for cidr, _ := range defaultCidrsStringMap {
								if _, ok := cidrMap[cidr]; !ok {
									firewallCidrs = append(firewallCidrs, cidr)
								}
							}
							firewallEp.Prefixes.PrivatePrefix.Cidrs = firewallCidrs
						}
					}
				}
			}
		}
		log.Printf("input cidrs: %v", firewallEp.Prefixes.PrivatePrefix.Cidrs)
		resEps = append(resEps, firewallEp)
	}
	return resEps
}

func loadEgressNat(d *schema.ResourceData) (ngfw.EgressNATConfig, error) {
	firewallEgressNat := ngfw.EgressNATConfig{}
	egressNat := d.Get("egress_nat").([]interface{})
	if len(egressNat) > 0 {
		egressNatMap := egressNat[0].(map[string]interface{})
		enabled := egressNatMap["enabled"].(bool)
		firewallEgressNat.Enabled = enabled
		egressNatSettings := egressNatMap["settings"].([]interface{})
		if len(egressNatSettings) > 0 {
			egressNatSettingsMap := egressNatSettings[0].(map[string]interface{})
			ipPoolType := egressNatSettingsMap["ip_pool_type"].(string)
			firewallEgressNat.Settings = &ngfw.EgressNATSettings{
				IPPoolType: ipPoolType,
			}
			if ipPoolType == "BYOIP" {
				if egressNatSettingsMap["ipam_pool_id"] == nil || egressNatSettingsMap["ipam_pool_id"].(string) == "" {
					return ngfw.EgressNATConfig{}, fmt.Errorf("ipam_pool_id is required when ip_pool_type is BYOIP")
				}
				ipamPoolId := egressNatSettingsMap["ipam_pool_id"].(string)
				firewallEgressNat.Settings = &ngfw.EgressNATSettings{
					IPPoolType: ipPoolType,
					IPAMPoolId: &ipamPoolId,
				}
			}
		} else {
			firewallEgressNat.Settings = nil
		}
	}
	return firewallEgressNat, nil
}

func setChangeProtection(d *schema.ResourceData) []string {
	changeProtection := setToSlice(d.Get("change_protection"))
	changeProtection = checkNilSlice(changeProtection)
	if len(changeProtection) == 0 {
		changeProtection = []string{"GlobalFirewallAdmin"}
	}
	return changeProtection
}

func setTags(d *schema.ResourceData) ([]tag.Details, error) {
	fwName := d.Get("name").(string)
	tags := loadTags(d.Get("tags"))
	if tags == nil {
		tags = make([]tag.Details, 0)
	}
	tagFwName, err := getFirewallName(tags)
	if err == nil && fwName != tagFwName {
		return nil, fmt.Errorf("firewall name mismatch: expected %s, got %s", fwName, tagFwName)
	}
	if err != nil {
		tags = append(tags, tag.Details{
			Key:   "FirewallName",
			Value: fwName,
		})
	}
	return tags, nil
}

func loadNgfw(d *schema.ResourceData, curEps []firewall.EndpointConfig) (ngfw.Info, error) {
	firewallEgressNat, err := loadEgressNat(d)
	if err != nil {
		return ngfw.Info{}, err
	}
	resEps := loadEndpoints(firewallEgressNat, d, curEps)
	allowListAccounts := setToSlice(d.Get("allowlist_accounts"))
	customerZoneIdList := setToSlice(d.Get("az_list"))
	firewallUserIdConfig, err := loadUserIdConfig(d)
	if err != nil {
		return ngfw.Info{}, err
	}
	firewallPrivateAccess, err := loadPrivateAccessConfig(d)
	if err != nil {
		return ngfw.Info{}, err
	}
	tags, err := setTags(d)
	if err != nil {
		return ngfw.Info{}, err
	}
	return ngfw.Info{
		Description:           d.Get("description").(string),
		Rulestack:             d.Get(RulestackName).(string),
		GlobalRulestack:       d.Get("global_rulestack").(string),
		MultiVpc:              d.Get("multi_vpc").(bool),
		EndpointMode:          d.Get("endpoint_mode").(string),
		LinkId:                d.Get("link_id").(string),
		LinkStatus:            d.Get("link_status").(string),
		Tags:                  tags,
		AllowListAccounts:     checkNilSlice(allowListAccounts),
		ChangeProtection:      setChangeProtection(d),
		CustomerZoneIdList:    checkNilSlice(customerZoneIdList),
		Endpoints:             resEps,
		EgressNAT:             &firewallEgressNat,
		PrivateAccess:         firewallPrivateAccess,
		UserID:                firewallUserIdConfig,
		UpdateToken:           d.Get("update_token").(string),
		DeploymentUpdateToken: d.Get("deployment_update_token").(string),
		Id:                    d.Get("firewall_id").(string),
	}, nil
}

func saveEgressNat(ctx context.Context, d *schema.ResourceData, o ngfw.ReadResponse) {
	egressNat := make([]interface{}, 0)
	egressNatMap := make(map[string]interface{})
	if o.Firewall.EgressNAT.Settings != nil {
		egressNatMap["settings"] = []interface{}{
			map[string]interface{}{
				"ip_pool_type": o.Firewall.EgressNAT.Settings.IPPoolType,
			},
		}
		if o.Firewall.EgressNAT.Settings.IPAMPoolId != nil {
			egressNatMap["ipam_pool_id"] = *o.Firewall.EgressNAT.Settings.IPAMPoolId
		}
	} else {
		egressNatMap["settings"] = nil
	}
	egressNatMap["enabled"] = o.Firewall.EgressNAT.Enabled
	egressNat = append(egressNat, egressNatMap)
	d.Set("egress_nat", egressNat)
	tflog.Info(ctx, fmt.Sprintf("Saved egress nat: %v", egressNat))
}

func saveUserIdConfig(d *schema.ResourceData, o ngfw.ReadResponse) {
	if o.Firewall.UserID == nil {
		d.Set("user_id", nil)
	} else {
		userId := make([]interface{}, 0)
		userIdMap := make(map[string]interface{})
		if o.Firewall.UserID.CustomIncludeExcludeNetwork != nil {
			includeExcludeNetwork := o.Firewall.UserID.CustomIncludeExcludeNetwork
			subnetFilterList := make([]interface{}, 0)
			for _, includeNetwork := range includeExcludeNetwork {
				subnetFilter := make(map[string]interface{})
				subnetFilter["enabled"] = includeNetwork.Enabled
				subnetFilter["name"] = includeNetwork.Name
				subnetFilter["network_address"] = includeNetwork.NetworkAddress
				subnetFilter["discovery_include"] = includeNetwork.DiscoveryInclude
				subnetFilterList = append(subnetFilterList, subnetFilter)
			}
			userIdMap["custom_include_exclude_network"] = subnetFilterList
		}
		userIdMap["user_id_status"] = o.Firewall.UserID.UserIDStatus
		userIdMap["secret_key_arn"] = o.Firewall.UserID.SecretKeyARN
		userIdMap["port"] = o.Firewall.UserID.Port
		userIdMap["collector_name"] = o.Firewall.UserID.CollectorName
		if o.Firewall.UserID.AgentName != "" {
			userIdMap["agent_name"] = o.Firewall.UserID.AgentName
		} else {
			userIdMap["agent_name"] = "agent-" + o.Firewall.Id
		}
		userIdMap["enabled"] = o.Firewall.UserID.Enabled
		userId = append(userId, userIdMap)
		d.Set("user_id", userId)
	}
}

func savePrivateAccessConfig(d *schema.ResourceData, o ngfw.ReadResponse) {
	if o.Firewall.PrivateAccess == nil {
		d.Set("private_access", nil)
	} else {
		privAccess := make([]interface{}, 0)
		privAccessMap := make(map[string]interface{})
		if o.Firewall.PrivateAccess != nil {
			privAccessMap["type"] = o.Firewall.PrivateAccess.Type
			privAccessMap["resource_id"] = o.Firewall.PrivateAccess.ResourceID
		}
		privAccess = append(privAccess, privAccessMap)
		d.Set("private_access", privAccess)
	}
}

func saveEndpoints(d *schema.ResourceData, o ngfw.ReadResponse) {
	resultEps := make([]interface{}, 0)
	responseEps := o.Firewall.Endpoints
	if len(responseEps) > 0 {
		for _, responseEp := range responseEps {
			ep := make(map[string]interface{})
			ep["endpoint_id"] = responseEp.EndpointId
			ep["egress_nat_enabled"] = responseEp.EgressNATEnabled
			ep["vpc_id"] = responseEp.VpcId
			ep["account_id"] = responseEp.AccountId
			ep["zone_id"] = responseEp.ZoneId
			ep["subnet_id"] = responseEp.SubnetId
			ep["status"] = responseEp.Status
			ep["rejected_reason"] = responseEp.RejectedReason
			ep["mode"] = responseEp.Mode
			responseCidrs := responseEp.Prefixes.PrivatePrefix.Cidrs
			cidrs := make([]interface{}, 0)
			if len(responseCidrs) > 0 {
				for _, cidr := range responseCidrs {
					cidrs = append(cidrs, cidr)
				}
				ep["prefixes"] = []interface{}{
					map[string]interface{}{
						"private_prefix": []interface{}{
							map[string]interface{}{
								"cidrs": schema.NewSet(schema.HashString, cidrs),
							},
						},
					},
				}
			}
			resultEps = append(resultEps, ep)
		}
	}
	d.Set("endpoints", resultEps)
}

func getFirewallName(tags []tag.Details) (string, error) {
	if len(tags) == 0 {
		return "", fmt.Errorf("firewall name wasn't found in tags")
	}
	for _, tag := range tags {
		if tag.Key == "FirewallName" {
			return tag.Value, nil
		}
	}
	return "", fmt.Errorf("firewall name wasn't found in tags")
}

func saveNgfw(ctx context.Context, d *schema.ResourceData, o ngfw.ReadResponse) error {
	saveEgressNat(ctx, d, o)
	saveEndpoints(d, o)
	saveUserIdConfig(d, o)
	savePrivateAccessConfig(d, o)
	stat := []interface{}{
		map[string]interface{}{
			"firewall_status":                o.Status.FirewallStatus,
			"failure_reason":                 o.Status.FailureReason,
			"rulestack_status":               o.Status.RulestackStatus,
			"device_rulestack_commit_status": o.Status.DeviceRuleStackCommitStatus,
		},
	}
	d.Set("firewall_id", o.Firewall.Id)
	fwName, err := getFirewallName(o.Firewall.Tags)
	if err != nil {
		return fmt.Errorf("firewall name cannot be empty")
	}
	d.Set("name", fwName)
	d.Set("description", o.Firewall.Description)
	d.Set(RulestackName, o.Firewall.Rulestack)
	d.Set("global_rulestack", o.Firewall.GlobalRulestack)
	d.Set("endpoint_service_name", o.Firewall.EndpointServiceName)
	d.Set("endpoint_mode", o.Firewall.EndpointMode)
	d.Set("link_id", o.Firewall.LinkId)
	d.Set("link_status", o.Firewall.LinkStatus)
	d.Set(TagsName, dumpTags(o.Firewall.Tags))
	d.Set("update_token", o.Firewall.UpdateToken)
	d.Set("deployment_update_token", o.Firewall.DeploymentUpdateToken)
	d.Set("allowlist_accounts", sliceToSet(o.Firewall.AllowListAccounts))
	d.Set("change_protection", sliceToSet(o.Firewall.ChangeProtection))
	d.Set("az_list", sliceToSet(o.Firewall.CustomerZoneIdList))
	d.Set("multi_vpc", o.Firewall.MultiVpc)
	d.Set("status", stat)
	return nil
}

// Id functions.
func buildNgfwId(a, b string) string {
	return strings.Join([]string{a, b}, IdSeparator)
}

func parseNgfwId(v string) (string, string, error) {
	tok := strings.Split(v, IdSeparator)
	if len(tok) != 2 {
		return "", "", fmt.Errorf("expecting 2 tokens, got %d", len(tok))
	}

	return tok[0], tok[1], nil
}
