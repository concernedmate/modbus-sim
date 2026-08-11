#!/bin/bash
set -e

go clean -testcache
go test -race ./modbus