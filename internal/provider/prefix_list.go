package provider

import (
    "context"
    "fmt"
    "strings"

    "github.com/paloaltonetworks/cloud-ngfw-aws-go"
    "github.com/paloaltonetworks/cloud-ngfw-aws-go/object/prefix"

    "github.com/hashicorp/terraform-plugin-log/tflog"
    "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Data source.
func dataSourcePrefixList() *schema.Resource {
    return &schema.Resource{
        Description: "Data source for retrieving prefix list information.",

        ReadContext: readPrefixListDataSource,

        Schema: prefixListSchema(false, nil),
    }
}

func readPrefixListDataSource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    svc := prefix.NewClient(meta.(*awsngfw.Client))

    style := d.Get(ConfigTypeName).(string)
    d.Set(ConfigTypeName, style)

    stack := d.Get(RulestackName).(string)
    name := d.Get("name").(string)

    id := configTypeId(style, buildPrefixListId(stack, name))

    req := prefix.ReadInput{
        Rulestack: stack,
        Name: name,
    }
    switch style {
    case CandidateConfig:
        req.Candidate = true
    case RunningConfig:
        req.Running = true
    }

    tflog.Info(
        ctx, "read prefix list",
        "ds", true,
        ConfigTypeName, style,
        RulestackName, req.Rulestack,
        "name", req.Name,
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

    var info *prefix.Info
    switch style {
    case CandidateConfig:
        info = res.Response.Candidate
    case RunningConfig:
        info = res.Response.Running
    }
    savePrefixList(d, stack, name, *info)

    return nil
}

// Resource.
func resourcePrefixList() *schema.Resource {
    return &schema.Resource{
        Description: "Resource for prefix list manipulation.",

        CreateContext: createPrefixList,
        ReadContext: readPrefixList,
        UpdateContext: updatePrefixList,
        DeleteContext: deletePrefixList,

        Importer: &schema.ResourceImporter{
            State: schema.ImportStatePassthrough,
        },

        Schema: prefixListSchema(true, []string{ConfigTypeName}),
    }
}

func createPrefixList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    svc := prefix.NewClient(meta.(*awsngfw.Client))
    o := loadPrefixList(d)
    tflog.Info(
        ctx, "create prefix list",
        RulestackName, o.Rulestack,
        "name", o.Name,
    )

    if err := svc.Create(ctx, o); err != nil {
        return diag.FromErr(err)
    }

    d.SetId(buildPrefixListId(o.Rulestack, o.Name))

    return readPrefixList(ctx, d, meta)
}

func readPrefixList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    svc := prefix.NewClient(meta.(*awsngfw.Client))
    stack, name, err := parsePrefixListId(d.Id())
    if err != nil {
        return diag.Errorf("Error in parsing ID %q: %s", d.Id(), err)
    }

    req := prefix.ReadInput{
        Rulestack: stack,
        Name: name,
        Candidate: true,
    }
    tflog.Info(
        ctx, "read prefix list",
        RulestackName, req.Rulestack,
        "name", name,
    )

    res, err := svc.Read(ctx, req)
    if err != nil {
        if isObjectNotFound(err) {
            d.SetId("")
            return nil
        }
        return diag.FromErr(err)
    }

    savePrefixList(d, stack, name, *res.Response.Candidate)

    return nil
}

func updatePrefixList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    svc := prefix.NewClient(meta.(*awsngfw.Client))
    o := loadPrefixList(d)
    tflog.Info(
        ctx, "update prefix list",
        RulestackName, o.Rulestack,
        "name", o.Name,
    )

    if err := svc.Update(ctx, o); err != nil {
        return diag.FromErr(err)
    }

    return readPrefixList(ctx, d, meta)
}

func deletePrefixList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    svc := prefix.NewClient(meta.(*awsngfw.Client))
    stack, name, err := parsePrefixListId(d.Id())
    if err != nil {
        return diag.Errorf("Error in parsing ID %q: %s", d.Id(), err)
    }

    tflog.Info(
        ctx, "delete prefix list",
        RulestackName, stack,
        "name", name,
    )

    if err := svc.Delete(ctx, stack, name); err != nil && !isObjectNotFound(err) {
        return diag.FromErr(err)
    }

    d.SetId("")
    return nil
}

// Schema handling.
func prefixListSchema(isResource bool, rmKeys []string) map[string] *schema.Schema {
    ans := map[string] *schema.Schema{
        ConfigTypeName: configTypeSchema(),
        RulestackName: rsSchema(),
        "name": {
            Type: schema.TypeString,
            Required: true,
            Description: "The name.",
            ForceNew: true,
        },
        "description": {
            Type: schema.TypeString,
            Optional: true,
            Description: "The description.",
        },
        "prefix_list": {
            Type: schema.TypeSet,
            Required: true,
            Description: "The prefix list.",
            MinItems: 1,
            Elem: &schema.Schema{
                Type: schema.TypeString,
            },
        },
        "audit_comment": {
            Type: schema.TypeString,
            Optional: true,
            Description: "The audit comment.",
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
        computed(ans, "", []string{RulestackName, "name"})
    }

    return ans
}

func loadPrefixList(d *schema.ResourceData) prefix.Info {
    return prefix.Info{
        Rulestack: d.Get(RulestackName).(string),
        Name: d.Get("name").(string),
        Description: d.Get("description").(string),
        PrefixList: setToSlice(d.Get("prefix_list")),
        AuditComment: d.Get("audit_comment").(string),
    }
}

func savePrefixList(d *schema.ResourceData, stack, name string, o prefix.Info) {
    d.Set(RulestackName, stack)
    d.Set("name", name)
    d.Set("description", o.Description)
    d.Set("prefix_list", sliceToSet(o.PrefixList))
    d.Set("audit_comment", o.AuditComment)
    d.Set("update_token", o.UpdateToken)
}

// Id functions.
func buildPrefixListId(a, b string) string {
    return strings.Join([]string{a, b}, IdSeparator)
}

func parsePrefixListId(v string) (string, string, error) {
    tok := strings.Split(v, IdSeparator)
    if len(tok) != 2 {
        return "", "", fmt.Errorf("Expecting 2 tokens, got %d", len(tok))
    }

    return tok[0], tok[1], nil
}
