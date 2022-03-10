data "cloudngfwaws_predefined_url_category_override" "example" {
  rulestack = cloudngfwaws_rulestack.r.name
  name      = "foobar"
}

resource "cloudngfwaws_rulestack" "r" {
  name        = "my-rulestack"
  scope       = "Local"
  account_id  = "12345"
  description = "Made by Terraform"
  profile_config {
    anti_spyware = "BestPractice"
  }
}
