package kilt

import (
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/go-akka/configuration/hocon"
	"sort"

	"github.com/go-akka/configuration"
)

func renderHoconValue(v *hocon.HoconValue) interface{} {
	if v.IsObject() {
		obj := v.GetObject()
		items := obj.Items()
		if len(items) == 0 {
			return nil
		}

		dics := map[string]interface{}{}
		for k, v := range items {
			dics[k] = renderHoconValue(v)
		}

		return dics
	} else if v.IsArray() {
		arr := v.GetArray()
		if len(arr) == 0 {
			return nil
		}

		var items []interface{}
		for _, v := range arr {
			items = append(items, renderHoconValue(v))
		}
		return items
	} else {
		return v.GetString()
	}
}

func extractBuild(config *configuration.Config) (*Build, error) {
	b := new(Build)

	b.Image = config.GetString("build.image")

	b.EntryPoint = gabs.New()
	b.EntryPoint.Set(make([]interface{}, 0))
	rawEntryPoint := config.GetValue("build.entry_point").GetArray()
	if rawEntryPoint != nil {
		for _, c := range rawEntryPoint {
			b.EntryPoint.ArrayAppend(renderHoconValue(c))
		}
	}

	b.Command = gabs.New()
	b.Command.Set(make([]interface{}, 0))
	rawCommand := config.GetValue("build.command").GetArray()
	if rawCommand != nil {
		for _, c := range rawCommand {
			b.Command.ArrayAppend(renderHoconValue(c))
		}
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
