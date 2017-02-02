#!/bin/bash -e
rm -rf build
mkdir build
env GOOS=linux GOARCH=amd64 go build -o build/orca-executions-linux-amd64 $(glide nv)
env GOOS=darwin GOARCH=amd64 go build -o build/orca-executions-darwin-amd64 $(glide nv) 
