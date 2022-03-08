resource "cloudngfwaws_security_rule" "example" {
  rulestack   = cloudngfwaws_rulestack.r.name
  rule_list   = "LocalRule"
  priority    = 3
  name        = "tf-security-rule"
  description = "Also configured by Terraform"
  source {
    cidrs = ["any"]
  }
  destination {
    cidrs = ["192.168.0.0/16"]
  }
  negate_destination = true
  applications       = ["any"]
  category {}
  action        = "Allow"
  logging       = true
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
