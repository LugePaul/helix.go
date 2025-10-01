#!/usr/bin/env bash

go clean -testcache

go test -test.v ./...

integrations=$( cat ./ecosystem.json | jq -r '.integrations[] | .id' | cat )
for mod in $integrations; do
  cd ./integration/$mod

  ./scripts/tests.sh

  cd ../../
done
