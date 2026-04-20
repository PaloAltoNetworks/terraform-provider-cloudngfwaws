provider "cloudngfwaws" {
  mp_region_host    = "api-devint-1.us-east-1.ngfwaas.com"
  mp_region         = "us-east-1"
  account_admin_arn = "arn:aws:iam::228220987174:role/jp-programmatic-access-role"
  region            = "us-east-2"
}


terraform {
  required_providers {
    cloudngfwaws = {
      source = "paloaltonetworks/cloudngfwaws"
    }
  }
}
