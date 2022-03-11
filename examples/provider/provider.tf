terraform {
  required_providers {
    cloudngfwaws = {
      source  = "paloaltonetworks/terraform-provider-cloudngfwaws"
      version = "1.0.0"
    }
  }
}

provider "cloudngfwaws" {
  region           = "eu-west-1"
  json_config_file = "~/.cloudngfwaws_creds.json"
}
