package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go/api/prefix"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Data source.
func TestAccDataSourcePrefixList(t *testing.T) {
	name := fmt.Sprintf("tf%s", acctest.RandString(8))

	o1 := prefix.Info{
		Description:  "This is my first description",
		PrefixList:   []string{"192.168.0.0"},
		AuditComment: "data source acctest",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixList(name, o1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_prefix_list.test", "name", name,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_prefix_list.test", "description", o1.Description,
					),
				),
			},
		},
	})
}

// Resource.
func TestAccResourcePrefixList(t *testing.T) {
	name := fmt.Sprintf("tf%s", acctest.RandString(8))

	o1 := prefix.Info{
		Description:  "This is my first description",
		PrefixList:   []string{"192.168.0.0"},
		AuditComment: "resource acctest",
	}
	o2 := prefix.Info{
		Description:  "Another description",
		PrefixList:   []string{"192.168.1.0", "10.2.3.0"},
		AuditComment: "second audit comment",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixList(name, o1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_prefix_list.test", "name", name,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_prefix_list.test", "description", o1.Description,
					),
				),
			},
			{
				Config: testAccPrefixList(name, o2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_prefix_list.test", "name", name,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_prefix_list.test", "description", o2.Description,
					),
				),
			},
		},
	})
}

func testAccPrefixList(name string, x prefix.Info) string {
	var buf strings.Builder

	buf.WriteString(testAccRulestackConfig("r", nil))

	buf.WriteString(fmt.Sprintf(`
data "cloudngfwaws_prefix_list" "test" {
    %s = cloudngfwaws_rulestack.r.name
    name = cloudngfwaws_prefix_list.test.name
}

resource "cloudngfwaws_prefix_list" "test" {
    %s = cloudngfwaws_rulestack.r.name
    name = %q
    description = %q
    prefix_list = %s
    audit_comment = %q
}`, RulestackName, RulestackName, name, x.Description, sliceToString(x.PrefixList), x.AuditComment))

	return buf.String()
}
