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

func shouldSkip(info *kilt.TargetInfo, configuration *Configuration, hints *InstrumentationHints) bool {
	containerNameData := info.ContainerName.Data()
	var containerName string
	switch containerNameData.(type) {
	case string:
		containerName = containerNameData.(string)
	default:
		containerName = info.ContainerName.String()
	}

	isForceIncluded := containerInConfig(containerName, hints.IncludeContainersNamed)
	isExcluded := containerInConfig(containerName, hints.IgnoreContainersNamed)

	return (configuration.OptIn && !isForceIncluded && !hints.HasGlobalInclude) || (!configuration.OptIn && isExcluded)
}

func applyParametersPatch(ctx context.Context, template *gabs.Container, configuration *Configuration) (*gabs.Container, error) {
	patchConfig := kilt.PatchConfig{
		ParametrizeEnvars: configuration.ParameterizeEnvars,
	}

	k := kilt.NewKiltHoconWithConfig(configuration.Kilt, configuration.RecipeConfig)
	container := gabs.New()
	container.Set(make(map[string]interface{}))
	build, _ := k.Patch(container, &patchConfig, new(kilt.TargetInfo))

	parameters := build.EnvParameters
	if parameters == nil {
		return template, nil
	}

	template.Merge(build.EnvParameters)
	return template, nil
}

func applyTaskDefinitionPatch(ctx context.Context, name string, resource, parameters *gabs.Container, configuration *Configuration, hints *InstrumentationHints) (*gabs.Container, error) {
	l := log.Ctx(ctx)

	successes := 0
	containers := make(map[string]kilt.BuildResource)
	k := kilt.NewKiltHoconWithConfig(configuration.Kilt, configuration.RecipeConfig)

	taskPatch, err := k.Task()
	if err != nil {
		return nil, fmt.Errorf("could not get task definition patch: %w", err)
	}

	if taskPatch.PidMode != "" {
		if !resource.Exists("Properties") {
			_, err := resource.Set(map[string]interface{}{}, "Properties")
			if err != nil {
				return nil, fmt.Errorf("could not add properties to task definition: %w", err)
			}
		}

		_, err = resource.Set(taskPatch.PidMode, "Properties", "PidMode")
		if err != nil {
			return nil, fmt.Errorf("could not set PidMode: %w", err)
		}
	}

	patchConfig := kilt.PatchConfig{
		ParametrizeEnvars: configuration.ParameterizeEnvars,
	}

	if resource.Exists("Properties", "ContainerDefinitions") {
		for _, container := range resource.S("Properties", "ContainerDefinitions").Children() {
			info := extractContainerInfo(ctx, resource, name, container, parameters, configuration)
			l.Info().Msgf("extracted info for container: %+v", info)
			if shouldSkip(info, configuration, hints) {
				l.Info().Msgf("skipping container due to hints in tags")
				continue
			}
			patch, err := k.Patch(container, &patchConfig, info)
			if err != nil {
				return nil, fmt.Errorf("could not construct kilt patch: %w", err)
			}
			l.Info().Msgf("created patch for container: %v", patch)
			successes += 1

			for _, appendResource := range patch.Resources {
				containers[appendResource.Name] = appendResource
			}
		}
		err := appendContainers(resource, containers, configuration.ImageAuthSecret, configuration.LogGroup, name)
		if err != nil {
			return nil, fmt.Errorf("could not append container: %w", err)
		}
	}
	if successes == 0 {
		return resource, fmt.Errorf("could not patch a single container in the task")
	}
	return resource, nil
}

func appendContainers(resource *gabs.Container, containers map[string]kilt.BuildResource, imageAuth string, logGroup string, name string) error {
	for _, inject := range containers {
		appended := map[string]interface{}{
			"Name":       inject.Name,
			"Image":      inject.Image,
			"EntryPoint": inject.EntryPoint,
		}
		if len(inject.EnvironmentVariables) > 0 {
			appended["Environment"] = inject.EnvironmentVariables
		}
		if len(imageAuth) > 0 {
			appended["RepositoryCredentials"] = map[string]interface{}{
				"CredentialsParameter": imageAuth,
			}
		}
		if len(logGroup) > 0 {
			appended["LogConfiguration"] = prepareLogConfiguration(name, logGroup)
		}
		_, err := resource.Set(appended, "Properties", "ContainerDefinitions", "-")
		if err != nil {
			return fmt.Errorf("could not inject %s: %w", inject.Name, err)
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
