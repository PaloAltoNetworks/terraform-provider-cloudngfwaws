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

func setToSlice(v interface{}) []string {
    if v == nil {
        return nil
    }

    vs, ok := v.(*schema.Set)
    if !ok {
        return nil
    }

    list := vs.List()
    if len(list) == 0 {
        return nil
    }

    ans := make([]string, len(list))
    for i := range list {
        ans[i] = list[i].(string)
    }

    return ans
}

func sliceToSet(s []string) *schema.Set {
    var items []interface{}

    if len(s) > 0 {
        items = make([]interface{}, len(s))

        for i := range s {
            items[i] = s[i]
        }
    }

    return schema.NewSet(schema.HashString, items)
}
