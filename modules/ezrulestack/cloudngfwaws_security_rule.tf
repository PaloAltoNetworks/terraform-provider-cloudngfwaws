resource "cloudngfwaws_security_rule" "this" {
  for_each = { for security_rule in var.security_rules: security_rule.name => security_rule }
  rulestack   = var.rulestack.name
  rule_list   = each.value.rule_list
  priority    = each.value.priority
  name        = each.value.name
  description = lookup(each.value, "description", null)
  enabled     = lookup(each.value, "enabled", null)
  dynamic "source" {
    for_each = length(keys(lookup(each.value, "source", {}))) == 0 ? [] : [lookup(each.value, "source", {})]
    content {
        cidrs         = lookup(source.value, "cidrs", [])
        countries     = lookup(source.value, "countries", [])
        feeds         = lookup(source.value, "feeds", [])
        prefix_lists  = lookup(source.value, "prefix_list", [])
    }
  }
  negate_source = lookup(each.value, "negate_source", null)
  dynamic "destination" {
    for_each = length(keys(lookup(each.value, "destination", {}))) == 0 ? [] : [lookup(each.value, "destination", {})]
    content {
      cidrs         = lookup(destination.value, "cidrs", ["192.168.0.0/16"])
      countries     = lookup(destination.value, "countries", [])
      feeds         = lookup(destination.value, "feeds", [])
      prefix_lists  = lookup(destination.value, "prefix_list", [])
      fqdn_lists  = lookup(destination.value, "fqdn_list", [])
    }
  }
  negate_destination = lookup(each.value, "negate_destination", null)
  dynamic "category" {
    for_each = length(keys(lookup(each.value, "category", {}))) == 0 ? [] : [lookup(each.value, "category", {})]
    content {
        feeds        = lookup(category.value, "feeds", [])
        url_category_names  = lookup(category.value, "url_category_names", [])
    }
  }
  applications       = lookup(each.value, "applications", [])
  protocol = lookup(each.value, "protocol", null)
  prot_port_list     = lookup(each.value, "prot_port_list", [])
  action        = each.value.action
  logging       = lookup(each.value, "logging", null)
  decryption_rule_type  = lookup(each.value, "decryption_rule_type", null)
  audit_comment = lookup(each.value, "audit_comment", null)
  depends_on = [
    cloudngfwaws_intelligent_feed.this,
    cloudngfwaws_prefix_list.this,
    cloudngfwaws_fqdn_list.this,
    cloudngfwaws_custom_url_category.this,
    cloudngfwaws_predefined_url_category_override.this,
    cloudngfwaws_rulestack.this
  ]
}
