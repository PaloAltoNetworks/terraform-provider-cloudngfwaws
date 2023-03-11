resource "cloudngfwaws_fqdn_list" "this" {
  for_each = { for fqdn_list in var.fqdn_lists: fqdn_list.name => fqdn_list }

  rulestack   = var.rulestack.name
  name        = each.value.name
  description = lookup(each.value, "description", null)
  fqdn_list = each.value.fqdn_list
  audit_comment = lookup(each.value, "audit_comment", null)
  depends_on = [
      cloudngfwaws_rulestack.this
  ]
}
