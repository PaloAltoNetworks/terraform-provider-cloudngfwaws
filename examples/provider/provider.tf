terraform {
  required_providers {
    cloudngfwaws = {
      source  = "terraform.local/local/cloudngfwaws"
      version = "1.0.0"
    }
  }
}

provider "cloudngfwaws" {
  json_config_file = "~/.cloudngfwaws_creds.json"
}
