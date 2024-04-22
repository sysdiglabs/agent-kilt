ROOT := $(abspath ../..)

# Default name
SERVERLESS_PATCHER_IMAGE ?= sysdig/serverless-patcher:latest

# Use the latest workload-agent by default
WORKLOAD_AGENT_IMAGE ?= quay.io/sysdig/workload-agent:latest

serverless-patcher:
	@echo "Building agent-kilt"
	make -C $(ROOT)/runtimes/cloudformation cmd/handler/handler
	mv $(ROOT)/runtimes/cloudformation/cmd/handler/handler .
	@echo "Building $(SERVERLESS_PATCHER_IMAGE) on $(shell uname)/$(shell uname -m)"
	docker build -t $(SERVERLESS_PATCHER_IMAGE) \
		--build-arg KILT_DEFINITION="$(shell base64 res/cfn-macro.kilt.cfg | tr -d \\n)" \
		--build-arg KILT_DEFINITION_TYPE="base64" \
		--build-arg SYSDIG_WORKLOAD_AGENT_IMAGE="$(WORKLOAD_AGENT_IMAGE)" \
		.

check-jq:
	@command -v jq >/dev/null 2>&1 || { echo >&2 "jq is not installed. Please install it to proceed."; exit 1; }

test: check-jq serverless-patcher
	@echo "Testing $(SERVERLESS_PATCHER_IMAGE)"

	$(call run_test,'base',)

	$(call run_test,'full-recipe',-e SYSDIG_ACCESS_KEY="qwerty" -e SYSDIG_AGENT_NICE_VALUE_INCREMENT="42" -e SYSDIG_COLLECTOR_HOST="sysdig.com" -e SYSDIG_COLLECTOR_PORT="4321" -e SYSDIG_LOGGING="scream" -e SYSDIG_ORCHESTRATOR_HOST="orchestrator.com" -e SYSDIG_ORCHESTRATOR_PORT="1234" -e SYSDIG_WORKLOAD_AGENT_IMAGE="worky:latest")

	$(call run_test,'log-group',-e KILT_LOG_GROUP="friedChicken")

	$(call run_test,'no-hints',-e KILT_DISABLE_REPO_HINTS="true")

	$(call run_test,'opt-in-include-all',-e KILT_OPT_IN="true")

	$(call run_test,'opt-in-include-containers',-e KILT_OPT_IN="true")

	$(call run_test,'opt-in-include-none',-e KILT_OPT_IN="true")

	$(call run_test,'opt-out-ignore-all',)

	$(call run_test,'opt-out-ignore-containers',)

	$(call run_test,'parameterized',-e SYSDIG_ORCHESTRATOR_HOST="orchestrator.com" -e SYSDIG_ORCHESTRATOR_PORT="1234" -e KILT_PARAMETERIZE_ENVARS="true")

	$(call run_test,'sidecar',-e SYSDIG_SIDECAR_MODE="auto")

# Run serverless-patcher and test the result.
# If the test fails, it provides the .out file in the folder test/template/out.
# Args:
#  - $(1) test name, must match with the base name of the src/out json files
#  - $(2) additional envars to configure serverless patcher
define run_test
	@echo 'Test name: $(1), Additional envars: $(2)'
	$(eval host_path := $(ROOT)/docker/serverless-patcher/test/templates)
	$(eval guest_path_src := /templates/src/$(1).json)
	$(eval guest_path_out := /templates/out/$(1).out.json)
	$(eval guest_path_expected := /templates/out/$(1).expected.json)

	@# Patch the src template and wait for the container to finish
	docker run --rm \
		-e KILT_MODE="local" \
		-e KILT_SRC_TEMPLATE=$(guest_path_src) \
		-e KILT_OUT_TEMPLATE=$(guest_path_out) \
		$(2) \
		-v $(host_path):/templates \
		$(SERVERLESS_PATCHER_IMAGE)

	@# Test the output against the expected template (sorting required)
	docker run --rm \
		-v $(host_path):/templates alpine \
			/bin/sh -c "apk add jq && \
				jq '.Resources.TaskDefinition.Properties.ContainerDefinitions |= map(.Environment |= if . then sort_by(.Name) else . end)' $(guest_path_out) > sorted_out.json && \
				jq '.Resources.TaskDefinition.Properties.ContainerDefinitions |= map(.Environment |= if . then sort_by(.Name) else . end)' $(guest_path_expected) > sorted_expected.json && \
				diff -d sorted_out.json sorted_expected.json && \
					(echo 'Test $(1) SUCCEEDED' && rm $(guest_path_out)) || \
					(echo '### Test $(1) FAILED: the patched template obtained from $(guest_path_src) differs from $(guest_path_expected) ###' && cp sorted_out.json $(guest_path_out) && exit 1)"
endef

.PHONY: test
