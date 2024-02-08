#!/bin/sh -x

if [ $# -ne 1 ]; then
    echo "output is expected"
fi

env | grep GO

go build -ldflags="${GOLDFLAGS}" -o "${1}/kaoto" cmd/main.go