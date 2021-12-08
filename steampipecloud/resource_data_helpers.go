package steampipecloud

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/iancoleman/strcase"
	"github.com/turbot/terraform-provider-steampipecloud/helpers"
)

// given the resource data and a list of properties, construct a map of property values
func mapFromResourceData(d *schema.ResourceData, properties []interface{}) map[string]interface{} {
	// each element in the 'properties' array is either a map defining explicit name mappings, or a string containing the terraform property name.
	// this is converted to the turbot property name by performing a snake case -> lowerCamelCase conversion
	// to build the output map:
	// 1) extract the value from ResourceData using the terraform property name
	// 2) add the property to a map using the turbot property name
	var propertyMap = map[string]interface{}{}
	for _, element := range properties {
		terraformToTurbotMap, ok := element.(map[string]string)
		if ok {
			helpers.MergeMaps(propertyMap, mapFromResourceDataWithPropertyMap(d, terraformToTurbotMap))
		} else {
			// otherwise perform automatic mapping from snake case (Terraform format) to lowerCamelCase (Turbot format).
			terraformProperty := element.(string)
			value, propertySet := d.GetOk(terraformProperty)
			// if property is set, map it
			if propertySet {
				var turbotProperty = strcase.ToLowerCamel(terraformProperty)
				propertyMap[turbotProperty] = value
			}
		}
	}
	return propertyMap
}

// given the resource data and a map of properties (terraform property name -> output property name), construct a map of property values
func mapFromResourceDataWithPropertyMap(d *schema.ResourceData, terraformToTurbotMap map[string]string) map[string]interface{} {
	var resourcePropertyMap = map[string]interface{}{}
	for terraform, turbot := range terraformToTurbotMap {
		// get schema for property
		value, propertySet := d.GetOk(terraform)
		if propertySet {
			resourcePropertyMap[turbot] = value
		}
	}
	return resourcePropertyMap
}
