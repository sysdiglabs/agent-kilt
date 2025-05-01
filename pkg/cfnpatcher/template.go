package cfnpatcher

import (
	"context"
	"fmt"
	"os"

	"github.com/Jeffail/gabs/v2"
	"github.com/rs/zerolog/log"
)

func fillContainerInfo(ctx context.Context, container *gabs.Container, parameters *gabs.Container, configuration *Configuration) {
	l := log.Ctx(ctx)

	hasOverriddenEntrypoint := container.Exists("EntryPoint")
	hasOverriddenCommand := container.Exists("Command")

	if hasOverriddenEntrypoint && hasOverriddenCommand {
		return
	}

	if !container.Exists("Image") {
		return
	}

	var image string
	// If the image is a reference to a parameter, resolve it
	if container.Exists("Image", "Ref") {
		l.Info().Str("image", container.S("Image").String()).Msg("retrieving image from template parameters")

		parameterName, ok := container.S("Image", "Ref").Data().(string)
		if ok {
			image, ok := parameters.S(parameterName).Data().(string)
			if ok {
				l.Info().Str("image", container.S("Image").String()).Msgf("found image %s", image)
			} else {
				l.Warn().Str("image", container.S("Image").String()).Msg("could not resolve the image parameter")
			}
		} else {
			l.Warn().Str("image", container.S("Image").String()).Msg("could not find the name of the image parameter")
		}
	} else {
		tag := container.S("Image").Data()
		switch v := tag.(type) {
		case string:
			// If the image is a string, use it as is
			image = v
		default:
			// Otherwise, convert it to a raw string
			image = fmt.Sprintf("%v", v)
		}
	}

	if configuration.UseRepositoryHints {
		os.Setenv("HOME", "/tmp") // crane requires $HOME variable
		repoInfo, err := GetConfigFromRepository(image)
		if err != nil {
			l.Err(err).Msg("The entrypoint and command of the image must be manually set in the original template")
		} else {
			// Use the image's entrypoint if the task definition does not override it
			if repoInfo.Entrypoint != nil && !hasOverriddenEntrypoint {
				l.Info().Str("image", container.S("Image").String()).Msgf("using default entrypoint %s", repoInfo.Entrypoint)
				container.Set(repoInfo.Entrypoint, "EntryPoint")
			}
			// Use the image's command if the task definition overrides neither the entrypoint nor the command
			if repoInfo.Command != nil && !hasOverriddenCommand && !hasOverriddenEntrypoint {
				l.Info().Str("image", container.S("Image").String()).Msgf("using default command %s", repoInfo.Command)
				container.Set(repoInfo.Command, "Command")
			}
		}
	}
}
