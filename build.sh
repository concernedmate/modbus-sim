#!/bin/bash
docker stop modbus
docker rm modbus
rm -r ./bin
set -e

echo 'start building binaries'
mkdir bin

docker build -t modbus .
docker run -d --name modbus modbus

docker cp modbus:/usr/app/app-win64-latest.exe ./bin/modbus-win64.exe
docker cp modbus:/usr/app/app-linux64-latest ./bin/modbus-linux64

docker stop modbus
docker rm modbus

echo 'finished building binaries'

