# syntax=docker/dockerfile:1

# Final Stage
FROM golang:1.22

LABEL maintainer="Sysdig <support@sysdig.com>"

# Kilt default configuration
ARG  KILT_DEFINITION_TYPE
ENV  KILT_DEFINITION_TYPE=$KILT_DEFINITION_TYPE

ARG  KILT_DEFINITION
ENV  KILT_DEFINITION=$KILT_DEFINITION

ARG  SYSDIG_WORKLOAD_AGENT_IMAGE
ENV  SYSDIG_WORKLOAD_AGENT_IMAGE=$SYSDIG_WORKLOAD_AGENT_IMAGE

ENV  SYSDIG_LOGGING='info'

# Kilt Recipe Configuration
ENV  KILT_RECIPE_CONFIG=''

WORKDIR /usr/src/app/kilt

COPY handler .
COPY entrypoint.sh .

RUN  chmod 755 entrypoint.sh

ENTRYPOINT [ "./entrypoint.sh" ]
