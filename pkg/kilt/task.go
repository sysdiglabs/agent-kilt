package kilt

import (
	"github.com/go-akka/configuration"
)

func extractTask(config *configuration.Config) (*Task, error) {
	var task = new(Task)

	if config.HasPath("task.pid_mode") {
		task.PidMode = config.GetString("task.pid_mode")
	}

	return task, nil
}
