package kilt

import "github.com/Jeffail/gabs/v2"

type Build struct {
	Sidecars map[string]*gabs.Container
}

type Task struct {
	PidMode string // the only value is `task` right now
}

type PatchConfig struct {
	ParametrizeEnvars bool
}
