package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go/v2/api"
	lp "github.com/paloaltonetworks/cloud-ngfw-aws-go/v2/api/logprofile"

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

func validateNgfwLogProfile(ctx context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if diff.Get("account_id").(string) != "" || diff.Get("cloud_watch_metric_namespace").(string) != "" || len(diff.Get("cloudwatch_metric_fields").([]interface{})) > 0 {
		if diff.Get("account_id").(string) == "" || diff.Get("cloud_watch_metric_namespace").(string) == "" {
			return fmt.Errorf("if cloudwatch_metric_fields or account_id or cloud_watch_metric_namespace is set, both account_id and cloud_watch_metric_namespace must be set \nOr if you are using an old deployment please use provider version 2.0.20 or below")
		}
	}
	return nil
}

func readNgfwLogProfileDataSource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)

	firewallId := d.Get("firewall_id").(string)

	req := lp.ReadInput{
		FirewallId: firewallId,
	}

	res, err := svc.ReadFirewallLogProfile(ctx, req)
	if err != nil {
		if isObjectNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(buildNgfwLogProfileId("log_profile", firewallId))

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

		Schema:        ngfwLogProfileSchema(true, nil),
		CustomizeDiff: validateNgfwLogProfile,
	}
}

func createUpdateNgfwLogProfile(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)
	o := loadNgfwLogProfile(d)

	if err := svc.UpdateFirewallLogProfile(ctx, o); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(buildNgfwLogProfileId("log_profile", o.FirewallId))

	return readNgfwLogProfile(ctx, d, meta)
}

func readNgfwLogProfile(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*api.ApiClient)

	firewallId := d.Get("firewall_id").(string)

	req := lp.ReadInput{
		FirewallId: firewallId,
	}

	res, err := svc.ReadFirewallLogProfile(ctx, req)
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
	cloudwatch_metric_fields := []string{"Dataplane_CPU_Utilization", "Dataplane_Packet_Buffer_Utilization", "Connection_Per_Second",
		"Session_Throughput_Kbps", "Session_Throughput_Pps", "Session_Active", "Session_Utilization",
		"BytesIn", "BytesOut", "PktsIn", "PktsOut"}
	ans := map[string]*schema.Schema{
		"ngfw": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of the NGFW.",
			ForceNew:    true,
		},
		"firewall_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The Firewall Id for the NGFW.",
			ForceNew:    true,
		},
		"account_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The unique ID of the account.",
		},
		"cloud_watch_metric_namespace": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The CloudWatch metric namespace.",
		},
		"advanced_threat_log": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Enable advanced threat logging.",
		},
		"cloudwatch_metric_fields": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Cloudwatch metric fields.",
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				Description:  addStringInSliceValidation("Allowed metrics fields:", cloudwatch_metric_fields),
				ValidateFunc: validation.StringInSlice(cloudwatch_metric_fields, false),
			},
		},
		"log_destination": {
			Type:        schema.TypeList,
			Optional:    true,
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
		"log_config": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Log configuration details.",
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"log_destination": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "The log destination details.",
					},
					"log_destination_type": {
						Type:         schema.TypeString,
						Required:     true,
						Description:  addStringInSliceValidation("The log destination type.", destinationTypes),
						ValidateFunc: validation.StringInSlice(destinationTypes, false),
					},
					"log_type": {
						Type: schema.TypeSet,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Required:    true,
						Description: "The list of different log types that are wanted",
					},
					"role_type": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Type of Role for log configuration",
					},
					"account_id": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Type of Role for log configuration",
					},
				},
			},
		},
		"update_token": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The update token.",
		},
		"region": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The region of the NGFW.",
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
	logProfile := lp.Info{}

	list := d.Get("log_config").([]interface{})
	if len(list) > 0 {
		x := list[0].(map[string]interface{})
		logConfig := lp.LogConfig{}
		logConfig.LogDestination = x["log_destination"].(string)
		logConfig.LogDestinationType = x["log_destination_type"].(string)
		logConfig.LogType = setToSlice(x["log_type"])
		if x["role_type"] != nil {
			logConfig.RoleType = x["role_type"].(string)
		}
		if x["account_id"] != nil {
			logConfig.AccountId = x["account_id"].(string)
		}
		if logConfig.LogDestination == "" && logConfig.LogDestinationType == "" && len(logConfig.LogType) == 0 {
			logProfile.LogConfig = nil
		} else {
			logProfile.LogConfig = &logConfig
		}
	}
	cwMetrics := lp.CloudwatchMetrics{}
	if d.Get("account_id") != nil {
		cwMetrics.AccountId = d.Get("account_id").(string)
	}
	if d.Get("cloud_watch_metric_namespace") != nil {
		cwMetrics.Namespace = d.Get("cloud_watch_metric_namespace").(string)
	}

	if len(d.Get("cloudwatch_metric_fields").([]interface{})) > 0 {
		cwMetrics.Metrics = setToSlice(d.Get("cloudwatch_metric_fields"))
	}
	if cwMetrics.Namespace == "" && len(cwMetrics.Metrics) == 0 && cwMetrics.AccountId == "" {
		logProfile.CloudwatchMetrics = nil
	} else {
		logProfile.CloudwatchMetrics = &cwMetrics
	}
	metricFieldList := d.Get("cloudwatch_metric_fields").([]interface{})
	metricFields := make([]string, 0)
	if len(metricFieldList) > 0 {
		for _, metricField := range metricFieldList {
			metricFields = append(metricFields, metricField.(string))
		}
		logProfile.CloudWatchMetricsFields = metricFields
	}
	logProfile.AdvancedThreatLog = d.Get("advanced_threat_log").(bool)
	logProfile.Region = d.Get("region").(string)
	logProfile.UpdateToken = d.Get("update_token").(string)
	logProfile.FirewallId = d.Get("firewall_id").(string)

	return logProfile
}
func saveNgfwLogProfile(d *schema.ResourceData, o lp.Info) {
	if o.LogConfig != nil {
		logConfig := make([]interface{}, 0)
		logConfigMap := make(map[string]interface{})
		logConfigMap["log_destination"] = o.LogConfig.LogDestination
		logConfigMap["log_destination_type"] = o.LogConfig.LogDestinationType
		logConfigMap["log_type"] = sliceToSet(o.LogConfig.LogType)
		if o.LogConfig.RoleType != "" {
			logConfigMap["role_type"] = o.LogConfig.RoleType
		}
		if o.LogConfig.AccountId != "" {
			logConfigMap["account_id"] = o.LogConfig.AccountId
		}
		logConfig = append(logConfig, logConfigMap)
		d.Set("log_config", logConfig)
	}
	if o.CloudwatchMetrics != nil {
		d.Set("cloud_watch_metric_namespace", o.CloudwatchMetrics.Namespace)
		d.Set("account_id", o.CloudwatchMetrics.AccountId)
		if o.CloudwatchMetrics.Metrics != nil {
			d.Set("cloudwatch_metric_fields", o.CloudwatchMetrics.Metrics)
		}
	}
	d.Set("firewall_id", o.FirewallId)
	d.Set("region", o.Region)
	d.Set("advanced_threat_log", o.AdvancedThreatLog)
	d.Set("update_token", o.UpdateToken)
}

// Id functions.
func buildNgfwLogProfileId(a, b string) string {
	return strings.Join([]string{a, b}, IdSeparator)
}
