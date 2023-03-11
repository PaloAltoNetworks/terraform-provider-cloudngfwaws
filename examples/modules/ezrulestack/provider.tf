terraform {
  required_providers {
    cloudngfwaws = {
      source  = "paloaltonetworks/cloudngfwaws"
      version = "1.0.10"
    }
  }
}

provider "cloudngfwaws" {
  json_config_file = "~/.cloudngfwaws_creds.json"
}
