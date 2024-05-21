package provider

import (
	"fmt"
	"strings"

	"github.com/paloaltonetworks/cloud-ngfw-aws-go/api/response"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func addStringInSliceValidation(desc string, values []string) string {
	var b strings.Builder
	b.Grow(len(desc) + 20*len(values))
	b.WriteString(desc)

	if len(values) > 0 {
		b.WriteString(" Valid values are")

		for i := range values {
			if i != 0 && len(values) > 2 {
				b.WriteString(",")
			}
			b.WriteString(" ")
			if i == len(values)-1 {
				b.WriteString("or ")
			}
			b.WriteString("`")
			b.WriteString(values[i])
			b.WriteString("`")
		}
		b.WriteString(".")
	}

	return b.String()
}

func addIntBetweenValidation(desc string, low, high int) string {
	return fmt.Sprintf("%s The number must be between [%d, %d] incluside.", desc, low, high)
}

func addProviderParamDescription(desc, envName, jsonName string) string {
	var buf strings.Builder

	buf.WriteString(desc)
	if envName != "" {
		buf.WriteString(" Environment variable: `")
		buf.WriteString(envName)
		buf.WriteString("`.")
	}
	if jsonName != "" {
		buf.WriteString(" JSON conf file variable: `")
		buf.WriteString(jsonName)
		buf.WriteString("`.")
	}

	return buf.String()
}

func configTypeId(a, b string) string {
	return strings.Join([]string{a, b}, IdSeparator)
}

func configFolder(v interface{}) map[string]interface{} {
	if v != nil {
		if ilist, ok := v.([]interface{}); ok && ilist != nil && len(ilist) == 1 {
			if ans, ok := ilist[0].(map[string]interface{}); ok && ans != nil {
				return ans
			}
		}
	}

	return nil
}

func computed(sm map[string]*schema.Schema, parent string, omits []string) {
	for key, s := range sm {
		stop := false
		for _, o := range omits {
			if parent == "" {
				if o == key {
					stop = true
					break
				}
			} else if o == parent+"."+key {
				stop = true
				break
			}
		}
		if stop {
			continue
		}
		s.Computed = true
		s.Required = false
		s.Optional = false
		s.MinItems = 0
		s.MaxItems = 0
		s.Default = nil
		s.DiffSuppressFunc = nil
		s.DefaultFunc = nil
		s.ConflictsWith = nil
		s.ExactlyOneOf = nil
		s.AtLeastOneOf = nil
		s.ValidateFunc = nil
		//s.RequiredWith = nil
		if s.Type == schema.TypeList || s.Type == schema.TypeSet {
			switch et := s.Elem.(type) {
			case *schema.Resource:
				var path string
				if parent == "" {
					path = key
				} else {
					path = parent + "." + key
				}
				computed(et.Schema, path, omits)
			}
		}
	}
}

func isObjectNotFound(e error) bool {
	if e2, ok := e.(*response.Status); ok {
		return e2.ObjectNotFound()
	}

	return false
}

func isBadRequest(e error) bool {
	if e2, ok := e.(*response.Status); ok {
		return e2.Code == 400
	}
	return false
}

func Contains(inputStr string, strList []string) bool {
	for _, str := range strList {
		if inputStr == str {
			return true
		}
	}
	return false
}

/*
func structToMap(data interface{}) (map[string]interface{}, error) {
	mapData := make(map[string]interface{})

	values := reflect.ValueOf(data)
	keys := values.Type()

	for i := 0; i < values.NumField(); i++ {
		mapData[keys.Field(i).Name] = values.Field(i).Interface()
	}
	return mapData, nil
}
*/

func PtrToString(s string) *string {
	return &s
}
