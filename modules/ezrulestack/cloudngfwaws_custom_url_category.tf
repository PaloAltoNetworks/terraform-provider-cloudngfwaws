resource "cloudngfwaws_custom_url_category" "this" {
  for_each = { for custom_url_category in var.custom_url_categories: custom_url_category.name => custom_url_category }

  rulestack   = var.rulestack.name
  name        = each.value.name
  description = lookup(each.value, "description", null)
  url_list    = each.value.url_list
  action      = lookup(each.value, "action", null)
  audit_comment = lookup(each.value, "audit_comment", null)
  depends_on = [
      cloudngfwaws_rulestack.this
  ]
}