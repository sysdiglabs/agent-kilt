package kilt

import "github.com/Jeffail/gabs/v2"

type Build struct {
	Sidecars map[string]*gabs.Container
}

type PatchConfig struct {
	ParametrizeEnvars bool
}
