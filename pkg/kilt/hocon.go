package kilt

import (
	"encoding/json"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/go-akka/configuration"
)

var defaults = `
build {
	entry_point: ${original.entry_point}
	command: ${original.command}
	image: ${original.image}

	mount: []
}
`

type KiltHocon struct {
	definition    string
	config        string
	sidecarConfig interface{}
}

func NewKiltHocon(definition string) *KiltHocon {
	return NewKiltHoconWithConfig(definition, "{}", nil)
}

func NewKiltHoconWithConfig(definition string, recipeConfig string, sidecarConfig interface{}) *KiltHocon {
	h := new(KiltHocon)
	h.definition = definition
	h.config = recipeConfig
	h.sidecarConfig = sidecarConfig
	return h
}

func (k *KiltHocon) prepareFullStringConfig(container *gabs.Container, groupName string) (*configuration.Config, error) {
	rawVars := ""

	jsonDoc, err := json.Marshal(container.S("Image"))
	if err != nil {
		return nil, fmt.Errorf("could not serialize container image: %w", err)
	}
	rawVars += "original.image:" + string(jsonDoc) + "\n"

	jsonDoc, err = json.Marshal(container.S("Name"))
	if err != nil {
		return nil, fmt.Errorf("could not serialize container name: %w", err)
	}
	rawVars += "original.container_name:" + string(jsonDoc) + "\n"

	jsonDoc, err = json.Marshal(groupName)
	if err != nil {
		return nil, fmt.Errorf("could not serialize container group name: %w", err)
	}
	rawVars += "original.container_group_name:" + string(jsonDoc) + "\n"

	jsonDoc, err = json.Marshal(container.S("EntryPoint"))
	if err != nil {
		return nil, fmt.Errorf("could not serialize container entry point: %w", err)
	}
	rawVars += "original.entry_point:" + string(jsonDoc) + "\n"

	jsonDoc, err = json.Marshal(container.S("Command"))
	if err != nil {
		return nil, fmt.Errorf("could not serialize container command: %w", err)
	}
	rawVars += "original.command:" + string(jsonDoc) + "\n"

	rawEnvMap := make(map[string]interface{})
	for _, env := range container.S("Environment").Children() {
		rawEnvMap[env.S("Name").Data().(string)] = env.S("Value")
	}
	jsonDoc, err = json.Marshal(rawEnvMap)
	if err != nil {
		return nil, fmt.Errorf("could not serialize container environment variables: %w", err)
	}
	rawVars += "original.environment_variables:" + string(jsonDoc) + "\n"

	sidecarConfig := []byte("{}")
	if k.sidecarConfig != nil {
		sidecarConfig, err = json.Marshal(k.sidecarConfig)
		if err != nil {
			return nil, fmt.Errorf("could not serialize sidecar configuration: %w", err)
		}
	}

	configString := string(rawVars) + "\n" +
		"config:" + k.config + "\n" +
		"sidecar_config:" + string(sidecarConfig) + "\n" +
		defaults + "\n" +
		k.definition

	return configuration.ParseString(configString), nil
}

func (k *KiltHocon) patchContainerDefinitions(containers *gabs.Container, patchConfig *PatchConfig, groupName string, filter func(container *gabs.Container) bool) error {
	sidecars := make(map[string]*gabs.Container)

	for _, container := range containers.Children() {
		if filter(container) {
			config, err := k.prepareFullStringConfig(container, groupName)
			if err != nil {
				return fmt.Errorf("could not assemble full config: %w", err)
			}
			build, err := applyPatch(container, config, patchConfig)
			if err != nil {
				return fmt.Errorf("could not patch container definition %v: %w", container, err)
			}

			for name, sidecar := range build.Sidecars {
				sidecars[name] = sidecar
			}
		}
	}

	for sidecarName, sidecar := range sidecars {
		err := containers.ArrayAppend(sidecar)
		if err != nil {
			return fmt.Errorf("could not inject %s: %w", sidecarName, err)
		}
	}
	return nil
}

func (k *KiltHocon) PatchCfnTemplate(template *gabs.Container, patchConfig *PatchConfig) error {
	container := gabs.New()
	container.Set(make(map[string]interface{}))
	config, err := k.prepareFullStringConfig(container, "")
	if err != nil {
		return fmt.Errorf("could not assemble full config: %w", err)
	}
	params := getTaskParameters(config, patchConfig)
	err = template.Merge(params)
	if err != nil {
		return fmt.Errorf("could not merge parameters: %w", err)
	}
	return nil
}

func (k *KiltHocon) PatchTaskDefinition(taskdef *gabs.Container, patchConfig *PatchConfig, groupName string, filter func(container *gabs.Container) bool) error {
	container := gabs.New()
	container.Set(make(map[string]interface{}))
	config, err := k.prepareFullStringConfig(container, "")
	if err != nil {
		return fmt.Errorf("could not assemble full config: %w", err)
	}

	if config.HasPath("task.pid_mode") {
		pidMode := config.GetString("task.pid_mode")
		_, err = taskdef.Set(pidMode, "Properties", "PidMode")
		if err != nil {
			return fmt.Errorf("could not set PidMode: %w", err)
		}
	}

	containerDefinitions := taskdef.S("Properties", "ContainerDefinitions")
	if containerDefinitions != nil {
		err := k.patchContainerDefinitions(containerDefinitions, patchConfig, groupName, filter)
		if err != nil {
			return err
		}

		taskdef.Set(containerDefinitions, "Properties", "ContainerDefinitions")
	}
	return nil
}
