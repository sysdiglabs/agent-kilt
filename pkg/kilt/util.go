package kilt

import (
	"github.com/go-akka/configuration"
)

func extractToStringMap(config *configuration.Config, path string) map[string]string {
	value := make(map[string]string)

	if config.HasPath(path) && config.IsObject(path) {
		obj := config.GetNode(path).GetObject()

		for k, v := range obj.Items() {
			value[k] = v.GetString()
		}
	}

	return value
}
