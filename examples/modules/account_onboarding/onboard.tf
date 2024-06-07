module "account_onboarding" {
  source        = "github.com/PaloAltoNetworks/terraform-provider-cloudngfwaws/modules/account_onboarding"
  account_ids   = ["12345678"]
  cft_role_name = "panw_cft_role"  
}
