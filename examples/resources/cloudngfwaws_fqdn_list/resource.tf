resource "cloudngfwaws_fqdn_list" "example" {
  rulestack   = cloudngfwaws_rulestack.r.name
  name        = "tf-fqdn-list"
  description = "Also configured by Terraform"
  fqdn_list = [
    "example.com",
    "foobar.org",
  ]
  audit_comment = "initial config"
}

resource "cloudngfwaws_rulestack" "r" {
  name        = "terraform-rulestack"
  scope       = "Local"
  account_id  = "123456789"
  description = "Made by Terraform"
  profile_config {
    anti_spyware = "BestPractice"
  }
}
