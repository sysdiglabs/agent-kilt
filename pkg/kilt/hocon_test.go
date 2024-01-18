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
	containers := gabs.Wrap(make([]interface{}, 0))
	err := containers.ArrayAppend(container.Data())
	if err != nil {
		panic(err)
	}

	return containers, containerGroupName
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

func yes(container *gabs.Container) bool {
	return true
}

func TestSimpleBuild(t *testing.T) {
	containers, groupName := readInput("./fixtures/input.json")
	definitionString, _ := os.ReadFile("./fixtures/kilt.cfg")

	k := NewKiltHocon(string(definitionString))
	err := k.patchContainerDefinitions(containers, &PatchConfig{}, groupName, yes)
	if err != nil {
		panic(err)
	}
	container := containers.S("0")

	assert.Equal(t, "busybox:latest", toStringOrEmpty(container.S("Image").Data()))
	assert.Equal(t, "/falco/pdig", toStringOrEmpty(container.S("EntryPoint").Children()[0].Data()))
	assert.Equal(t, "true", *getEnvByName(container, "TEST"))
	numContainers, _ := containers.ArrayCount()
	assert.Equal(t, 2, numContainers)
}

func TestEnvironmentVariables(t *testing.T) {
	containers, groupName := readInput("./fixtures/env_vars_input.json")
	definitionString, _ := os.ReadFile("./fixtures/kilt_env_vars.cfg")

	k := NewKiltHocon(string(definitionString))
	err := k.patchContainerDefinitions(containers, &PatchConfig{}, groupName, yes)
	if err != nil {
		panic(err)
	}

	container := containers.S("0")
	assert.Equal(t, "true", *getEnvByName(container, "PREEXISTING"))
}
