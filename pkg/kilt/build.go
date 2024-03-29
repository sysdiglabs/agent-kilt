package kilt

import (
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/go-akka/configuration/hocon"
	"regexp"
	"sort"
	"strings"

	"github.com/go-akka/configuration"
)

func renderHoconValue(v *hocon.HoconValue) interface{} {
	if v == nil {
		return nil
	}
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

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9][^\W_]?`)

func getParameterName(envarName string) string {
	// Use the envar name if it does not contain any non-alphanumeric chars
	if found := nonAlphanumericRegex.MatchString(envarName); !found {
		return envarName
	}

	// Otherwise, try to make it more readable, e.g. MY_AWESOME_ENVAR becomes myAwesomeEnvar
	parameterName := nonAlphanumericRegex.ReplaceAllStringFunc(strings.ToLower(envarName), func(str string) string {
		return strings.ToUpper(str[1:])
	})
	return parameterName
}

func patchEnvironment(container *gabs.Container, env *hocon.HoconValue, overwrite bool, parametrize bool) error {
	if env == nil || !env.IsObject() {
		return nil
	}
	envMap := make(map[string]interface{})
	existingVars := make(map[string]struct{})

	existingEnv := container.S("Environment").Children()
	for _, v := range existingEnv {
		var varName string
		switch v.S("Name").Data().(type) {
		case string:
			varName = v.S("Name").Data().(string)
		default:
			return fmt.Errorf("could not parse environment variable name: %v", v.S("Name").Data())
		}

		envMap[varName] = v.S("Value").Data()
		existingVars[varName] = struct{}{}
	}

	for k, v := range env.GetObject().Items() {
		if _, ok := envMap[k]; ok && !overwrite {
			continue
		}
		envMap[k] = renderHoconValue(v)
	}

	if len(envMap) == 0 {
		return nil
	}

	keys := make([]string, 0, len(envMap))
	for k, _ := range envMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	_, err := container.Set(make([]interface{}, 0), "Environment")
	if err != nil {
		return fmt.Errorf("could not set empty environment: %w", err)
	}
	for _, k := range keys {
		keyValue := make(map[string]interface{})
		keyValue["Name"] = k
		v := envMap[k]
		switch v.(type) {
		case string:
			if _, ok := existingVars[k]; !ok && parametrize {
				keyValue["Value"] = map[string]interface{}{"Ref": getParameterName(k)}
			} else {
				keyValue["Value"] = v
			}
		default:
			keyValue["Value"] = v
		}

		_, err := container.Set(keyValue, "Environment", "-")
		if err != nil {
			return fmt.Errorf("could not add environment variable %v: %w", keyValue, err)
		}
	}

	return nil
}

func getTaskParameters(config *configuration.Config, patchConfig *PatchConfig) *gabs.Container {
	if !patchConfig.ParametrizeEnvars {
		return nil
	}

	env := config.GetValue("build.environment_variables")
	if env == nil || !env.IsObject() {
		return nil
	}

	taskParameters := gabs.New()
	taskParameters.Set(make(map[string]interface{}))
	for k, v := range env.GetObject().Items() {
		keyStripped := getParameterName(k)
		taskParameters.Set("String", "Parameters", keyStripped, "Type")
		taskParameters.Set(renderHoconValue(v), "Parameters", keyStripped, "Default")
	}

	return taskParameters
}

func applyPatch(container *gabs.Container, config *configuration.Config, patchConfig *PatchConfig) (map[string]*gabs.Container, error) {
	_, err := container.Set(renderHoconValue(config.GetValue("build.image")), "Image")
	if err != nil {
		return nil, fmt.Errorf("could not set image: %w", err)
	}

	entryPoint := gabs.New()
	entryPoint.Set(make([]interface{}, 0))
	rawEntryPoint := config.GetValue("build.entry_point").GetArray()
	if rawEntryPoint != nil {
		for _, c := range rawEntryPoint {
			entryPoint.ArrayAppend(renderHoconValue(c))
		}
	}
	_, err = container.Set(entryPoint.Data(), "EntryPoint")
	if err != nil {
		return nil, fmt.Errorf("could not set entry point: %w", err)
	}

	command := gabs.New()
	command.Set(make([]interface{}, 0))
	rawCommand := config.GetValue("build.command").GetArray()
	if rawCommand != nil {
		for _, c := range rawCommand {
			command.ArrayAppend(renderHoconValue(c))
		}
	}
	_, err = container.Set(command, "Command")
	if err != nil {
		return nil, fmt.Errorf("could not set command: %w", err)
	}

	capabilities := config.GetStringList("build.capabilities")
	if capabilities != nil {
		for _, c := range capabilities {
			err = container.ArrayAppend(c, "LinuxParameters", "Capabilities", "Add")
			if err != nil {
				return nil, fmt.Errorf("could not append to LinuxParameters.Capabilities.Add: %w", err)
			}
		}
	}

	env := config.GetValue("build.environment_variables")
	err = patchEnvironment(container, env, true, patchConfig.ParametrizeEnvars)
	if err != nil {
		return nil, err
	}

	sidecars := make(map[string]*gabs.Container)
	if config.IsArray("build.mount") {
		mounts := config.GetValue("build.mount").GetArray()
		sidecarConfig := gabs.New()
		sidecarConfig.Set(renderHoconValue(config.GetValue("sidecar_config")))

		for k, m := range mounts {
			if m.IsObject() {
				mount := m.GetObject()

				sidecarName := mount.GetKey("name").GetString()
				sidecarImage := mount.GetKey("image").GetString()
				if sidecarName == "" || sidecarImage == "" {
					return nil, fmt.Errorf("error at build.mount.%d: name and image are required ", k)
				}

				if len(mount.GetKey("volumes").GetStringList()) > 0 {
					addVolume := map[string]interface{}{
						"ReadOnly":        true,
						"SourceContainer": sidecarName,
					}

					err := container.ArrayAppend(addVolume, "VolumesFrom")
					if err != nil {
						return nil, fmt.Errorf("could not add VolumesFrom directive: %w", err)
					}
				}

				sidecar := gabs.New()
				sidecar.Set(map[string]interface{}{
					"Name":  sidecarName,
					"Image": sidecarImage,
				})

				sidecarEntryPoint := mount.GetKey("entry_point").GetStringList()
				if sidecarEntryPoint != nil && len(sidecarEntryPoint) > 0 {
					sidecar.Set(sidecarEntryPoint, "EntryPoint")
				}

				sidecarEnv := mount.GetKey("environment_variables")
				sidecarEnvKv := make([]interface{}, 0)
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

						sidecarEnvKv = append(sidecarEnvKv, keyValue)
					}
				}

				if len(sidecarEnvKv) > 0 {
					sidecar.Set(sidecarEnvKv, "Environment")
				}

				err := patchEnvironment(sidecar, config.GetValue("original.environment_variables"), false, false)
				if err != nil {
					return nil, err
				}

				err = patchEnvironment(sidecar, env, false, patchConfig.ParametrizeEnvars)
				if err != nil {
					return nil, err
				}

				err = sidecar.Merge(sidecarConfig)
				if err != nil {
					return nil, fmt.Errorf("could not merge sidecar configuration: %w", err)
				}
				sidecars[sidecarName] = sidecar
			}
		}
	}

	return sidecars, nil
}
