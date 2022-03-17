package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go/object/certificate"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Data source.
func TestAccDataSourceCertificate(t *testing.T) {
	name := fmt.Sprintf("tf%s", acctest.RandString(8))

	o1 := certificate.Info{
		Description:  "This is my first description",
		SelfSigned:   true,
		AuditComment: "data source acctest",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig(name, o1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_certificate.test", "name", name,
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_certificate.test", "self_signed", "true",
					),
					resource.TestCheckResourceAttr(
						"data.cloudngfwaws_certificate.test", "audit_comment", o1.AuditComment,
					),
				),
			},
		},
	})
}

// Resource.
func TestAccResourceCertificate(t *testing.T) {
	name := fmt.Sprintf("tf%s", acctest.RandString(8))

	o1 := certificate.Info{
		Description:  "This is my first description",
		SelfSigned:   true,
		AuditComment: "data source acctest",
	}
	o2 := certificate.Info{
		Description:  "Another one goes here",
		SignerArn:    "arn:123456789",
		AuditComment: "data source acctest second time around",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig(name, o1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudngfwaws_certificate.test", "name", name,
					),
					resource.TestCheckResourceAttr(
						"cloudngfwaws_certificate.test", "self_signed", "true",
					),
					resource.TestCheckResourceAttr(
						"cloudngfwaws_certificate.test", "signer_arn", "",
					),
					resource.TestCheckResourceAttr(
						"cloudngfwaws_certificate.test", "audit_comment", o1.AuditComment,
					),
				),
			},
			{
				Config: testAccCertificateConfig(name, o2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudngfwaws_certificate.test", "name", name,
					),
					resource.TestCheckResourceAttr(
						"cloudngfwaws_certificate.test", "self_signed", "false",
					),
					resource.TestCheckResourceAttr(
						"cloudngfwaws_certificate.test", "signer_arn", o2.SignerArn,
					),
					resource.TestCheckResourceAttr(
						"cloudngfwaws_certificate.test", "audit_comment", o2.AuditComment,
					),
				),
			},
		},
	})
}

func testAccCertificateConfig(name string, x certificate.Info) string {
	var buf strings.Builder

	buf.WriteString(testAccRulestackConfig("r", nil))

	buf.WriteString(fmt.Sprintf(`
data "cloudngfwaws_certificate" "test" {
    %s = cloudngfwaws_rulestack.r.name
    name = cloudngfwaws_certificate.test.name
}

resource "cloudngfwaws_certificate" "test" {
    %s = cloudngfwaws_rulestack.r.name
    name = %q
    description = %q
    audit_comment = %q
`, RulestackName, RulestackName, name, x.Description, x.AuditComment))

	if x.SelfSigned {
		buf.WriteString("    self_signed = true\n")
	}

	if x.SignerArn != "" {
		buf.WriteString(fmt.Sprintf("    signer_arn = %q\n", x.SignerArn))
	}
	buf.WriteString("\n}")

	return buf.String()
}
