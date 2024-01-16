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
	definition string
	config     string
}

func NewKiltHocon(definition string) *KiltHocon {
	return NewKiltHoconWithConfig(definition, "{}")
}

func NewKiltHoconWithConfig(definition string, recipeConfig string) *KiltHocon {
	h := new(KiltHocon)
	h.definition = definition
	h.config = recipeConfig
	return h
}

func (k *KiltHocon) prepareFullStringConfig(info *TargetInfo) (*configuration.Config, error) {
	rawVars := ""

	jsonDoc, err := json.Marshal(info.Image)
	if err != nil {
		return nil, fmt.Errorf("could not serialize container image: %w", err)
	}
	rawVars += "original.image:" + string(jsonDoc) + "\n"

	jsonDoc, err = json.Marshal(info.ContainerName)
	if err != nil {
		return nil, fmt.Errorf("could not serialize container name: %w", err)
	}
	rawVars += "original.container_name:" + string(jsonDoc) + "\n"

	jsonDoc, err = json.Marshal(info.ContainerGroupName)
	if err != nil {
		return nil, fmt.Errorf("could not serialize container group name: %w", err)
	}
	rawVars += "original.container_group_name:" + string(jsonDoc) + "\n"

	jsonDoc, err = json.Marshal(info.EntryPoint)
	if err != nil {
		return nil, fmt.Errorf("could not serialize container entry point: %w", err)
	}
	rawVars += "original.entry_point:" + string(jsonDoc) + "\n"

	jsonDoc, err = json.Marshal(info.Command)
	if err != nil {
		return nil, fmt.Errorf("could not serialize container command: %w", err)
	}
	rawVars += "original.command:" + string(jsonDoc) + "\n"

	rawEnv, err := json.Marshal(info.EnvironmentVariables)
	if err != nil {
		return nil, fmt.Errorf("could not serialize container environment variables: %w", err)
	}

	configString := string(rawVars) + "\n" +
		"config:" + k.config + "\n" +
		defaults + "build.environment_variables: " + string(rawEnv) + "\n" +
		k.definition

	return configuration.ParseString(configString), nil
}

func (k *KiltHocon) Patch(container *gabs.Container, info *TargetInfo) (*Build, error) {
	config, err := k.prepareFullStringConfig(info)
	if err != nil {
		return nil, fmt.Errorf("could not assemble full config: %w", err)
	}

	return applyPatch(container, config)
}

func (k *KiltHocon) Task() (*Task, error) {
	config, err := k.prepareFullStringConfig(&TargetInfo{})
	if err != nil {
		return nil, fmt.Errorf("could not assemble full config: %w", err)
	}
	return extractTask(config)
}
