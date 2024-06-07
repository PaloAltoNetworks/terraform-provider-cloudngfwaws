package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	ngfw "github.com/paloaltonetworks/cloud-ngfw-aws-go"
	"github.com/paloaltonetworks/cloud-ngfw-aws-go/api"
	"github.com/paloaltonetworks/cloud-ngfw-aws-go/ngfw/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var resourceTimeout = 120 * time.Minute

func init() {
	schema.DescriptionKind = schema.StringMarkdown

	schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
		desc := s.Description

		if s.Default != nil {
			desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
		}

		if s.Deprecated != "" {
			desc += " " + s.Deprecated
		}

		return strings.TrimSpace(desc)
	}
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: providerSchema(),

			DataSourcesMap: map[string]*schema.Resource{
				"cloudngfwaws_app_id_version":                   dataSourceAppIdVersion(),
				"cloudngfwaws_app_id_versions":                  dataSourceAppIdVersions(),
				"cloudngfwaws_certificate":                      dataSourceCertificate(),
				"cloudngfwaws_country":                          dataSourceCountry(),
				"cloudngfwaws_custom_url_category":              dataSourceCustomUrlCategory(),
				"cloudngfwaws_fqdn_list":                        dataSourceFqdnList(),
				"cloudngfwaws_ngfw":                             dataSourceNgfw(),
				"cloudngfwaws_ngfws":                            dataSourceNgfws(),
				"cloudngfwaws_ngfw_log_profile":                 dataSourceNgfwLogProfile(),
				"cloudngfwaws_intelligent_feed":                 dataSourceIntelligentFeed(),
				"cloudngfwaws_predefined_url_categories":        dataSourcePredefinedUrlCategories(),
				"cloudngfwaws_predefined_url_category_override": dataSourcePredefinedUrlCategoryOverride(),
				"cloudngfwaws_prefix_list":                      dataSourcePrefixList(),
				"cloudngfwaws_rulestack":                        dataSourceRulestack(),
				"cloudngfwaws_security_rule":                    dataSourceSecurityRule(),
				"cloudngfwaws_validate_rulestack":               dataSourceValidateRulestack(),
				"cloudngfwaws_account":                          dataSourceAccount(),
				"cloudngfwaws_accounts":                         dataSourceAccounts(),
			},

			ResourcesMap: map[string]*schema.Resource{
				"cloudngfwaws_certificate":                      resourceCertificate(),
				"cloudngfwaws_commit_rulestack":                 resourceCommitRulestack(),
				"cloudngfwaws_custom_url_category":              resourceCustomUrlCategory(),
				"cloudngfwaws_fqdn_list":                        resourceFqdnList(),
				"cloudngfwaws_ngfw":                             resourceNgfw(),
				"cloudngfwaws_ngfw_log_profile":                 resourceNgfwLogProfile(),
				"cloudngfwaws_intelligent_feed":                 resourceIntelligentFeed(),
				"cloudngfwaws_predefined_url_category_override": resourcePredefinedUrlCategoryOverride(),
				"cloudngfwaws_prefix_list":                      resourcePrefixList(),
				"cloudngfwaws_rulestack":                        resourceRulestack(),
				"cloudngfwaws_security_rule":                    resourceSecurityRule(),
				"cloudngfwaws_account":                          resourceAccount(),
				"cloudngfwaws_account_onboarding":               resourceAccountOnboarding(),
				"cloudngfwaws_account_onboarding_stack":         resourceAccountOnboardingStack(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

func providerSchema() map[string]*schema.Schema {
	protoOpts := []string{"https", "http"}

	return map[string]*schema.Schema{
		"host": {
			Type:     schema.TypeString,
			Optional: true,
			Description: addProviderParamDescription(
				"The hostname of the API (default: `api.us-east-1.aws.cloudngfw.paloaltonetworks.com`).",
				"CLOUDNGFWAWS_HOST",
				"host",
			),
		},
		"mp_region_host": {
			Type:     schema.TypeString,
			Optional: true,
			Description: addProviderParamDescription(
				"AWS management plane MP region host",
				"CLOUDNGFWAWS_MP_REGION_HOST",
				"mp_region_host",
			),
		},
		"access_key": {
			Type:     schema.TypeString,
			Optional: true,
			Description: addProviderParamDescription(
				"(Used for the initial `sts assume role`) AWS access key.",
				"CLOUDNGFWAWS_ACCESS_KEY",
				"access-key",
			),
		},
		"secret_key": {
			Type:     schema.TypeString,
			Optional: true,
			Description: addProviderParamDescription(
				"(Used for the initial `sts assume role`) AWS secret key.",
				"CLOUDNGFWAWS_SECRET_KEY",
				"secret-key",
			),
		},
		"profile": {
			Type:     schema.TypeString,
			Optional: true,
			Description: addProviderParamDescription(
				"(Used for the initial `sts assume role`) AWS PROFILE.",
				"CLOUDNGFWAWS_PROFILE",
				"profile",
			),
		},
		"sync_mode": {
			Type:     schema.TypeBool,
			Optional: true,
			Description: addProviderParamDescription(
				"Enable synchronous mode while creating resources",
				"CLOUDNGFWAWS_SYNC_MODE",
				"sync_mode",
			),
		},
		"region": {
			Type:     schema.TypeString,
			Optional: true,
			Description: addProviderParamDescription(
				"AWS region.",
				"CLOUDNGFWAWS_REGION",
				"region",
			),
		},
		"mp_region": {
			Type:     schema.TypeString,
			Optional: true,
			Description: addProviderParamDescription(
				"AWS management plane region.",
				"CLOUDNGFWAWS_MP_REGION",
				"mp_region",
			),
		},
		"arn": {
			Type:     schema.TypeString,
			Optional: true,
			Description: addProviderParamDescription(
				"The ARN allowing firewall, rulestack, and global rulestack admin permissions. Global rulestack admin permissions can be enabled only if the AWS account is onboarded by AWS Firewall Manager. Use 'lfa_arn' and 'lra_arn' if you want to enable only firewall and rulestack admin permissions.",
				"CLOUDNGFWAWS_ARN",
				"arn",
			),
		},
		"lfa_arn": {
			Type:     schema.TypeString,
			Optional: true,
			Description: addProviderParamDescription(
				"The ARN allowing firewall admin permissions. This is preferentially used over the `arn` param if both are specified.",
				"CLOUDNGFWAWS_LFA_ARN",
				"lfa-arn",
			),
		},
		"lra_arn": {
			Type:     schema.TypeString,
			Optional: true,
			Description: addProviderParamDescription(
				"The ARN allowing rulestack admin permissions. This is preferentially used over the `arn` param if both are specified.",
				"CLOUDNGFWAWS_LRA_ARN",
				"lra-arn",
			),
		},
		"gra_arn": {
			Type:     schema.TypeString,
			Optional: true,
			Description: addProviderParamDescription(
				"The ARN allowing global rulestack admin permissions. Global rulestack admin permissions can be enabled only if the AWS account is onboarded by AWS Firewall Manager. 'gra_arn' is preferentially used over the `arn` param if both are specified.",
				"CLOUDNGFWAWS_GRA_ARN",
				"gra-arn",
			),
		},
		"account_admin_arn": {
			Type:     schema.TypeString,
			Optional: true,
			Description: addProviderParamDescription(
				"The ARN allowing account admin permissions.",
				"CLOUDNGFWAWS_ACCT_ADMIN_ARN",
				"account-admin-arn",
			),
		},
		"protocol": {
			Type:     schema.TypeString,
			Optional: true,
			Description: addStringInSliceValidation(
				addProviderParamDescription(
					"The protocol (defaults to `https`).",
					"CLOUDNGFWAWS_PROTOCOL",
					"protocol",
				),
				protoOpts,
			),
			ValidateFunc: validation.StringInSlice(protoOpts, false),
		},
		"timeout": {
			Type:     schema.TypeInt,
			Optional: true,
			Description: addProviderParamDescription(
				"The timeout for any single API call (default: `30`).",
				"CLOUDNGFWAWS_TIMEOUT",
				"timeout",
			),
		},
		"resource_timeout": {
			Type:     schema.TypeInt,
			Optional: true,
			Description: addProviderParamDescription(
				"The timeout for terraform resource create/update/delete operations (default: `7200s`).",
				"CLOUDNGFWAWS_RESOURCE_TIMEOUT",
				"timeout",
			),
		},
		"headers": {
			Type:     schema.TypeMap,
			Optional: true,
			Description: addProviderParamDescription(
				"Additional HTTP headers to send with API calls.",
				"CLOUDNGFWAWS_HEADERS",
				"headers",
			),
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"skip_verify_certificate": {
			Type:     schema.TypeBool,
			Optional: true,
			Description: addProviderParamDescription(
				"Skip verifying the SSL certificate.",
				"CLOUDNGFWAWS_SKIP_VERIFY_CERTIFICATE",
				"skip-verify-certificate",
			),
		},
		"logging": {
			Type:     schema.TypeList,
			Optional: true,
			Description: addProviderParamDescription(
				"The logging options for the provider.",
				"CLOUDNGFWAWS_LOGGING",
				"logging",
			),
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"json_config_file": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Retrieve provider configuration from this JSON file.",
		},
	}
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		var lc uint32

		lm := map[string]uint32{
			"quiet":   ngfw.LogQuiet,
			"login":   ngfw.LogLogin,
			"get":     ngfw.LogGet,
			"patch":   ngfw.LogPatch,
			"post":    ngfw.LogPost,
			"put":     ngfw.LogPut,
			"delete":  ngfw.LogDelete,
			"action":  ngfw.LogAction,
			"path":    ngfw.LogPath,
			"send":    ngfw.LogSend,
			"receive": ngfw.LogReceive,
		}

		var hdrs map[string]string
		hconfig := d.Get("headers").(map[string]interface{})
		if len(hconfig) > 0 {
			hdrs = make(map[string]string)
			for key, val := range hconfig {
				hdrs[key] = val.(string)
			}
		}

		if ll := d.Get("logging").([]interface{}); len(ll) > 0 {
			for i := range ll {
				s := ll[i].(string)
				if v, ok := lm[s]; !ok {
					return nil, diag.Errorf("Unknown logging artifact specified: %s", s)
				} else {
					lc |= v
				}
			}
		}

		con := &aws.Client{
			Host:                  d.Get("host").(string),
			MPRegionHost:          d.Get("mp_region_host").(string),
			AccessKey:             d.Get("access_key").(string),
			SecretKey:             d.Get("secret_key").(string),
			Profile:               d.Get("profile").(string),
			SyncMode:              d.Get("sync_mode").(bool),
			ResourceTimeout:       d.Get("resource_timeout").(int),
			Region:                d.Get("region").(string),
			MPRegion:              d.Get("mp_region").(string),
			Arn:                   d.Get("arn").(string),
			LfaArn:                d.Get("lfa_arn").(string),
			LraArn:                d.Get("lra_arn").(string),
			GraArn:                d.Get("gra_arn").(string),
			AcctAdminArn:          d.Get("account_admin_arn").(string),
			AuthType:              aws.AuthTypeIAMRole,
			Protocol:              d.Get("protocol").(string),
			Timeout:               d.Get("timeout").(int),
			Headers:               hdrs,
			SkipVerifyCertificate: d.Get("skip_verify_certificate").(bool),
			Logging:               lc,
			AuthFile:              d.Get("json_config_file").(string),

			CheckEnvironment: true,
			Agent:            p.UserAgent("terraform-provider-cloudngfwaws", version),
		}

		if err := con.Setup(); err != nil {
			return nil, diag.FromErr(err)
		}

		con.HttpClient.Transport = logging.NewTransport("CloudNgfwAws", con.HttpClient.Transport)

		InitLogger(InfoLevel)
		api.SetLogger(Logger)

		apiClient := api.NewAPIClient(con, ctx, 5000, "", false)
		api.Logger.Infof("sync_mode:%+v", apiClient.IsSyncModeEnabled(ctx))
		resourceTimeout = time.Duration(apiClient.GetResourceTimeout(ctx)) * time.Second
		api.Logger.Infof("resource_timeout:%+v", resourceTimeout)
		return apiClient, nil
	}
}
