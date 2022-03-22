# Retrieve the feed information every day at midnight.
resource "cloudngfwaws_intelligent_feed" "example" {
  rulestack   = cloudngfwaws_rulestack.r.name
  name        = "tf-feed"
  description = "Also configured by Terraform"
  url         = "https://foobar.net"
  type        = "URL_LIST"
  frequency   = "DAILY"
  time        = 0
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
