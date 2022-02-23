package provider

import (
    "context"

    "github.com/paloaltonetworks/cloud-ngfw-aws-go"
    "github.com/paloaltonetworks/cloud-ngfw-aws-go/rulestack"

    //"github.com/hashicorp/terraform-plugin-log/tflog"
    "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Data source.

// Resource.
func resourceRulestack() *schema.Resource {
    return &schema.Resource{
        Description: "Resource for rulestack manipulation.",

        CreateContext: createRulestack,
        ReadContext: readRulestack,
        UpdateContext: updateRulestack,
        DeleteContext: deleteRulestack,

        Importer: &schema.ResourceImporter{
            State: schema.ImportStatePassthrough,
        },

        Schema: rulestackSchema(true, nil),
    }
}

func createRulestack(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    svc := rulestack.NewClient(meta.(*awsngfw.Client))
    o := loadRulestack(d)

    if err := svc.Create(ctx, o); err != nil {
        return diag.FromErr(err)
    }

    d.SetId(o.Name)

    return readRulestack(ctx, d, meta)
}

func readRulestack(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    svc := rulestack.NewClient(meta.(*awsngfw.Client))
    name := d.Id()
    req := rulestack.ReadInput{
        Name: name,
        Candidate: true,
    }

    res, err := svc.Read(ctx, req)
    if err != nil {
        if isObjectNotFound(err) {
            d.SetId("")
            return nil
        }
        return diag.FromErr(err)
    }

    saveRulestack(d, res.Response.Name, res.Response.Candidate)

    return nil
}

func updateRulestack(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    svc := rulestack.NewClient(meta.(*awsngfw.Client))
    o := loadRulestack(d)

    if err := svc.Update(ctx, o); err != nil {
        return diag.FromErr(err)
    }

    d.SetId(o.Name)
    return readRulestack(ctx, d, meta)
}

func deleteRulestack(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    svc := rulestack.NewClient(meta.(*awsngfw.Client))
    name := d.Id()

    if err := svc.Delete(ctx, name); err != nil && !isObjectNotFound(err) {
        return diag.FromErr(err)
    }

    d.SetId("")
    return nil
}

// Schema handling.
func rulestackSchema(isResource bool, rmKeys []string) map[string] *schema.Schema {
    ans := map[string] *schema.Schema{
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
        "scope": {
            Type: schema.TypeString,
            Optional: true,
            Description: "The scope.",
        },
        "account_id": {
            Type: schema.TypeString,
            Optional: true,
            Description: "The account ID.",
        },
        "account_group": {
            Type: schema.TypeString,
            Optional: true,
            Description: "Account group.",
        },
        "minimum_app_id_version": {
            Type: schema.TypeString,
            Optional: true,
            Computed: true,
            Description: "Minimum App-ID version number.",
        },
        "tags": {
            Type: schema.TypeList,
            Optional: true,
            Description: "Tags.",
            Elem: &schema.Schema{
                Type: schema.TypeString,
            },
        },
        "profile_config": {
            Type: schema.TypeList,
            Required: true,
            MaxItems: 1,
            MinItems: 1,
            Elem: &schema.Resource{
                Schema: map[string] *schema.Schema{
                    "anti_spyware": {
                        Type: schema.TypeString,
                        Optional: true,
                        Description: "Anti-spyware profile setting.",
                        Default: "BestPractice",
                    },
                    "anti_virus": {
                        Type: schema.TypeString,
                        Optional: true,
                        Description: "Anti-virus profile setting.",
                        Default: "BestPractice",
                    },
                    "vulnerability": {
                        Type: schema.TypeString,
                        Optional: true,
                        Description: "Vulnerability profile setting.",
                        Default: "BestPractice",
                    },
                    "url_filtering": {
                        Type: schema.TypeString,
                        Optional: true,
                        Description: "URL filtering profile setting.",
                        Default: "None",
                    },
                    "file_blocking": {
                        Type: schema.TypeString,
                        Optional: true,
                        Description: "File blocking profile setting.",
                        Default: "BestPractice",
                    },
                    "outbound_trust_certificate": {
                        Type: schema.TypeString,
                        Optional: true,
                        Description: "Outbound trust certificate.",
                    },
                    "outbound_untrust_certificate": {
                        Type: schema.TypeString,
                        Optional: true,
                        Description: "Outbound untrust certificate.",
                    },
                },
            },
        },
    }

    for _, rmKey := range rmKeys {
        delete(ans, rmKey)
    }

    if !isResource {
        computed(ans, "", []string{"name"})
    }

    return ans
}

func loadRulestack(d *schema.ResourceData) rulestack.Info {
    p := configFolder(d, "profile_config")

    return rulestack.Info{
        Name: d.Get("name").(string),
        Entry: rulestack.Details{
            Scope: d.Get("scope").(string),
            AccountId: d.Get("account_id").(string),
            AccountGroup: d.Get("account_group").(string),
            MinimumAppIdVersion: d.Get("minimum_app_id_version").(string),
            Profile: rulestack.ProfileConfig{
                AntiSpyware: p["anti_spyware"].(string),
                AntiVirus: p["anti_virus"].(string),
                Vulnerability: p["vulnerability"].(string),
                UrlFiltering: p["url_filtering"].(string),
                FileBlocking: p["file_blocking"].(string),
                OutboundTrustCertificate: p["outbound_trust_certificate"].(string),
                OutboundUntrustCertificate: p["outbound_untrust_certificate"].(string),
            },
        },
    }
}

func saveRulestack(d *schema.ResourceData, name string, o *rulestack.Details) {
    pc := map[string] interface{}{
        "anti_spyware": o.Profile.AntiSpyware,
        "anti_virus": o.Profile.AntiVirus,
        "vulnerability": o.Profile.Vulnerability,
        "url_filtering": o.Profile.UrlFiltering,
        "file_blocking": o.Profile.FileBlocking,
        "outbound_trust_certificate": o.Profile.OutboundTrustCertificate,
        "outbound_untrust_certificate": o.Profile.OutboundUntrustCertificate,
    }

    d.Set("name", name)
    d.Set("scope", o.Scope)
    d.Set("account_id", o.AccountId)
    d.Set("account_group", o.AccountGroup)
    d.Set("minimum_app_id_version", o.MinimumAppIdVersion)
    d.Set("profile_config", []interface{}{pc})
}
