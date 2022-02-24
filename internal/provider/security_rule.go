package provider

import (
    "context"
    "fmt"
    "strconv"
    "strings"

    "github.com/paloaltonetworks/cloud-ngfw-aws-go"
    "github.com/paloaltonetworks/cloud-ngfw-aws-go/rule/security"

    "github.com/hashicorp/terraform-plugin-log/tflog"
    "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Data source.
func dataSourceSecurityRule() *schema.Resource {
    return &schema.Resource{
        Description: "Data source for retrieving security rule information.",

        ReadContext: readSecurityRuleDataSource,

        Schema: securityRuleSchema(false, nil),
    }
}

func readSecurityRuleDataSource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    svc := security.NewClient(meta.(*awsngfw.Client))

    stack := d.Get("rulestack").(string)
    rlist := d.Get("rule_list").(string)
    priority := d.Get("priority").(int)

    id := buildSecurityRuleId(stack, rlist, priority)

    req := security.ReadInput{
        Rulestack: stack,
        RuleList: rlist,
        Priority: priority,
        Candidate: true,
    }
    tflog.Info(
        ctx, "read security rule",
        "ds", true,
        "rulestack", req.Rulestack,
        "rule_list", req.RuleList,
        "priority", req.Priority,
    )

    res, err := svc.Read(ctx, req)
    if err != nil {
        if isObjectNotFound(err) {
            d.SetId("")
            return nil
        }
        return diag.FromErr(err)
    }

    d.SetId(id)
    saveSecurityRule(d, stack, rlist, priority, *res.Response.Candidate)

    return nil
}

// Resource.
func resourceSecurityRule() *schema.Resource {
    return &schema.Resource{
        Description: "Resource for security rule manipulation.",

        CreateContext: createSecurityRule,
        ReadContext: readSecurityRule,
        UpdateContext: updateSecurityRule,
        DeleteContext: deleteSecurityRule,

        Importer: &schema.ResourceImporter{
            State: schema.ImportStatePassthrough,
        },

        Schema: securityRuleSchema(true, nil),
    }
}

func createSecurityRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    svc := security.NewClient(meta.(*awsngfw.Client))
    o := loadSecurityRule(d)
    tflog.Info(
        ctx, "create security rule",
        "rulestack", o.Rulestack,
        "rule_list", o.RuleList,
        "priority", o.Priority,
        "name", o.Entry.Name,
    )

    if err := svc.Create(ctx, o); err != nil {
        return diag.FromErr(err)
    }

    d.SetId(buildSecurityRuleId(o.Rulestack, o.RuleList, o.Priority))

    return readSecurityRule(ctx, d, meta)
}

func readSecurityRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    svc := security.NewClient(meta.(*awsngfw.Client))
    stack, rlist, priority, err := parseSecurityRuleId(d.Id())
    if err != nil {
        return diag.Errorf("Error in parsing ID %q: %s", d.Id(), err)
    }

    req := security.ReadInput{
        Rulestack: stack,
        RuleList: rlist,
        Priority: priority,
        Candidate: true,
    }
    tflog.Info(
        ctx, "read security rule",
        "rulestack", req.Rulestack,
        "rule_list", req.RuleList,
        "priority", req.Priority,
    )

    res, err := svc.Read(ctx, req)
    if err != nil {
        if isObjectNotFound(err) {
            d.SetId("")
            return nil
        }
        return diag.FromErr(err)
    }

    saveSecurityRule(d, stack, rlist, priority, *res.Response.Candidate)

    return nil
}

func updateSecurityRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    svc := security.NewClient(meta.(*awsngfw.Client))
    o := loadSecurityRule(d)
    tflog.Info(
        ctx, "update security rule",
        "rulestack", o.Rulestack,
        "rule_list", o.RuleList,
        "priority", o.Priority,
    )

    if err := svc.Update(ctx, o); err != nil {
        return diag.FromErr(err)
    }

    return readSecurityRule(ctx, d, meta)
}

func deleteSecurityRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    svc := security.NewClient(meta.(*awsngfw.Client))
    stack, rlist, priority, err := parseSecurityRuleId(d.Id())
    if err != nil {
        return diag.Errorf("Error in parsing ID %q: %s", d.Id(), err)
    }

    tflog.Info(
        ctx, "delete rulestack",
        "rulestack", stack,
        "rule_list", rlist,
        "priority", priority,
    )

    if err := svc.Delete(ctx, stack, rlist, priority); err != nil && !isObjectNotFound(err) {
        return diag.FromErr(err)
    }

    d.SetId("")
    return nil
}

