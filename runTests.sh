#!/bin/bash

curr_dir=`pwd`
echo "Current Directory: ${curr_dir}"
cd ${curr_dir}
#go build -o testcases/tests testcases/tests.go

go clean -testcache
go test -count=1

