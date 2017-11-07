#!/bin/bash

function build () 
{
    echo "start build"
    go build 
    docker build -f dockerfile -t apigateway:latest .
    sleep 3
    docker run -i -t -d -p 80:80 -p 8080:8080 apigateway:latest 
    echo "end build"
}

