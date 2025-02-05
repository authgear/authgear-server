#!/bin/bash

## To run multiple processes in a container, we need to make sure
## the main process reaping the child processes when the container exits.
##
## bash is pre-installed and it can run and reap child processes easily,
## as opposed to using systemd.
##
## This script is inspired by https://docs.docker.com/engine/containers/multi-service_container/#use-a-wrapper-script
## One important difference is we use `wait` instead of `wait -n`.
## According to the documentation https://www.gnu.org/software/bash/manual/html_node/Job-Control-Builtins.html#index-wait
## `wait -n` only wait for the first background job to complete and return the exit status of that job.
## This IS NOT what we want because we want to wait PostgreSQL and Redis to exits properly, as
## they will save files when they are interrupted by signal.
##
## In my testing, with `wait`, I can see the exit message of both PostgreSQL and Redis.

set -e

postgres &

redis-server /etc/redis/redis.conf &

nginx &

wait
