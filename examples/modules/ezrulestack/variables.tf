variable "rulestack" { default = {} }
variable "security_rules" { default = {} }
variable "feeds" { default = {} }
variable "prefix_lists" { default = {} }
variable "fqdn_lists" { default = {} }
variable "custom_url_categories" { default = {} }
variable "certificates" { default = {} }
variable "commit" {
  type    = bool
  default = true
}