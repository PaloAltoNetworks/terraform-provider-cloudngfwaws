terraform {
  required_providers {
    cloudngfwaws = {
      source  = "paloaltonetworks/terraform-provider-cloudngfwaws"
      version = "1.0.8"
    }
  }
}

provider "cloudngfwaws" {
  json_config_file = "~/.cloudngfwaws_creds.json"
}
