resource "awsngfw_certificate" "example" {
  rulestack     = awsngfw_rulestack.r.name
  name          = "tf-cert"
  description   = "Also configured by Terraform"
  self_signed   = true
  audit_comment = "initial config"
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
