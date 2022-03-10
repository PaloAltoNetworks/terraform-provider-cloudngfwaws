resource "cloudngfwaws_predefined_url_category_override" "example" {
  rulestack = cloudngfwaws_rulestack.r.name
  name      = "foobar"
  action    = "block"
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
