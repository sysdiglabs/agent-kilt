package cfnpatcher

import (
	"fmt"
	"github.com/Jeffail/gabs/v2"
)

func getOptTags(template *gabs.Container) map[string]string {
	optTags := make(map[string]string)
	if !template.Exists("Properties", "Tags") {
		return optTags
	}
	for _, tag := range template.S("Properties", "Tags").Children() {
		if tag.Exists("Key") && tag.Exists("Value") {
			k, ok := tag.S("Key").Data().(string)
			if !ok {
				panic(fmt.Errorf("tag has an unsupported key type: %s", tag.String()))
			}
			if isOptTagKey(k) {
				v, ok := tag.S("Value").Data().(string)
				if !ok {
					panic(fmt.Errorf("OptIn/OptOut tag %s has an unsupported value type: %s", k, v))
				}
				optTags[k] = v
			}
		}
	}
	return optTags
}

func hasType(resource *gabs.Container, resourceType string) bool {
	exists := resource.Exists("Type")
	if !exists {
		return false
	}
	value, ok := resource.S("Type").Data().(string)
	return ok && value == resourceType
}

func isTaskDefinition(resource *gabs.Container) bool {
	return hasType(resource, "AWS::ECS::TaskDefinition")
}

func matchFargate(resource *gabs.Container) bool {
	return isTaskDefinition(resource) && requiresFargate(resource)
}

func requiresFargate(resource *gabs.Container) bool {
	exists := resource.Exists("Properties", "RequiresCompatibilities")
	if !exists {
		return false
	}
	for _, value := range resource.S("Properties", "RequiresCompatibilities").Children() {
		compat, ok := value.Data().(string)
		if ok && compat == "FARGATE" {
			return true
		}
	}
	return false
}
