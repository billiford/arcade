#!/bin/bash

set -o errexit

go get -v -t -d ./...
if [ -f Gopkg.toml ]; then
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
    dep ensure
fi
