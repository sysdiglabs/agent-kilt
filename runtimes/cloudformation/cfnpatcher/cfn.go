package cfnpatcher

import (
	"context"

	"github.com/Jeffail/gabs/v2"
	"github.com/rs/zerolog/log"
)

type Configuration struct {
	Kilt               string
	OptIn              bool
	RecipeConfig       string
	UseRepositoryHints bool
	LogGroup           string
	ParameterizeEnvars bool
	SidecarConfig      string
}

type InstrumentationHints struct {
	IgnoreContainersNamed  []string
	IncludeContainersNamed []string
	HasGlobalInclude       bool
}

func Patch(ctx context.Context, configuration *Configuration, fragment, templateParameters []byte) ([]byte, error) {
	l := log.Ctx(ctx)
	template, err := gabs.ParseJSON(fragment)
	if err != nil {
		l.Error().Err(err).Msg("failed to parse input fragment")
		return nil, err
	}

	if configuration.ParameterizeEnvars {
		l.Info().Msg("parameterizing recipe envars")
		applyParametersPatch(ctx, template, configuration)
	}

	var parameters *gabs.Container
	if len(templateParameters) > 0 {
		parameters, err = gabs.ParseJSON(templateParameters)
		if err != nil {
			l.Error().Err(err).Msg("failed to parse input templateParameters")
			return nil, err
		}
	}

	for name, resource := range template.S("Resources").ChildrenMap() {
		if matchFargate(resource) {
			optTags := getOptTags(resource)
			if isIgnored(optTags, configuration.OptIn) {
				l.Info().Str("resource", name).Msg("ignored resource due to tag")
				continue
			}

			l.Info().Str("resource", name).Msg("patching task definition")
			hints := extractHintsFromTags(optTags)
			_, err = applyTaskDefinitionPatch(ctx, name, resource, parameters, configuration, hints)
			if err != nil {
				l.Error().Err(err).Str("resource", name).Msgf("could not patch resource")
			}
		}
	}

	return template.Bytes(), nil
}
