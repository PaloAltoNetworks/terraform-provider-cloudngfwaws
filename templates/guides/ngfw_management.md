---
page_title: "cloudngfwaws: NGFW Management Guide"
subcategory: ""
description: |-
  Step-by-step guide for creating and managing Cloud NGFW firewalls using V1 and V2 schemas.
---

# NGFW Management Guide

This guide covers the complete lifecycle of Cloud NGFW firewall management using Terraform,
including schema selection, firewall creation, endpoint management, and Egress NAT configuration.

## Upgrading the Provider

This section covers what to expect and what to change when upgrading to the latest
version of the `cloudngfwaws` Terraform provider.

### Step 1 — Update the Version Constraint

In your `required_providers` block, relax or bump the version constraint to allow the new version:

```terraform
terraform {
  required_providers {
    cloudngfwaws = {
      source  = "paloaltonetworks/cloudngfwaws"
      version = "~> 2.0"   # update to match the target release
    }
  }
}
```

### Step 2 — Re-initialize

```shell
terraform init -upgrade
```

This downloads the new provider binary and updates the lock file.

### Step 3 — Run a Plan and Review

```shell
terraform plan
```

Review the output carefully before applying. The sections below describe changes.

---

#### New firewalls must use V2 schema

Firewall resources created with the old `subnet_mapping` + `endpoint_mode` fields
(V1 schema) continue to work for existing deployments and can still be managed in-place.

