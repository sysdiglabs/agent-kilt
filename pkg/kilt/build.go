package kilt

import (
	"fmt"
	"sort"

	"github.com/go-akka/configuration"
)

func extractBuild(config *configuration.Config) (*Build, error) {
	b := new(Build)

	b.Image = config.GetString("build.image")
	b.EntryPoint = config.GetStringList("build.entry_point")
	if b.EntryPoint == nil {
		b.EntryPoint = make([]string, 0)
	}
	b.Command = config.GetStringList("build.command")
	if b.Command == nil {
		b.Command = make([]string, 0)
	}

	b.Capabilities = config.GetStringList("build.capabilities")
	if b.Capabilities == nil {
		b.Capabilities = make([]string, 0)
	}

	b.EnvironmentVariables = extractToStringMap(config, "build.environment_variables")

	if config.IsArray("build.mount") {
		mounts := config.GetValue("build.mount").GetArray()

		for k, m := range mounts {
			if m.IsObject() {
				mount := m.GetObject()

				resource := BuildResource{
					Name:       mount.GetKey("name").GetString(),
					Image:      mount.GetKey("image").GetString(),
					Volumes:    mount.GetKey("volumes").GetStringList(),
					EntryPoint: mount.GetKey("entry_point").GetStringList(),
				}

				sidecarEnv := mount.GetKey("environment_variables")
				if sidecarEnv != nil && sidecarEnv.IsObject() {
					obj := sidecarEnv.GetObject()
					items := obj.Items()
					keys := make([]string, 0, len(items))
					for k := range items {
						keys = append(keys, k)
					}
					sort.Strings(keys)

					for _, k := range keys {
						keyValue := make(map[string]interface{})
						keyValue["Name"] = k
						keyValue["Value"] = items[k].GetString()

						resource.EnvironmentVariables = append(resource.EnvironmentVariables, keyValue)
					}
				}

				if resource.Image == "" || len(resource.Volumes) == 0 || len(resource.EntryPoint) == 0 {
					return nil, fmt.Errorf("error at build.mount.%d: image, volumes and entry_point are all required ", k)
				}

				b.Resources = append(b.Resources, resource)
			}
		}
	}

	return b, nil
}
