package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go/v2/api/security"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Data source.
func TestAccDataSourceSecurityRule(t *testing.T) {
	priority := acctest.RandIntRange(1, 99)

	o1 := security.Details{
		Name:        fmt.Sprintf("tf%s", acctest.RandString(8)),
		Description: "This is my first description",
		Enabled:     true,
		Source: security.SourceDetails{
			Cidrs: []string{"any"},
		},
		Destination: security.DestinationDetails{
			Cidrs: []string{"192.168.0.0/16"},
		},
		NegateDestination: true,
		Applications:      []string{"any"},
		Protocol:          "application-default",
		AuditComment:      "data source acctest",
		Action:            "Allow",
		Logging:           true,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityRuleConfig(priority, o1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "priority", fmt.Sprintf("%d", priority),
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "name", o1.Name,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "description", o1.Description,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "enabled", fmt.Sprintf("%t", o1.Enabled),
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "negate_source", fmt.Sprintf("%t", o1.NegateSource),
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "negate_destination", fmt.Sprintf("%t", o1.NegateDestination),
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "protocol", o1.Protocol,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "audit_comment", o1.AuditComment,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "action", o1.Action,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "logging", fmt.Sprintf("%t", o1.Logging),
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "decryption_rule_type", o1.DecryptionRuleType,
					),
				),
			},
		},
	})
}

// Resource.
func TestAccResourceSecurityRule(t *testing.T) {
	priority := acctest.RandIntRange(1, 99)

	o1 := security.Details{
		Name:        fmt.Sprintf("tf%s", acctest.RandString(8)),
		Description: "This is my first description",
		Enabled:     true,
		Source: security.SourceDetails{
			Cidrs: []string{"any"},
		},
		Destination: security.DestinationDetails{
			Cidrs: []string{"192.168.0.0/16"},
		},
		NegateDestination: true,
		Applications:      []string{"any"},
		Protocol:          "application-default",
		AuditComment:      "first audit comment",
		Action:            "Allow",
		Logging:           true,
	}
	o2 := security.Details{
		Name:        fmt.Sprintf("tf%s", acctest.RandString(8)),
		Description: "Worlds apart",
		Enabled:     false,
		Source: security.SourceDetails{
			Cidrs: []string{"192.168.0.0/16"},
		},
		NegateSource: true,
		Destination: security.DestinationDetails{
			Cidrs: []string{"any"},
		},
		Applications:       []string{"any"},
		Protocol:           "application-default",
		AuditComment:       "second times the charm",
		Action:             "DenySilent",
		Logging:            false,
		DecryptionRuleType: "SSLOutboundInspection",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityRuleConfig(priority, o1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "priority", fmt.Sprintf("%d", priority),
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "name", o1.Name,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "description", o1.Description,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "enabled", fmt.Sprintf("%t", o1.Enabled),
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "negate_source", fmt.Sprintf("%t", o1.NegateSource),
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "negate_destination", fmt.Sprintf("%t", o1.NegateDestination),
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "protocol", o1.Protocol,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "audit_comment", o1.AuditComment,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "action", o1.Action,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "logging", fmt.Sprintf("%t", o1.Logging),
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "decryption_rule_type", o1.DecryptionRuleType,
					),
				),
			},
			{
				Config: testAccSecurityRuleConfig(priority, o2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "priority", fmt.Sprintf("%d", priority),
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "name", o2.Name,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "description", o2.Description,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "enabled", fmt.Sprintf("%t", o2.Enabled),
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "negate_source", fmt.Sprintf("%t", o2.NegateSource),
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "negate_destination", fmt.Sprintf("%t", o2.NegateDestination),
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "protocol", o2.Protocol,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "audit_comment", o2.AuditComment,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "action", o2.Action,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "logging", fmt.Sprintf("%t", o2.Logging),
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_security_rule.test", "decryption_rule_type", o2.DecryptionRuleType,
					),
				),
			},
		},
	})
}

func testAccSecurityRuleConfig(priority int, x security.Details) string {
	var buf strings.Builder
	var src, dst, cat strings.Builder

	buf.WriteString(testAccRulestackConfig("r", nil))

	// Source.
	src.WriteString("    source {\n")
	if x.Source.Cidrs != nil {
		src.WriteString(fmt.Sprintf("        cidrs = %s\n", sliceToString(x.Source.Cidrs)))
	}
	if x.Source.Countries != nil {
		src.WriteString(fmt.Sprintf("        countries = %s\n", sliceToString(x.Source.Countries)))
	}
	if x.Source.Feeds != nil {
		src.WriteString(fmt.Sprintf("        feeds = %s\n", sliceToString(x.Source.Feeds)))
	}
	if x.Source.PrefixLists != nil {
		src.WriteString(fmt.Sprintf("        prefix_lists = %s\n", sliceToString(x.Source.PrefixLists)))
	}
	src.WriteString("    }\n")

	// Destination.
	dst.WriteString("    destination {\n")
	if x.Destination.Cidrs != nil {
		dst.WriteString(fmt.Sprintf("        cidrs = %s\n", sliceToString(x.Destination.Cidrs)))
	}
	if x.Destination.Countries != nil {
		dst.WriteString(fmt.Sprintf("        countries = %s\n", sliceToString(x.Destination.Countries)))
	}
	if x.Destination.Feeds != nil {
		dst.WriteString(fmt.Sprintf("        feeds = %s\n", sliceToString(x.Destination.Feeds)))
	}
	if x.Destination.PrefixLists != nil {
		dst.WriteString(fmt.Sprintf("        prefix_lists = %s\n", sliceToString(x.Destination.PrefixLists)))
	}
	if x.Destination.FqdnLists != nil {
		dst.WriteString(fmt.Sprintf("        fqdn_lists = %s\n", sliceToString(x.Destination.FqdnLists)))
	}
	dst.WriteString("    }\n")

	// Category.
	cat.WriteString("    category {\n")
	if x.Category.UrlCategoryNames != nil {
		cat.WriteString(fmt.Sprintf("        url_category_names = %s\n", sliceToString(x.Category.UrlCategoryNames)))
	}
	if x.Category.Feeds != nil {
		cat.WriteString(fmt.Sprintf("        feeds = %s\n", sliceToString(x.Category.Feeds)))
	}
	cat.WriteString("    }\n")

	buf.WriteString(fmt.Sprintf(`
data "cloudngfwaws_security_rule" "test" {
    %s = cloudngfwaws_rulestack.r.name
    %s = cloudngfwaws_security_rule.test.%s
    priority = cloudngfwaws_security_rule.test.priority
}

resource "cloudngfwaws_security_rule" "test" {
    %s = cloudngfwaws_rulestack.r.name
    %s = "LocalRule"
    priority = %d
    name = %q
    description = %q
    enabled = %t
%s
    negate_source = %t
%s
    negate_destination = %t
    applications = %s
%s
    protocol = %q
    audit_comment = %q
    action = %q
    logging = %t
    decryption_rule_type = %q
}`, RulestackName, RuleListName, RuleListName, RulestackName, RuleListName, priority, x.Name, x.Description, x.Enabled, src.String(), x.NegateSource, dst.String(), x.NegateDestination, sliceToString(x.Applications), cat.String(), x.Protocol, x.AuditComment, x.Action, x.Logging, x.DecryptionRuleType))

	return buf.String()
}
