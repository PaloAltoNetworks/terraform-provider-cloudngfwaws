package provider

// The token separator for Terraform IDs.
const IdSeparator = ":"

// Various param name constants that show up in multiple resources / data sources.
const (
	RulestackName  = "rulestack"
	RuleListName   = "rule_list"
	ConfigTypeName = "config_type"
	TagsName       = "tags"
	ScopeName      = "scope"
)

// Valid values for ConfigTypeName within data sources.
const (
	CandidateConfig = "candidate"
	RunningConfig   = "running"
)

// Account onboarding related constants.
const (
	DefaultMPRegion     = "us-east-1"
	DefaultCFTStackName = "PaloAltoNetworksCrossAccountRoleSetup-Programmatic"
	DefaultMPRegionHost = "api.us-east-1.aws.cloudngfw.paloaltonetworks.com"
)

// Firewall resource constans
const (
	CustomerManagedEndpointMode = "CustomerManaged"
	ServiceManagedEndpointMode  = "ServiceManaged"
)
