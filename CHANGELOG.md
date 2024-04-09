## 2.0.7 (Apr 09, 2024)

* fix global rulestack creation issues

## 2.0.6 (Feb 25, 2024)

* provide resource_timeout field to override default provider values

## 2.0.5 (Feb 19, 2024)

* provide sync_mode flag that enables terraform to wait for ngfw resource creation

## 2.0.4 (Sep 21, 2023)

* Mark link_id field as computed to support migration of existing firewalls

## 2.0.3 (Sep 15, 2023)

* Added link_id and link_status fields to cloudngfwaws_ngfw resource

## 2.0.2 (Aug 17, 2023)

* Added support for specifying AWS profile in cloudngfwaws provider

## 2.0.1 (Mar 17, 2023)

* Update ezrulestack module documentation

## 2.0.0 (Mar 10, 2023)

* Add the module ezrulestack for quick rulestack creation with dependencies resolved 
  implicity.

## 1.0.10 (Feb 9, 2023)

* Fix the issue where firewalls are getting disrupted when migrating to multi_vpc 
  enabled build.

## 1.0.9 (Jan 27, 2023)

* Added support for multi vpc and multi account

## 1.0.8 (July 11, 2022)

* Added ProtPortList support

## 1.0.7 (July 8, 2022)

* Added XFF support

## 1.0.6 (July 7, 2022)

* Fix deployment issues

## 1.0.5 (July 07, 2022)

Bug Fixes:

* Update the documentation with global rulestack admin dependency on Firewall Manager

## 1.0.4 (June 13, 2022)

Bug Fixes:

* `cloudngfwaws_rulestack`: Commit after deleting a rulestack

## 1.0.3 (June 9, 2022)

* Bug fixes

## 1.0.2 (June 8, 2022)

* Bug fixes
* Documentation fixes

## 1.0.1 (May 20, 2022)

* Fix deployment issues

## 1.0.0 (May 20, 2022)

New Data Sources:

* `cloudngfwaws_app_id_version` / `cloudngfwaws_app_id_versions`
* `cloudngfwaws_certificate`
* `cloudngfwaws_country`
* `cloudngfwaws_custom_url_category`
* `cloudngfwaws_fqdn_list`
* `cloudngfwaws_intelligent_feed`
* `cloudngfwaws_ngfw` / `cloudngfwaws_ngfws`
* `cloudngfwaws_predefined_url_categories`
* `cloudngfwaws_predefined_url_category_override`
* `cloudngfwaws_prefix_list`
* `cloudngfwaws_rulestack`
* `cloudngfwaws_security_rule`
* `cloudngfwaws_validate_rulestack`

New Data Sources:

* `cloudngfwaws_certificate`
* `cloudngfwaws_commit_rulestack`
* `cloudngfwaws_custom_url_category`
* `cloudngfwaws_fqdn_list`
* `cloudngfwaws_intelligent_feed`
* `cloudngfwaws_ngfw`
* `cloudngfwaws_ngfw_log_profile`
* `cloudngfwaws_predefined_url_category_override`
* `cloudngfwaws_prefix_list`
* `cloudngfwaws_rulestack`
* `cloudngfwaws_security_rule`
