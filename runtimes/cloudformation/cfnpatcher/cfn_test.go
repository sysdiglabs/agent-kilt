package cfnpatcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/Jeffail/gabs/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	diff "github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

var optInTests = [...]string{
	"respect_ignores/opt_in_check",
	"respect_ignores/opt_in_ignored",
	"respect_ignores/opt_in_multiple_containers",
	"respect_ignores/opt_in_single_container",
}

var optPanicTests = [...]string{
	"respect_ignores/panic_opt_exclude",
	"respect_ignores/panic_opt_exclude_container",
	"respect_ignores/panic_opt_include",
	"respect_ignores/panic_opt_include_container",
}

var defaultTests = [...]string{
	"respect_ignores/opt_out_default",
	"respect_ignores/opt_out_ignored",
	"respect_ignores/opt_out_ignore_multiple_containers",
	"respect_ignores/opt_out_ignore_single_container",

	"patching/command",
	"patching/entrypoint",
	"patching/ref",
	"patching/ref_command",
	"patching/ref_env",
	"patching/ref_tags",
	"patching/tags",
	"patching/volumes_from",
}

var parameterizedEnvarsTests = [...]string{
	"patching/parameterize_env_add",
	"patching/parameterize_env_merge",
}

var sidecarEnvTests = [...]string{
	"sidecar_env/overlap",
	"sidecar_env/ref_env",
	"sidecar_env/volumes_from",
}

var taskPidModeTests = [...]string{
	"task_pid_mode/command",
}

const defaultConfig = `
build {
	entry_point: ["/kilt/run", "--"]
	command: [] ${?original.entry_point} ${?original.command}
	mount: [
		{
			name: "KiltImage"
			image: "KILT:latest"
			volumes: ["/kilt"]
			entry_point: ["/kilt/wait"]
		}
	]
	capabilities: ["SYS_PTRACE"]
}
`

const parameterizeEnvarsConfig = `
build {
	entry_point: ["/kilt/run", "--"]
	command: [] ${?original.entry_point} ${?original.command}
	environment_variables: {
		"SO_LONG_AND_THANKS": "ForAllTheFish"
	}
	mount: [
		{
			name: "KiltImage"
			image: "KILT:latest"
			volumes: ["/kilt"]
			entry_point: ["/kilt/wait"]
		}
	]
	capabilities: ["SYS_PTRACE"]
}
`

const sidecarEnvConfig = `
build {
	entry_point: ["/kilt/run", "--"]
	command: [] ${?original.entry_point} ${?original.command}
	mount: [
		{
			name: "KiltImage"
			image: "KILT:latest"
			volumes: ["/kilt"]
			entry_point: ["/kilt/wait"]
			environment_variables: {
				"MEANING_OF_LIFE": "42"
			}
		}
	]
	capabilities: ["SYS_PTRACE"]
}
`

const taskPidModeConfig = `
build {
	entry_point: ["/kilt/run", "--"]
	command: [] ${?original.entry_point} ${?original.command}
	mount: [
		{
			name: "KiltImage"
			image: "KILT:latest"
			volumes: ["/kilt"]
			entry_point: ["/kilt/wait"]
		}
	]
	capabilities: ["SYS_PTRACE"]
}
task {
	pid_mode: "task"
}
`

func runTest(t *testing.T, name string, context context.Context, config Configuration) {
	fragment, err := ioutil.ReadFile("fixtures/" + name + ".json")
	if err != nil {
		t.Fatalf("cannot find fixtures/%s.json", name)
	}
	templateParameters := make([]byte, 0)
	result, err := Patch(context, &config, fragment, templateParameters)
	if err != nil {
		t.Fatalf("error patching: %s", err.Error())
	}
	expected, err := ioutil.ReadFile("fixtures/" + name + ".patched.json")
	if err != nil {
		// To regenerate test simply delete patched variant
		_ = ioutil.WriteFile("fixtures/"+name+".patched.json", result, 0644)
		return
	}

	fmt.Printf("result: %s\n", result)
	fmt.Printf("expected: %s\n", expected)

	differ := diff.New()
	d, err := differ.Compare(expected, result)

	if d.Modified() {
		var expectedJson map[string]interface{}
		t.Log("Found differences!")
		_ = json.Unmarshal(result, &expectedJson) // would error during diff
		formatter := formatter.NewAsciiFormatter(expectedJson, formatter.AsciiFormatterConfig{
			ShowArrayIndex: true,
			Coloring:       true,
		})
		diffString, _ := formatter.Format(d)
		fmt.Println(diffString)
		t.Fail()
	}
}

func TestPatchingOptIn(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()

	for _, testName := range optInTests {
		t.Run(testName, func(t *testing.T) {
			runTest(t, testName, l.WithContext(context.Background()),
				Configuration{
					Kilt:               defaultConfig,
					OptIn:              true,
					RecipeConfig:       "{}",
					UseRepositoryHints: false,
				})
		})
	}
}

func TestPatching(t *testing.T) {
	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()

	for _, testName := range defaultTests {
		t.Run(testName, func(t *testing.T) {
			runTest(t, testName, l.WithContext(context.Background()),
				Configuration{
					Kilt:               defaultConfig,
					OptIn:              false,
					RecipeConfig:       "{}",
					UseRepositoryHints: false,
				})
		})
	}
}

