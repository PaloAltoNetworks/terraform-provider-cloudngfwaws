resource "cloudngfwaws_prefix_list" "this" {
  for_each = { for prefix_list in var.prefix_lists: prefix_list.name => prefix_list }

  rulestack   = var.rulestack.name
  name        = each.value.name
  description = lookup(each.value, "description", null)
  prefix_list = each.value.prefix_list
  audit_comment = lookup(each.value, "audit_comment", null)
  depends_on = [
      cloudngfwaws_rulestack.this
  ]
}
