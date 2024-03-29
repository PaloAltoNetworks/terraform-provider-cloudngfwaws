---
page_title: "cloudngfwaws: cloudngfwaws_country Data Source"
subcategory: ""
description: |-
  Data source get a list of countries and their country codes.
---

# cloudngfwaws_country

Data source get a list of countries and their country codes.


## Admin Permission Type

* `Rulestack`


## Example Usage

```terraform
data "cloudngfwaws_country" "example" {}
```


<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `max_results` (Number) Max number of results. Defaults to `100`.
- `token` (String) Pagination token.

### Read-Only

- `codes` (Map of String) The country code (as the key) and description (as the value).
- `id` (String) The ID of this resource.
- `next_token` (String) Token for the next page of results.
