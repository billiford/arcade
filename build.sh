#!/bin/bash

set -o errexit

CGO_ENABLED=0 go build "${@}" cmd/arcade/arcade.go
