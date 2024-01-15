package kilt

type TargetInfo struct {
	Image                string            `json:"image"`
	ContainerName        string            `json:"container_name"`
	ContainerGroupName   string            `json:"container_group_name"`
	EntryPoint           []string          `json:"entry_point"`
	Command              []string          `json:"command"`
	EnvironmentVariables map[string]string `json:"environment_variables"`
	Metadata             map[string]string `json:"metadata"`
}

type BuildResource struct {
	Name                 string
	Image                string
	Volumes              []string
	EntryPoint           []string
	EnvironmentVariables []map[string]interface{}
}

type Build struct {
	Image                string
	EntryPoint           []string
	Command              []string
	EnvironmentVariables map[string]string
	Capabilities         []string

	Resources []BuildResource
}

type Task struct {
	PidMode string // the only value is `task` right now
}

type LanguageInterface interface {
	Build(info *TargetInfo) (*Build, error)
	Task() (*Task, error)
}
