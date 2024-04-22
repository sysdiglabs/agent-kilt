#!/usr/bin/env bash

echo "Kilt Definition (type $KILT_DEFINITION_TYPE): $KILT_DEFINITION"

if [ -z "$KILT_RECIPE_CONFIG" ];
then
	echo "Using default Recipe Configuration"
	KILT_RECIPE_CONFIG=$(cat <<-EOF
	{
		"agent_image": "$SYSDIG_WORKLOAD_AGENT_IMAGE",
		"collector_host": "$SYSDIG_COLLECTOR_HOST",
		"collector_port": "$SYSDIG_COLLECTOR_PORT",
		"sidecar": "$SYSDIG_SIDECAR_MODE",
		"sysdig_access_key": "$SYSDIG_ACCESS_KEY",
		"sysdig_agent_nice_value_increment": "$SYSDIG_AGENT_NICE_VALUE_INCREMENT",
		"sysdig_logging": "$SYSDIG_LOGGING",
		"orchestrator_host": "$SYSDIG_ORCHESTRATOR_HOST",
		"orchestrator_port": "$SYSDIG_ORCHESTRATOR_PORT",
		"priority": "$SYSDIG_PRIORITY"
	}
	EOF
	)
else
	echo "Using custom Recipe Configuration"
fi

echo "Recipe Configuration: $KILT_RECIPE_CONFIG"

./handler
