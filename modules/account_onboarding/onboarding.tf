resource "cloudngfwaws_account" "account" {
  for_each = toset(var.account_ids)
  account_id = each.value
}

resource "cloudngfwaws_account_onboarding_stack" "account_onboarding_stack" {
  for_each = toset(var.account_ids)
  account_id = each.value
  onboarding_cft = file("${path.module}/onboarding_template.yaml")
  endpoint_mode = var.endpoint_mode
  decryption_cert = var.decryption_cert
  cloudwatch_namespace = var.cloudwatch_namespace
  cloudwatch_log_group = var.cloudwatch_log_group
  auditlog_group = var.auditlog_group
  kinesis_firehose = var.kinesis_firehose
  s3_bucket = var.s3_bucket
  cft_role_name = var.cft_role_name
  trusted_account = cloudngfwaws_account.account[each.key].trusted_account
  external_id = cloudngfwaws_account.account[each.key].external_id
  sns_topic_arn = cloudngfwaws_account.account[each.key].sns_topic_arn
}

resource "cloudngfwaws_account_onboarding" "account_onboarding" {
  for_each = toset(var.account_ids)
  account_id = each.value
  depends_on = [ cloudngfwaws_account_onboarding_stack.account_onboarding_stack ]
}
