package kilt

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleBuild(t *testing.T) {
	targetInfoString, _ := ioutil.ReadFile("./fixtures/input.json")
	definitionString, _ := ioutil.ReadFile("./fixtures/kilt.cfg")
	k := NewKiltHocon(string(definitionString))
	info := new(TargetInfo)
	_ = json.Unmarshal(targetInfoString, info)
	b, _ := k.Build(info)

	assert.Equal(t, "busybox:latest", b.Image)
	assert.Equal(t, "/falco/pdig", b.EntryPoint[0])
	assert.Equal(t, "true", b.EnvironmentVariables["TEST"])
	assert.Equal(t, 1, len(b.Resources))
}

func TestEnvironmentVariables(t *testing.T) {
	targetInfoString, _ := ioutil.ReadFile("./fixtures/env_vars_input.json")
	definitionString, _ := ioutil.ReadFile("./fixtures/kilt_env_vars.cfg")

	k := NewKiltHocon(string(definitionString))
	info := new(TargetInfo)
	_ = json.Unmarshal(targetInfoString, info)
	b, _ := k.Build(info)

	assert.Containsf(t, b.EnvironmentVariables, "PREEXISTING", "does not contain preexisting vars")
	assert.Equal(t, "true", b.EnvironmentVariables["PREEXISTING"])
}
