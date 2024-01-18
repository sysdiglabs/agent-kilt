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

func readInput(path string) (*gabs.Container, string) {
	targetInfoString, _ := os.ReadFile(path)
	gabsInfo, _ := gabs.ParseJSON(targetInfoString)
	container := gabs.New()
	container.Set(gabsInfo.S("image").Data(), "Image")
	container.Set(gabsInfo.S("name").Data(), "Name")
	containerGroupName := toStringOrEmpty(gabsInfo.S("container_group_name").Data())
	container.Set(gabsInfo.S("entry_point").Data(), "EntryPoint")
	container.Set(gabsInfo.S("command").Data(), "Command")
	for k, v := range gabsInfo.S("environment_variables").ChildrenMap() {
		env := make(map[string]interface{})
		env["Name"] = k
		env["Value"] = v.Data()
		container.ArrayAppend(env, "Environment")
	}
	return container, containerGroupName
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
	container, groupName := readInput("./fixtures/input.json")
	definitionString, _ := os.ReadFile("./fixtures/kilt.cfg")

	k := NewKiltHocon(string(definitionString))
	b, _ := k.PatchContainerDefinition(container, &PatchConfig{}, groupName)

	assert.Equal(t, "busybox:latest", toStringOrEmpty(container.S("Image").Data()))
	assert.Equal(t, "/falco/pdig", toStringOrEmpty(container.S("EntryPoint").Children()[0].Data()))
	assert.Equal(t, "true", *getEnvByName(container, "TEST"))
	assert.Equal(t, 1, len(b.Sidecars))
}

func TestEnvironmentVariables(t *testing.T) {
	container, groupName := readInput("./fixtures/env_vars_input.json")
	definitionString, _ := os.ReadFile("./fixtures/kilt_env_vars.cfg")

	k := NewKiltHocon(string(definitionString))
	k.PatchContainerDefinition(container, &PatchConfig{}, groupName)

	assert.Equal(t, "true", *getEnvByName(container, "PREEXISTING"))
}