However, **creating a new firewall with V1 schema is no longer supported**. New
`cloudngfwaws_ngfw` resources must use `az_list` and `endpoints` (V2 schema).
See [V2 Schema — New Firewalls](#v2-schema--new-firewalls) below.

#### Log profile schema change — `log_destination` → `log_config`

The `cloudngfwaws_ngfw_log_profile` resource now supports a `log_config` block
(V2 schema) in addition to the existing `log_destination` block (V1 schema).
Existing `log_destination` configurations continue to work, but new deployments
should use `log_config` for full feature support including cross-account logging
and multiple log types per destination. See [Log Profile Management](#log-profile-management) below.

---

## Schema Overview

Cloud NGFW supports two resource schemas:

| | V1 Schema | V2 Schema |
|---|---|---|
| **Use case** | Existing deployments | New firewalls only |
| **Subnet configuration** | `subnet_mapping` | `az_list` + `endpoints` |
| **Create new firewalls** | Not supported | Required |
| **Egress NAT** | Supported | Supported |
| **Security zones** | `security_zones` block | Per-endpoint `prefixes` + `egress_nat_enabled` |
| **Endpoints during creation** | Not applicable | Not supported — add in update |

---

## V1 Schema — Existing Deployments

> V1 schema is for customers who already have firewalls deployed with Terraform.
> **Do not use V1 to create new firewalls.**

### Prerequisites

- An existing `cloudngfwaws_ngfw` resource in your Terraform state using `subnet_mapping`
- `Firewall` admin permission

---

### Step 1 — Verify Existing Firewall State

Before making any changes, confirm the current state matches your infrastructure.

A clean plan with no changes confirms your state is accurate. Resolve any drift before proceeding.

---

### Step 2 — Enable Egress NAT on an Existing Firewall (V1)

Egress NAT allows firewall-initiated traffic to use AWS-managed or customer-provided public IPs.
It can be enabled on an existing V1 firewall without recreating the resource.

**When to use:**
- Outbound traffic from protected workloads needs a predictable public IP
- Compliance requirements mandate source IP visibility for egress traffic

**Configuration options:**

| `ip_pool_type` | Description |
|---|---|
| `AWSService` | AWS manages the public IP pool automatically |
| `BYOIP` | Bring your own IP pool — requires `ipam_pool_id` |

**Steps:**

1. Add the `egress_nat` block to your existing resource:

```terraform
resource "cloudngfwaws_ngfw" "example" {
  name          = "example-instance"
  vpc_id        = "vpc-0a1b2c3d4e5f00001"
  account_id    = "111111111111"
  description   = "Example firewall"
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
    Env = "production"
  }
}
```

4. Confirm Egress NAT is active by checking the firewall status in the AWS console or via:

```shell
terraform show | grep -A 5 "egress_nat"
```

**To disable Egress NAT:** set `enabled = false` and re-apply.

---

### Step 3 — Configure Security Zones on an Existing Firewall (V1)

Security zones allow you to control Egress NAT and private CIDR routing on a per-endpoint basis.

> **Prerequisite:** Endpoints must be in `ACCEPTED` state before security zones can be configured.
> Attempting to configure security zones before endpoint acceptance will result in an error.

**When to use:**
- You need different Egress NAT behavior for different endpoints (e.g. enabled for one VPC, disabled for another)
- You need to define which private CIDRs are routed through the firewall per endpoint

**Steps:**

1. Check that all endpoints are in `ACCEPTED` state:

```shell
terraform show | grep -A 10 "attachment"
```

Look for `status = "ACCEPTED"` in each `attachment` block. If any endpoint shows
`PENDING` or `REJECTED`, wait or investigate before continuing.

2. Retrieve the `endpoint_id` from the status output. It will look like `vpce-0a1b2c3d4e5f00001`.

3. Add the `security_zones` block to your resource, referencing the accepted endpoint ID:

```terraform
resource "cloudngfwaws_ngfw" "example" {
  name          = "example-instance"
  vpc_id        = "vpc-0a1b2c3d4e5f00001"
  account_id    = "111111111111"
  description   = "Example firewall"
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

  # Configure after endpoint is ACCEPTED
  security_zones {
    endpoint_id        = "vpce-0a1b2c3d4e5f00001"   # from status.attachment[*].endpoint_id
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
    Env = "production"
  }
}
```

**To add or remove private prefixes:** update the `cidrs` list and re-apply.
**To toggle Egress NAT per zone:** change `egress_nat_enabled` and re-apply.

---

## V2 Schema — New Firewalls

> New firewalls can **only** be created using the V2 schema.
> The V2 schema uses `az_list` for availability zone selection and `endpoints`
> for VPC attachment — not `subnet_mapping`.

### Prerequisites

- `Firewall` admin permission
- Target VPCs and subnets already exist in AWS
- A committed rulestack (recommended — ensures Terraform waits for a successful commit before creating the firewall)

---

### Step 1 — Create the Firewall (V2)

Create the firewall resource specifying availability zones.
**Do not add `endpoints` in this step.** Endpoints must be added in a separate update
after the firewall reaches `CREATE_COMPLETE` state.

**Why:** The Cloud NGFW API does not support adding endpoints atomically with firewall creation.
Attempting to include endpoints in the create call will be rejected.

```terraform
resource "cloudngfwaws_ngfw" "example" {
  name               = "my-firewall"
  description        = "Production firewall"
  az_list            = ["use1-az1", "use1-az4"]
  allowlist_accounts = ["111111111111"]

  rulestack = cloudngfwaws_commit_rulestack.rs.rulestack

  tags = {
    Env = "production"
  }
}

resource "cloudngfwaws_commit_rulestack" "rs" {
  rulestack = "my-rulestack"
}
```

Once deployed, confirm the firewall has reached `CREATE_COMPLETE` state:

```shell
terraform show | grep firewall_status
```

> Firewall creation typically takes 5–15 minutes. Do not proceed to Step 2 until
> `firewall_status = "CREATE_COMPLETE"`.

---

### Step 2 — Add Endpoints

Endpoints attach the firewall to specific VPCs in allowlisted accounts. Add them as
an update after the firewall is created.

**Configuration options for `mode`:**

| Value | Description |
|---|---|
| `ServiceManaged` | AWS creates and manages the VPC endpoint automatically |
| `CustomerManaged` | You create and manage the VPC endpoint in your account |

**Steps:**

1. Add `endpoints` blocks to the existing resource:

```terraform
resource "cloudngfwaws_ngfw" "example" {
  name               = "my-firewall"
  description        = "Production firewall"
  az_list            = ["use1-az1", "use1-az4"]
  allowlist_accounts = ["111111111111"]

  rulestack = cloudngfwaws_commit_rulestack.rs.rulestack

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
    Env = "production"
  }
}
```

3. Wait for each endpoint to reach `ACCEPTED` state before proceeding:

```shell
terraform show | grep -A 15 "endpoints"
```

Look for `status = "ACCEPTED"` on each endpoint. A status of `REJECTED` means the
endpoint request was declined — check `rejected_reason` for details.

> Endpoint acceptance typically takes 2–5 minutes per endpoint.

---

### Step 3 — Enable Egress NAT (V2)

Enable Egress NAT at the firewall level once at least one endpoint is accepted.
This applies globally to all endpoints unless overridden per-endpoint in Step 4.

> **Prerequisite:** At least one endpoint must be in `ACCEPTED` state.

```terraform
resource "cloudngfwaws_ngfw" "example" {
  name               = "my-firewall"
  description        = "Production firewall"
  az_list            = ["use1-az1", "use1-az4"]
  allowlist_accounts = ["111111111111"]

  rulestack = cloudngfwaws_commit_rulestack.rs.rulestack

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
    Env = "production"
  }
}
```

**To disable Egress NAT:** set `enabled = false` and re-apply.

---

### Step 4 — Configure Per-Endpoint Egress NAT and Private Prefixes (V2)

Once endpoints are accepted, you can configure Egress NAT and private CIDR routing
individually per endpoint within the `endpoints` block.

> **Prerequisite:** The target endpoint must be in `ACCEPTED` state.
> The `endpoint_id` field is **read-only** — it is populated automatically by the provider
> after the endpoint is accepted. Do not set it manually.

**When to use:**
- Different VPCs need different Egress NAT behavior (e.g. enabled for a production VPC, disabled for a dev VPC)
- You need to restrict which private CIDR ranges are routed through the firewall per VPC

**Steps:**

1. Retrieve the `endpoint_id` for each accepted endpoint:

```shell
terraform show | grep -A 20 "endpoints"
```

The `endpoint_id` field will appear as a computed value, e.g. `endpoint_id = "vpce-0a1b2c3d4e5f00001"`.

2. Update the relevant `endpoints` block with `egress_nat_enabled` and `prefixes`:

```terraform
resource "cloudngfwaws_ngfw" "example" {
  name               = "my-firewall"
  description        = "Production firewall"
  az_list            = ["use1-az1", "use1-az4"]
  allowlist_accounts = ["111111111111"]

  rulestack = cloudngfwaws_commit_rulestack.rs.rulestack

  # Endpoint with Egress NAT enabled and private prefixes configured
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

  # Endpoint with Egress NAT disabled
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
    Env = "production"
  }
}
```

**To add private prefixes:** add CIDRs to the `cidrs` list and re-apply.
**To remove private prefixes:** remove CIDRs from the `cidrs` list and re-apply.
**To toggle per-endpoint Egress NAT:** change `egress_nat_enabled` and re-apply.

---

## Log Profile Management

The `cloudngfwaws_ngfw_log_profile` resource controls where firewall logs are delivered.
Two schemas are supported — use V1 for existing configurations, V2 for new deployments.

| | V1 Schema | V2 Schema |
|---|---|---|
| **Block** | `log_destination` | `log_config` |
| **Log types per block** | One | Multiple (Set) |
| **Cross-account logging** | Not supported | Supported via `role_type` + `account_id` |
| **Max destinations** | Unlimited blocks | 1 block (consolidates all destinations) |

---

### V1 Schema — `log_destination` (Existing Deployments)

> V1 uses one `log_destination` block per log type. Existing configurations
> using this schema continue to work and do not need to be migrated.

**Supported destination types:** `S3`, `CloudWatchLogs`, `KinesisDataFirehose`
**Supported log types:** `TRAFFIC`, `THREAT`, `DECRYPTION`

**Example — V1 log profile with multiple destinations:**

```terraform
resource "cloudngfwaws_ngfw_log_profile" "example" {
  ngfw       = cloudngfwaws_ngfw.example.name
  account_id = cloudngfwaws_ngfw.example.account_id

  log_destination {
    destination_type = "S3"
    destination      = "my-s3-bucket"
    log_type         = "TRAFFIC"
  }

  log_destination {
    destination_type = "CloudWatchLogs"
    destination      = "my-log-group"
    log_type         = "THREAT"
  }

  log_destination {
    destination_type = "KinesisDataFirehose"
    destination      = "my-firehose-stream"
    log_type         = "DECRYPTION"
  }
}
```

**To add a destination:** add another `log_destination` block and re-apply.
**To remove a destination:** remove the block and re-apply.
**To change the destination:** update `destination` or `destination_type` in the block and re-apply.

---

### V2 Schema — `log_config` (New Deployments)

> V2 consolidates all destination configuration into a single `log_config` block
> and supports multiple log types per destination. Use V2 for new deployments.

**Key differences from V1:**
- `log_type` is a **set** — multiple log types can be sent to a single destination
- Supports cross-account log delivery via `role_type` and `account_id`
- Only one `log_config` block is allowed per resource

**Example — V2 log profile, same-account delivery:**

```terraform
resource "cloudngfwaws_ngfw_log_profile" "example" {
  ngfw       = cloudngfwaws_ngfw.example.name
  account_id = cloudngfwaws_ngfw.example.account_id

  log_config {
    log_destination_type = "S3"
    log_destination      = "my-s3-bucket"
    log_type             = ["TRAFFIC", "THREAT", "DECRYPTION"]
  }
}
```

**Example — V2 log profile with cross-account delivery:**

```terraform
resource "cloudngfwaws_ngfw_log_profile" "example" {
  ngfw       = cloudngfwaws_ngfw.example.name
  account_id = cloudngfwaws_ngfw.example.account_id

  log_config {
    log_destination_type = "CloudWatchLogs"
    log_destination      = "arn:aws:logs:us-east-1:222222222222:log-group:my-log-group"
    log_type             = ["TRAFFIC", "THREAT"]
    role_type            = "CrossAccount"
    account_id           = "222222222222"
  }
}
```

**Example — V2 log profile with advanced threat logging and CloudWatch metrics:**

```terraform
resource "cloudngfwaws_ngfw_log_profile" "example" {
  ngfw                        = cloudngfwaws_ngfw.example.name
  account_id                  = cloudngfwaws_ngfw.example.account_id
  advanced_threat_log         = true
  cloud_watch_metric_namespace = "CloudNGFW"

  log_config {
    log_destination_type = "KinesisDataFirehose"
    log_destination      = "my-firehose-stream"
    log_type             = ["TRAFFIC", "THREAT", "DECRYPTION"]
  }
}
```

---

### Migrating from V1 to V2

If you want to migrate an existing V1 `log_destination` configuration to V2 `log_config`:

1. Replace the `log_destination` blocks with a single `log_config` block, consolidating
   all log types into the `log_type` set:

```terraform
# Before (V1)
resource "cloudngfwaws_ngfw_log_profile" "example" {
  ngfw       = cloudngfwaws_ngfw.example.name
  account_id = cloudngfwaws_ngfw.example.account_id

  log_destination {
    destination_type = "S3"
    destination      = "my-s3-bucket"
    log_type         = "TRAFFIC"
  }

  log_destination {
    destination_type = "S3"
    destination      = "my-s3-bucket"
    log_type         = "THREAT"
  }
}

# After (V2)
resource "cloudngfwaws_ngfw_log_profile" "example" {
  ngfw       = cloudngfwaws_ngfw.example.name
  account_id = cloudngfwaws_ngfw.example.account_id

  log_config {
    log_destination_type = "S3"
    log_destination      = "my-s3-bucket"
    log_type             = ["TRAFFIC", "THREAT"]
  }
}
```

2. Run `terraform plan` to confirm only the log profile resource is updated.
3. Apply — the firewall will remain running throughout; log delivery updates without interruption.

---

## Troubleshooting

### Endpoint status is REJECTED

Check the `rejected_reason` field on the endpoint:

```shell
terraform show | grep -A 20 "endpoints" | grep rejected_reason
```

Common causes:
- The VPC or subnet does not exist in the specified account
- The `allowlist_accounts` list does not include the endpoint's `account_id`
- The subnet CIDR conflicts with an existing endpoint in the same AZ

### Security zones configuration fails

Ensure the endpoint referenced by `endpoint_id` is in `ACCEPTED` state. Security zones
cannot be applied to endpoints in `PENDING` or `REJECTED` state.

### Firewall creation times out

The default Terraform timeout for firewall creation is set in the `timeouts` block.
If your environment takes longer, extend the timeout:

```terraform
resource "cloudngfwaws_ngfw" "example" {
  # ...
  timeouts {
    create = "30m"
  }
}
```

## Import

To import an existing firewall into Terraform state:

```shell
# Format: <account_id>:<firewall_name>
terraform import cloudngfwaws_ngfw.example 111111111111:my-firewall
```
