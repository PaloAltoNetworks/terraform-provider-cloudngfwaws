resource "cloudngfwaws_rulestack_tag" "example" {
  rulestack = cloudngfwaws_rulestack.rs1.name
  tags = {
    Foo  = "bar",
    Tag1 = "value1",
  }
}

resource "cloudngfwaws_rulestack" "rs1" {
  name        = "terraform-rulestack"
  scope       = "Local"
  account_id  = "123456789"
  description = "Made by Terraform"
  profile_config {
    anti_spyware = "BestPractice"
  }
}
