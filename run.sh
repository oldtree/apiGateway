#!/bin/bash

echo $GOPATH
chmod -R 777 apiGateway
nohup ./apiGateway > app.log &
 
