---
page_title: "terraform-provider-cloudngfwaws: cloudngfwaws_ngfw Resource"
subcategory: ""
description: |-
  Resource for NGFW manipulation.
---

# cloudngfwaws_ngfw

Resource for NGFW manipulation.

-> **NOTE:** Having the `rulestack` param reference the rulestack name from `cloudngfwaws_commit_rulestack` ensures that Terraform will only try to spin up a NGFW instance if the commit is successful.

## Admin Permission Type

* `Firewall`

## Configuration Guide

---

### V1 Schema — Existing Deployments Only

> **Important:** V1 schema is for existing customers who already have firewalls deployed with Terraform.
> New firewalls must be created using the [V2 schema](#v2-schema--new-firewalls).

---

#### 1. Managing an Existing Firewall (no configuration changes)

Use the V1 schema as-is. No steps required beyond ensuring your existing state is in sync.

**Steps:**

1. Verify there is no unintended drift:
   2. If the plan is clean, no action needed. If drift is detected, review and apply:
   
**Full example — existing V1 firewall:**

```terraform
resource "cloudngfwaws_ngfw" "example" {
  name          = "example-instance"
  vpc_id        = aws_vpc.example.id
  account_id    = "111111111111"
  description   = "Example description"
  endpoint_mode = "ServiceManaged"

  subnet_mapping {
    subnet_id = aws_subnet.subnet1.id
  }

  subnet_mapping {
    subnet_id = aws_subnet.subnet2.id
  }

  rulestack = cloudngfwaws_commit_rulestack.rs.rulestack

  tags = {
    Foo = "bar"
  }
}

resource "cloudngfwaws_commit_rulestack" "rs" {
  rulestack = "my-rulestack"
}
```

---

#### 2. Configuring Egress NAT on an Existing Firewall (V1)

Egress NAT can be added to an existing V1 firewall without recreating the resource.

> `ip_pool_type` accepts `AWSService` or `BYOIP`. Use `BYOIP` together with `ipam_pool_id`
> if bringing your own IP pool.

**Steps:**

1. Add the `egress_nat` block to your existing resource.

**Full example — existing V1 firewall with Egress NAT enabled:**

```terraform
resource "cloudngfwaws_ngfw" "example" {
  name          = "example-instance"
  vpc_id        = "vpc-0a1b2c3d4e5f00001"
  account_id    = "111111111111"
  description   = "Example description"
  endpoint_mode = "CustomerManaged"

  subnet_mapping {
    availability_zone = "us-east-1a"
  }

  subnet_mapping {
    availability_zone = "us-east-1c"
  }

  rulestack = "my-rulestack"

  egress_nat {
    enabled = true
    settings {
      ip_pool_type = "AWSService"
    }
  }

  tags = {
    Foo = "bar"
  }
}
```

**To disable Egress NAT:** set `enabled = false` and re-apply.

---

#### 3. Configuring Security Zones on an Existing Firewall (V1)

Security zones let you enable or disable Egress NAT per endpoint and add or remove private CIDR prefixes.

> **Prerequisite:** Endpoints must be successfully created and in `ACCEPTED` state before
> security zones can be configured. Check `status.attachment[*].status` in Terraform state
> or the AWS console before proceeding.

**Steps:**

1. Confirm endpoint status is `ACCEPTED`:
   ```shell
   terraform show | grep -A 10 "attachment"
   ```
2. Copy the `endpoint_id` value from the `status.attachment` output.
3. Add the `security_zones` block to your existing resource referencing that endpoint ID.

**Full example — existing V1 firewall with Egress NAT and security zones:**

```terraform
resource "cloudngfwaws_ngfw" "example" {
  name          = "example-instance"
  vpc_id        = "vpc-0a1b2c3d4e5f00001"
  account_id    = "111111111111"
  description   = "Example description"
  endpoint_mode = "CustomerManaged"

  subnet_mapping {
    availability_zone = "us-east-1a"
  }

  subnet_mapping {
    availability_zone = "us-east-1c"
  }

  rulestack = "my-rulestack"

  egress_nat {
    enabled = true
    settings {
      ip_pool_type = "AWSService"
    }
  }

  # Add after endpoint is ACCEPTED — use endpoint_id from status.attachment[*].endpoint_id
  security_zones {
    endpoint_id        = "vpce-0a1b2c3d4e5f00001"
    egress_nat_enabled = true
    prefixes {
      private_prefix {
        cidrs = [
          "10.0.0.0/8",
          "172.16.0.0/12",
          "192.168.0.0/16",
          "100.64.0.0/10"
        ]
      }
    }
  }

  tags = {
    Foo = "bar"
  }
}
```

**To remove private prefixes:** remove the CIDR entries from `cidrs` and re-apply.
**To disable Egress NAT for a specific zone:** set `egress_nat_enabled = false` and re-apply.

---

### V2 Schema — New Firewalls

> **Important:** New firewalls can only be created using the V2 schema. Use `az_list`
> instead of `subnet_mapping`, and `endpoints` instead of `endpoint_mode`/`subnet_mapping`.

---

#### 1. Creating a New Firewall (V2)

Firewall creation uses `az_list` to specify availability zones.
**Do not include `endpoints` during creation** — they must be added in a separate update after the firewall is running.

**Steps:**

1. Define the resource with `az_list` and no `endpoints` block.
4. Proceed to **Step 2** once the firewall reaches `RUNNING` state.

**Full example — new V2 firewall (creation only):**

```terraform
resource "cloudngfwaws_ngfw" "example" {
  name               = "my-firewall"
  description        = "My new firewall"
  az_list            = ["use1-az1", "use1-az4"]
  allowlist_accounts = ["111111111111"]

  tags = {
    Owner = "my-team"
  }
}
```

---

#### 2. Adding Endpoints to a V2 Firewall

Endpoints connect the firewall to customer VPCs. They must be added in a separate
a separate update after the firewall is running.

**Steps:**

1. Confirm the firewall status is `RUNNING`:
   ```shell
   terraform show | grep firewall_status
   ```
2. Add one or more `endpoints` blocks to the existing resource.
5. Wait for each endpoint's `status` to reach `ACCEPTED` before proceeding to configure
   Egress NAT or private prefixes:
   ```shell
   terraform show | grep -A 10 "endpoints"
   ```

**Full example — V2 firewall with endpoints added:**

```terraform
resource "cloudngfwaws_ngfw" "example" {
  name               = "my-firewall"
  description        = "My new firewall"
  az_list            = ["use1-az1", "use1-az4"]
  allowlist_accounts = ["111111111111"]

  endpoints {
    account_id = "111111111111"
    vpc_id     = "vpc-0a1b2c3d4e5f00002"
    subnet_id  = "subnet-0a1b2c3d4e5f00001"
    mode       = "ServiceManaged"
  }

  endpoints {
    account_id = "111111111111"
    vpc_id     = "vpc-0a1b2c3d4e5f00003"
    subnet_id  = "subnet-0a1b2c3d4e5f00002"
    mode       = "ServiceManaged"
  }

  tags = {
    Owner = "my-team"
  }
}
```

---

#### 3. Configuring Egress NAT on a V2 Firewall

Egress NAT can be enabled at the firewall level once at least one endpoint is accepted.

> **Prerequisite:** At least one endpoint must be in `ACCEPTED` state.

**Steps:**

1. Add the `egress_nat` block to the resource.

**Full example — V2 firewall with Egress NAT enabled:**

```terraform
resource "cloudngfwaws_ngfw" "example" {
  name               = "my-firewall"
  description        = "My new firewall"
  az_list            = ["use1-az1", "use1-az4"]
  allowlist_accounts = ["111111111111"]

  endpoints {
    account_id = "111111111111"
    vpc_id     = "vpc-0a1b2c3d4e5f00002"
    subnet_id  = "subnet-0a1b2c3d4e5f00001"
    mode       = "ServiceManaged"
  }

  endpoints {
    account_id = "111111111111"
    vpc_id     = "vpc-0a1b2c3d4e5f00003"
    subnet_id  = "subnet-0a1b2c3d4e5f00002"
    mode       = "ServiceManaged"
  }

  egress_nat {
    enabled = true
    settings {
      ip_pool_type = "AWSService"
    }
  }

  tags = {
    Owner = "my-team"
  }
}
```

**To disable Egress NAT:** set `enabled = false` and re-apply.

---

#### 4. Configuring Private Prefixes and Per-Endpoint Egress NAT (V2)

Once an endpoint is accepted, you can enable or disable Egress NAT and configure private
CIDR prefixes on a per-endpoint basis within the `endpoints` block.

> **Prerequisite:** The endpoint must be in `ACCEPTED` state. The `endpoint_id`
> is a read-only computed value — retrieve it from Terraform state after apply:
> ```shell
> terraform show | grep -A 15 "endpoints"
> ```

**Steps:**

1. Update the relevant `endpoints` block with `egress_nat_enabled` and `prefixes`.
   The `endpoint_id` field is read-only and is populated automatically by the provider
   once the endpoint is accepted — do not set it manually.

**Full example — V2 firewall with per-endpoint Egress NAT and private prefixes:**

```terraform
resource "cloudngfwaws_ngfw" "example" {
  name               = "my-firewall"
  description        = "My new firewall"
  az_list            = ["use1-az1", "use1-az4"]
  allowlist_accounts = ["111111111111"]

  endpoints {
    account_id         = "111111111111"
    vpc_id             = "vpc-0a1b2c3d4e5f00002"
    subnet_id          = "subnet-0a1b2c3d4e5f00001"
    mode               = "ServiceManaged"
    egress_nat_enabled = true
    prefixes {
      private_prefix {
        cidrs = [
          "10.0.0.0/8",
          "172.16.0.0/12",
          "192.168.0.0/16",
          "100.64.0.0/10"
        ]
      }
    }
  }

  endpoints {
    account_id         = "111111111111"
    vpc_id             = "vpc-0a1b2c3d4e5f00003"
    subnet_id          = "subnet-0a1b2c3d4e5f00002"
    mode               = "ServiceManaged"
    egress_nat_enabled = false
  }

  egress_nat {
    enabled = true
    settings {
      ip_pool_type = "AWSService"
    }
  }

  tags = {
    Owner = "my-team"
  }
}
```

**To remove private prefixes:** remove the CIDR entries from `cidrs` and re-apply.
**To disable per-endpoint Egress NAT:** set `egress_nat_enabled = false` and re-apply.

---

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The NGFW name.

### Optional

- `account_id` (String) The Account Id.
- `allowlist_accounts` (Set of String) The list of allowed accounts for this NGFW.
- `app_id_version` (String) App-ID version number.
- `automatic_upgrade_app_id_version` (Boolean) Automatic App-ID upgrade version number. Defaults to `true`.
- `az_list` (Set of String) The list of availability zone IDs for this NGFW.
- `change_protection` (Set of String) Enables or disables change protection for the NGFW.
- `description` (String) The NGFW description.
- `egress_nat` (Block List) (see [below for nested schema](#nestedblock--egress_nat))
- `endpoint_mode` (String) Set endpoint mode from the following options. Valid values are `ServiceManaged` or `CustomerManaged`.
- `endpoints` (Block List) (see [below for nested schema](#nestedblock--endpoints))
- `firewall_id` (String) The Firewall ID.
- `global_rulestack` (String) The global rulestack for this NGFW.
- `link_id` (String) The link ID.
- `multi_vpc` (Boolean) Share NGFW with Multiple VPCs. This feature can be enabled only if the endpoint_mode is CustomerManaged.
- `private_access` (Block List) (see [below for nested schema](#nestedblock--private_access))
- `rulestack` (String) The rulestack for this NGFW.
- `security_zones` (Block List) (see [below for nested schema](#nestedblock--security_zones))
- `subnet_mapping` (Block List) Subnet mappings. (see [below for nested schema](#nestedblock--subnet_mapping))
- `tags` (Map of String) The tags.
- `tier` (String) Firewall Instance Tier. Allowed values are 'base', 'standard', or 'premium'.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `user_id` (Block List) (see [below for nested schema](#nestedblock--user_id))
- `vpc_id` (String) The VPC ID for the NGFW.

### Read-Only

- `deployment_update_token` (String) The update token.
- `endpoint_service_name` (String) The endpoint service name.
- `id` (String) The ID of this resource.
- `link_status` (String) The link status.
- `status` (List of Object) (see [below for nested schema](#nestedatt--status))
- `update_token` (String) The update token.

<a id="nestedblock--egress_nat"></a>
### Nested Schema for `egress_nat`

Required:

- `enabled` (Boolean) Enable egress NAT

Optional:

- `settings` (Block List) (see [below for nested schema](#nestedblock--egress_nat--settings))

<a id="nestedblock--egress_nat--settings"></a>
### Nested Schema for `egress_nat.settings`

Optional:

- `ip_pool_type` (String) Set ip pool type from the following options. Valid values are `AWSService` or `BYOIP`.
- `ipam_pool_id` (String) The IP pool ID

<a id="nestedblock--endpoints"></a>
### Nested Schema for `endpoints`

Required:

- `mode` (String) The endpoint mode. Valid values are `ServiceManaged` or `CustomerManaged`.

Optional:

- `account_id` (String) The account id.
- `egress_nat_enabled` (Boolean) Enable egress NAT
- `prefixes` (Block List) (see [below for nested schema](#nestedblock--endpoints--prefixes))
- `subnet_id` (String) The subnet id.
- `vpc_id` (String) The vpc id.
- `zone_id` (String) The AZ id.

Read-Only:

- `endpoint_id` (String) Endpoint ID of the security zone
- `rejected_reason` (String) The rejected reason.
- `status` (String) The attachment status.

<a id="nestedblock--endpoints--prefixes"></a>
### Nested Schema for `endpoints.prefixes`

Optional:

- `private_prefix` (Block List) (see [below for nested schema](#nestedblock--endpoints--prefixes--private_prefix))

<a id="nestedblock--endpoints--prefixes--private_prefix"></a>
### Nested Schema for `endpoints.prefixes.private_prefix`

Optional:

- `cidrs` (Set of String)

<a id="nestedblock--private_access"></a>
### Nested Schema for `private_access`

Required:

- `resource_id` (String) AWS ResourceID
- `type` (String) Type of Private Access

<a id="nestedblock--security_zones"></a>
### Nested Schema for `security_zones`

Required:

- `endpoint_id` (String) Endpoint ID of the security zone

Optional:

- `account_id` (String) The account id.
- `egress_nat_enabled` (Boolean) Enable egress NAT
- `mode` (String) The endpoint mode. Valid values are `ServiceManaged` or `CustomerManaged`.
- `prefixes` (Block List) (see [below for nested schema](#nestedblock--security_zones--prefixes))
- `subnet_id` (String) The subnet id.
- `vpc_id` (String) The vpc id.
- `zone_id` (String) The AZ id.

Read-Only:

- `rejected_reason` (String) The rejected reason.
- `status` (String) The attachment status.

<a id="nestedblock--security_zones--prefixes"></a>
### Nested Schema for `security_zones.prefixes`

Optional:

- `private_prefix` (Block List) (see [below for nested schema](#nestedblock--security_zones--prefixes--private_prefix))

<a id="nestedblock--security_zones--prefixes--private_prefix"></a>
### Nested Schema for `security_zones.prefixes.private_prefix`

Optional:

- `cidrs` (Set of String)

<a id="nestedblock--subnet_mapping"></a>
### Nested Schema for `subnet_mapping`

Optional:

- `availability_zone` (String) The availability zone, for when the endpoint mode is customer managed.
- `availability_zone_id` (String) The availability zone ID, for when the endpoint mode is customer managed.
- `subnet_id` (String) The subnet id, for when the endpoint mode is service managed.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `default` (String)
- `delete` (String)
- `read` (String)
- `update` (String)

<a id="nestedblock--user_id"></a>
### Nested Schema for `user_id`

Required:

- `enabled` (Boolean) Enable UserID Config
- `port` (Number) The Port

Optional:

- `agent_name` (String) Agent Name for UserID
- `collector_name` (String) The Collector Name
- `custom_include_exclude_network` (Block List) List of Custom Include Exclude Networks (see [below for nested schema](#nestedblock--user_id--custom_include_exclude_network))
- `secret_key_arn` (String) AWS Secret Key ARN

Read-Only:

- `user_id_status` (String) Status and State of UserID Configuration

<a id="nestedblock--user_id--custom_include_exclude_network"></a>
### Nested Schema for `user_id.custom_include_exclude_network`

Required:

- `discovery_include` (Boolean) Include or exclude this subnet from user-id configuration
- `enabled` (Boolean) Enable this specific custom include/exclude network
- `name` (String) Name of subnet filter
- `network_address` (String) Network IP address of the subnet filter

<a id="nestedatt--status"></a>
### Nested Schema for `status`

Read-Only:

- `attachment` (List of Object) (see [below for nested schema](#nestedobjatt--status--attachment))
- `device_rulestack_commit_status` (String)
- `failure_reason` (String)
- `firewall_status` (String)
- `rulestack_status` (String)

<a id="nestedobjatt--status--attachment"></a>
### Nested Schema for `status.attachment`

Read-Only:

- `endpoint_id` (String)
- `rejected_reason` (String)
- `status` (String)
- `subnet_id` (String)

## Import

Import is supported using the following syntax:

```shell
# import name is <account_id>:<name>
terraform import cloudngfwaws_ngfw.example 12345678:example-instance
```
