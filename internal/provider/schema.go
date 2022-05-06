package provider

import (
	"github.com/paloaltonetworks/cloud-ngfw-aws-go/tag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func scopeSchema() *schema.Schema {
	scopes := []string{"Local", "Global"}

	return &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  addStringInSliceValidation("The rulestack's scope. A local rulestack will require that you've retrieved a LRA JWT. A global rulestack will require that you've retrieved a GRA JWT.", scopes),
		Default:      scopes[0],
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(scopes, false),
	}
}

func rsSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The rulestack.",
		ForceNew:    true,
	}
}

func gRsSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "The global rulestack.",
		ForceNew:    true,
	}
}

func ruleListSchema() *schema.Schema {
	opts := []string{"PreRule", "PostRule", "LocalRule"}

	return &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  addStringInSliceValidation("The rulebase.", opts),
		Default:      "PreRule",
		ValidateFunc: validation.StringInSlice(opts, false),
		ForceNew:     true,
	}
}

func configTypeSchema() *schema.Schema {
	opts := []string{"candidate", "running"}

	return &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  addStringInSliceValidation("Retrieve either the candidate or running config.", opts),
		Default:      "candidate",
		ValidateFunc: validation.StringInSlice(opts, false),
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

func tagsSchema(isOptional bool) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeMap,
		Optional:    isOptional,
		Computed:    !isOptional,
		Description: "The tags.",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
}

func loadTags(v interface{}) []tag.Details {
	if v == nil {
		return nil
	}

	v2, ok := v.(map[string]interface{})
	if !ok || len(v2) == 0 {
		return nil
	}

	ans := make([]tag.Details, 0, len(v2))
	for k, v := range v2 {
		ans = append(ans, tag.Details{
			Key:   k,
			Value: v.(string),
		})
	}

	return ans
}

func dumpTags(list []tag.Details) map[string]interface{} {
	if len(list) == 0 {
		return nil
	}

	ans := make(map[string]interface{})
	for _, x := range list {
		ans[x.Key] = x.Value
	}

	return ans
}
