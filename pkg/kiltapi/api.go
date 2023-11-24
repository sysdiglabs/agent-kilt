package kiltapi

import (
	"github.com/sysdiglabs/agent-kilt/pkg/hocon"
	"github.com/sysdiglabs/agent-kilt/pkg/kilt"
)

func NewKiltFromHocon(definition string) *kilt.Kilt {
	impl := hocon.NewKiltHocon(definition)
	return kilt.NewKilt(impl)
}

func NewKiltFromHoconWithConfig(definition string, config string) *kilt.Kilt {
	impl := hocon.NewKiltHoconWithConfig(definition, config)
	return kilt.NewKilt(impl)
}
