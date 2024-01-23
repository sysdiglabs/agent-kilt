package kilt

import (
	"github.com/Jeffail/gabs/v2"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func toStringOrEmpty(c interface{}) string {
	switch c.(type) {
	case string:
		return c.(string)
	default:
		return ""
	}
}

func readInput(path string) *TargetInfo {
	targetInfoString, _ := os.ReadFile(path)
	gabsInfo, _ := gabs.ParseJSON(targetInfoString)
	info := new(TargetInfo)
	info.Image = gabsInfo.S("image")
	info.ContainerName = gabsInfo.S("container_name")
	info.ContainerGroupName = toStringOrEmpty(gabsInfo.S("container_group_name").Data())
	info.EntryPoint = gabsInfo.S("entry_point")
	info.Command = gabsInfo.S("command")
	info.EnvironmentVariables = make(map[string]*gabs.Container)
	for k, v := range gabsInfo.S("environment_variables").ChildrenMap() {
		info.EnvironmentVariables[k] = v
	}
	return info
}

func getEnvByName(container *gabs.Container, name string) *string {
	for _, env := range container.S("Environment").Children() {
		if env.S("Name").Data().(string) == name {
			value := env.S("Value").Data().(string)
			return &value
		}
	}
	return nil
}

func TestSimpleBuild(t *testing.T) {
	info := readInput("./fixtures/input.json")
	definitionString, _ := os.ReadFile("./fixtures/kilt.cfg")

	k := NewKiltHocon(string(definitionString))
	container := gabs.New()
	b, _ := k.Patch(container, &PatchConfig{}, info)

	assert.Equal(t, "busybox:latest", toStringOrEmpty(container.S("Image").Data()))
	assert.Equal(t, "/falco/pdig", toStringOrEmpty(container.S("EntryPoint").Children()[0].Data()))
	assert.Equal(t, "true", *getEnvByName(container, "TEST"))
	assert.Equal(t, 1, len(b.Resources))
}

func TestEnvironmentVariables(t *testing.T) {
	info := readInput("./fixtures/env_vars_input.json")
	definitionString, _ := os.ReadFile("./fixtures/kilt_env_vars.cfg")

	k := NewKiltHocon(string(definitionString))
	container := gabs.New()
	k.Patch(container, &PatchConfig{}, info)

	assert.Equal(t, "true", *getEnvByName(container, "PREEXISTING"))
}
