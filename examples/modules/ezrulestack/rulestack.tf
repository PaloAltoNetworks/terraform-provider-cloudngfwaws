module "rulestack" {
  source                = "./modules/ezrulestack"
  rulestack             = var.my_rulestack
  security_rules        = var.my_security_rules
  feeds                 = var.my_feeds
  prefix_lists          = var.my_prefix_lists
  fqdn_lists            = var.my_fqdn_lists
  custom_url_categories = var.my_custom_url_categories
  certificates          = var.my_certificates
  commit                = var.commit
}