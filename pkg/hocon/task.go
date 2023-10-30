package hocon

import (
	"github.com/go-akka/configuration"
	"github.com/sysdiglabs/agent-kilt/pkg/kilt"
)

func extractTask(config *configuration.Config) (*kilt.Task, error) {
	var task = new(kilt.Task)

	if config.HasPath("task.pid_mode") {
		task.PidMode = config.GetString("task.pid_mode")
	}

	return task, nil
}