// Schema handling.
func securityRuleSchema(isResource bool, rmKeys []string) map[string] *schema.Schema {
    ans := map[string] *schema.Schema{
        "rulestack": rsSchema(),
        "rule_list": ruleListSchema(),
        "priority": {
            Type: schema.TypeInt,
            Required: true,
            Description: "The rule priority.",
            ForceNew: true,
        },
        "name": {
            Type: schema.TypeString,
            Required: true,
            Description: "The name.",
        },
        "description": {
            Type: schema.TypeString,
            Optional: true,
            Description: "The description.",
        },
        "enabled": {
            Type: schema.TypeBool,
            Optional: true,
            Description: "Set to false to disable this rule.",
            Default: true,
        },
        "source": {
            Type: schema.TypeList,
            Required: true,
            Description: "The source spec.",
            MinItems: 1,
            MaxItems: 1,
            Elem: &schema.Resource{
                Schema: map[string] *schema.Schema{
                    "cidrs": {
                        Type: schema.TypeList,
                        Optional: true,
                        Description: "List of CIDRs.",
                        Elem: &schema.Schema{
                            Type: schema.TypeString,
                        },
                    },
                    "countries": {
                        Type: schema.TypeList,
                        Optional: true,
                        Description: "List of countries.",
                        Elem: &schema.Schema{
                            Type: schema.TypeString,
                        },
                    },
                    "feeds": {
                        Type: schema.TypeList,
                        Optional: true,
                        Description: "List of feeds.",
                        Elem: &schema.Schema{
                            Type: schema.TypeString,
                        },
                    },
                    "prefix_lists": {
                        Type: schema.TypeList,
                        Optional: true,
                        Description: "List of prefix list.",
                        Elem: &schema.Schema{
                            Type: schema.TypeString,
                        },
                    },
                },
            },
        },
        "negate_source": {
            Type: schema.TypeBool,
            Optional: true,
            Description: "Negate the source definition.",
        },
        "destination": {
            Type: schema.TypeList,
            Required: true,
            Description: "The destination spec.",
            MinItems: 1,
            MaxItems: 1,
            Elem: &schema.Resource{
                Schema: map[string] *schema.Schema{
                    "cidrs": {
                        Type: schema.TypeList,
                        Optional: true,
                        Description: "List of CIDRs.",
                        Elem: &schema.Schema{
                            Type: schema.TypeString,
                        },
                    },
                    "countries": {
                        Type: schema.TypeList,
                        Optional: true,
                        Description: "List of countries.",
                        Elem: &schema.Schema{
                            Type: schema.TypeString,
                        },
                    },
                    "feeds": {
                        Type: schema.TypeList,
                        Optional: true,
                        Description: "List of feeds.",
                        Elem: &schema.Schema{
                            Type: schema.TypeString,
                        },
                    },
                    "prefix_lists": {
                        Type: schema.TypeList,
                        Optional: true,
                        Description: "List of prefix list.",
                        Elem: &schema.Schema{
                            Type: schema.TypeString,
                        },
                    },
                    "fqdn_lists": {
                        Type: schema.TypeList,
                        Optional: true,
                        Description: "List of FQDN lists.",
                        Elem: &schema.Schema{
                            Type: schema.TypeString,
                        },
                    },
                },
            },
        },
        "negate_destination": {
            Type: schema.TypeBool,
            Optional: true,
            Description: "Negate the destination definition.",
        },
        "applications": {
            Type: schema.TypeList,
            Required: true,
            Description: "The list of applications.",
            Elem: &schema.Schema{
                Type: schema.TypeString,
            },
        },
        "category": {
            Type: schema.TypeList,
            Required: true,
            Description: "The category spec.",
            MinItems: 1,
            MaxItems: 1,
            Elem: &schema.Resource{
                Schema: map[string] *schema.Schema{
                    "url_category_names": {
                        Type: schema.TypeList,
                        Optional: true,
                        Description: "List of URL category names.",
                        Elem: &schema.Schema{
                            Type: schema.TypeString,
                        },
                    },
                    "feeds": {
                        Type: schema.TypeList,
                        Optional: true,
                        Description: "List of feeds.",
                        Elem: &schema.Schema{
                            Type: schema.TypeString,
                        },
                    },
                },
            },
        },
        "protocol": {
            Type: schema.TypeString,
            Optional: true,
            Description: "The protocol.",
            Default: "application-default",
        },
        "audit_comment": {
            Type: schema.TypeString,
            Optional: true,
            Description: "The audit comment.",
        },
        "action": {
            Type: schema.TypeString,
            Required: true,
            Description: "The action to take.",
            ValidateFunc: validation.StringInSlice(
                []string{"Allow", "DenySilent", "DenyResetServer", "DenyResetBoth"},
                false,
            ),
        },
        "logging": {
            Type: schema.TypeBool,
            Optional: true,
            Description: "Enable logging at end.",
            Default: true,
        },
        "decryption_rule_type": {
            Type: schema.TypeString,
            Optional: true,
            Description: "Decryption rule type.",
            ValidateFunc: validation.StringInSlice(
                []string{"", "SSLOutboundInspection"},
                false,
            ),
        },
        "tag": {
            Type: schema.TypeList,
            Computed: true,
            Description: "Tags.",
            Elem: &schema.Resource{
                Schema: map[string] *schema.Schema{
                    "key": {
                        Type: schema.TypeString,
                        Computed: true,
                        Description: "The key.",
                    },
                    "value": {
                        Type: schema.TypeString,
                        Computed: true,
                        Description: "The value.",
                    },
                },
            },
        },
        "update_token": {
            Type: schema.TypeString,
            Computed: true,
            Description: "The update token.",
        },
    }

    for _, rmKey := range rmKeys {
        delete(ans, rmKey)
    }

    if !isResource {
        computed(ans, "", []string{"rulestack", "rule_list", "priority"})
    }

    return ans
}

