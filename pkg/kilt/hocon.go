package kilt

import (
	"encoding/json"
	"fmt"

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

type HoconProvided struct {
	Image string `json:"image"`
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
	rawVars, err := json.Marshal(info)
	if err != nil {
		return nil, fmt.Errorf("could not serialize info: %w", err)
	}
	rawEnv, _ := json.Marshal(info.EnvironmentVariables) // we would fail at info step

	configString := "original:" + string(rawVars) + "\n" +
		"config:" + k.config + "\n" +
		defaults + "build.environment_variables: " + string(rawEnv) + "\n" +
		k.definition

	return configuration.ParseString(configString), nil
}

func (k *KiltHocon) Build(info *TargetInfo) (*Build, error) {
	config, err := k.prepareFullStringConfig(info)
	if err != nil {
		return nil, fmt.Errorf("could not assemble full config: %w", err)
	}

	return extractBuild(config)
}

func (k *KiltHocon) Task() (*Task, error) {
	config, err := k.prepareFullStringConfig(&TargetInfo{})
	if err != nil {
		return nil, fmt.Errorf("could not assemble full config: %w", err)
	}
	return extractTask(config)
}
