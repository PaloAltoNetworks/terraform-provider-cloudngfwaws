# Retrieve the feed information every day at midnight.
resource "awsngfw_intelligent_feed" "example" {
  rulestack   = awsngfw_rulestack.r.name
  name        = "tf-feed"
  description = "Also configured by Terraform"
  url         = "https://foobar.net"
  type        = "URL_LIST"
  frequency   = "DAILY"
  time        = 0
}

resource "awsngfw_rulestack" "r" {
  name        = "terraform-rulestack"
  scope       = "Local"
  account_id  = "123456789"
  description = "Made by Terraform"
  profile_config {
    anti_spyware = "BestPractice"
  }
}
