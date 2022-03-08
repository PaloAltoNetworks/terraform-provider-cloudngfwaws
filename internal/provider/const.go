package provider

// The token separator for Terraform IDs.
const IdSeparator = ":"

// Various param name constants that show up in multiple resources / data sources.
const (
	RulestackName       = "rulestack"
	GlobalRulestackName = "globalrulestack"
	RuleListName        = "rule_list"
	ConfigTypeName      = "config_type"
	TagsName            = "tags"
)

// Valid values for ConfigTypeName within data sources.
const (
	CandidateConfig = "candidate"
	RunningConfig   = "running"
)

// Valid values for EndpointMode when creating firewalls
const (
	CustomerManaged = "CustomerManaged"
	ServiceManaged  = "ServiceManaged"
)
