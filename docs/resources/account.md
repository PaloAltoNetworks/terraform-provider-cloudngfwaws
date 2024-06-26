---
page_title: "cloudngfwaws: cloudngfwaws_account Resource"
subcategory: ""
description: |-
  Resource for Account manipulation.
---

# cloudngfwaws_account

Resource for Account manipulation.


## Admin Permission Type

* `Rulestack` (for `scope="Local"`)
* `Global Rulestack` (for `scope="Global"`)





<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `account_id` (String) The account ID
- `cft_url` (String) The CFT URL.
- `external_id` (String) The external ID of the account
- `onboarding_status` (String) The Account onboarding status
- `origin` (String) Origin of account onboarding
- `service_account_id` (String) The account ID of cloud NGFW service
- `sns_topic_arn` (String) The SNS topic ARN
- `trusted_account` (String) The trusted account ID

### Read-Only

- `id` (String) The ID of this resource.
- `update_token` (String) The update token.
