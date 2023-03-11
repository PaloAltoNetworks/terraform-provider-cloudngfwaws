variable "rulestack" { default = {} }
variable "security_rules" { default = {} }
variable "anti_spyware" { default = {} }
variable "anti_virus" { default = {} }
variable "vulnerability" { default = {} }
variable "url_filtering" { default = {} }
variable "file_blocking" { default = {} }
variable "outbound_trust_certificate" { default = {} }
variable "outbound_untrust_certificate" { default = {} }
variable "feeds" { default = {} }
variable "prefix_lists" { default = {} }
variable "fqdn_lists" { default = {} }
variable "custom_url_categories" { default = {} }
variable "certificates" { default = {} }
variable "commit" {
  type    = bool
  default = false
}