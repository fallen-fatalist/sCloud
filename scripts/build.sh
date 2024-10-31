#!/bin/sh
if ! go version; then
    echo "Go must be installed"
    exit 1
fi

if go build -o bin/S3CLoud cmd/web/*; then 
    echo "S3Cloud built successfully, binary file located in bin"
else
    echo "Error happened while building"
    exit 1
fi