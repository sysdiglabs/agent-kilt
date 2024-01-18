package main

import (
	"context"
	"encoding/json"
	"github.com/Jeffail/gabs/v2"
	"os"
	"strings"

	"github.com/sysdiglabs/agent-kilt/runtimes/cloudformation/config"

	"github.com/sysdiglabs/agent-kilt/runtimes/cloudformation/cfnpatcher"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type MacroInput struct {
	Region                  string          `json:"region"`
	AccountID               string          `json:"accountId"`
	RequestID               string          `json:"requestId"`
	TransformID             string          `json:"transformId"`
	TemplateParameterValues json.RawMessage `json:"templateParameterValues"`
	Fragment                json.RawMessage `json:"fragment"`
}

type MacroOutput struct {
	RequestID string          `json:"requestId"`
	Status    string          `json:"status"`
	Fragment  json.RawMessage `json:"fragment"`
}

func HandleRequest(configuration *cfnpatcher.Configuration, ctx context.Context, event MacroInput) (MacroOutput, error) {
	l := log.With().
		Str("region", event.Region).
		Str("account", event.AccountID).
		Str("requestId", event.RequestID).
		Str("transformId", event.TransformID).
		Logger()
	loggerCtx := l.WithContext(ctx)
	result, err := cfnpatcher.Patch(loggerCtx, configuration, event.Fragment, event.TemplateParameterValues)
	if err != nil {
		return MacroOutput{event.RequestID, "failure", result}, err
	}
	log.Info().Str("template", string(result)).Msg("processing complete")
	return MacroOutput{event.RequestID, "success", result}, nil
}

func PatchLocalFile(configuration *cfnpatcher.Configuration, ctx context.Context, inputFile string) ([]byte, error) {
	l := log.With().
		Str("region", "local").
		Logger()
	loggerCtx := l.WithContext(ctx)

	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		l.Error().Err(err).Msgf("cannot read file %s", inputFile)
		return nil, err
	}

	templateParameters := make([]byte, 0)
	result, err := cfnpatcher.Patch(loggerCtx, configuration, inputData, templateParameters)
	if err != nil {
		l.Error().Err(err).Msg("failed to patch local file")
		return nil, err
	}

	log.Info().Str("template", string(result)).Msg("processing complete")
	return result, nil
}

func GetConfig() *cfnpatcher.Configuration {
	definition := os.Getenv("KILT_DEFINITION")
	definitionType := os.Getenv("KILT_DEFINITION_TYPE")
	optIn := os.Getenv("KILT_OPT_IN")
	imageAuth := os.Getenv("KILT_IMAGE_AUTH_SECRET")
	recipeConfig := os.Getenv("KILT_RECIPE_CONFIG")
	disableRepoHints := os.Getenv("KILT_DISABLE_REPO_HINTS")
	logGroup := os.Getenv("KILT_LOG_GROUP")
	parameterizeEnvars := os.Getenv("KILT_PARAMETERIZE_ENVARS")
	sidecarEssential := os.Getenv("KILT_SIDECAR_ESSENTIAL")
	sidecarCpu := os.Getenv("KILT_SIDECAR_CPU")
	sidecarMemoryLimit := os.Getenv("KILT_SIDECAR_MEMORY_LIMIT")
	sidecarMemoryReservation := os.Getenv("KILT_SIDECAR_MEMORY_RESERVATION")
	sidecarConfig := os.Getenv("KILT_SIDECAR_CONFIG")

	var fullDefinition string
	switch definitionType {
	case config.S3:
		fullDefinition = config.FromS3(definition, false)
	case config.S3Gz:
		fullDefinition = config.FromS3(definition, true)
	case config.Http:
		fullDefinition = config.FromWeb(definition)
	case config.Base64:
		fullDefinition = config.FromBase64(definition, false)
	case config.Base64Gz:
		fullDefinition = config.FromBase64(definition, true)
	default:
		panic("unrecognized definition type - " + definitionType)
	}

	scObj := gabs.New()
	if sidecarConfig != "" {
		sc, err := gabs.ParseJSON([]byte(sidecarConfig))
		if err != nil {
			panic("cannot parse sidecar config: " + err.Error())
		}
		scObj = sc
	}

	if imageAuth != "" {
		_, err := scObj.Set(imageAuth, "RepositoryCredentials", "CredentialsParameter")
		if err != nil {
			panic("cannot set image auth secret in sidecar config: " + err.Error())
		}
	}

	if sidecarEssential != "" {
		_, err := scObj.Set(sidecarEssential, "Essential")
		if err != nil {
			panic("cannot set sidecar essential in sidecar config: " + err.Error())
		}
	}

	if sidecarCpu != "" {
		_, err := scObj.Set(sidecarCpu, "Cpu")
		if err != nil {
			panic("cannot set sidecar cpu in sidecar config: " + err.Error())
		}
	}

	if sidecarMemoryLimit != "" {
		_, err := scObj.Set(sidecarMemoryLimit, "Memory")
		if err != nil {
			panic("cannot set sidecar memory limit in sidecar config: " + err.Error())
		}
	}

	if sidecarMemoryReservation != "" {
		_, err := scObj.Set(sidecarMemoryReservation, "MemoryReservation")
		if err != nil {
			panic("cannot set sidecar memory reservation in sidecar config: " + err.Error())
		}
	}

	sc, err := json.Marshal(scObj)
	if err != nil {
		panic("cannot marshal sidecar config: " + err.Error())
	}

	sidecarConfig = string(sc)
	configuration := &cfnpatcher.Configuration{
		Kilt:               fullDefinition,
		OptIn:              optIn != "",
		RecipeConfig:       recipeConfig,
		UseRepositoryHints: disableRepoHints == "",
		LogGroup:           logGroup,
		ParameterizeEnvars: strings.ToLower(parameterizeEnvars) == "true",
		SidecarConfig:      sidecarConfig,
	}

	return configuration
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	configuration := GetConfig()
	switch os.Getenv("KILT_MODE") {
	case "local":
		result, err := PatchLocalFile(configuration, context.Background(), os.Getenv("KILT_SRC_TEMPLATE"))
		if err != nil {
			panic("cannot patch local file " + os.Getenv("KILT_SRC_TEMPLATE"))
		}

		err = os.WriteFile(os.Getenv("KILT_OUT_TEMPLATE"), result, 0644)
		if err != nil {
			panic("cannot write dst file " + os.Getenv("KILT_OUT_TEMPLATE"))
		}

	default:
		lambda.Start(
			func(ctx context.Context, event MacroInput) (MacroOutput, error) {
				return HandleRequest(configuration, ctx, event)
			})
	}
}
