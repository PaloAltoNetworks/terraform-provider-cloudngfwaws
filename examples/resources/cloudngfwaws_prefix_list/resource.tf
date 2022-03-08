resource "cloudngfwaws_prefix_list" "example" {
  rulestack   = cloudngfwaws_rulestack.r.name
  name        = "tf-prefix-list"
  description = "Also configured by Terraform"
  prefix_list = [
    "192.168.0.0",
    "10.1.5.0",
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
