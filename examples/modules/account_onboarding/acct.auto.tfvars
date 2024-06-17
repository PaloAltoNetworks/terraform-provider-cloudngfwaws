account_onboarding = {
  account_ids        = ["12345678"]
  cft_role_name       = "panw_cft_role"
  endpoint_mode = "yes"
  decryption_cert = "TagBasedCertificate"
  cloudwatch_namespace = "PaloAltoCloudNGFW"
  cloudwatch_log_group = "PaloAltoCloudNGFW"
  auditlog_group = "PaloAltoCloudNGFWAuditLog"
  kinesis_firehose = "PaloAltoCloudNGFW"
  s3_bucket = "None"    
}
