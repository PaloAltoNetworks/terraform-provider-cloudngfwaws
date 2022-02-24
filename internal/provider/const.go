package provider

// The token separator for Terraform IDs.
const IdSeparator = ":"

// Various param name constants that show up in multiple resources / data sources.
const (
    RulestackName = "rulestack"
    RuleListName = "rule_list"
    ConfigTypeName = "config_type"
)

// Valid values for ConfigTypeName within data sources.
const (
    CandidateConfig = "candidate"
    RunningConfig = "running"
)
