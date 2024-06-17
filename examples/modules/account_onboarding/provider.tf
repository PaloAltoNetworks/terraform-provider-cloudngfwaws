provider "cloudngfwaws" {
  mp_region_host    = "abcdefg.execute-api.us-east-1.amazonaws.com/prod"
  mp_region         = "us-east-1"
  account_admin_arn = "arn:aws:iam::87654321:role/fwaas_prog_onboard"
  region            = "eu-west-2"
}


terraform {
  required_providers {
    cloudngfwaws = {
      source = "paloaltonetworks/cloudngfwaws"
    }
  }
}
