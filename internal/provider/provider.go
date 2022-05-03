package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

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
				"cloudngfwaws_ngfw_tag":                         dataSourceNgfwTag(),
				"cloudngfwaws_intelligent_feed":                 dataSourceIntelligentFeed(),
				"cloudngfwaws_predefined_url_categories":        dataSourcePredefinedUrlCategories(),
				"cloudngfwaws_predefined_url_category_override": dataSourcePredefinedUrlCategoryOverride(),
				"cloudngfwaws_prefix_list":                      dataSourcePrefixList(),
				"cloudngfwaws_rulestack":                        dataSourceRulestack(),
				"cloudngfwaws_rulestack_tag":                    dataSourceRulestackTag(),
				"cloudngfwaws_security_rule":                    dataSourceSecurityRule(),
				"cloudngfwaws_validate_rulestack":               dataSourceValidateRulestack(),
			},

			ResourcesMap: map[string]*schema.Resource{
				"cloudngfwaws_certificate":                      resourceCertificate(),
				"cloudngfwaws_commit_rulestack":                 resourceCommitRulestack(),
				"cloudngfwaws_custom_url_category":              resourceCustomUrlCategory(),
				"cloudngfwaws_fqdn_list":                        resourceFqdnList(),
				"cloudngfwaws_ngfw":                             resourceNgfw(),
				"cloudngfwaws_ngfw_log_profile":                 resourceNgfwLogProfile(),
				"cloudngfwaws_ngfw_tag":                         resourceNgfwTag(),
				"cloudngfwaws_intelligent_feed":                 resourceIntelligentFeed(),
				"cloudngfwaws_predefined_url_category_override": resourcePredefinedUrlCategoryOverride(),
				"cloudngfwaws_prefix_list":                      resourcePrefixList(),
				"cloudngfwaws_rulestack":                        resourceRulestack(),
				"cloudngfwaws_rulestack_tag":                    resourceRulestackTag(),
				"cloudngfwaws_security_rule":                    resourceSecurityRule(),
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
				"The hostname of the API (default: `api.us-east-1.aws.cloudngfw.com`).",
				"CLOUDNGFWAWS_HOST",
				"host",
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
		"region": {
			Type:     schema.TypeString,
			Optional: true,
			Description: addProviderParamDescription(
				"AWS region.",
				"CLOUDNGFWAWS_REGION",
				"region",
			),
		},
		"arn": {
			Type:     schema.TypeString,
			Optional: true,
			Description: addProviderParamDescription(
				"The ARN allowing both firewall and rulestack admin permissions.",
				"CLOUDNGFWAWS_ARN",
				"arn",
			),
		},
		"lfa_arn": {
			Type:     schema.TypeString,
			Optional: true,
			Description: addProviderParamDescription(
				"The ARN allowing firewall admin permissions.",
				"CLOUDNGFWAWS_LFA_ARN",
				"lfa-arn",
			),
		},
		"lra_arn": {
			Type:     schema.TypeString,
			Optional: true,
			Description: addProviderParamDescription(
				"The ARN allowing rulestack admin permissions.",
				"CLOUDNGFWAWS_LRA_ARN",
				"lra-arn",
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
			"quiet":   awsngfw.LogQuiet,
			"login":   awsngfw.LogLogin,
			"get":     awsngfw.LogGet,
			"post":    awsngfw.LogPost,
			"put":     awsngfw.LogPut,
			"delete":  awsngfw.LogDelete,
			"path":    awsngfw.LogPath,
			"send":    awsngfw.LogSend,
			"receive": awsngfw.LogReceive,
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

		con := &awsngfw.Client{
			Host:                  d.Get("host").(string),
			AccessKey:             d.Get("access_key").(string),
			SecretKey:             d.Get("secret_key").(string),
			Region:                d.Get("region").(string),
			Arn:                   d.Get("arn").(string),
			LfaArn:                d.Get("lfa_arn").(string),
			LraArn:                d.Get("lra_arn").(string),
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

		if err := con.RefreshJwts(ctx); err != nil {
			return nil, diag.FromErr(err)
		}

		return con, nil
	}
}
