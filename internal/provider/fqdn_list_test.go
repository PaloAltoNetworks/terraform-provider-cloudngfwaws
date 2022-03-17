package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go/object/fqdn"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Data source.
func TestAccDataSourceFqdnList(t *testing.T) {
	name := fmt.Sprintf("tf%s", acctest.RandString(8))

	o1 := fqdn.Info{
		Description:  "This is my first description",
		FqdnList:     []string{"example.com"},
		AuditComment: "data source acctest",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFqdnListConfig(name, o1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_fqdn_list.test", "name", name,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_fqdn_list.test", "description", o1.Description,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_fqdn_list.test", "audit_comment", o1.AuditComment,
					),
				),
			},
		},
	})
}

// Resource.
func TestAccResourceFqdnList(t *testing.T) {
	name := fmt.Sprintf("tf%s", acctest.RandString(8))

	o1 := fqdn.Info{
		Description:  "This is my first description",
		FqdnList:     []string{"example.com"},
		AuditComment: "resource acctest",
	}
	o2 := fqdn.Info{
		Description:  "Another one goes here",
		FqdnList:     []string{"example.com", "paloaltonetworks.com"},
		AuditComment: "second audit comment",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFqdnListConfig(name, o1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudngfwaws_fqdn_list.test", "name", name,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_fqdn_list.test", "description", o1.Description,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_fqdn_list.test", "audit_comment", o1.AuditComment,
					),
				),
			},
			{
				Config: testAccFqdnListConfig(name, o2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudngfwaws_fqdn_list.test", "name", name,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_fqdn_list.test", "description", o2.Description,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_fqdn_list.test", "audit_comment", o2.AuditComment,
					),
				),
			},
		},
	})
}

func testAccFqdnListConfig(name string, x fqdn.Info) string {
	var buf strings.Builder

	buf.WriteString(testAccRulestackConfig("r", nil))

	buf.WriteString(fmt.Sprintf(`
data "cloudngfwaws_fqdn_list" "test" {
    %s = cloudngfwaws_rulestack.r.name
    name = cloudngfwaws_fqdn_list.test.name
}

resource "cloudngfwaws_fqdn_list" "test" {
    %s = cloudngfwaws_rulestack.r.name
    name = %q
    description = %q
    fqdn_list = %s
    audit_comment = %q
}`, RulestackName, RulestackName, name, x.Description, sliceToString(x.FqdnList), x.AuditComment))

	return buf.String()
}
