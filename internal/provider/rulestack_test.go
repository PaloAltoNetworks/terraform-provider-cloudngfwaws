package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go/rule/stack"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceRulestack_LocalWithAccountId(t *testing.T) {
	if testAccAccountId == "" {
		t.Skip(TestAccAccountIdNotDefined)
	}

	name := fmt.Sprintf("tf%s", acctest.RandString(6))
	o1 := stack.Details{
		Scope:       "Local",
		AccountId:   testAccAccountId,
		Description: "This is my first description",
		Profile: stack.ProfileConfig{
			AntiSpyware: "BestPractice",
		},
	}
	o2 := stack.Details{
		Scope:       o1.Scope,
		AccountId:   testAccAccountId,
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
				Config: testAccRulestackConfig(name, o1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudngfwaws_rulestack.test", "name", name,
					),
					resource.TestCheckResourceAttr(
						"cloudngfwaws_rulestack.test", "scope", o1.Scope,
					),
					resource.TestCheckResourceAttr(
						"cloudngfwaws_rulestack.test", "account_id", o1.AccountId,
					),
					resource.TestCheckResourceAttr(
						"cloudngfwaws_rulestack.test", "description", o1.Description,
					),
				),
			},
			{
				Config: testAccRulestackConfig(name, o2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudngfwaws_rulestack.test", "name", name,
					),
					resource.TestCheckResourceAttr(
						"cloudngfwaws_rulestack.test", "scope", o2.Scope,
					),
					resource.TestCheckResourceAttr(
						"cloudngfwaws_rulestack.test", "account_id", o2.AccountId,
					),
					resource.TestCheckResourceAttr(
						"cloudngfwaws_rulestack.test", "description", o2.Description,
					),
				),
			},
		},
	})
}

func testAccRulestackConfig(name string, x stack.Details) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf(`
resource "cloudngfwaws_rulestack" "test" {
    name = %q
    scope = %q
    account_id = %q
    description = %q
    profile_config {`, name, x.Scope, x.AccountId, x.Description))

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