func TestPatchingSidecarEnv(t *testing.T) {
	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()

	for _, testName := range sidecarEnvTests {
		t.Run(testName, func(t *testing.T) {
			runTest(t, testName, l.WithContext(context.Background()),
				Configuration{
					Kilt:               sidecarEnvConfig,
					OptIn:              false,
					RecipeConfig:       "{}",
					UseRepositoryHints: false,
				})
		})
	}
}

func TestPatchingTask(t *testing.T) {
	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()

	for _, testName := range taskPidModeTests {
		t.Run(testName, func(t *testing.T) {
			runTest(t, testName, l.WithContext(context.Background()),
				Configuration{
					Kilt:               taskPidModeConfig,
					OptIn:              false,
					RecipeConfig:       "{}",
					UseRepositoryHints: false,
				})
		})
	}
}

func TestPatchingForParameterizingEnvars(t *testing.T) {
	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()

	for _, testName := range defaultTests {
		t.Run(testName, func(t *testing.T) {
			runTest(t, testName, l.WithContext(context.Background()),
				Configuration{
					Kilt:               defaultConfig,
					OptIn:              false,
					RecipeConfig:       "{}",
					UseRepositoryHints: false,
					ParameterizeEnvars: true,
				})
		})
	}

	for _, testName := range parameterizedEnvarsTests {
		t.Run(testName, func(t *testing.T) {
			runTest(t, testName, l.WithContext(context.Background()),
				Configuration{
					Kilt:               parameterizeEnvarsConfig,
					OptIn:              false,
					RecipeConfig:       "{}",
					UseRepositoryHints: false,
					ParameterizeEnvars: true,
				})
		})
	}
}

func TestPatchingForLogGroup(t *testing.T) {
	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()

	tests := []struct {
		Name   string
		Config Configuration
	}{
		{
			"patching/log_group_empty",
			Configuration{
				Kilt:               defaultConfig,
				OptIn:              false,
				RecipeConfig:       "{}",
				UseRepositoryHints: false,
			},
		},
		{
			"patching/log_group",
			Configuration{
				Kilt:               defaultConfig,
				OptIn:              false,
				RecipeConfig:       "{}",
				UseRepositoryHints: false,
				LogGroup:           "test_logs",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			runTest(t, test.Name, l.WithContext(context.Background()), test.Config)
		})
	}
}

func TestOptTagPanic(t *testing.T) {
	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()

	for _, testName := range optPanicTests {
		t.Run(testName, func(t *testing.T) {
			defer func() { recover() }()

			runTest(t, testName, l.WithContext(context.Background()),
				Configuration{
					Kilt:               defaultConfig,
					OptIn:              true,
					RecipeConfig:       "{}",
					UseRepositoryHints: false,
				})
		})
	}
}

func TestIsFuncOptKey(t *testing.T) {
	tests := []struct {
		key string
		out bool
	}{
		{
			"kilt-ignore",
			true,
		},
		{
			"kilt-include",
			true,
		},
		{
			"kilt-ignore-containers",
			true,
		},
		{
			"kilt-include-containers",
			true,
		},
		{
			"so-long-and-thanks-for-all-the-fish",
			false,
		},
	}
	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			assert.Equal(t, test.out, isOptTagKey(test.key), "OptIn/Out key not recognized")
		})
	}
}

func TestGetOptTags(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected map[string]string
	}{
		{
			name: `no-properties`,
			json: `{
"McGuffin": {
	"Tags":[
		{
			"Key": "SoLong",
			"Value": "AndThanksForAllTheFish"
		}
	]}
}`,
			expected: make(map[string]string),
		},
		{
			name: `no-tags`,
			json: `{
"Properties": {
	"Accio":[
		{
			"Key": "SoLong",
			"Value": "AndThanksForAllTheFish"
		}
	]}
}`,
			expected: make(map[string]string),
		},
		{
			name: `no-opt-tags`,
			json: `{
"Properties": {
	"Tags":[
		{
			"Key": "SoLong",
			"Value": "AndThanksForAllTheFish"
		},
		{
			"Key": "TimeIsAnIllusion",
			"Value": "LunchtimeDoublySo"
		}
	]}
}`,
			expected: make(map[string]string),
		},
		{
			name: `all-opt-tags`,
			json: `{
"Properties": {
	"Tags": [
		{
			"Key": "SoLong",
			"Value": "AndThanksForAllTheFish"
		},
		{
			"Key": "TimeIsAnIllusion",
			"Value": "LunchtimeDoublySo"
		},
		{
			"Key": "kilt-ignore",
			"Value": "nanananananaBatman"
		},
		{
			"Key": "kilt-include",
			"Value": "gimmeGimmeGimmeFriedChicken"
		},
		{
			"Key": "kilt-ignore-containers",
			"Value": "expelliarmus"
		},
		{
			"Key": "kilt-include-containers",
			"Value": "accioContainer"
		}
	]}
}`,
			expected: map[string]string{
				"kilt-ignore":             "nanananananaBatman",
				"kilt-include":            "gimmeGimmeGimmeFriedChicken",
				"kilt-ignore-containers":  "expelliarmus",
				"kilt-include-containers": "accioContainer",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			jsonParsed, err := gabs.ParseJSON([]byte(tc.json))
			if err != nil {
				panic(err)
			}
			mm := getOptTags(jsonParsed)
			eq := reflect.DeepEqual(tc.expected, mm)
			if !eq {
				assert.Fail(t, "maps do not match")
			}
		})
	}
}
