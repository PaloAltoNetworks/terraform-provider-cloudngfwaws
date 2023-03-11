resource "cloudngfwaws_predefined_url_category_override" "this" {
  for_each = { for predefined_url_category_override in var.predefined_url_category_overrides: predefined_url_category_override.name => predefined_url_category_override }

  rulestack   = var.rulestack.name
  name        = each.value.name
  action      = lookup(each.value, "action", null)
  depends_on = [
      cloudngfwaws_rulestack.this
  ]
}