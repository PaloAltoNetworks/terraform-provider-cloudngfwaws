resource "awsngfw_custom_url_category" "example" {
  rulestack   = awsngfw_rulestack.r.name
  name        = "tf-custom-category"
  description = "Also configured by Terraform"
  url_list = [
    "example.com",
    "paloaltonetworks.com",
    "foobar.org",
  ]
  action = "alert"
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
