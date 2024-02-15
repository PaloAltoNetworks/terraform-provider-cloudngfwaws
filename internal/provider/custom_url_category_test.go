package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go/api/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Data source.
func TestAccDataSourceCustomUrlCategory(t *testing.T) {
	name := fmt.Sprintf("tf%s", acctest.RandString(8))

	o1 := url.Info{
		Description:  "Data source description",
		UrlList:      []string{"example.com", "foobar.org"},
		Action:       "allow",
		AuditComment: "data source audit_comment",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomUrlCategoryConfig(name, o1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_custom_url_category.test", "name", name,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_custom_url_category.test", "description", o1.Description,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_custom_url_category.test", "audit_comment", o1.AuditComment,
					),
				),
			},
		},
	})
}

// Resource.
func TestAccResourceCustomUrlCategory(t *testing.T) {
	name := fmt.Sprintf("tf%s", acctest.RandString(8))

	o1 := url.Info{
		Description:  "Resource description",
		UrlList:      []string{"example.com", "foobar.org"},
		Action:       "allow",
		AuditComment: "resource audit_comment",
	}
	o2 := url.Info{
		Description:  "Second description",
		UrlList:      []string{"example.com", "foobar.org", "library.net"},
		Action:       "block",
		AuditComment: "updated for second time",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomUrlCategoryConfig(name, o1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudngfwaws_custom_url_category.test", "name", name,
					),
					resource.TestCheckResourceAttr(
						"cloudngfwaws_custom_url_category.test", "description", o1.Description,
					),
					resource.TestCheckResourceAttr(
						"cloudngfwaws_custom_url_category.test", "action", o1.Action,
					),
					resource.TestCheckResourceAttr(
						"cloudngfwaws_custom_url_category.test", "audit_comment", o1.AuditComment,
					),
				),
			},
			{
				Config: testAccCustomUrlCategoryConfig(name, o2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudngfwaws_custom_url_category.test", "name", name,
					),
					resource.TestCheckResourceAttr(
						"cloudngfwaws_custom_url_category.test", "description", o2.Description,
					),
					resource.TestCheckResourceAttr(
						"cloudngfwaws_custom_url_category.test", "action", o2.Action,
					),
					resource.TestCheckResourceAttr(
						"cloudngfwaws_custom_url_category.test", "audit_comment", o2.AuditComment,
					),
				),
			},
		},
	})
}

func testAccCustomUrlCategoryConfig(name string, x url.Info) string {
	var buf strings.Builder

	buf.WriteString(testAccRulestackConfig("r", nil))

	buf.WriteString(fmt.Sprintf(`
data "cloudngfwaws_custom_url_category" "test" {
    %s = cloudngfwaws_rulestack.r.name
    name = cloudngfwaws_custom_url_category.test.name
}

resource "cloudngfwaws_custom_url_category" "test" {
    %s = cloudngfwaws_rulestack.r.name
    name = %q
    description = %q
    url_list = %s
    action = %q
    audit_comment = %q
}`, RulestackName, RulestackName, name, x.Description, sliceToString(x.UrlList), x.Action, x.AuditComment))

	return buf.String()
}
