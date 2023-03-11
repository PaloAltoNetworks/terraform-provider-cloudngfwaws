resource "cloudngfwaws_intelligent_feed" "this" {
  for_each = { for feed in var.feeds: feed.name => feed }

  rulestack   = var.rulestack.name 
  name        = each.value.name
  description = lookup(each.value, "description", null)
  certificate = lookup(each.value, "certificate", null)
  url         = each.value.url
  type        = lookup(each.value, "type", null)
  frequency   = lookup(each.value, "frequency", null)
  time        = lookup(each.value, "time", null)
  audit_comment = lookup(each.value, "audit_comment", null)
  depends_on = [
      cloudngfwaws_rulestack.this
  ]
}
