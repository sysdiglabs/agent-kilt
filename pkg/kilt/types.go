package kilt

import "github.com/Jeffail/gabs/v2"

type TargetInfo struct {
	Image                *gabs.Container   `json:"image"`
	ContainerName        *gabs.Container   `json:"container_name"`
	ContainerGroupName   string            `json:"container_group_name"`
	EntryPoint           *gabs.Container   `json:"entry_point"`
	Command              *gabs.Container   `json:"command"`
	EnvironmentVariables map[string]string `json:"environment_variables"`
}

type BuildResource struct {
	Name                 string
	Image                string
	Volumes              []string
	EntryPoint           []string
	EnvironmentVariables []map[string]interface{}
}

type Build struct {
	Image                *gabs.Container
	EntryPoint           *gabs.Container
	Command              *gabs.Container
	EnvironmentVariables map[string]string
	Capabilities         []string

	Resources []BuildResource
}

type Task struct {
	PidMode string // the only value is `task` right now
}
