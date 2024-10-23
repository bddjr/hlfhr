#!/bin/bash
set -e
cd $(dirname $0)
cd test
go build -trimpath -ldflags "-w -s"
./test