func loadSecurityRule(d *schema.ResourceData) security.Info {
    src := configFolder(d, "source")
    dst := configFolder(d, "destination")
    cat := configFolder(d, "category")

    /*
    var tlist []security.TagDetails
    ts := d.Get("tag").([]interface{})
    if len(ts) > 0 {
        tlist = make([]security.TagDetails, 0, len(ts))
        for i := range ts {
            x := ts[i].(map[string] interface{})
            tlist = append(tlist, security.TagDetails{
                Key: x["key"].(string),
                Value: x["value"].(string),
            })
        }
    }
    */

    return security.Info{
        Rulestack: d.Get("rulestack").(string),
        RuleList: d.Get("rule_list").(string),
        Priority: d.Get("priority").(int),
        Entry: security.Details{
            Name: d.Get("name").(string),
            Description: d.Get("description").(string),
            Enabled: d.Get("enabled").(bool),
            Source: security.SourceDetails{
                Cidrs: toStringSlice(src["cidrs"]),
                Countries: toStringSlice(src["countries"]),
                Feeds: toStringSlice(src["feeds"]),
                PrefixLists: toStringSlice(src["prefix_lists"]),
            },
            NegateSource: d.Get("negate_source").(bool),
            Destination: security.DestinationDetails{
                Cidrs: toStringSlice(dst["cidrs"]),
                Countries: toStringSlice(dst["countries"]),
                Feeds: toStringSlice(dst["feeds"]),
                PrefixLists: toStringSlice(dst["prefix_lists"]),
                FqdnLists: toStringSlice(dst["fqdn_lists"]),
            },
            NegateDestination: d.Get("negate_destination").(bool),
            Applications: toStringSlice(d.Get("applications")),
            Category: security.CategoryDetails{
                UrlCategoryNames: toStringSlice(cat["url_category_names"]),
                Feeds: toStringSlice(cat["feeds"]),
            },
            Protocol: d.Get("protocol").(string),
            AuditComment: d.Get("audit_comment").(string),
            Action: d.Get("action").(string),
            Logging: d.Get("logging").(bool),
            DecryptionRuleType: d.Get("decryption_rule_type").(string),
            //Tags: tlist,
            //UpdateToken: d.Get("update_token").(string),
        },
    }
}

func saveSecurityRule(d *schema.ResourceData, stack, rlist string, priority int, o security.Details) {
    src := map[string] interface{}{
        "cidrs": o.Source.Cidrs,
        "countries": o.Source.Countries,
        "feeds": o.Source.Feeds,
        "prefix_lists": o.Source.PrefixLists,
    }
    dst := map[string] interface{}{
        "cidrs": o.Destination.Cidrs,
        "countries": o.Destination.Countries,
        "feeds": o.Destination.Feeds,
        "prefix_lists": o.Destination.PrefixLists,
        "fqdn_lists": o.Destination.FqdnLists,
    }
    cat := map[string] interface{}{
        "url_category_names": o.Category.UrlCategoryNames,
        "feeds": o.Category.Feeds,
    }

    var tlist []interface{}
    if len(o.Tags) > 0 {
        tlist = make([]interface{}, 0, len(o.Tags))
        for _, x := range o.Tags {
            tlist = append(tlist, map[string] interface{}{
                "key": x.Key,
                "value": x.Value,
            })
        }
    }

    d.Set("rulestack", stack)
    d.Set("rule_list", rlist)
    d.Set("priority", priority)
    d.Set("name", o.Name)
    d.Set("description", o.Description)
    d.Set("enabled", o.Enabled)
    d.Set("source", []interface{}{src})
    d.Set("negate_source", o.NegateSource)
    d.Set("destination", []interface{}{dst})
    d.Set("negate_destination", o.NegateDestination)
    d.Set("applications", o.Applications)
    d.Set("category", []interface{}{cat})
    d.Set("protocol", o.Protocol)
    d.Set("audit_comment", o.AuditComment)
    d.Set("action", o.Action)
    d.Set("logging", o.Logging)
    d.Set("decryption_rule_type", o.DecryptionRuleType)
    d.Set("tag", tlist)
    d.Set("update_token", o.UpdateToken)
}

// Id functions.
func buildSecurityRuleId(a, b string, c int) string {
    return strings.Join([]string{a, b, strconv.Itoa(c)}, IdSeparator)
}

func parseSecurityRuleId(v string) (string, string, int, error) {
    tok := strings.Split(v, IdSeparator)
    if len(tok) != 3 {
        return "", "", 0, fmt.Errorf("Expecting 3 tokens, got %d", len(tok))
    }

    priority, err := strconv.Atoi(tok[2])
    if err != nil {
        return "", "", 0, err
    }

    return tok[0], tok[1], priority, nil
}
