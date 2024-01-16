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
	info.Image = toStringOrEmpty(gabsInfo.S("image").Data())
	info.ContainerName = toStringOrEmpty(gabsInfo.S("container_name").Data())
	info.ContainerGroupName = toStringOrEmpty(gabsInfo.S("container_group_name").Data())
	info.EntryPoint = gabsInfo.S("entry_point")
	info.Command = gabsInfo.S("command")
	info.EnvironmentVariables = make(map[string]string)
	for k, v := range gabsInfo.S("environment_variables").ChildrenMap() {
		info.EnvironmentVariables[k] = v.Data().(string)
	}
	return info
}

func TestSimpleBuild(t *testing.T) {
	info := readInput("./fixtures/input.json")
	definitionString, _ := os.ReadFile("./fixtures/kilt.cfg")

	k := NewKiltHocon(string(definitionString))
	b, _ := k.Build(info)

	assert.Equal(t, "busybox:latest", b.Image)
	assert.Equal(t, "/falco/pdig", b.EntryPoint.Children()[0].Data())
	assert.Equal(t, "true", b.EnvironmentVariables["TEST"])
	assert.Equal(t, 1, len(b.Resources))
}

func TestEnvironmentVariables(t *testing.T) {
	info := readInput("./fixtures/env_vars_input.json")
	definitionString, _ := os.ReadFile("./fixtures/kilt_env_vars.cfg")

	k := NewKiltHocon(string(definitionString))
	b, _ := k.Build(info)

	assert.Containsf(t, b.EnvironmentVariables, "PREEXISTING", "does not contain preexisting vars")
	assert.Equal(t, "true", b.EnvironmentVariables["PREEXISTING"])
}
