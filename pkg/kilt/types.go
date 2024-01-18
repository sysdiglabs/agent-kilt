package kilt

import "github.com/Jeffail/gabs/v2"

type BuildResource struct {
	Name                 string
	Image                string
	Volumes              []string
	EntryPoint           []string
	EnvironmentVariables []map[string]interface{}
}

type Build struct {
	EnvParameters *gabs.Container
	Resources     []BuildResource
}

type Task struct {
	PidMode string // the only value is `task` right now
}

type PatchConfig struct {
	ParametrizeEnvars bool
}
