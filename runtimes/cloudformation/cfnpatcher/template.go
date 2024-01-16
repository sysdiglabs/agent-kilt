package cfnpatcher

import (
	"context"
	"os"

	"github.com/Jeffail/gabs/v2"
	"github.com/rs/zerolog/log"

	"github.com/sysdiglabs/agent-kilt/pkg/kilt"
)

type TemplateInfo struct {
	TargetInfo *kilt.TargetInfo
	// Containers are not null when template values are complex
	EnvironmentVariables map[string]*gabs.Container
}

func GetValueFromTemplate(what *gabs.Container) (string, *gabs.Container) {
	var result string
	var fallback *gabs.Container

	switch v := what.Data().(type) {
	case string:
		result = v
		fallback = nil
	default:
		result = "placeholder: " + what.String()
		fallback = what
	}
	return result, fallback
}

func extractContainerInfo(ctx context.Context, group *gabs.Container, groupName string, container, parameters *gabs.Container, configuration *Configuration) *TemplateInfo {
	cfnInfo := new(TemplateInfo)
	info := new(kilt.TargetInfo)
	cfnInfo.TargetInfo = info
	l := log.Ctx(ctx)

	info.ContainerName = container.S("Name")
	info.ContainerGroupName = groupName
	info.EnvironmentVariables = make(map[string]string)
	cfnInfo.EnvironmentVariables = make(map[string]*gabs.Container)

	if container.Exists("Image") {
		info.Image = container.S("Image")
		if info.Image.Exists("Ref") {
			l.Info().Str("image", info.Image.String()).Msg("retrieving image from template parameters")

			parameterName, ok := info.Image.S("Ref").Data().(string)
			if ok {
				image, ok := parameters.S(parameterName).Data().(string)
				if ok {
					l.Info().Str("image", info.Image.String()).Msgf("found image %s", image)
					info.Image.Set(image)
				} else {
					l.Warn().Str("image", info.Image.String()).Msg("could not resolve the image parameter")
				}
			} else {
				l.Warn().Str("image", info.Image.String()).Msg("could not find the name of the image parameter")
			}
		}

		os.Setenv("HOME", "/tmp") // crane requires $HOME variable
		repoInfo, err := GetConfigFromRepository(info.Image.String())
		if err != nil {
			l.Warn().Str("image", info.Image.String()).Err(err).Msg("could not retrieve metadata from repository")
		} else {
			if configuration.UseRepositoryHints {
				l.Info().Str("image", info.Image.String()).Msgf("extracted info from remote repository: %+v", repoInfo)
				if repoInfo.Entrypoint != nil {
					info.EntryPoint = gabs.New()
					info.EntryPoint.Set(repoInfo.Entrypoint)
				}
				if repoInfo.Command != nil {
					info.Command = gabs.New()
					info.Command.Set(repoInfo.Command)
				}
			}
		}
	}

	if container.Exists("EntryPoint") {
		info.EntryPoint = gabs.New()
		info.EntryPoint.Set(container.S("EntryPoint").Children())
	} else {
		l.Warn().Str("image", info.Image.String()).Msg("no EntryPoint was specified")
	}

	if container.Exists("Command") {
		info.Command = gabs.New()
		info.Command.Set(container.S("Command").Children())
	} else {
		l.Warn().Str("image", info.Image.String()).Msg("no Command was specified")
	}

	if container.Exists("Environment") {
		for _, env := range container.S("Environment").Children() {
			k, ok := env.S("Name").Data().(string)
			if !ok {
				l.Fatal().Str("Fragment", env.S("Name").String()).Str("TaskDefinition", groupName).Msg("Environment has an unsupported value type. Expected string")
			}
			passthrough, templateVal := GetValueFromTemplate(env.S("Value"))

			cfnInfo.EnvironmentVariables[k] = templateVal
			info.EnvironmentVariables[k] = passthrough
		}
	}

	return cfnInfo
}
