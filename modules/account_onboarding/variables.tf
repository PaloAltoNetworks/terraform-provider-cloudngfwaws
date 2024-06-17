variable "account_ids" {
  description = "List of client account IDs"
  type = list(string)
}

variable "cft_role_name" {
  description = "Role name to run the account onboarding CFT in each account to be onboarded"
  type = string
}

variable "endpoint_mode" {
	description = "Controls whether cloud NGFW will create firewall endpoints automatitically in customer subnets"
	type = string
	validation {
		condition     = contains(["Yes", "No"], var.endpoint_mode)
		error_message = "Allowed values for endpoint_mode are Yes/No"
	}
	default = "Yes"
}

variable "decryption_cert" {
	description = "The CloudNGFW can decrypt inbound and outbound traffic by providing a certificate stored in secret Manager.  The role allows the service to access a certificate configured in the rulestack.  Only certificated tagged with PaloAltoCloudNGFW can be accessed"
	type = string
	validation {
		condition     = contains(["None", "TagBasedCertificate"], var.decryption_cert)
		error_message = "Allowed values for decryption_cert are None/TagBasedCertificate"
	}
	default = "TagBasedCertificate"
}

variable "cloudwatch_namespace" {
	description = "Cloudwatch Namespace"
	type = string
	default = "PaloAltoCloudNGFW"
}

variable "cloudwatch_log_group" {
	description = "Cloudwatch Log Group"
	type = string
	default = "PaloAltoCloudNGFW"
}

variable "auditlog_group" {
	description = "Audit Log Group Name"
	type = string
	default = "PaloAltoCloudNGFWAuditLog"
}

variable "kinesis_firehose" {
	description = "Kinesis Firehose for logging"
	type = string
	default = "PaloAltoCloudNGFW"
}

variable "s3_bucket" {
	description = "S3 Bucket Name for Logging. Logging roles provide access to create log contents in this bucket."
	type = string
	default = "None"
}

variable "sync_mode" {
	description = "Controls whether to wait for account onboarding to complete."
	type = bool
	default = true
}
