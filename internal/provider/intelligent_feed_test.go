package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go/api/feed"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Data source.
func TestAccDataSourceIntelligentFeed(t *testing.T) {
	name := fmt.Sprintf("tf%s", acctest.RandString(8))

	o1 := feed.Info{
		Description:  "This is my first description",
		Url:          "https://example.com",
		Type:         "URL_LIST",
		Frequency:    "DAILY",
		Time:         0,
		AuditComment: "data source acctest",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIntelligentFeed(name, o1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_intelligent_feed.test", "name", name,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_intelligent_feed.test", "description", o1.Description,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_intelligent_feed.test", "url", o1.Url,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_intelligent_feed.test", "type", o1.Type,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_intelligent_feed.test", "frequency", o1.Frequency,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_intelligent_feed.test", "time", fmt.Sprintf("%d", o1.Time),
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_intelligent_feed.test", "audit_comment", o1.AuditComment,
					),
				),
			},
		},
	})
}

// Resource.
func TestAccResourceIntelligentFeed(t *testing.T) {
	name := fmt.Sprintf("tf%s", acctest.RandString(8))

	o1 := feed.Info{
		Description:  "This is my first description",
		Url:          "https://example.com",
		Type:         "URL_LIST",
		Frequency:    "DAILY",
		Time:         acctest.RandIntRange(0, 24),
		AuditComment: "resource acctest",
	}
	o2 := feed.Info{
		Description:  "This is my first description",
		Url:          "https://foobar.net/list",
		Type:         "URL_LIST",
		Frequency:    "HOURLY",
		Time:         0,
		AuditComment: "second audit comment",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIntelligentFeed(name, o1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_intelligent_feed.test", "name", name,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_intelligent_feed.test", "description", o1.Description,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_intelligent_feed.test", "url", o1.Url,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_intelligent_feed.test", "type", o1.Type,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_intelligent_feed.test", "frequency", o1.Frequency,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_intelligent_feed.test", "time", fmt.Sprintf("%d", o1.Time),
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_intelligent_feed.test", "audit_comment", o1.AuditComment,
					),
				),
			},
			{
				Config: testAccIntelligentFeed(name, o2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_intelligent_feed.test", "name", name,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_intelligent_feed.test", "description", o2.Description,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_intelligent_feed.test", "url", o2.Url,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_intelligent_feed.test", "type", o2.Type,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_intelligent_feed.test", "frequency", o2.Frequency,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_intelligent_feed.test", "time", fmt.Sprintf("%d", o2.Time),
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_intelligent_feed.test", "audit_comment", o2.AuditComment,
					),
				),
			},
		},
	})
}

func testAccIntelligentFeed(name string, x feed.Info) string {
	var buf strings.Builder

	buf.WriteString(testAccRulestackConfig("r", nil))

	buf.WriteString(fmt.Sprintf(`
data "cloudngfwaws_intelligent_feed" "test" {
    %s = cloudngfwaws_rulestack.r.name
    name = cloudngfwaws_intelligent_feed.test.name
}

resource "cloudngfwaws_intelligent_feed" "test" {
    %s = cloudngfwaws_rulestack.r.name
    name = %q
    description = %q
    url = %q
    type = %q
    frequency = %q
    time = %d
    audit_comment = %q
}`, RulestackName, RulestackName, name, x.Description, x.Url, x.Type, x.Frequency, x.Time, x.AuditComment))

	return buf.String()
}
