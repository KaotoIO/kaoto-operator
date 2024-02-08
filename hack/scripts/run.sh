#!/bin/sh -x

env | grep GO
env | grep KAOTO

go run -ldflags="${GOLDFLAGS}" cmd/main.go run --leader-election=false --zap-devel