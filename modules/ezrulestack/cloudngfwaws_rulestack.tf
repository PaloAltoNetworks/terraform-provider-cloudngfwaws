resource "cloudngfwaws_rulestack" "this" {
  name        = var.rulestack.name
  scope       = var.rulestack.scope
  account_id  = lookup(var.rulestack, "account_id", null)
  account_group = lookup(var.rulestack, "account", null)
  description = lookup(var.rulestack, "description", null)
  profile_config {  
    anti_spyware                  = lookup(var.rulestack.profile_config, "anti_spyware", null)
    anti_virus                    = lookup(var.rulestack.profile_config, "anti_virus", null)
    vulnerability                 = lookup(var.rulestack.profile_config, "vulnerability", null)
    url_filtering                 = lookup(var.rulestack.profile_config, "url_filtering", null)
    file_blocking                 = lookup(var.rulestack.profile_config, "file_blocking", null)
    outbound_trust_certificate    = lookup(var.rulestack.profile_config, "outbound_trust_certificate", null)
    outbound_untrust_certificate  = lookup(var.rulestack.profile_config, "outbound_untrust_certificate", null)
  }
  lookup_x_forwarded_for = var.rulestack.lookup_x_forwarded_for
}
