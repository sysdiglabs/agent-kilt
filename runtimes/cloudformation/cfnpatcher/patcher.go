package cfnpatcher

import (
	"context"
	"fmt"
	"github.com/sysdiglabs/agent-kilt/pkg/kilt"

	"github.com/Jeffail/gabs/v2"
	"github.com/rs/zerolog/log"
)

func containerInConfig(name string, listOfNames []string) bool {
	for _, n := range listOfNames {
		if n == name {
			return true
		}
	}
	return false
}

func shouldSkip(container *gabs.Container, configuration *Configuration, hints *InstrumentationHints) bool {
	containerNameData := container.S("Name").Data()
	var containerName string
	switch containerNameData.(type) {
	case string:
		containerName = containerNameData.(string)
	default:
		containerName = container.S("Name").String()
	}

	isForceIncluded := containerInConfig(containerName, hints.IncludeContainersNamed)
	isExcluded := containerInConfig(containerName, hints.IgnoreContainersNamed)

	return (configuration.OptIn && !isForceIncluded && !hints.HasGlobalInclude) || (!configuration.OptIn && isExcluded)
}

func applyParametersPatch(ctx context.Context, template *gabs.Container, configuration *Configuration) (*gabs.Container, error) {
	patchConfig := kilt.PatchConfig{
		ParametrizeEnvars: configuration.ParameterizeEnvars,
	}

	k := kilt.NewKiltHoconWithConfig(configuration.Kilt, configuration.RecipeConfig, nil)
	err := k.PatchCfnTemplate(template, &patchConfig)
	if err != nil {
		return nil, err
	}
	return template, nil
}

func applyTaskDefinitionPatch(ctx context.Context, name string, resource, parameters *gabs.Container, configuration *Configuration, hints *InstrumentationHints) (*gabs.Container, error) {
	l := log.Ctx(ctx)

	sidecarConfig := gabs.New()
	err := applyConfiguration(sidecarConfig, configuration, name)
	if err != nil {
		return nil, fmt.Errorf("could not apply sidecar configuration: %w", err)
	}

	patchConfig := kilt.PatchConfig{
		ParametrizeEnvars: configuration.ParameterizeEnvars,
	}

	k := kilt.NewKiltHoconWithConfig(configuration.Kilt, configuration.RecipeConfig, sidecarConfig)
	err = k.PatchTaskDefinition(resource, &patchConfig, name, func(container *gabs.Container) bool {
		if shouldSkip(container, configuration, hints) {
			l.Info().Msgf("skipping container due to hints in tags")
			return false
		}

		fillContainerInfo(ctx, container, parameters, configuration)
		return true
	})

	if err != nil {
		return nil, fmt.Errorf("could not patch task definition: %w", err)
	}
	return resource, nil
}

func applyConfiguration(container *gabs.Container, configuration *Configuration, taskName string) error {
	if len(configuration.ImageAuthSecret) > 0 {
		_, err := container.Set(configuration.ImageAuthSecret, "RepositoryCredentials", "CredentialsParameter")
		if err != nil {
			return fmt.Errorf("could not set image auth secret: %w", err)
		}
	}
	if len(configuration.LogGroup) > 0 {
		_, err := container.Set(prepareLogConfiguration(taskName, configuration.LogGroup), "LogConfiguration")
		if err != nil {
			return fmt.Errorf("could not set log configuration: %w", err)
		}
	}
	return nil
}

func prepareLogConfiguration(taskName string, logGroup string) map[string]interface{} {
	// assuming that all given log configurations are for the awslogs driver
	config := map[string]interface{}{
		"LogDriver": "awslogs",
		"Options": map[string]interface{}{
			"awslogs-region": map[string]interface{}{
				"Ref": "AWS::Region",
			},
			"awslogs-group":         logGroup,
			"awslogs-stream-prefix": taskName,
		},
	}

	return config
}
