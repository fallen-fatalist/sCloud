#!/bin/sh
if ! go version; then
    echo "Go must be installed"
    exit 1
fi

if go build -o triple-s cmd/web/*; then 
    echo "S3Cloud built successfully, binary file located in project directory"
else
    echo "Error happened while building"
    exit 1
fi