package provider

import (
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func rsSchema() *schema.Schema {
    return &schema.Schema{
        Type: schema.TypeString,
        Required: true,
        Description: "The rulestack.",
        ForceNew: true,
    }
}

func ruleListSchema() *schema.Schema {
    return &schema.Schema{
        Type: schema.TypeString,
        Optional: true,
        Description: "The rulebase.",
        Default: "PreRule",
        ValidateFunc: validation.StringInSlice(
            []string{"PreRule", "PostRule", "LocalRule"},
            false,
        ),
        ForceNew: true,
    }
}

func toStringSlice(v interface{}) []string {
    if v == nil {
        return nil
    }

    vlist, ok := v.([]interface{})
    if !ok {
        return nil
    }

    ans := make([]string, len(vlist))
    for i := range vlist {
        ans[i] = vlist[i].(string)
    }

    return ans
}
