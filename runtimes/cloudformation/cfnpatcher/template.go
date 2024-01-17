package cfnpatcher

import (
	"context"
	"os"

	"github.com/Jeffail/gabs/v2"
	"github.com/rs/zerolog/log"

	"github.com/sysdiglabs/agent-kilt/pkg/kilt"
)

func extractContainerInfo(ctx context.Context, group *gabs.Container, groupName string, container, parameters *gabs.Container, configuration *Configuration) *kilt.TargetInfo {
	info := new(kilt.TargetInfo)
	l := log.Ctx(ctx)

	info.ContainerName = container.S("Name")
	info.ContainerGroupName = groupName

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

		if configuration.UseRepositoryHints {
			os.Setenv("HOME", "/tmp") // crane requires $HOME variable
			image, ok := info.Image.Data().(string)
			if ok {
				repoInfo, err := GetConfigFromRepository(image)
				if err != nil {
					l.Warn().Str("image", image).Err(err).Msg("could not retrieve metadata from repository")
				} else {
					l.Info().Str("image", image).Msgf("extracted info from remote repository: %+v", repoInfo)
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

	info.EnvironmentVariables = make([]map[string]*gabs.Container, 0)
	if container.Exists("Environment") {
		for _, env := range container.S("Environment").Children() {
			info.EnvironmentVariables = append(info.EnvironmentVariables, env.ChildrenMap())
		}
	}

	return info
}
