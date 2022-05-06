package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go/rule/stack"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Data source.
func TestAccDataSourceRulestack(t *testing.T) {
	o1 := stack.Details{
		Description: "This is my first description",
		Profile: stack.ProfileConfig{
			AntiSpyware: "BestPractice",
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRulestackConfig("test", &o1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.cloudngfwaws_rulestack.test", "name",
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_rulestack.test", ScopeName, "Local",
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_rulestack.test", "account_id", testAccAccountId,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_rulestack.test", "account_group", testAccAccountGroup,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_rulestack.test", "description", o1.Description,
					),
				),
			},
		},
	})
}

// Resource.
func TestAccResourceRulestack(t *testing.T) {
	o1 := stack.Details{
		Description: "This is my first description",
		Profile: stack.ProfileConfig{
			AntiSpyware: "BestPractice",
		},
	}
	o2 := stack.Details{
		Description: "Second description",
		Profile: stack.ProfileConfig{
			AntiVirus: "BestPractice",
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRulestackConfig("test", &o1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudngfwaws_rulestack.test", ScopeName, "Local",
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_rulestack.test", "account_id", testAccAccountId,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_rulestack.test", "account_group", testAccAccountGroup,
					),
					resource.TestCheckResourceAttr(
						"cloudngfwaws_rulestack.test", "description", o1.Description,
					),
				),
			},
			{
				Config: testAccRulestackConfig("test", &o2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudngfwaws_rulestack.test", ScopeName, "Local",
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_rulestack.test", "account_id", testAccAccountId,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_rulestack.test", "account_group", testAccAccountGroup,
					),
					resource.TestCheckResourceAttr(
						"cloudngfwaws_rulestack.test", "description", o2.Description,
					),
				),
			},
		},
	})
}

func testAccRulestackConfig(id string, x *stack.Details) string {
	var buf strings.Builder

	name := fmt.Sprintf("tf%s", acctest.RandString(8))

	if x == nil {
		x = &stack.Details{
			Description: "Acctest description",
			Profile: stack.ProfileConfig{
				AntiSpyware: "BestPractice",
			},
		}
	}

	if id == "test" {
		buf.WriteString(`
data "cloudngfwaws_rulestack" "test" {
    name = cloudngfwaws_rulestack.test.name
}`)
	}

	buf.WriteString(fmt.Sprintf(`
resource "cloudngfwaws_rulestack" %q {
    name = %q
    scope = "Local"
    account_id = %q
    account_group = %q
    description = %q
    profile_config {`, id, name, testAccAccountId, testAccAccountGroup, x.Description))

	if x.Profile.AntiSpyware != "" {
		buf.WriteString(fmt.Sprintf("\n        anti_spyware = %q", x.Profile.AntiSpyware))
	}
	if x.Profile.AntiVirus != "" {
		buf.WriteString(fmt.Sprintf("\n        anti_virus = %q", x.Profile.AntiVirus))
	}
	if x.Profile.Vulnerability != "" {
		buf.WriteString(fmt.Sprintf("\n        vulnerability = %q", x.Profile.Vulnerability))
	}
	if x.Profile.UrlFiltering != "" {
		buf.WriteString(fmt.Sprintf("\n        url_filtering = %q", x.Profile.UrlFiltering))
	}
	if x.Profile.FileBlocking != "" {
		buf.WriteString(fmt.Sprintf("\n        file_blocking = %q", x.Profile.FileBlocking))
	}
	if x.Profile.OutboundTrustCertificate != "" {
		buf.WriteString(fmt.Sprintf("\n        outbound_trust_certificate = %q", x.Profile.OutboundTrustCertificate))
	}
	if x.Profile.OutboundUntrustCertificate != "" {
		buf.WriteString(fmt.Sprintf("\n        outbound_untrust_certificate = %q", x.Profile.OutboundUntrustCertificate))
	}

	buf.WriteString(`
    }
}`)

	return buf.String()
}
