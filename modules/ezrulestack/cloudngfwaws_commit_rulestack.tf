resource "cloudngfwaws_commit_rulestack" "this" {
  count = var.commit == true ? 1 : 0
  rulestack = var.rulestack.name
  depends_on = [
          cloudngfwaws_security_rule.this
  ]
}
