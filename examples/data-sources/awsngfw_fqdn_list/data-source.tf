data "awsngfw_fqdn_list" "example" {
  rulestack = awsngfw_rulestack.r.name
  name      = "foobar"
}

resource "awsngfw_rulestack" "r" {
  name        = "my-rulestack"
  scope       = "Local"
  account_id  = "12345"
  description = "Made by Terraform"
  profile_config {
    anti_spyware = "BestPractice"
  }
}
