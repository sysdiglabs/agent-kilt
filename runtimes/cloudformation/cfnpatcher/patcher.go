package cfnpatcher

import (
	"context"
	"fmt"
	"sort"

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
	k := kilt.NewKiltHoconWithConfig(configuration.Kilt, configuration.RecipeConfig)
	build, _ := k.Build(new(kilt.TargetInfo))
	for k, v := range build.EnvironmentVariables {
		keyStripped := getParameterName(k)
		template.Set("String", "Parameters", keyStripped, "Type")
		template.Set(v, "Parameters", keyStripped, "Default")
	}
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

	if resource.Exists("Properties", "ContainerDefinitions") {
		for _, container := range resource.S("Properties", "ContainerDefinitions").Children() {
			info := extractContainerInfo(ctx, resource, name, container, parameters, configuration)
			l.Info().Msgf("extracted info for container: %+v", info)
			if shouldSkip(info, configuration, hints) {
				l.Info().Msgf("skipping container due to hints in tags")
				continue
			}
			patch, err := k.Build(info)
			if err != nil {
				return nil, fmt.Errorf("could not construct kilt patch: %w", err)
			}
			l.Info().Msgf("created patch for container: %v", patch)
			err = applyContainerDefinitionPatch(l.WithContext(ctx), container, patch, configuration)
			if err != nil {
				l.Warn().Str("resource", name).Err(err).Msg("skipped patching container in task definition")
			} else {
				successes += 1
			}

			for _, appendResource := range patch.Resources {
				existingSidecarVars := make(map[string]struct{})

				for _, kv := range appendResource.EnvironmentVariables {
					existingSidecarVars[kv["Name"].(string)] = struct{}{}
				}

				for k, v := range patch.EnvironmentVariables {
					if _, ok := existingSidecarVars[k]; ok {
						continue
					}

					keyValue := make(map[string]interface{})
					keyValue["Name"] = k

					if configuration.ParameterizeEnvars && !v.Exists("Ref") {
						keyValue["Value"] = map[string]interface{}{"Ref": getParameterName(k)}
					} else {
						keyValue["Value"] = v
					}

					appendResource.EnvironmentVariables = append(appendResource.EnvironmentVariables, keyValue)
				}
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

func applyContainerDefinitionPatch(ctx context.Context, container *gabs.Container, patch *kilt.Build, configuration *Configuration) error {
	l := log.Ctx(ctx)

	_, err := container.Set(patch.EntryPoint, "EntryPoint")
	if err != nil {
		return fmt.Errorf("could not set EntryPoint: %w", err)
	}
	_, err = container.Set(patch.Command, "Command")
	if err != nil {
		return fmt.Errorf("could not set Command: %w", err)
	}

	_, err = container.Set(patch.Image, "Image")
	if err != nil {
		return fmt.Errorf("could not set Command: %w", err)
	}

	if !container.Exists("VolumesFrom") {
		_, err = container.Set([]interface{}{}, "VolumesFrom")
		if err != nil {
			return fmt.Errorf("could not set VolumesFrom: %w", err)
		}
	}

	for _, newContainer := range patch.Resources {
		// Skip containers with no volumes - just injecting sidecars
		if len(newContainer.Volumes) == 0 {
			l.Info().Msgf("Skipping injection of %s because it has no volumes specified", newContainer.Name)
			continue
		}
		addVolume := map[string]interface{}{
			"ReadOnly":        true,
			"SourceContainer": newContainer.Name,
		}

		_, err = container.Set(addVolume, "VolumesFrom", "-")
		if err != nil {
			return fmt.Errorf("could not add VolumesFrom directive: %w", err)
		}
	}

	if len(patch.EnvironmentVariables) > 0 {
		_, err = container.Set([]interface{}{}, "Environment")

		if err != nil {
			return fmt.Errorf("could not add environment variable container: %w", err)
		}
	}

	keys := make([]string, 0, len(patch.EnvironmentVariables))
	for k := range patch.EnvironmentVariables {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		keyValue := make(map[string]interface{})
		keyValue["Name"] = k
		v := patch.EnvironmentVariables[k]
		if configuration.ParameterizeEnvars && !v.Exists("Ref") {
			keyValue["Value"] = map[string]interface{}{"Ref": getParameterName(k)}
		} else {
			keyValue["Value"] = v
		}

		_, err = container.Set(keyValue, "Environment", "-")

		if err != nil {
			return fmt.Errorf("could not add environment variable %v: %w", keyValue, err)
		}

	}

	if len(patch.Capabilities) > 0 {
		capabilities := make([]interface{}, len(patch.Capabilities))
		for i, v := range patch.Capabilities {
			capabilities[i] = v
		}
		// We need to add capabilities to the container
		if !container.Exists("LinuxParameters") {
			emptyMap := make(map[string]interface{})
			_, err = container.Set(emptyMap, "LinuxParameters")
			if err != nil {
				return fmt.Errorf("could not add LinuxParameters: %w", err)
			}
		}

		if !container.Exists("LinuxParameters", "Capabilities") {
			emptyMap := make(map[string]interface{})
			_, err = container.Set(emptyMap, "LinuxParameters", "Capabilities")
			if err != nil {
				return fmt.Errorf("could not add LinuxParameters.Capabilities: %w", err)
			}
		}

		if !container.Exists("LinuxParameters", "Capabilities", "Add") {
			emptyList := make([]interface{}, 0)
			_, err = container.Set(emptyList, "LinuxParameters", "Capabilities", "Add")
			if err != nil {
				return fmt.Errorf("could not add LinuxParameters.Capabilities.Add: %w", err)
			}
		}

		err := container.ArrayConcat(capabilities, "LinuxParameters", "Capabilities", "Add")
		if err != nil {
			return fmt.Errorf("could not append to LinuxParameters.Capabilities.Add: %w", err)
		}
	}

	return nil
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
