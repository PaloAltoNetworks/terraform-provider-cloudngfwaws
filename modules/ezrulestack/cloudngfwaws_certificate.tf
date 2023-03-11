resource "cloudngfwaws_certificate" "this" {
  for_each = { for certificate in var.certificates: certificate.name => certificate }

  rulestack   = var.rulestack.name
  name        = each.value.name
  description = lookup(each.value, "descriptione", "")
  scope = var.rulestack.scope
  self_signed   = lookup(each.value, "self_signed", null)
  signer_arn = lookup(each.value, "signer_arn", null)
  audit_comment = lookup(each.value, "audit_comment", "")
  depends_on = [
      cloudngfwaws_rulestack.this
  ]
}