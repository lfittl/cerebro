#!/bin/bash

export PORT=8080
export HOST_IP=$(ip route show 0.0.0.0/0 | grep -Eo 'via \S+' | awk '{ print $2 }')
export ETCD_ENDPOINT="http://$HOST_IP:4001"
export DOCKER_ENDPOINT="unix:///var/run/docker.sock"

exec "$@"
