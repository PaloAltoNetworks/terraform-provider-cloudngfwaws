my_rulestack = {
  name        = "ezrulestack1"
  scope       = "Local"
  account_id  = 123456789
  description = "Made by Terraform"
  profile_config = {
    anti_spyware = "None"
  }
  lookup_x_forwarded_for = "SecurityPolicy"
}

my_security_rules = [
  {
    name        = "tf-security-rule1"
    rule_list   = "LocalRule"
    priority    = "1"
    description = "Also configured by Terraform"
    source = {
      cidrs       = ["any"]
      feeds       = ["feed1"]
      prefix_list = ["prefix-list1"]
    }
    negate_source = false
    destination = {
      prefix_list = ["prefix-list2"]
    }
    category = {
      url_category_names = []
    }
    applications  = ["any"]
    protocol      = "any"
    action        = "Allow"
    logging       = true
    audit_comment = "initial config"
  },
  {
    name      = "tf-security-rule2"
    rule_list = "LocalRule"
    priority  = "2"
    enabled   = false
    source = {
      feeds = ["feed2"]
    }
    destination = {
      cidrs     = ["192.168.2.0/24"]
      fqdn_list = ["fqdn1"]
    }
    negate_destination = false
    category = {
      url_category_names = ["tf-custom-category"]
    }
    applications   = ["any"]
    prot_port_list = ["TCP:90", "TCP:93", "UDP:100"]
    action         = "Allow"
    logging        = false
  }
]

my_feeds = [
  {
    name        = "feed1"
    description = "Also configured by Terraform"
    url         = "https://foobar.net"
    type        = "IP_LIST"
    frequency   = "DAILY"
    time        = 10
  },
  {
    name        = "feed2"
    description = "Also configured by Terraform"
    url         = "https://foobar2.net"
    type        = "IP_LIST"
    frequency   = "HOURLY"
    time        = 0
  }
]

my_prefix_lists = [
  {
    name        = "prefix-list1"
    description = "Also configured by Terraform"
    prefix_list = [
      "192.168.0.0",
      "10.1.5.0",
    ]
    audit_comment = "initial config"
  },
  {
    name          = "prefix-list2"
    description   = "Also configured by Terraform"
    prefix_list   = ["10.0.1.0/24"]
    audit_comment = "initial config"
  }
]

my_fqdn_lists = [
  {
    name        = "fqdn1"
    description = "Also configured by Terraform"
    fqdn_list = [
      "example1.com",
      "foobar1.org",
    ]
    audit_comment = "initial config"
  },
  {
    name        = "fqdn2"
    description = "Also configured by Terraform"
    fqdn_list = [
      "example2.com",
      "foobar2.org",
    ]
    audit_comment = "initial config"
  }
]

my_custom_url_categories = [
  {
    name        = "tf-custom-category"
    description = "Configured by Terraform"
    url_list = [
      "example.com",
      "foobar.org",
    ]
    # action = "alert" 
  }
]

my_certificates = [
  {
    name          = "tf-cert"
    description   = "Configured by Terraform"
    signer_arn    = "arn:aws:secretsmanager:us-east-1:123456789:secret:tf-test-inbound-cert-abcdef"
    audit_comment = "initial config"
  }
]